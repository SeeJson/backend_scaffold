package backend

import (
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/SeeJson/backend_scaffold/app"
	"github.com/SeeJson/backend_scaffold/cerror"
	"github.com/SeeJson/backend_scaffold/docgen"
	"github.com/SeeJson/backend_scaffold/ginplus"
	"github.com/SeeJson/backend_scaffold/model"
)

type LoginRequest struct {
	Name     string `json:"name" note:"登录名"`
	Password string `json:"password" note:"密码"`
}

type LoginResponse struct {
	ginplus.CommonResponse

	Data LoginData `json:"data"`
}

type LoginData struct {
	ID   int64  `json:"uid" note:"用户id"`
	Name string `json:"name"`
	Role uint32 `json:"role"`
	Perm uint64 `json:"perm"`
}

func (s *Server) LoginHandler() Handler {
	return Handler{
		doc: &docgen.DocInfo{
			Index:   1,
			Section: sectionUser,
			Name:    "登录",
			Desc:    "后台登录",
			Req:     LoginRequest{},
			Rsp: LoginResponse{
				CommonResponse: ginplus.OkResponse,
				Data:           LoginData{},
			},
		},

		handler: func(c *gin.Context) {
			logger := app.GetLogger(c)

			var req LoginRequest
			err := c.ShouldBindJSON(&req)
			if err != nil {
				logger.Errorf("fail to bind request params: %v", err)
				ginplus.ResError(c, cerror.ErrInvalidRequest)
				return
			}

			user, err := model.GetUserByName(s.db, req.Name)
			if err != nil {
				logger.WithField("err", err).Error("GetUserByName error")
				if model.IsNotFoundError(err) {
					ginplus.ResError(c, cerror.ErrUserPasswordWrong)
				} else {
					ginplus.ResError(c, cerror.ErrInternalServerError)
				}
				return
			}
			if err := bcrypt.CompareHashAndPassword([]byte(user.PwdHash), []byte(user.Salt+req.Password)); err != nil {
				ginplus.ResError(c, cerror.ErrUserPasswordWrong)
				return
			}

			sess := Session{
				Uid:     user.Uid,
				Name:    user.Name,
				Role:    user.Role,
				Perm:    user.Perm,
				LoginTS: time.Now().UnixNano(),
			}

			resp := LoginResponse{
				CommonResponse: ginplus.OkResponse,
			}
			resp.Data.ID = user.Uid
			resp.Data.Name = user.Name
			resp.Data.Perm = user.Perm
			resp.Data.Role = user.Role

			_, err = s.sessionManager.SaveSession(c, sess)
			if err != nil {
				logger.WithField("err", err).WithField("sess", sess).Error("SaveSession error")
			}
			//resp.Data.Role = user.Perm
			//resp.Data.NeedChangePassword = user.ResetPassword

			logger.WithField("uid", user.Uid).Info("user login")
			ginplus.ResSuccess(c, resp)
		},
	}
}

func (s *Server) LogoutHandler() Handler {
	return Handler{
		doc: &docgen.DocInfo{
			Index:   2,
			Section: sectionUser,
			Name:    "登出",
			Desc:    "登出",
			Rsp:     ginplus.CommonResponse{},
		},

		handler: func(c *gin.Context) {
			logger := app.GetLogger(c)
			sess := GetSession(c)
			if sess == nil {
				ginplus.ResError(c, cerror.ErrUnauthorized)
				return
			}

			s.sessionManager.ClearSession(c, sess)

			logger.Info("logout")
			ginplus.ResOK(c)
		},
	}
}
