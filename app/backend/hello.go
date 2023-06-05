package backend

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"

	"github.com/SeeJson/backend_scaffold/docgen"
	"github.com/SeeJson/backend_scaffold/ginplus"
)

// 标记事件
// @Description 标记事件 tag: 0:未标记 1:已标记 2:已忽略
// @Tags 事件管理
// @Accept  json
// @Produce  json
// @Param data body applet.HolaRequest true "event id"
// @Failure 400 {object} applet.ErrorResponse
// @Failure 500 {object} applet.ErrorResponse
// @Success 200 {object} applet.HolaResponse
// @Router /event/tag [post]
func (s *Server) hello(c *gin.Context) {
	pp.Println("in api hekko")

	ginplus.ResSuccess(c, "hello")
}

func (s *Server) todoHandler() Handler {
	return Handler{
		handler: func(c *gin.Context) {
			ginplus.ResTodo(c)
		},
	}
}

type ErrorResponse struct {
	Code int
}

type HolaRequest struct {
	Str string
	BB  []string
}

type HolaResponse struct {
	ID     int64
	Status int8
}

func (s *Server) HolaHandler() Handler {
	return Handler{
		doc: &docgen.DocInfo{
			Section: sectionUser,
			Name:    "测试登录",
			Desc:    "测试是否正常登录",
			Rsp:     ginplus.CommonResponse{},
		},

		handler: func(c *gin.Context) {
			pp.Println("in api hola")

			resp := ginplus.OkResponse
			resp.Msg = "hola! amigo!"
			ginplus.ResSuccess(c, resp)
		},
	}
}

func (s *Server) hola(c *gin.Context) {
}

func (s *Server) callback(c *gin.Context) {
	pp.Println("in api callback")
	pp.Println(c.Request.URL.String())

	bodyBuf, err := ioutil.ReadAll(c.Request.Body)
	pp.Println(string(bodyBuf), err)

	for key, value := range c.Request.PostForm {
		pp.Printf("%v = %v \n", key, value)
	}

	ginplus.ResSuccess(c, "hello")
}
