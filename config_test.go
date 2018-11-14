package confl_test

// Config - your configuration struct
type Config struct {
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	Token    string   `yaml:"token"`
	In       ConfigIn `yaml:"in"`
}

// ConfigIn -
type ConfigIn struct {
	Addr string `yaml:"addr"`
	Pass string `yaml:"pass"`
}
