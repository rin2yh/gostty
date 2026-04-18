package ghostty_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/rin2yh/gostty/pkg/ghostty"
)

func TestMain(m *testing.M) {
	if err := ghostty.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "ghostty.Init failed: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestInit(t *testing.T) {
	if err := ghostty.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
}

func TestGetInfo(t *testing.T) {
	info := ghostty.GetInfo()
	if info.Version == "" {
		t.Error("GetInfo returned empty version")
	}
	t.Logf("ghostty version=%q buildMode=%d", info.Version, info.BuildMode)
}

func TestConfigLifecycle(t *testing.T) {
	cfg, err := ghostty.NewConfig()
	if err != nil {
		t.Fatalf("NewConfig failed: %v", err)
	}
	cfg.LoadDefaultFiles()
	cfg.Finalize()
	cfg.Free()
}

func TestAppLifecycle(t *testing.T) {
	cfg, err := ghostty.NewConfig()
	if err != nil {
		t.Fatalf("NewConfig failed: %v", err)
	}
	cfg.Finalize()
	defer cfg.Free()

	var wakeupCalled bool
	app, err := ghostty.NewApp(ghostty.RuntimeCallbacks{
		Wakeup: func() { wakeupCalled = true },
	}, cfg)
	if err != nil {
		t.Fatalf("NewApp failed: %v", err)
	}
	defer app.Free()

	app.Tick()
	app.SetFocus(true)
	_ = app.NeedsConfirmQuit()
	_ = wakeupCalled
}
