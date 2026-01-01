package response

type ResponseOk[T any] struct {
	Ok   bool   `json:"ok"`
	Data T      `json:"data"`
	Msg  string `json:"msg"`
}

type ResponseError struct {
	Error string `json:"error"`
	Ok    bool   `json:"ok"`
}
