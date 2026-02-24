package utils

func Ptr[T any](v T) *T {
	return &v
}

func StringPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
