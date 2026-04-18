//go:build darwin

package main

import "github.com/guigui-gui/guigui"

type Root struct {
	guigui.DefaultWidget
	terminal TerminalWidget
}

func (r *Root) Build(ctx *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&r.terminal)
	return nil
}

func (r *Root) Layout(ctx *guigui.Context, wb *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	layouter.LayoutWidget(&r.terminal, wb.Bounds())
}
