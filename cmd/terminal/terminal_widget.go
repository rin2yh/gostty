//go:build darwin

package main

import (
	"fmt"
	"image"
	"sync"
	"unsafe"

	"github.com/guigui-gui/guigui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/rin2yh/gostty/pkg/ghostty"
)

var ebitenToGhosttyKey = map[ebiten.Key]ghostty.Key{
	// Writing system keys
	ebiten.KeyBackquote:     ghostty.KeyBackquote,
	ebiten.KeyBackslash:     ghostty.KeyBackslash,
	ebiten.KeyBracketLeft:   ghostty.KeyBracketLeft,
	ebiten.KeyBracketRight:  ghostty.KeyBracketRight,
	ebiten.KeyComma:         ghostty.KeyComma,
	ebiten.KeyDigit0:        ghostty.KeyDigit0,
	ebiten.KeyDigit1:        ghostty.KeyDigit1,
	ebiten.KeyDigit2:        ghostty.KeyDigit2,
	ebiten.KeyDigit3:        ghostty.KeyDigit3,
	ebiten.KeyDigit4:        ghostty.KeyDigit4,
	ebiten.KeyDigit5:        ghostty.KeyDigit5,
	ebiten.KeyDigit6:        ghostty.KeyDigit6,
	ebiten.KeyDigit7:        ghostty.KeyDigit7,
	ebiten.KeyDigit8:        ghostty.KeyDigit8,
	ebiten.KeyDigit9:        ghostty.KeyDigit9,
	ebiten.KeyEqual:         ghostty.KeyEqual,
	ebiten.KeyIntlBackslash: ghostty.KeyIntlBackslash,
	ebiten.KeyA:             ghostty.KeyA,
	ebiten.KeyB:             ghostty.KeyB,
	ebiten.KeyC:             ghostty.KeyC,
	ebiten.KeyD:             ghostty.KeyD,
	ebiten.KeyE:             ghostty.KeyE,
	ebiten.KeyF:             ghostty.KeyF,
	ebiten.KeyG:             ghostty.KeyG,
	ebiten.KeyH:             ghostty.KeyH,
	ebiten.KeyI:             ghostty.KeyI,
	ebiten.KeyJ:             ghostty.KeyJ,
	ebiten.KeyK:             ghostty.KeyK,
	ebiten.KeyL:             ghostty.KeyL,
	ebiten.KeyM:             ghostty.KeyM,
	ebiten.KeyN:             ghostty.KeyN,
	ebiten.KeyO:             ghostty.KeyO,
	ebiten.KeyP:             ghostty.KeyP,
	ebiten.KeyQ:             ghostty.KeyQ,
	ebiten.KeyR:             ghostty.KeyR,
	ebiten.KeyS:             ghostty.KeyS,
	ebiten.KeyT:             ghostty.KeyT,
	ebiten.KeyU:             ghostty.KeyU,
	ebiten.KeyV:             ghostty.KeyV,
	ebiten.KeyW:             ghostty.KeyW,
	ebiten.KeyX:             ghostty.KeyX,
	ebiten.KeyY:             ghostty.KeyY,
	ebiten.KeyZ:             ghostty.KeyZ,
	ebiten.KeyMinus:         ghostty.KeyMinus,
	ebiten.KeyPeriod:        ghostty.KeyPeriod,
	ebiten.KeyQuote:         ghostty.KeyQuote,
	ebiten.KeySemicolon:     ghostty.KeySemicolon,
	ebiten.KeySlash:         ghostty.KeySlash,

	// Functional keys
	ebiten.KeyAltLeft:      ghostty.KeyAltLeft,
	ebiten.KeyAltRight:     ghostty.KeyAltRight,
	ebiten.KeyBackspace:    ghostty.KeyBackspace,
	ebiten.KeyCapsLock:     ghostty.KeyCapsLock,
	ebiten.KeyContextMenu:  ghostty.KeyContextMenu,
	ebiten.KeyControlLeft:  ghostty.KeyControlLeft,
	ebiten.KeyControlRight: ghostty.KeyControlRight,
	ebiten.KeyEnter:        ghostty.KeyEnter,
	ebiten.KeyMetaLeft:     ghostty.KeyMetaLeft,
	ebiten.KeyMetaRight:    ghostty.KeyMetaRight,
	ebiten.KeyShiftLeft:    ghostty.KeyShiftLeft,
	ebiten.KeyShiftRight:   ghostty.KeyShiftRight,
	ebiten.KeySpace:        ghostty.KeySpace,
	ebiten.KeyTab:          ghostty.KeyTab,

	// Control pad
	ebiten.KeyDelete:   ghostty.KeyDelete,
	ebiten.KeyEnd:      ghostty.KeyEnd,
	ebiten.KeyHome:     ghostty.KeyHome,
	ebiten.KeyInsert:   ghostty.KeyInsert,
	ebiten.KeyPageDown: ghostty.KeyPageDown,
	ebiten.KeyPageUp:   ghostty.KeyPageUp,

	// Arrow pad
	ebiten.KeyArrowDown:  ghostty.KeyArrowDown,
	ebiten.KeyArrowLeft:  ghostty.KeyArrowLeft,
	ebiten.KeyArrowRight: ghostty.KeyArrowRight,
	ebiten.KeyArrowUp:    ghostty.KeyArrowUp,

	// Numpad
	ebiten.KeyNumLock:        ghostty.KeyNumLock,
	ebiten.KeyNumpad0:        ghostty.KeyNumpad0,
	ebiten.KeyNumpad1:        ghostty.KeyNumpad1,
	ebiten.KeyNumpad2:        ghostty.KeyNumpad2,
	ebiten.KeyNumpad3:        ghostty.KeyNumpad3,
	ebiten.KeyNumpad4:        ghostty.KeyNumpad4,
	ebiten.KeyNumpad5:        ghostty.KeyNumpad5,
	ebiten.KeyNumpad6:        ghostty.KeyNumpad6,
	ebiten.KeyNumpad7:        ghostty.KeyNumpad7,
	ebiten.KeyNumpad8:        ghostty.KeyNumpad8,
	ebiten.KeyNumpad9:        ghostty.KeyNumpad9,
	ebiten.KeyNumpadAdd:      ghostty.KeyNumpadAdd,
	ebiten.KeyNumpadDecimal:  ghostty.KeyNumpadDecimal,
	ebiten.KeyNumpadDivide:   ghostty.KeyNumpadDivide,
	ebiten.KeyNumpadEnter:    ghostty.KeyNumpadEnter,
	ebiten.KeyNumpadEqual:    ghostty.KeyNumpadEqual,
	ebiten.KeyNumpadMultiply: ghostty.KeyNumpadMultiply,
	ebiten.KeyNumpadSubtract: ghostty.KeyNumpadSubtract,

	// Function keys
	ebiten.KeyEscape:      ghostty.KeyEscape,
	ebiten.KeyF1:          ghostty.KeyF1,
	ebiten.KeyF2:          ghostty.KeyF2,
	ebiten.KeyF3:          ghostty.KeyF3,
	ebiten.KeyF4:          ghostty.KeyF4,
	ebiten.KeyF5:          ghostty.KeyF5,
	ebiten.KeyF6:          ghostty.KeyF6,
	ebiten.KeyF7:          ghostty.KeyF7,
	ebiten.KeyF8:          ghostty.KeyF8,
	ebiten.KeyF9:          ghostty.KeyF9,
	ebiten.KeyF10:         ghostty.KeyF10,
	ebiten.KeyF11:         ghostty.KeyF11,
	ebiten.KeyF12:         ghostty.KeyF12,
	ebiten.KeyF13:         ghostty.KeyF13,
	ebiten.KeyF14:         ghostty.KeyF14,
	ebiten.KeyF15:         ghostty.KeyF15,
	ebiten.KeyF16:         ghostty.KeyF16,
	ebiten.KeyF17:         ghostty.KeyF17,
	ebiten.KeyF18:         ghostty.KeyF18,
	ebiten.KeyF19:         ghostty.KeyF19,
	ebiten.KeyF20:         ghostty.KeyF20,
	ebiten.KeyPause:       ghostty.KeyPause,
	ebiten.KeyScrollLock:  ghostty.KeyScrollLock,
	ebiten.KeyPrintScreen: ghostty.KeyPrintScreen,
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
	nsViewPtr   unsafe.Pointer
	lastBounds  image.Rectangle

	wakeupCh         chan struct{}
	justPressedKeys  []ebiten.Key
	justReleasedKeys []ebiten.Key
	inputChars       []rune
}

