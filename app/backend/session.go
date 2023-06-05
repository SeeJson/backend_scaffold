package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"

	"github.com/SeeJson/backend_scaffold/app"
	"github.com/SeeJson/backend_scaffold/cache"
	"github.com/SeeJson/backend_scaffold/cerror"
)

const (
	loginErrPrefix = "login_error_cnt"
	userPrefix     = "sk_user"
	sessionPrefix  = "sk_sess"
)

const defaultMaxIdleTime = 900

type SessionConfig struct {
	MaxIdleTime     int `yaml:"max_idle_time"`
	MaxSessionAlive int `yaml:"max_session_alive"` // 最多同时存在的session数

	Cookie struct {
		Key    string `yaml:"key"`
		Domain string `yaml:"domain"`
		MaxAge int    `yaml:"max_age"`
	}

	JWT struct {
		Secret string        `yaml:"secret"`
		Expire time.Duration `yaml:"expire"`
	}
}

type Session struct {
	Sid string `json:"sid,omitempty"`

	Uid     int64  `json:"uid,omitempty"`
	Name    string `json:"name,omitempty"`
	Role    uint32 `json:"role,omitempty"`
	Perm    uint64 `json:"perm,omitempty"`
	LoginTS int64  `json:"login_ts,omitempty"`
}

func (s *Session) CheckRole(r uint32) bool {
	return (s.Role & r) > 0
}

func sessionKey(sid string) string {
	return sessionPrefix + ":" + sid
}
func userKey(phone string, ts int64) string {
	return fmt.Sprintf("%s:%s:%d", userPrefix, phone, ts)
}
func userKeyPattern(phone string) string {
	return fmt.Sprintf("%s:%s:*", userPrefix, phone)
}

type SessionManager struct {
	conf SessionConfig

	cache *cache.Cache
}

func NewSessionManager(conf SessionConfig, cacher *cache.Cache) *SessionManager {
	return &SessionManager{
		conf:  conf,
		cache: cacher,
	}
}

func (m *SessionManager) SaveSession(c *gin.Context, session Session) (token string, err error) {
	conf := m.conf
	cacher := m.cache

	buf, _ := json.Marshal(&session)

	ttl := conf.MaxIdleTime
	if ttl <= 0 {
		ttl = defaultMaxIdleTime
	}

	// save session to redis
	sid := strings.Replace(uuid.New().String(), "-", "", -1)
	if err = cacher.Setx(sessionKey(sid), string(buf), ttl); err != nil {
		return "", err
	}
	session.Sid = sid

	// set sid in cookie
	c.SetCookie(
		conf.Cookie.Key, sid,
		conf.Cookie.MaxAge,
		"",
		conf.Cookie.Domain,
		false, false,
	)
	token, err = m.saveToken(c, &session)
	if err != nil {
		app.GetLogger(c).Infof("gen new jwt token error:%v", err)
		return "", err
	}

	// 踢掉该用户超过限制数的最早的session
	if err = cacher.Setx(userKey(session.Name, session.LoginTS), sid, ttl); err != nil {
		app.GetLogger(c).Errorf("fail to save session: %v", err)
		return "", err
	}

	if maxSessionAlive := conf.MaxSessionAlive; maxSessionAlive > 0 {
		userSessionKeys, _ := cacher.Keys(userKeyPattern(session.Name))

		if len(userSessionKeys) > maxSessionAlive {
			// 干掉前面的session
			sort.Strings(userSessionKeys)
			for i := 0; i < len(userSessionKeys)-maxSessionAlive; i++ {
				sid, _ := cacher.Get(userSessionKeys[i])
				cacher.Del(sessionKey(sid))
				cacher.Del(userSessionKeys[i])

				app.GetLogger(c).Warn("delete older session", userSessionKeys[i])
			}
		}
	}
	return token, nil
}

func (m *SessionManager) saveToken(c *gin.Context, session *Session) (tokenString string, err error) {
	now := time.Now()
	expireAt := now.Add(m.conf.JWT.Expire)

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, &jwt.MapClaims{
		"iat":  now.Unix(),
		"exp":  expireAt.Unix(),
		"sid":  session.Sid,
		"uid":  session.Uid,
		"role": session.Role,
		"perm": session.Perm,
	})

	tokenString, err = token.SignedString([]byte(m.conf.JWT.Secret))
	if err != nil {
		return "", err
	}

	app.GetLogger(c).Infof("update jwt token:%s", tokenString)

	c.Header("x-auth-token", tokenString)
	return tokenString, nil
}

func (m *SessionManager) parseToken(c *gin.Context, secret string) (sid string, err error) {
	tokenString := GetToken(c)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	sid, ok = claims["sid"].(string)
	if !ok {
		return "", fmt.Errorf("no sid in token")
	}

	return sid, nil
}

func GetSession(c *gin.Context) *Session {
	if v, exists := c.Get("sess"); exists {
		if sess, ok := v.(*Session); ok && sess != nil {
			return sess
		}
	}
	return nil
}

// GetToken 获取用户令牌
func GetToken(c *gin.Context) string {
	//var token string
	//auth := c.GetHeader("Authorization")
	//prefix := "Bearer "
	//if auth != "" && strings.HasPrefix(auth, prefix) {
	//	token = auth[len(prefix):]
	//}
	//return token
	return c.GetHeader("x-auth-token")
}

func (m *SessionManager) GetSession(c *gin.Context) (*Session, error) {
	if sess := GetSession(c); sess != nil {
		return sess, nil
	}

	cfg := m.conf
	cache := m.cache

	sid, err := m.parseToken(c, cfg.JWT.Secret)
	if err != nil {
		//app.GetLogger(c).WithField("err", err).Errorf("parse jwt error")

		sid, err = c.Cookie(cfg.Cookie.Key)
		if err != nil {
			app.GetLogger(c).WithField("err", err).Errorf("get cookie error")
			return nil, cerror.ErrTokenInvalid
		}
		//app.GetLogger(c).Info("get sid from cookie")
	}

	bb, err := cache.Get(sessionKey(sid))
	if err != nil {
		if err == redis.ErrNil {
			return nil, cerror.ErrSessionExpired
		}
		return nil, err
	}

	var sess Session

	if err := json.Unmarshal([]byte(bb), &sess); err != nil {
		return nil, err
	}

	c.Set("sess", &sess)

	sess.Sid = sid
	return &sess, nil
}

func (m *SessionManager) ClearSession(c *gin.Context, session *Session) {
	cfg := m.conf
	cache := m.cache

	sid := session.Sid
	cache.Del(sessionKey(sid))

	c.SetCookie(
		cfg.Cookie.Key, sid,
		-1, "", cfg.Cookie.Domain, false, false,
	)

	cache.Del(userKey(session.Name, session.LoginTS))
}

func (m *SessionManager) RemoveUser(c context.Context, userName string) {
	cacher := m.cache

	userSessionKeys, _ := cacher.Keys(userKeyPattern(userName))

	for i := 0; i < len(userSessionKeys); i++ {
		sid, _ := cacher.Get(userSessionKeys[i])
		cacher.Del(sessionKey(sid))
		cacher.Del(userSessionKeys[i])

		app.GetLogger(c).Warn("remove session", userSessionKeys[i])
	}
}

func (m *SessionManager) RefreshSession(session *Session) {
	conf := m.conf
	cacher := m.cache

	//pp.Println(conf)
	ttl := conf.MaxIdleTime
	if ttl <= 0 {
		ttl = defaultMaxIdleTime
	}

	cacher.TTL(userKey(session.Name, session.LoginTS), ttl)
	cacher.TTL(sessionKey(session.Sid), ttl)
}
