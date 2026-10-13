package utils

// App error codes — stable strings sent to clients as error.code.
const (
	CodeNotFound        = "NOT_FOUND"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeTokenExpired    = "TOKEN_EXPIRED"
	CodeForbidden       = "FORBIDDEN"
	CodeOutOfScope      = "OUT_OF_SCOPE"
	CodeAccountInactive = "ACCOUNT_INACTIVE"
	CodeInvalidInput    = "INVALID_INPUT"
	CodeConflict        = "CONFLICT"
	CodeRateLimited     = "RATE_LIMITED"
	CodeInternal        = "INTERNAL"
)

// AppErr is a typed error mapped to an HTTP status by error writers.
type AppErr struct {
	Status  int
	Code    string
	Message string
}

func (e *AppErr) Error() string { return e.Message }

func NewErrNotFound(msg string) error        { return &AppErr{Status: 404, Code: CodeNotFound, Message: msg} }
func NewErrUnauthorized(msg string) error    { return &AppErr{Status: 401, Code: CodeUnauthorized, Message: msg} }
func NewErrTokenExpired(msg string) error    { return &AppErr{Status: 401, Code: CodeTokenExpired, Message: msg} }
func NewErrForbidden(msg string) error       { return &AppErr{Status: 403, Code: CodeForbidden, Message: msg} }
func NewErrOutOfScope(msg string) error      { return &AppErr{Status: 403, Code: CodeOutOfScope, Message: msg} }
func NewErrAccountInactive(msg string) error { return &AppErr{Status: 403, Code: CodeAccountInactive, Message: msg} }
func NewErrInvalidInput(msg string) error    { return &AppErr{Status: 400, Code: CodeInvalidInput, Message: msg} }
func NewErrConflict(msg string) error        { return &AppErr{Status: 409, Code: CodeConflict, Message: msg} }
func NewErrRateLimited(msg string) error     { return &AppErr{Status: 429, Code: CodeRateLimited, Message: msg} }
func NewErrInternal(msg string) error        { return &AppErr{Status: 500, Code: CodeInternal, Message: msg} }