func (t *TerminalWidget) Build(ctx *guigui.Context, adder *guigui.ChildAdder) error {
	if !hasMainWindow() {
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

	t.nsViewPtr = createNSView(0, 0, cw, ch)
	if t.nsViewPtr == nil {
		return fmt.Errorf("failed to create NSView")
	}

	surface, err := ghostty.NewSurface(t.app, ghostty.SurfaceConfig{
		NSView:      t.nsViewPtr,
		ScaleFactor: scale,
	})
	if err != nil {
		return fmt.Errorf("ghostty surface: %w", err)
	}
	t.surface = surface

	t.surface.SetSize(uint32(cw*scale), uint32(ch*scale))
	t.surface.SetFocus(true)

	return nil
}

func (t *TerminalWidget) Tick(ctx *guigui.Context, wb *guigui.WidgetBounds) error {
	if t.surface == nil {
		return nil
	}

	select {
	case <-t.wakeupCh:
		t.app.Tick()
	default:
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
		updateNSViewFrame(t.nsViewPtr, x, y, w, h)

		t.surface.SetSize(uint32(bounds.Dx()), uint32(bounds.Dy()))
		t.surface.SetContentScale(scale, scale)
	}

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
	for _, key := range t.justPressedKeys {
		if gk, ok := ebitenToGhosttyKey[key]; ok {
			t.surface.Key(ghostty.KeyEvent{
				Action:  ghostty.InputActionPress,
				Mods:    mods,
				Keycode: uint32(gk),
			})
		}
	}

	t.justReleasedKeys = inpututil.AppendJustReleasedKeys(t.justReleasedKeys[:0])
	for _, key := range t.justReleasedKeys {
		if gk, ok := ebitenToGhosttyKey[key]; ok {
			t.surface.Key(ghostty.KeyEvent{
				Action:  ghostty.InputActionRelease,
				Mods:    mods,
				Keycode: uint32(gk),
			})
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
	// ghostty は NSView の Metal レイヤーに独立して描画する
}
