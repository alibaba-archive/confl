package etcd

import "errors"

var (
	// unexpected directory type of etcd's key
	ErrorUnexpectedDir = errors.New("unexpected dir type")
)
