package backend

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql" // for mysql
	"github.com/k0kubun/pp"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/SeeJson/backend_scaffold/apispec"
	"github.com/SeeJson/backend_scaffold/app"
	"github.com/SeeJson/backend_scaffold/cache"
	"github.com/SeeJson/backend_scaffold/cerror"
	"github.com/SeeJson/backend_scaffold/common"
	"github.com/SeeJson/backend_scaffold/docgen"
	"github.com/SeeJson/backend_scaffold/ginplus"
	"github.com/SeeJson/backend_scaffold/upload"
	"github.com/SeeJson/backend_scaffold/utils"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Config struct {
	Cache cache.Config
	DB    struct {
		DSN   string
		Debug bool
	}
	Log utils.LogConfig

	Session SessionConfig
	Upload  upload.Config

	Debug bool
}

func (c Config) New() (srv *Server, err error) {

	c.Log.InitLog()

	srv = &Server{
		conf:              c,
		permissionSetting: make(map[string]uint64),
		docs:              make([]*docgen.DocInfo, 0, 10),
		ab:                apispec.NewApiSpecBuilder("后台服务API接口", "后台服务API接口文档", "v0.0.1"),
	}
	pp.Println(c)

	gc := gorm.Config{}
	srv.db, err = gorm.Open(mysql.Open(c.DB.DSN), &gc)
	if err != nil {
		return nil, err
	}
	if c.DB.Debug {
		srv.db = srv.db.Debug()
	}
	srv.uploader, err = c.Upload.New()
	if err != nil {
		return nil, err
	}

	cacher, err := c.Cache.New()
	if err != nil {
		return nil, err
	}

	srv.sessionManager = NewSessionManager(c.Session, cacher)
	srv.opWriter = app.NewOpWriter(srv.db)

	return srv, nil
}

type Server struct {
	conf Config

	db             *gorm.DB
	sessionManager *SessionManager

	uploader *upload.Uploader
	opWriter *app.OpWriter

	permissionSetting map[string]uint64

	docs []*docgen.DocInfo
	ab   *apispec.ApiSpecBuilder
}

// 校验session
func (s *Server) Authentication(c *gin.Context) {
	session, err := s.sessionManager.GetSession(c)
	//pp.Println(session, err)
	if err != nil {
		if _, ok := err.(*cerror.APIError); ok {
			ginplus.ResError(c, err)
		} else if err == http.ErrNoCookie {
			ginplus.ResError(c, cerror.ErrUnauthorized)
		} else {
			app.GetLogger(c).WithField("err", err).Error("get session error")
			ginplus.ResError(c, cerror.ErrInternalServerError)
		}

		c.Abort()
		return
	}
	s.sessionManager.RefreshSession(session)

	logger := app.GetLogger(c).WithField("uid", session.Uid)
	app.SetLogger(c, logger)
	c.Next()
}

// 校验session
func (s *Server) MaybeAuthentication(c *gin.Context) {
	session, err := s.sessionManager.GetSession(c)
	if err == nil && session != nil {
		s.sessionManager.RefreshSession(session)

		logger := app.GetLogger(c).WithField("uid", session.Uid)
		app.SetLogger(c, logger)
	}

	c.Next()
}

func (s *Server) Run(addr string) error {
	router := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "x-auth-token")
	corsConfig.AllowAllOrigins = true
	router.Use(cors.New(corsConfig))

	router.Use(app.ErrorHandler())
	router.Use(app.TraceMiddleware())
	router.Use(app.NoCached())

	routePrefix := "/backend/api/v1"
	noAuthRouter := router.Group(routePrefix)
	{
		noAuthRouter.GET("/hello", s.hello)
		noAuthRouter.POST("/callback", s.callback)
		noAuthRouter.GET("/cb", s.callback)

		s.setupRoute(noAuthRouter, "POST", "/upload", 0, s.UploadHandler())
		s.setupRoute(noAuthRouter, "POST", "/login", 0, s.LoginHandler())
		//s.setupRoute(noAuthRouter, "POST", "/reset/password", 0, s.ResetPasswordHandler())
	}

	maybeRouter := router.Group(routePrefix)
	maybeRouter.Use(s.MaybeAuthentication)
	{
	}

	gr := router.Group(routePrefix)
	gr.Use(s.Authentication)
	gr.Use(s.CheckPermission)
	{
		s.setupRoute(gr, "GET", "/hola", 0, s.HolaHandler())
		s.setupRoute(gr, "POST", "/logout", 0, s.LogoutHandler())

		s.setupRoute(gr, "GET", "/tasks", 0, s.ListAddressHandler())
		s.setupRoute(gr, "GET", "/task", 0, s.TaskDetailHandler())
		s.setupRoute(gr, "GET", "/task/:id", 0, s.TaskDetail2Handler())
		s.setupRoute(gr, "POST", "/task", 0, s.NewTaskHandler())
		s.setupRoute(gr, "PUT", "/task", 0, s.UpdateTaskHandler())
		s.setupRoute(gr, "PUT", "/task/status", 0, s.UpdateTaskStatusHandler())
		s.setupRoute(gr, "DELETE", "/task", 0, s.DeleteTaskHandler())
	}

	docGen := docgen.DocGenerator{
		Title:      "后台API文档",
		Dir:        ".",
		Sections:   sections,
		EnumGetter: common.GetEnum,
	}
	_ = docGen
	common.RegisterEnumSpec(s.ab)
	cerror.RegisterErrorCodeSpec(s.ab)
	s.addRoute2Spec()

	s.ab.WriteFile("api_spec.yaml")

	mds, _ := docgen.ApiSpec2Markdown(s.ab.Build())
	ioutil.WriteFile("doc.md", []byte(mds), 0644)

	//docGen.GenDoc("backend.md", s.docs)
	//app.GenDoc("backend.md", "后台API文档", sections, s.docs)
	return router.Run(addr)
}

type Handler struct {
	handler gin.HandlerFunc

	doc *docgen.DocInfo
}

func (s *Server) addRoute2Spec() {
	sort.Slice(s.docs, func(i, j int) bool {
		if s.docs[i].Section == s.docs[j].Section {
			return s.docs[i].Index < s.docs[j].Index
		}
		return sections[s.docs[i].Section].Index < sections[s.docs[j].Section].Index
	})
	for _, doc := range s.docs {
		r := apispec.RouteInfo{
			Name:            doc.Name,
			Desc:            doc.Desc,
			Method:          doc.Method,
			Uri:             doc.Uri,
			Permissions:     doc.Permissions,
			RequestFormat:   doc.RequestFormat,
			Request:         doc.Req,
			SuccessResponse: doc.Rsp,
			SuccessHttpCode: 200,
		}
		s.ab.AddRoute(doc.Section, r)
	}
}

func (s *Server) setupRoute(router *gin.RouterGroup, method, uri string, roles uint64, handler Handler) {
	s.setupRBACRoute(router, method, uri, roles, handler.handler)
	s.setupRouteDoc(method, router.BasePath()+uri, roles, handler.doc)
}

func (s *Server) UploadHandler() Handler {
	return Handler{
		doc: &docgen.DocInfo{
			Section:       sectionOther,
			Index:         0,
			Name:          "上传图片",
			Desc:          "上传图片",
			RequestFormat: docgen.DocInfoReqFormatForm,
			Req:           app.UploadRequest{},
			Rsp:           app.UploadResponse{},
		},
		handler: app.NewUploadHandler(s.uploader),
	}
}
