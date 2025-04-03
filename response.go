package go_routeros

type Response struct {
	// Type response type returned by RouterOS: !re, !empty, !done, !trap, or !fatal
	Type string
	// Data response data returned by RouterOS
	Data map[string]string
	// Err an error occurred, returned by RouterOS in a !trap or !fatal response
	Err error
}
