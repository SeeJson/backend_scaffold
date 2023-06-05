package backend

import (
	"github.com/gin-gonic/gin"

	"github.com/SeeJson/backend_scaffold/app"
	"github.com/SeeJson/backend_scaffold/cerror"
)

func (s *Server) setupRBACRoute(router *gin.RouterGroup, method, uri string, roles uint64, handlers ...gin.HandlerFunc) {
	router.Handle(method, uri, handlers...)
	s.permissionSetting[app.JoinRouter(method, router.BasePath()+uri)] = roles
}

func (s *Server) CheckPermission(c *gin.Context) {
	sess := GetSession(c)
	if sess == nil {
		c.Error(cerror.ErrUnauthorized)
		c.Abort()
		return
	}
	mask, ok := s.permissionSetting[app.JoinRouter(c.Request.Method, c.FullPath())]
	if ok && (mask > 0) && (mask&uint64(sess.Role) == 0) {
		c.Error(cerror.ErrNoPermission)
		c.Abort()
		return
	}

	c.Next()
}
