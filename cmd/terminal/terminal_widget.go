//go:build darwin

package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego/objc"
	"github.com/guigui-gui/guigui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/rin2yh/gostty/pkg/ghostty"
)

var debugEnabled = os.Getenv("GOSTTY_DEBUG") != ""

func debugf(format string, args ...any) {
	if !debugEnabled {
		return
	}
	log.Printf("[gostty] "+format, args...)
}

// ebitenToMacKeycode maps ebiten.Key to the macOS native virtual keycode.
// Ghostty's embedded apprt looks up the physical key by matching KeyEvent.keycode
// against the platform-native column in keycodes.zig (ghostty/src/input/keycodes.zig).
var ebitenToMacKeycode = map[ebiten.Key]uint32{
	// Writing system keys
	ebiten.KeyA:             0x00,
	ebiten.KeyB:             0x0b,
	ebiten.KeyC:             0x08,
	ebiten.KeyD:             0x02,
	ebiten.KeyE:             0x0e,
	ebiten.KeyF:             0x03,
	ebiten.KeyG:             0x05,
	ebiten.KeyH:             0x04,
	ebiten.KeyI:             0x22,
	ebiten.KeyJ:             0x26,
	ebiten.KeyK:             0x28,
	ebiten.KeyL:             0x25,
	ebiten.KeyM:             0x2e,
	ebiten.KeyN:             0x2d,
	ebiten.KeyO:             0x1f,
	ebiten.KeyP:             0x23,
	ebiten.KeyQ:             0x0c,
	ebiten.KeyR:             0x0f,
	ebiten.KeyS:             0x01,
	ebiten.KeyT:             0x11,
	ebiten.KeyU:             0x20,
	ebiten.KeyV:             0x09,
	ebiten.KeyW:             0x0d,
	ebiten.KeyX:             0x07,
	ebiten.KeyY:             0x10,
	ebiten.KeyZ:             0x06,
	ebiten.KeyDigit0:        0x1d,
	ebiten.KeyDigit1:        0x12,
	ebiten.KeyDigit2:        0x13,
	ebiten.KeyDigit3:        0x14,
	ebiten.KeyDigit4:        0x15,
	ebiten.KeyDigit5:        0x17,
	ebiten.KeyDigit6:        0x16,
	ebiten.KeyDigit7:        0x1a,
	ebiten.KeyDigit8:        0x1c,
	ebiten.KeyDigit9:        0x19,
	ebiten.KeyMinus:         0x1b,
	ebiten.KeyEqual:         0x18,
	ebiten.KeyBracketLeft:   0x21,
	ebiten.KeyBracketRight:  0x1e,
	ebiten.KeyBackslash:     0x2a,
	ebiten.KeySemicolon:     0x29,
	ebiten.KeyQuote:         0x27,
	ebiten.KeyBackquote:     0x32,
	ebiten.KeyComma:         0x2b,
	ebiten.KeyPeriod:        0x2f,
	ebiten.KeySlash:         0x2c,
	ebiten.KeyIntlBackslash: 0x0a,

	// Functional keys
	ebiten.KeyAltLeft:      0x3a,
	ebiten.KeyAltRight:     0x3d,
	ebiten.KeyBackspace:    0x33,
	ebiten.KeyCapsLock:     0x39,
	ebiten.KeyContextMenu:  0x6e,
	ebiten.KeyControlLeft:  0x3b,
	ebiten.KeyControlRight: 0x3e,
	ebiten.KeyEnter:        0x24,
	ebiten.KeyMetaLeft:     0x37,
	ebiten.KeyMetaRight:    0x36,
	ebiten.KeyShiftLeft:    0x38,
	ebiten.KeyShiftRight:   0x3c,
	ebiten.KeySpace:        0x31,
	ebiten.KeyTab:          0x30,

	// Control pad
	ebiten.KeyDelete:   0x75,
	ebiten.KeyEnd:      0x77,
	ebiten.KeyHome:     0x73,
	ebiten.KeyInsert:   0x72,
	ebiten.KeyPageDown: 0x79,
	ebiten.KeyPageUp:   0x74,

	// Arrow pad
	ebiten.KeyArrowDown:  0x7d,
	ebiten.KeyArrowLeft:  0x7b,
	ebiten.KeyArrowRight: 0x7c,
	ebiten.KeyArrowUp:    0x7e,

	// Numpad
	ebiten.KeyNumLock:        0x47,
	ebiten.KeyNumpad0:        0x52,
	ebiten.KeyNumpad1:        0x53,
	ebiten.KeyNumpad2:        0x54,
	ebiten.KeyNumpad3:        0x55,
	ebiten.KeyNumpad4:        0x56,
	ebiten.KeyNumpad5:        0x57,
	ebiten.KeyNumpad6:        0x58,
	ebiten.KeyNumpad7:        0x59,
	ebiten.KeyNumpad8:        0x5b,
	ebiten.KeyNumpad9:        0x5c,
	ebiten.KeyNumpadAdd:      0x45,
	ebiten.KeyNumpadDecimal:  0x41,
	ebiten.KeyNumpadDivide:   0x4b,
	ebiten.KeyNumpadEnter:    0x4c,
	ebiten.KeyNumpadEqual:    0x51,
	ebiten.KeyNumpadMultiply: 0x43,
	ebiten.KeyNumpadSubtract: 0x4e,

	// Function keys
	ebiten.KeyEscape: 0x35,
	ebiten.KeyF1:     0x7a,
	ebiten.KeyF2:     0x78,
	ebiten.KeyF3:     0x63,
	ebiten.KeyF4:     0x76,
	ebiten.KeyF5:     0x60,
	ebiten.KeyF6:     0x61,
	ebiten.KeyF7:     0x62,
	ebiten.KeyF8:     0x64,
	ebiten.KeyF9:     0x65,
	ebiten.KeyF10:    0x6d,
	ebiten.KeyF11:    0x67,
	ebiten.KeyF12:    0x6f,
	ebiten.KeyF13:    0x69,
	ebiten.KeyF14:    0x6b,
	ebiten.KeyF15:    0x71,
	ebiten.KeyF16:    0x6a,
	ebiten.KeyF17:    0x40,
	ebiten.KeyF18:    0x4f,
	ebiten.KeyF19:    0x50,
	ebiten.KeyF20:    0x5a,
	// Pause, ScrollLock, PrintScreen have no mapping on macOS.
}

