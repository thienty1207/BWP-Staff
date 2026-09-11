package httperror

// AppError is a safe, explicit error that an HTTP handler may return to its
// client.
type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (err *AppError) Error() string {
	if err == nil {
		return ""
	}
	return err.Message
}
