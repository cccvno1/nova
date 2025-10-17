package errors

type Code int

const (
	Success Code = 0

	// 通用错误 1xxx
	ErrBadRequest       Code = 1001
	ErrUnauthorized     Code = 1002
	ErrForbidden        Code = 1003
	ErrNotFound         Code = 1004
	ErrMethodNotAllowed Code = 1005
	ErrConflict         Code = 1006
	ErrTooManyRequests  Code = 1007
	ErrInternalServer   Code = 1008

	// 参数验证错误 2xxx
	ErrInvalidParams Code = 2001
	ErrBindJSON      Code = 2002
	ErrBindQuery     Code = 2003
	ErrBindForm      Code = 2004

	// 业务错误 3xxx
	ErrDatabase         Code = 3001
	ErrRecordNotFound   Code = 3002
	ErrRecordExists     Code = 3003
	ErrRecordInvalidate Code = 3004

	// 认证授权错误 4xxx
	ErrTokenInvalid     Code = 4001
	ErrTokenExpired     Code = 4002
	ErrTokenMissing     Code = 4003
	ErrPermissionDenied Code = 4004
)

var codeText = map[Code]string{
	Success: "success",

	ErrBadRequest:       "bad request",
	ErrUnauthorized:     "unauthorized",
	ErrForbidden:        "forbidden",
	ErrNotFound:         "not found",
	ErrMethodNotAllowed: "method not allowed",
	ErrConflict:         "conflict",
	ErrTooManyRequests:  "too many requests",
	ErrInternalServer:   "internal server error",

	ErrInvalidParams: "invalid parameters",
	ErrBindJSON:      "failed to bind json",
	ErrBindQuery:     "failed to bind query",
	ErrBindForm:      "failed to bind form",

	ErrDatabase:         "database error",
	ErrRecordNotFound:   "record not found",
	ErrRecordExists:     "record already exists",
	ErrRecordInvalidate: "record is invalid",

	ErrTokenInvalid:     "token is invalid",
	ErrTokenExpired:     "token is expired",
	ErrTokenMissing:     "token is missing",
	ErrPermissionDenied: "permission denied",
}

func (c Code) String() string {
	if msg, ok := codeText[c]; ok {
		return msg
	}
	return "unknown error"
}

func (c Code) HTTPStatus() int {
	switch c {
	case Success:
		return 200
	case ErrBadRequest, ErrInvalidParams, ErrBindJSON, ErrBindQuery, ErrBindForm:
		return 400
	case ErrUnauthorized, ErrTokenInvalid, ErrTokenExpired, ErrTokenMissing:
		return 401
	case ErrForbidden, ErrPermissionDenied:
		return 403
	case ErrNotFound, ErrRecordNotFound:
		return 404
	case ErrMethodNotAllowed:
		return 405
	case ErrConflict, ErrRecordExists:
		return 409
	case ErrTooManyRequests:
		return 429
	default:
		return 500
	}
}
