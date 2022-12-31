package api

type ApiResponse struct {
	Code        int
	Type        string
	Description string
	Message     string
}

func NewOk(message string) ApiResponse {
	return ApiResponse{
		Code:        200,
		Type:        "OK",
		Description: "Everything went ok.",
		Message:     message,
	}
}
