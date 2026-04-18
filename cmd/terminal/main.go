//go:build darwin

package main

import (
	"fmt"
	"image"
	"os"
	"unsafe"

	"github.com/guigui-gui/guigui"
	"github.com/rin2yh/gostty/pkg/ghostty"
)

func main() {
	wakeupCh := make(chan struct{}, 1)

	if err := ghostty.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "ghostty init:", err)
		os.Exit(1)
	}

	cfg, err := ghostty.NewConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ghostty config:", err)
		os.Exit(1)
	}
	cfg.LoadDefaultFiles()
	cfg.Finalize()
	defer cfg.Free()

	callbacks := ghostty.RuntimeCallbacks{
		Wakeup: func() {
			select {
			case wakeupCh <- struct{}{}:
			default:
			}
		},
		CloseSurface:         func(s *ghostty.Surface, processAlive bool) {},
		WriteClipboard:       func(s *ghostty.Surface, cb ghostty.Clipboard, contents []ghostty.ClipboardContent, confirm bool) {},
		ReadClipboard:        func(s *ghostty.Surface, cb ghostty.Clipboard, req unsafe.Pointer) bool { return false },
		ConfirmReadClipboard: func(s *ghostty.Surface, content string, req unsafe.Pointer, reqType ghostty.ClipboardRequest) {},
	}

	app, err := ghostty.NewApp(callbacks, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ghostty app:", err)
		os.Exit(1)
	}
	defer app.Free()

	root := &Root{}
	root.terminal.wakeupCh = wakeupCh
	root.terminal.app = app
	defer root.terminal.Dispose()

	op := &guigui.RunOptions{
		Title:         "gostty",
		WindowSize:    image.Pt(1024, 768),
		WindowMinSize: image.Pt(400, 300),
	}
	if err := guigui.Run(root, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
