package writer

// Writer it's an interface used to abstract the logic
// of writing results to a Key Value without having to
// deal with types or internal structures
type Writer interface {
	// Write sets the value with the key to the internal structure,
	// the value will be casted to the correct type of each
	// implementation and an error can be returned normally for
	// repeated keys
	Write(key string, value interface{}) error

	// Sync writes the content of the writer
	// to the internal system. Each Writter may have
	// a different implementation of it with different
	// output formats
	Sync() error
}
