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

// ghosttyGoActionCB is declared in app.go as taking ghostty_app_t, but the
// Go export uses unsafe.Pointer. This is intentional: ghostty_app_t is
// typedef void*, so both map to the same ABI type and are compatible.
//
//export ghosttyGoActionCB
func ghosttyGoActionCB(appPtr unsafe.Pointer, target C.ghostty_target_s, action C.ghostty_action_s) C.bool {
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
	var items []ClipboardContent
	if n > 0 {
		items = make([]ClipboardContent, n)
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

// handleValue recovers a *T from a void* holding a cgo.Handle value.
func handleValue[T any](userdata unsafe.Pointer) *T {
	if userdata == nil {
		return nil
	}
	v, ok := cgo.Handle(uintptr(userdata)).Value().(*T)
	if !ok {
		return nil
	}
	return v
}

func appFromHandle(userdata unsafe.Pointer) *App {
	return handleValue[App](userdata)
}

func surfaceFromHandle(userdata unsafe.Pointer) *Surface {
	return handleValue[Surface](userdata)
}
