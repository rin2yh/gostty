package ghostty

// #include "ghostty.h"
// #include <stdbool.h>
//
// // Forward declarations of Go-exported callback functions.
// extern void ghosttyGoWakeupCB(void*);
// extern bool ghosttyGoActionCB(ghostty_app_t, ghostty_target_s, ghostty_action_s);
// extern bool ghosttyGoReadClipboardCB(void*, ghostty_clipboard_e, void*);
// extern void ghosttyGoConfirmReadClipboardCB(void*, const char*, void*, ghostty_clipboard_request_e);
// extern void ghosttyGoWriteClipboardCB(void*, ghostty_clipboard_e, const ghostty_clipboard_content_s*, size_t, bool);
// extern void ghosttyGoCloseSurfaceCB(void*, bool);
//
// // Convert a cgo.Handle (uintptr) to void* without unsafe.Pointer in Go.
// static void* ghosttyHandleToPtr(uintptr_t h) { return (void*)h; }
//
// // Build a runtime config with all callbacks wired to Go dispatch functions.
// static ghostty_runtime_config_s ghosttyMakeRuntimeConfig(void* userdata) {
//     ghostty_runtime_config_s cfg;
//     cfg.userdata                   = userdata;
//     cfg.supports_selection_clipboard = true;
//     cfg.wakeup_cb                  = ghosttyGoWakeupCB;
//     cfg.action_cb                  = ghosttyGoActionCB;
//     cfg.read_clipboard_cb          = ghosttyGoReadClipboardCB;
//     cfg.confirm_read_clipboard_cb  = ghosttyGoConfirmReadClipboardCB;
//     cfg.write_clipboard_cb         = ghosttyGoWriteClipboardCB;
//     cfg.close_surface_cb           = ghosttyGoCloseSurfaceCB;
//     return cfg;
// }
import "C"
import (
	"errors"
	"runtime/cgo"
	"unsafe"
)

// RuntimeCallbacks holds the Go callback functions for the ghostty runtime.
// Surface-level callbacks (ReadClipboard, WriteClipboard, CloseSurface) receive
// the *Surface that triggered the event. Wakeup is app-level.
type RuntimeCallbacks struct {
	// Wakeup is called when the app event loop should wake up.
	Wakeup func()

	// Action is called when the app or a surface performs an action.
	Action func(target Target, action Action) bool

	// ReadClipboard is called when a surface requests clipboard contents.
	// req is an opaque pointer that must be passed to Surface.CompleteClipboardRequest.
	ReadClipboard func(surface *Surface, cb Clipboard, req unsafe.Pointer) bool

	// ConfirmReadClipboard is called to ask the user to confirm a clipboard read.
	ConfirmReadClipboard func(surface *Surface, content string, req unsafe.Pointer, reqType ClipboardRequest)

	// WriteClipboard is called when a surface wants to write to the clipboard.
	WriteClipboard func(surface *Surface, cb Clipboard, contents []ClipboardContent, confirm bool)

	// CloseSurface is called when a surface requests to be closed.
	CloseSurface func(surface *Surface, processAlive bool)
}

// App wraps ghostty_app_t.
type App struct {
	ptr       C.ghostty_app_t
	handle    cgo.Handle
	callbacks RuntimeCallbacks
}

// NewApp creates a new ghostty app with the given runtime callbacks and config.
// cfg may be nil to use defaults. Call cfg.Finalize() before passing it here.
func NewApp(cb RuntimeCallbacks, cfg *Config) (*App, error) {
	app := &App{callbacks: cb}
	app.handle = cgo.NewHandle(app)

	rtcfg := C.ghosttyMakeRuntimeConfig(C.ghosttyHandleToPtr(C.uintptr_t(app.handle)))

	var cfgPtr C.ghostty_config_t
	if cfg != nil {
		cfgPtr = cfg.ptr
	}

	app.ptr = C.ghostty_app_new(&rtcfg, cfgPtr)
	if app.ptr == nil {
		app.handle.Delete()
		return nil, errors.New("ghostty_app_new failed")
	}
	return app, nil
}

// Free releases all resources associated with the app.
func (a *App) Free() {
	C.ghostty_app_free(a.ptr)
	a.handle.Delete()
	a.ptr = nil
}

// Tick processes pending events. Call this whenever the wakeup callback fires.
func (a *App) Tick() {
	C.ghostty_app_tick(a.ptr)
}

// SetFocus sets the application focus state.
func (a *App) SetFocus(focused bool) {
	C.ghostty_app_set_focus(a.ptr, C.bool(focused))
}

// NeedsConfirmQuit reports whether quitting requires user confirmation.
func (a *App) NeedsConfirmQuit() bool {
	return bool(C.ghostty_app_needs_confirm_quit(a.ptr))
}
