package ghostty_test

import (
	"testing"

	"github.com/rin2yh/gostty/pkg/ghostty"
)

func TestInit(t *testing.T) {
	if err := ghostty.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
}

func TestGetInfo(t *testing.T) {
	ghostty.Init() //nolint:errcheck
	info := ghostty.GetInfo()
	if info.Version == "" {
		t.Error("GetInfo returned empty version")
	}
	t.Logf("ghostty version=%q buildMode=%d", info.Version, info.BuildMode)
}

func TestConfigLifecycle(t *testing.T) {
	ghostty.Init() //nolint:errcheck

	cfg := ghostty.NewConfig()
	if cfg == nil {
		t.Fatal("NewConfig returned nil")
	}
	cfg.LoadDefaultFiles()
	cfg.Finalize()
	cfg.Free()
}

func TestAppLifecycle(t *testing.T) {
	ghostty.Init() //nolint:errcheck

	cfg := ghostty.NewConfig()
	if cfg == nil {
		t.Fatal("NewConfig returned nil")
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
