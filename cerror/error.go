package cerror

import (
	"fmt"
	"net/http"
)

// APIError 统一的api错误定义
type APIError struct {
	HTTPCode int
	Code     int
	Msg      string
	Detail   string
}

func (err *APIError) Error() string {
	return fmt.Sprintf("http code: %v, status code: %v, status: %v", err.HTTPCode, err.Code, err.Msg)
}

func newAPIError(code int, statusCode int, status string) *APIError {
	return &APIError{
		HTTPCode: code,
		Code:     statusCode,
		Msg:      status,
	}
}
func newAPIError2(code int, statusCode int, status string, detail string) *APIError {
	return &APIError{
		HTTPCode: code,
		Code:     statusCode,
		Msg:      status,
		Detail:   detail,
	}
}

const (
	StatusCodeSuccess = 0
)

var (
	ErrInternalServerError = newAPIError(http.StatusInternalServerError, 10100, "internal_server_error")
	ErrUnauthorized        = newAPIError(http.StatusUnauthorized, 10101, "unauthorized")
	ErrInvalidRequest      = newAPIError(http.StatusBadRequest, 10102, "invalid_request")
	ErrSessionExpired      = newAPIError2(http.StatusUnauthorized, 10103, "session_expired", "登录过期")
	ErrDuplicatedEntry     = newAPIError(http.StatusOK, 10104, "duplicated_entry")
	ErrUserPasswordWrong   = newAPIError2(http.StatusOK, 10106, "user_or_password_incorrect", "用户名或密码错误")
	ErrTokenInvalid        = newAPIError2(http.StatusBadRequest, 10107, "invalid_token", "未登录")
	ErrNoPermission        = newAPIError(http.StatusBadRequest, 10108, "no_permission")
	ErrNotFound            = newAPIError(http.StatusBadRequest, 10109, "not_found")

	ErrInvalidStatus = newAPIError(http.StatusBadRequest, 10121, "invalid_status")

	ErrMaxLimit        = newAPIError(http.StatusBadRequest, 10121, "more than the max limit")
	ErrReachMaxLogin   = newAPIError(http.StatusBadRequest, 10200, "reach_max_login_limit")
	ErrOutdatedSession = newAPIError(http.StatusUnauthorized, 10201, "outdated_session")
	ErrBadCookie       = newAPIError(http.StatusBadRequest, 10202, "bad_cookie")
	ErrBadSession      = newAPIError(http.StatusBadRequest, 10203, "bad_session")
	ErrReachMaxConnect = newAPIError(http.StatusBadRequest, 10204, "reach_max_connections_limit")
)
