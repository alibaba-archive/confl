package confl

import "log"

// defautlOnError default handle for error case
func defautlOnError(err error) {
	if err != nil {
		log.Println(err)
	}
}

// Hook the handler of update events
// pass the value of configuration(not ptr) to avoiding changing it
type Hook func(oldCfg, newCfg interface{})

// Watcher concern about the changes of configuration
type Watcher interface {
	// Config return the copy of configuration struct
	// Example:
	//   cfg := Config().(MyConfigStruct)
	Config() interface{}
	// Close close the watcher
	Close() error
	// AddHook add hooks for the update events of configuration
	AddHook(...Hook)
	// OnError add error handle for error cases
	OnError(func(error))
	// Watch start watch the update events
	// It is blocked until the watcher is closed
	Watch()
}
