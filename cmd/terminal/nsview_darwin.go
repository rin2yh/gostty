//go:build darwin && !ios

package main

import (
	"fmt"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

type nsPoint struct{ X, Y float64 }
type nsSize struct{ Width, Height float64 }
type nsRect struct {
	Origin nsPoint
	Size   nsSize
}

var (
	class_NSApplication objc.Class
	class_NSView        objc.Class
)

var (
	sel_sharedApplication   = objc.RegisterName("sharedApplication")
	sel_mainWindow          = objc.RegisterName("mainWindow")
	sel_contentView         = objc.RegisterName("contentView")
	sel_frame               = objc.RegisterName("frame")
	sel_alloc               = objc.RegisterName("alloc")
	sel_initWithFrame       = objc.RegisterName("initWithFrame:")
	sel_addSubview          = objc.RegisterName("addSubview:")
	sel_setFrame            = objc.RegisterName("setFrame:")
	sel_removeFromSuperview = objc.RegisterName("removeFromSuperview")
)

func init() {
	if _, err := purego.Dlopen("/System/Library/Frameworks/AppKit.framework/AppKit", purego.RTLD_LAZY|purego.RTLD_GLOBAL); err != nil {
		panic(fmt.Errorf("nsview: failed to dlopen AppKit: %w", err))
	}
	class_NSApplication = objc.GetClass("NSApplication")
	class_NSView = objc.GetClass("NSView")
}

func mainWindowContentView() objc.ID {
	app := objc.ID(class_NSApplication).Send(sel_sharedApplication)
	win := app.Send(sel_mainWindow)
	if win == 0 {
		return 0
	}
	return win.Send(sel_contentView)
}

func getContentViewSize() (float64, float64) {
	cv := mainWindowContentView()
	if cv == 0 {
		return 0, 0
	}
	r := objc.Send[nsRect](cv, sel_frame)
	return r.Size.Width, r.Size.Height
}

func createNSView(x, y, w, h float64) objc.ID {
	cv := mainWindowContentView()
	if cv == 0 {
		return 0
	}
	r := nsRect{nsPoint{x, y}, nsSize{w, h}}
	view := objc.ID(class_NSView).Send(sel_alloc)
	view = objc.Send[objc.ID](view, sel_initWithFrame, r)
	// NOTE: setWantsLayer は呼ばない。ghostty_surface_new 内で
	//       layer を IOSurfaceLayer に差し替えた "後" に wantsLayer=YES される仕様のため、
	//       ここで先に呼ぶと layer-backed view としての内部状態が残り描画が不定になる。
	cv.Send(sel_addSubview, view)
	return view
}

func updateNSViewFrame(view objc.ID, x, y, w, h float64) {
	view.Send(sel_setFrame, nsRect{nsPoint{x, y}, nsSize{w, h}})
}

func removeNSView(view objc.ID) {
	view.Send(sel_removeFromSuperview)
}
