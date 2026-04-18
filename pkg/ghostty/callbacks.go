package ghostty

// #include "ghostty.h"
import "C"
import (
	"runtime/cgo"
	"unsafe"
)

//export ghosttyGoWakeupCB
func ghosttyGoWakeupCB(userdata unsafe.Pointer) {
	app := appFromHandle(userdata)
	if app != nil && app.callbacks.Wakeup != nil {
		app.callbacks.Wakeup()
	}
}

//export ghosttyGoActionCB
func ghosttyGoActionCB(appPtr unsafe.Pointer, target C.ghostty_target_s, action C.ghostty_action_s) C.bool {
	// Convert unsafe.Pointer → C.ghostty_app_t (both are void* at runtime).
	appTyped := *(*C.ghostty_app_t)(unsafe.Pointer(&appPtr))
	ud := C.ghostty_app_userdata(appTyped)
	app := appFromHandle(ud)
	if app == nil || app.callbacks.Action == nil {
		return C.bool(false)
	}

	t := Target{Tag: TargetTag(target.tag)}
	if target.tag == C.GHOSTTY_TARGET_SURFACE {
		// The target union's sole member is ghostty_surface_t (void*).
		// Read it as unsafe.Pointer then convert to C.ghostty_surface_t.
		rawPtr := *(*unsafe.Pointer)(unsafe.Pointer(&target.target))
		surfTyped := *(*C.ghostty_surface_t)(unsafe.Pointer(&rawPtr))
		ud2 := C.ghostty_surface_userdata(surfTyped)
		t.Surface = surfaceFromHandle(ud2)
	}

	a := Action{Tag: ActionTag(action.tag)}
	return C.bool(app.callbacks.Action(t, a))
}

//export ghosttyGoReadClipboardCB
func ghosttyGoReadClipboardCB(userdata unsafe.Pointer, clipboard C.ghostty_clipboard_e, req unsafe.Pointer) C.bool {
	surf := surfaceFromHandle(userdata)
	if surf == nil || surf.app.callbacks.ReadClipboard == nil {
		return C.bool(false)
	}
	return C.bool(surf.app.callbacks.ReadClipboard(surf, Clipboard(clipboard), req))
}

//export ghosttyGoConfirmReadClipboardCB
func ghosttyGoConfirmReadClipboardCB(userdata unsafe.Pointer, content *C.char, req unsafe.Pointer, reqType C.ghostty_clipboard_request_e) {
	surf := surfaceFromHandle(userdata)
	if surf == nil || surf.app.callbacks.ConfirmReadClipboard == nil {
		return
	}
	surf.app.callbacks.ConfirmReadClipboard(surf, C.GoString(content), req, ClipboardRequest(reqType))
}

//export ghosttyGoWriteClipboardCB
func ghosttyGoWriteClipboardCB(userdata unsafe.Pointer, clipboard C.ghostty_clipboard_e, contents *C.ghostty_clipboard_content_s, count C.size_t, confirm C.bool) {
	surf := surfaceFromHandle(userdata)
	if surf == nil || surf.app.callbacks.WriteClipboard == nil {
		return
	}

	n := int(count)
	items := make([]ClipboardContent, n)
	if n > 0 {
		// Treat the C array as a Go slice for iteration.
		slice := (*[1 << 28]C.ghostty_clipboard_content_s)(unsafe.Pointer(contents))[:n:n]
		for i, c := range slice {
			items[i] = ClipboardContent{
				MIME: C.GoString(c.mime),
				Data: C.GoString(c.data),
			}
		}
	}
	surf.app.callbacks.WriteClipboard(surf, Clipboard(clipboard), items, bool(confirm))
}

//export ghosttyGoCloseSurfaceCB
func ghosttyGoCloseSurfaceCB(userdata unsafe.Pointer, processAlive C.bool) {
	surf := surfaceFromHandle(userdata)
	if surf == nil || surf.app.callbacks.CloseSurface == nil {
		return
	}
	surf.app.callbacks.CloseSurface(surf, bool(processAlive))
}

// appFromHandle recovers an *App from a void* holding a cgo.Handle value.
func appFromHandle(userdata unsafe.Pointer) *App {
	if userdata == nil {
		return nil
	}
	h := cgo.Handle(uintptr(userdata))
	app, ok := h.Value().(*App)
	if !ok {
		return nil
	}
	return app
}

// surfaceFromHandle recovers a *Surface from a void* holding a cgo.Handle value.
func surfaceFromHandle(userdata unsafe.Pointer) *Surface {
	if userdata == nil {
		return nil
	}
	h := cgo.Handle(uintptr(userdata))
	surf, ok := h.Value().(*Surface)
	if !ok {
		return nil
	}
	return surf
}
