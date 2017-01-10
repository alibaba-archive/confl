package confl

type Configuration interface {
	// return the path of config in background storage
	Path() string
	// bytes to struct
	Unmarshal([]byte) error
}
