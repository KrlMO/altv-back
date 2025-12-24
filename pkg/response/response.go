package response

type ResponseOk[T any] struct {
	Ok   bool `json:"ok"`
	Data T    `json:"data"`
}

type ResponseError struct {
	Error string `json:"error"`
	Ok    bool   `json:"ok"`
}
