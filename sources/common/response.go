package common

type Response struct {
	Key, Value string
	NextIndex  uint64
	Error      error
}
