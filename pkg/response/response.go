package response

type ResponseOk[T any] struct {
	Data T      `json:"data"`
	Msg  string `json:"msg"`
}

type ResponseError struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}
