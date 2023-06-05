package ginplus

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
	"github.com/pkg/errors"

	"github.com/SeeJson/backend_scaffold/cerror"
)

// 定义上下文中的键
const (
	prefix = "ginadmin"
	// UserIDKey 存储上下文中的键(用户ID)
	UserIDKey = prefix + "/user_id"
	// TraceIDKey 存储上下文中的键(跟踪ID)
	TraceIDKey = prefix + "/trace_id"
	// ResBodyKey 存储上下文中的键(响应Body数据)
	ResBodyKey = prefix + "/res_body"

	UIDKey        = prefix + "/uid"
	RoleKey       = prefix + "/role"
	UserKey       = prefix + "/user"
	PermissionKey = prefix + "/permission"

	CHNKey = prefix + "/chn"
)

// NewContext 封装上线文入口
//func NewContext(c *gin.Context) context.Context {
//	parent := context.Background()
//
//	if v := GetTraceID(c); v != "" {
//		parent = icontext.NewTraceID(parent, v)
//		parent = logger.NewTraceIDContext(parent, GetTraceID(c))
//	}
//
//	if v := GetUserID(c); v != "" {
//		parent = icontext.NewUserID(parent, v)
//		parent = logger.NewUserIDContext(parent, v)
//	}
//
//	return parent
//}

// GetToken 获取用户令牌
func GetToken(c *gin.Context) string {
	var token string
	auth := c.GetHeader("Authorization")
	prefix := "Bearer "
	if auth != "" && strings.HasPrefix(auth, prefix) {
		token = auth[len(prefix):]
	}
	return token
}

func GetRole(c *gin.Context) int8 {
	r := c.GetInt(RoleKey)
	return int8(r)
}

//GetPageIndex 获取分页的页索引
//func GetPageIndex(c *gin.Context) int {
//	defaultVal := 1
//	if v := c.Query("current"); v != "" {
//		if iv := util.S(v).DefaultInt(defaultVal); iv > 0 {
//			return iv
//		}
//	}
//	return defaultVal
//}
//
//// GetPageSize 获取分页的页大小(最大50)
//func GetPageSize(c *gin.Context) int {
//	defaultVal := 10
//	if v := c.Query("pageSize"); v != "" {
//		if iv := util.S(v).DefaultInt(defaultVal); iv > 0 {
//			if iv > 50 {
//				iv = 50
//			}
//			return iv
//		}
//	}
//	return defaultVal
//}

//// GetPaginationParam 获取分页查询参数
//func GetPaginationParam(c *gin.Context) *schema.PaginationParam {
//	return &schema.PaginationParam{
//		PageIndex: GetPageIndex(c),
//		PageSize:  GetPageSize(c),
//	}
//}

// GetTraceID 获取追踪ID
func GetTraceID(c *gin.Context) string {
	return c.GetString(TraceIDKey)
}

// GetUserID 获取用户ID
func GetUserID(c *gin.Context) string {
	return c.GetString(UserIDKey)
}

// GetUserID 获取用户ID
func GetUID(c *gin.Context) string {
	return c.GetString(UIDKey)
}

// GetUserID 获取用户ID
func GetChn(c *gin.Context) string {
	return c.GetString(CHNKey)
}

// SetUserID 设定用户ID
func SetUserID(c *gin.Context, userID string) {
	c.Set(UserIDKey, userID)
}

// ResPage 响应分页数据
//func ResPage(c *gin.Context, v interface{}, pr *schema.PaginationResult) {
//	list := schema.HTTPList{
//		List: v,
//		Pagination: &schema.HTTPPagination{
//			Current:  GetPageIndex(c),
//			PageSize: GetPageSize(c),
//		},
//	}
//	if pr != nil {
//		list.Pagination.Total = pr.Total
//	}
//
//	ResSuccess(c, list)
//}

type CommonResponse struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Error string `json:"error"`
}

var OkResponse = CommonResponse{
	Code: 0,
	Msg:  "ok",
}

type IDResponse struct {
	CommonResponse

	ID int64 `json:"id"`
}

// ResOK 响应OK
func ResOK(c *gin.Context) {
	ResSuccess(c, OkResponse)
}

