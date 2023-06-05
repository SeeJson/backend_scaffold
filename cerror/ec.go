package cerror

import (
	"strconv"

	"github.com/SeeJson/backend_scaffold/apispec"
)

type EC struct {
	*APIError
	desc string
}

func (e *EC) Code() string {
	return strconv.Itoa(e.APIError.Code)
}
func (e *EC) Msg() string {
	return e.APIError.Msg
}
func (e *EC) Desc() string {
	return e.desc
}

func RegisterErrorCodeSpec(ab *apispec.ApiSpecBuilder) {
	ab.AddErrorCode(&EC{ErrInternalServerError, ""})
	ab.AddErrorCode(&EC{ErrUnauthorized, ""})
	ab.AddErrorCode(&EC{ErrInvalidRequest, ""})
	ab.AddErrorCode(&EC{ErrSessionExpired, ""})
	ab.AddErrorCode(&EC{ErrDuplicatedEntry, ""})
	ab.AddErrorCode(&EC{ErrUserPasswordWrong, ""})
	ab.AddErrorCode(&EC{ErrTokenInvalid, ""})
	ab.AddErrorCode(&EC{ErrNoPermission, ""})
	ab.AddErrorCode(&EC{ErrNotFound, ""})
	ab.AddErrorCode(&EC{ErrInvalidStatus, ""})

	ab.AddErrorCode(&EC{ErrMaxLimit, ""})
	ab.AddErrorCode(&EC{ErrReachMaxLogin, ""})
	ab.AddErrorCode(&EC{ErrOutdatedSession, ""})
	ab.AddErrorCode(&EC{ErrBadCookie, ""})
	ab.AddErrorCode(&EC{ErrBadSession, ""})
	ab.AddErrorCode(&EC{ErrReachMaxConnect, ""})
}
