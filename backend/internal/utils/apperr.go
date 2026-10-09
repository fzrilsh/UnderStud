package utils

// App error codes — stable strings sent to clients as error.code.
const (
	CodeNotFound     = "NOT_FOUND"
	CodeUnauthorized = "UNAUTHORIZED"
	CodeInvalidInput = "INVALID_INPUT"
	CodeConflict     = "CONFLICT"
	CodeInternal     = "INTERNAL"
)

// AppErr is a typed error mapped to an HTTP status by error writers.
type AppErr struct {
	Status  int
	Code    string
	Message string
}

func (e *AppErr) Error() string { return e.Message }

func NewErrNotFound(msg string) error     { return &AppErr{Status: 404, Code: CodeNotFound, Message: msg} }
func NewErrUnauthorized(msg string) error { return &AppErr{Status: 401, Code: CodeUnauthorized, Message: msg} }
func NewErrInvalidInput(msg string) error { return &AppErr{Status: 400, Code: CodeInvalidInput, Message: msg} }
func NewErrConflict(msg string) error     { return &AppErr{Status: 409, Code: CodeConflict, Message: msg} }
func NewErrInternal(msg string) error     { return &AppErr{Status: 500, Code: CodeInternal, Message: msg} }