func currentMods() ghostty.Mods {
	var mods ghostty.Mods
	if ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight) {
		mods |= ghostty.ModsShift
	}
	if ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight) {
		mods |= ghostty.ModsCtrl
	}
	if ebiten.IsKeyPressed(ebiten.KeyAltLeft) || ebiten.IsKeyPressed(ebiten.KeyAltRight) {
		mods |= ghostty.ModsAlt
	}
	if ebiten.IsKeyPressed(ebiten.KeyMetaLeft) || ebiten.IsKeyPressed(ebiten.KeyMetaRight) {
		mods |= ghostty.ModsSuper
	}
	return mods
}

type TerminalWidget struct {
	guigui.DefaultWidget
	surfaceOnce sync.Once
	surfaceErr  error
	app         *ghostty.App
	surface     *ghostty.Surface
	nsViewID    objc.ID
	lastBounds  image.Rectangle

	wakeupCh         chan struct{}
	justPressedKeys  []ebiten.Key
	justReleasedKeys []ebiten.Key
	inputChars       []rune
}

func (t *TerminalWidget) Build(ctx *guigui.Context, adder *guigui.ChildAdder) error {
	if mainWindowContentView() == 0 {
		return nil
	}
	t.surfaceOnce.Do(func() {
		t.surfaceErr = t.initSurface(ctx)
	})
	if t.surfaceErr != nil {
		return t.surfaceErr
	}
	ctx.SetFocused(t, true)
	return nil
}

func (t *TerminalWidget) initSurface(ctx *guigui.Context) error {
	scale := ctx.DeviceScale()
	cw, ch := getContentViewSize()
	if cw <= 0 || ch <= 0 {
		cw, ch = 512, 384
	}
	debugf("initSurface: contentView size = (%.0f, %.0f), scale = %.2f", cw, ch, scale)

	var (
		viewID  objc.ID
		surface *ghostty.Surface
		initErr error
	)
	ebiten.RunOnMainThread(func() {
		viewID = createNSView(0, 0, cw, ch)
		if viewID == 0 {
			initErr = fmt.Errorf("failed to create NSView")
			return
		}
		debugf("createNSView: nsview id = 0x%x", uintptr(viewID))

		// ObjC handle is not a Go pointer; bitcast via intermediate *unsafe.Pointer to avoid unsafeptr vet warning.
		nsview := *(*unsafe.Pointer)(unsafe.Pointer(&viewID))
		s, err := ghostty.NewSurface(t.app, ghostty.SurfaceConfig{
			NSView:      nsview,
			ScaleFactor: scale,
		})
		if err != nil {
			removeNSView(viewID)
			viewID = 0
			initErr = fmt.Errorf("ghostty surface: %w", err)
			return
		}
		surface = s
	})
	if initErr != nil {
		return initErr
	}

	t.nsViewID = viewID
	t.surface = surface
	t.surface.SetSize(uint32(cw*scale), uint32(ch*scale))
	t.surface.SetFocus(true)

	return nil
}

