package webirr

// ApiResponse is the common WeBirr merchant API response wrapper.
type ApiResponse[T any] struct {
	Error     string `json:"error"`
	Res       T      `json:"res"`
	ErrorCode string `json:"errorCode"`
}
