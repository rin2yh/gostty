package ghostty

// #cgo CFLAGS: -I../../ghostty/include -DGHOSTTY_STATIC
// #cgo darwin LDFLAGS: ${SRCDIR}/../../ghostty/macos/GhosttyKit.xcframework/macos-arm64_x86_64/ghostty-internal.a
// #cgo darwin LDFLAGS: -framework Foundation -framework AppKit -framework Carbon
// #cgo darwin LDFLAGS: -framework Metal -framework MetalKit -framework QuartzCore
// #cgo darwin LDFLAGS: -framework IOSurface -framework CoreText -framework CoreGraphics
// #cgo darwin LDFLAGS: -lc++
// #include "ghostty.h"
import "C"
import (
	"errors"
	"sync"
)

var (
	initOnce sync.Once
	initErr  error
)

// Init initializes the ghostty library. Safe to call multiple times;
// initialization happens exactly once.
func Init() error {
	initOnce.Do(func() {
		if ret := C.ghostty_init(0, nil); ret != C.GHOSTTY_SUCCESS {
			initErr = errors.New("ghostty_init failed")
		}
	})
	return initErr
}

// BuildMode is the Zig build optimization mode.
type BuildMode int

const (
	BuildModeDebug        BuildMode = 0
	BuildModeReleaseSafe  BuildMode = 1
	BuildModeReleaseFast  BuildMode = 2
	BuildModeReleaseSmall BuildMode = 3
)

// Info contains build information about the compiled ghostty library.
type Info struct {
	BuildMode BuildMode
	Version   string
}

// GetInfo returns build information about the ghostty library.
func GetInfo() Info {
	i := C.ghostty_info()
	return Info{
		BuildMode: BuildMode(i.build_mode),
		Version:   C.GoStringN(i.version, C.int(i.version_len)),
	}
}