func (t *TerminalWidget) Tick(ctx *guigui.Context, wb *guigui.WidgetBounds) error {
	select {
	case <-t.wakeupCh:
	default:
	}
	t.app.Tick()

	if t.surface == nil {
		return nil
	}

	bounds := wb.Bounds()
	if !bounds.Empty() && bounds != t.lastBounds {
		t.lastBounds = bounds
		scale := ctx.DeviceScale()

		_, ch := getContentViewSize()
		x := float64(bounds.Min.X) / scale
		y := ch - float64(bounds.Max.Y)/scale
		w := float64(bounds.Dx()) / scale
		h := float64(bounds.Dy()) / scale
		debugf("bounds changed: x=%.1f y=%.1f w=%.1f h=%.1f", x, y, w, h)
		viewID := t.nsViewID
		ebiten.RunOnMainThread(func() {
			updateNSViewFrame(viewID, x, y, w, h)
		})

		t.surface.SetSize(uint32(bounds.Dx()), uint32(bounds.Dy()))
		t.surface.SetContentScale(scale, scale)
	}

	// NOTE: surface.Draw() はメインスレッドに移さない。
	//       ghostty の IOSurfaceLayer.setSurface が内部で必要に応じて
	//       main queue に dispatch_async するため、毎フレーム同期ディスパッチすると UI が詰まる。
	t.surface.Draw()
	guigui.RequestRedraw(t)
	return nil
}

func (t *TerminalWidget) HandleButtonInput(ctx *guigui.Context, wb *guigui.WidgetBounds) guigui.HandleInputResult {
	if t.surface == nil {
		return guigui.HandleInputResult{}
	}

	mods := currentMods()

	t.justPressedKeys = inpututil.AppendJustPressedKeys(t.justPressedKeys[:0])
	t.justReleasedKeys = inpututil.AppendJustReleasedKeys(t.justReleasedKeys[:0])
	type keyBatch struct {
		keys   []ebiten.Key
		action ghostty.InputAction
	}
	for _, b := range [2]keyBatch{
		{t.justPressedKeys, ghostty.InputActionPress},
		{t.justReleasedKeys, ghostty.InputActionRelease},
	} {
		for _, key := range b.keys {
			if kc, ok := ebitenToMacKeycode[key]; ok {
				t.surface.Key(ghostty.KeyEvent{
					Action:  b.action,
					Mods:    mods,
					Keycode: kc,
				})
			}
		}
	}

	t.inputChars = ebiten.AppendInputChars(t.inputChars[:0])
	if len(t.inputChars) > 0 {
		t.surface.SendText(string(t.inputChars))
	}

	return guigui.HandleInputByWidget(t)
}

func (t *TerminalWidget) HandlePointingInput(ctx *guigui.Context, wb *guigui.WidgetBounds) guigui.HandleInputResult {
	if t.surface == nil || !wb.IsHitAtCursor() {
		return guigui.HandleInputResult{}
	}

	bounds := wb.Bounds()
	cx, cy := ebiten.CursorPosition()
	rx := float64(cx - bounds.Min.X)
	ry := float64(cy - bounds.Min.Y)
	mods := currentMods()

	t.surface.MousePos(rx, ry, mods)

	type btnPair struct {
		eb ebiten.MouseButton
		gb ghostty.MouseButton
	}
	for _, b := range []btnPair{
		{ebiten.MouseButtonLeft, ghostty.MouseLeft},
		{ebiten.MouseButtonRight, ghostty.MouseRight},
		{ebiten.MouseButtonMiddle, ghostty.MouseMiddle},
	} {
		if inpututil.IsMouseButtonJustPressed(b.eb) {
			t.surface.MouseButton(ghostty.MouseStatePress, b.gb, mods)
		}
		if inpututil.IsMouseButtonJustReleased(b.eb) {
			t.surface.MouseButton(ghostty.MouseStateRelease, b.gb, mods)
		}
	}

	wx, wy := ebiten.Wheel()
	if wx != 0 || wy != 0 {
		t.surface.MouseScroll(wx, -wy, 0)
	}

	return guigui.HandleInputByWidget(t)
}

func (t *TerminalWidget) Draw(ctx *guigui.Context, wb *guigui.WidgetBounds, dst *ebiten.Image) {
	// NSView の Metal レイヤーが独自に描画するため、ebiten の dst への書き込みは不要
}

func (t *TerminalWidget) Dispose() {
	viewID := t.nsViewID
	surface := t.surface
	t.nsViewID = 0
	t.surface = nil
	if viewID == 0 && surface == nil {
		return
	}
	// surface.Free と removeNSView を同じ main thread closure で順序保証する。
	// IOSurfaceLayer 側の dispatch_async が view 解放後に走らないようにするため。
	ebiten.RunOnMainThread(func() {
		if surface != nil {
			surface.Free()
		}
		if viewID != 0 {
			removeNSView(viewID)
		}
	})
}