var TodoResponse = CommonResponse{
	Code: -1,
	Msg:  "todo",
}

func ResTodo(c *gin.Context) {
	ResSuccess(c, TodoResponse)
}

func ResID(c *gin.Context, id int64) {
	ResSuccess(c, IDResponse{
		CommonResponse: OkResponse,
		ID:             id,
	})
}

// ResSuccess 响应成功
func ResSuccess(c *gin.Context, v interface{}) {
	//logger := app.GetLogger(c)
	//logger.WithField("rsp", v).Info("return response")

	//ResJSON(c, http.StatusOK, v)
	c.JSON(http.StatusOK, v)
}

// ResJSON 响应JSON数据
func ResJSON(c *gin.Context, status int, v interface{}) {
	buf, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	c.Set(ResBodyKey, buf)
	c.Data(status, "application/json; charset=utf-8", buf)
	c.Abort()
}

// ResError 响应错误
func ResError(c *gin.Context, err error, msgs ...string) {
	statusCode := cerror.ErrInternalServerError.HTTPCode
	res := CommonResponse{
		Code: cerror.ErrInternalServerError.Code,
		Msg:  cerror.ErrInternalServerError.Msg,
	}

	oerr := errors.Cause(err)
	pp.Println(err)
	if cerr, ok := oerr.(*cerror.APIError); ok {
		res.Code = cerr.Code
		res.Msg = cerr.Msg
		res.Error = cerr.Msg
		statusCode = cerr.HTTPCode

		if len(msgs) > 0 && msgs[0] != "" {
			res.Msg = msgs[0]
		} else if cerr.Detail != "" {
			res.Msg = cerr.Detail
		}
	} else {
		pp.Println("ResError", err, msgs)
	}

	pp.Println("ResError", res.Code, res.Error, res.Msg)
	ResJSON(c, statusCode, res)
}

//func ResError(c *gin.Context, err error, status ...int) {
//	statusCode := cerror.ErrInternalServerError.HTTPCode
//	res := CommonResponse{
//		Code: cerror.ErrInternalServerError.Code,
//		Msg:  cerror.ErrInternalServerError.Msg,
//	}
//
//	oerr := errors.Cause(err)
//	if cerr, ok := oerr.(*cerror.APIError); ok {
//		res.Code = cerr.Code
//		res.Msg = cerr.Msg
//		statusCode = cerr.HTTPCode
//	} else {
//		pp.Println("ResError", err, status)
//	}
//
//	if len(status) > 0 {
//		statusCode = status[0]
//	}
//
//	//if statusCode == 500 && err != nil {
//	//span := logger.StartSpan(NewContext(c))
//	//span = span.WithField("stack", fmt.Sprintf("%+v", err))
//	//span.Errorf(err.Error())
//	//}
//
//	ResJSON(c, statusCode, res)
//}

type IDRequest struct {
	ID int64 `json:"id" form:"id" uri:"id"`
}

type IDsRequest struct {
	IDs []int64 `json:"ids" form:"ids"`
}

type IDPager struct {
	ID int64 `json:"id" form:"id"`
	Pager
}

type IDName struct {
	ID   int64  `json:"id" form:"id"`
	Name string `json:"name" form:"name"`
}

type Pager struct {
	Offset int `json:"offset" form:"offset" note:"偏移量， 从0开始"`
	Limit  int `json:"limit" form:"limit" note:"个数"`
}

type SearchRequest struct {
	Search string `json:"search" form:"search" note:"搜索内容"`

	Pager
}

type IDSearchRequest struct {
	ID     int64  `json:"id" form:"id" note:"id"`
	Search string `json:"search" form:"search" note:"搜索内容"`

	Pager
}

type StartEnd struct {
	Start string `json:"start" form:"start" note:"格式:2006-01-02"`
	End   string `json:"end" form:"end" note:"格式:2006-01-02"`
}

func (req *StartEnd) Check() string {
	_, err := time.Parse("2006-01-02", req.Start)
	if err != nil {
		return "开始日期错误"
	}
	_, err = time.Parse("2006-01-02", req.End)
	if err != nil {
		return "结束日期错误"
	}

	return ""
}
