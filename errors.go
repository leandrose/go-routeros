package go_routeros

type RouterOSError struct {
	message string
}

func (e *RouterOSError) Error() string {
	return e.message
}
