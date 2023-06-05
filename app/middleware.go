package app

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"

	"github.com/SeeJson/backend_scaffold/cerror"
	"github.com/SeeJson/backend_scaffold/ginplus"
)

const TraceIDKey = "/TraceID"

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		traceID := c.GetHeader("X-RequestExample-Id")
		if traceID == "" {
			traceID = xid.New().String()
		}

		c.Set(TraceIDKey, traceID)
		logger := GetLogger(c).
			WithField("tid", traceID).
			WithField("method", c.Request.Method).
			WithField("uri", c.FullPath())

		SetLogger(c, logger)

		c.Next()
	}
}

type ErrorResponse struct {
	Status string `json:"status" example:"some_error"`
	Code   int    `json:"code" example:"21030100"`
}

// ErrorHandler 对错误结果统一处理
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		errs := c.Errors.ByType(gin.ErrorTypeAny)
		if len(errs) > 0 {
			err := errs.Last().Err
			switch err.(type) {
			case *cerror.APIError:
				parsedErr := err.(*cerror.APIError)
				ginplus.ResError(c, parsedErr)
				//c.JSON(parsedErr.HTTPCode, ErrorResponse{
				//	Code:   parsedErr.Code,
				//	Status: parsedErr.Msg})
				return
			default:
			}
		}
	}
}

// SkipperFunc 定义中间件跳过函数
type SkipperFunc func(*gin.Context) bool

// AllowMethodAndPathPrefixSkipper 检查请求方法和路径是否包含指定的前缀，如果不包含则跳过
func AllowMethodAndPathPrefixSkipper(prefixes ...string) SkipperFunc {
	return func(c *gin.Context) bool {
		path := JoinRouter(c.Request.Method, c.Request.URL.Path)
		pathLen := len(path)

		for _, p := range prefixes {
			if pl := len(p); pathLen >= pl && path[:pl] == p {
				return true
			}
		}
		return false
	}
}

// JoinRouter 拼接路由
func JoinRouter(method, path string) string {
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}
	return fmt.Sprintf("%s%s", strings.ToUpper(method), path)
}

func MiddlewareWithSkipper(mw gin.HandlerFunc, skippers ...SkipperFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, sk := range skippers {
			if sk(c) {
				c.Next()
				return
			}
		}

		mw(c)
	}
}

// 限制并发连接数
func MaxConnections(n int) gin.HandlerFunc {
	sem := make(chan struct{}, n)
	acquire := func() bool {
		select {
		case sem <- struct{}{}:
			return true
		default:
			return false
		}
	}
	release := func() { <-sem }
	return func(c *gin.Context) {
		// before request
		if !acquire() {
			c.Error(cerror.ErrReachMaxConnect)
			c.Abort()
			return
		}
		defer release() // after request
		c.Next()
	}
}

// ErrorHandler 对错误结果统一处理
func NoCached() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		c.Next()
	}
}
