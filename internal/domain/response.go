package domain

// Response represents standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SuccessResponse creates a success response
func SuccessResponse(message string, data interface{}) *Response {
	return &Response{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse creates an error response
func ErrorResponse(message string, err error) *Response {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	return &Response{
		Success: false,
		Message: message,
		Error:   errMsg,
	}
}
