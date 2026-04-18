package ghostty

// #include "ghostty.h"
// #include <stdlib.h>
//
// // Convert a cgo.Handle (uintptr) to void* without unsafe.Pointer in Go.
// static void* ghosttyHandleToPtr(uintptr_t h) { return (void*)h; }
//
// // Set the NSView pointer in the macOS platform union.
// static void ghosttySetNSView(ghostty_surface_config_s* cfg, void* nsview) {
//     cfg->platform.macos.nsview = nsview;
// }
import "C"
import (
	"errors"
	"runtime/cgo"
	"unsafe"
)

// SurfaceConfig holds the configuration for creating a new surface.
type SurfaceConfig struct {
	// NSView is the macOS NSView pointer to render into.
	NSView unsafe.Pointer
	// ScaleFactor is the display scale factor (e.g. 2.0 for Retina).
	ScaleFactor float64
	// FontSize overrides the configured font size. 0 uses the config default.
	FontSize float32
	// WorkDir sets the initial working directory.
	WorkDir string
	// Command runs a specific command instead of the default shell.
	Command string
	// Context is the surface context (window, tab, split).
	Context SurfaceContext
}

// Surface wraps ghostty_surface_t.
type Surface struct {
	ptr    C.ghostty_surface_t
	handle cgo.Handle
	app    *App
}

// NewSurface creates a new terminal surface attached to the given app.
// On macOS, cfg.NSView must be a valid *NSView pointer.
func NewSurface(app *App, cfg SurfaceConfig) (*Surface, error) {
	s := &Surface{app: app}
	s.handle = cgo.NewHandle(s)

	ccfg := C.ghostty_surface_config_new()
	ccfg.platform_tag = C.GHOSTTY_PLATFORM_MACOS
	C.ghosttySetNSView(&ccfg, cfg.NSView)
	ccfg.userdata = C.ghosttyHandleToPtr(C.uintptr_t(s.handle))
	ccfg.scale_factor = C.double(cfg.ScaleFactor)
	ccfg.font_size = C.float(cfg.FontSize)
	ccfg.context = C.ghostty_surface_context_e(cfg.Context)

	if cfg.WorkDir != "" {
		cs := C.CString(cfg.WorkDir)
		defer C.free(unsafe.Pointer(cs))
		ccfg.working_directory = cs
	}
	if cfg.Command != "" {
		cs := C.CString(cfg.Command)
		defer C.free(unsafe.Pointer(cs))
		ccfg.command = cs
	}

	s.ptr = C.ghostty_surface_new(app.ptr, &ccfg)
	if s.ptr == nil {
		s.handle.Delete()
		return nil, errors.New("ghostty_surface_new failed")
	}
	return s, nil
}

// Free releases all resources associated with the surface.
func (s *Surface) Free() {
	C.ghostty_surface_free(s.ptr)
	s.handle.Delete()
	s.ptr = nil
}

// Draw draws the surface contents.
func (s *Surface) Draw() {
	C.ghostty_surface_draw(s.ptr)
}

// Refresh requests a redraw of the surface.
func (s *Surface) Refresh() {
	C.ghostty_surface_refresh(s.ptr)
}

// SetSize notifies the surface of a new pixel size.
func (s *Surface) SetSize(w, h uint32) {
	C.ghostty_surface_set_size(s.ptr, C.uint32_t(w), C.uint32_t(h))
}

// SetFocus sets the keyboard focus state of the surface.
func (s *Surface) SetFocus(focused bool) {
	C.ghostty_surface_set_focus(s.ptr, C.bool(focused))
}

// SetContentScale notifies the surface of a display scale change.
func (s *Surface) SetContentScale(x, y float64) {
	C.ghostty_surface_set_content_scale(s.ptr, C.double(x), C.double(y))
}

// Key sends a key event to the surface. Returns true if the event was consumed.
func (s *Surface) Key(ev KeyEvent) bool {
	cev := C.ghostty_input_key_s{
		action:              C.ghostty_input_action_e(ev.Action),
		mods:                C.ghostty_input_mods_e(ev.Mods),
		consumed_mods:       C.ghostty_input_mods_e(ev.ConsumedMods),
		keycode:             C.uint32_t(ev.Keycode),
		unshifted_codepoint: C.uint32_t(ev.UnshiftedCodepoint),
		composing:           C.bool(ev.Composing),
	}
	if ev.Text != "" {
		cs := C.CString(ev.Text)
		defer C.free(unsafe.Pointer(cs))
		cev.text = cs
	}
	return bool(C.ghostty_surface_key(s.ptr, cev))
}

// MouseButton sends a mouse button event. Returns true if consumed.
func (s *Surface) MouseButton(state MouseButtonState, btn MouseButton, mods Mods) bool {
	return bool(C.ghostty_surface_mouse_button(
		s.ptr,
		C.ghostty_input_mouse_state_e(state),
		C.ghostty_input_mouse_button_e(btn),
		C.ghostty_input_mods_e(mods),
	))
}

// MousePos sends a mouse position update.
func (s *Surface) MousePos(x, y float64, mods Mods) {
	C.ghostty_surface_mouse_pos(s.ptr, C.double(x), C.double(y), C.ghostty_input_mods_e(mods))
}

// MouseScroll sends a scroll event.
func (s *Surface) MouseScroll(dx, dy float64, mods ScrollMods) {
	C.ghostty_surface_mouse_scroll(s.ptr, C.double(dx), C.double(dy), C.ghostty_input_scroll_mods_t(mods))
}

// SendText sends text directly to the terminal.
func (s *Surface) SendText(text string) {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	C.ghostty_surface_text(s.ptr, cs, C.uintptr_t(len(text)))
}

// RequestClose asks the surface to close itself.
func (s *Surface) RequestClose() {
	C.ghostty_surface_request_close(s.ptr)
}

// CompleteClipboardRequest completes a pending clipboard read request.
// req is the opaque pointer received in the ReadClipboard callback.
func (s *Surface) CompleteClipboardRequest(text string, req unsafe.Pointer, confirmed bool) {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	C.ghostty_surface_complete_clipboard_request(s.ptr, cs, req, C.bool(confirmed))
}
