package utils

// App error codes — stable strings sent to clients as error.code.
const (
	CodeNotFound        = "NOT_FOUND"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeTokenExpired    = "TOKEN_EXPIRED"
	CodeForbidden       = "FORBIDDEN"
	CodeOutOfScope      = "OUT_OF_SCOPE"
	CodeAccountInactive = "ACCOUNT_INACTIVE"
	CodeInvalidInput        = "INVALID_INPUT"
	CodeConflict            = "CONFLICT"
	CodeRateLimited         = "RATE_LIMITED"
	CodeOutOfRange          = "OUT_OF_RANGE"
	CodeLocationImprecise   = "LOCATION_IMPRECISE"
	CodeLocationUnavailable = "LOCATION_UNAVAILABLE"
	CodeChatHiddenByStudent = "CHAT_HIDDEN_BY_STUDENT"
	CodeChatClosedByStaff   = "CHAT_CLOSED_BY_STAFF"
	CodePayloadTooLarge     = "PAYLOAD_TOO_LARGE"
	CodeUnsupportedMedia    = "UNSUPPORTED_MEDIA"
	CodePreviewExpired      = "PREVIEW_EXPIRED"
	CodeAINotConfigured     = "AI_NOT_CONFIGURED"
	CodeAIFailed            = "AI_FAILED" // reserved: AI async, kegagalan tersimpan di ai_summaries.status
	CodeInternal            = "INTERNAL"
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
func NewErrOutOfRange(msg string) error          { return &AppErr{Status: 403, Code: CodeOutOfRange, Message: msg} }
func NewErrLocationImprecise(msg string) error   { return &AppErr{Status: 403, Code: CodeLocationImprecise, Message: msg} }
func NewErrLocationUnavailable(msg string) error { return &AppErr{Status: 403, Code: CodeLocationUnavailable, Message: msg} }
func NewErrChatHiddenByStudent(msg string) error { return &AppErr{Status: 403, Code: CodeChatHiddenByStudent, Message: msg} }
func NewErrChatClosedByStaff(msg string) error   { return &AppErr{Status: 403, Code: CodeChatClosedByStaff, Message: msg} }
func NewErrPayloadTooLarge(msg string) error     { return &AppErr{Status: 413, Code: CodePayloadTooLarge, Message: msg} }
func NewErrUnsupportedMedia(msg string) error    { return &AppErr{Status: 415, Code: CodeUnsupportedMedia, Message: msg} }
func NewErrPreviewExpired(msg string) error      { return &AppErr{Status: 410, Code: CodePreviewExpired, Message: msg} }
func NewErrAINotConfigured(msg string) error     { return &AppErr{Status: 503, Code: CodeAINotConfigured, Message: msg} }
func NewErrAIFailed(msg string) error            { return &AppErr{Status: 502, Code: CodeAIFailed, Message: msg} }
func NewErrInternal(msg string) error        { return &AppErr{Status: 500, Code: CodeInternal, Message: msg} }
