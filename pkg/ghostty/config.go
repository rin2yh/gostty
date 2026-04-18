package ghostty

// #include "ghostty.h"
// #include <stdlib.h>
import "C"
import (
	"errors"
	"unsafe"
)

// Config wraps ghostty_config_t.
type Config struct {
	ptr C.ghostty_config_t
}

// NewConfig creates a new ghostty configuration.
func NewConfig() (*Config, error) {
	ptr := C.ghostty_config_new()
	if ptr == nil {
		return nil, errors.New("ghostty_config_new failed")
	}
	return &Config{ptr: ptr}, nil
}

// LoadDefaultFiles loads the default ghostty configuration files.
func (c *Config) LoadDefaultFiles() {
	C.ghostty_config_load_default_files(c.ptr)
}

// LoadFile loads a configuration file from path.
func (c *Config) LoadFile(path string) {
	cs := C.CString(path)
	defer C.free(unsafe.Pointer(cs))
	C.ghostty_config_load_file(c.ptr, cs)
}

// Finalize finalizes the configuration. Must be called before use.
func (c *Config) Finalize() {
	C.ghostty_config_finalize(c.ptr)
}

// Free releases the configuration resources.
func (c *Config) Free() {
	if c.ptr == nil {
		return
	}
	C.ghostty_config_free(c.ptr)
	c.ptr = nil
}
