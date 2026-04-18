# libghostty Go ブリッジ実装記録

## 概要

ghostty（Zigで書かれたターミナルエミュレータ）のC API（libghostty）を
GoのcgoブリッジパッケージとしてラップしたGoパッケージ `pkg/ghostty` の実装記録。

---

## プロジェクト構成

```
gostty/
├── go.mod
├── ghostty/          (submodule)
├── guigui/           (submodule)
└── pkg/
    └── ghostty/
        ├── ghostty.go       # cgoディレクティブ + Init / GetInfo
        ├── types.go         # Go型定義（enum, struct等）
        ├── config.go        # Config型ラッパー
        ├── app.go           # App型ラッパー + Cコールバック登録
        ├── surface.go       # Surface型ラッパー
        ├── callbacks.go     # //export cgo関数（コールバック実装）
        └── ghostty_test.go  # ライフサイクルテスト
```

---

## ビルド手順

### 1. zigのセットアップ

プロジェクトルート（`gostty/`）で mise を使い、zig 0.15.2 をインストール。

```sh
# gostty/ ディレクトリで実行（ghostty/ submodule 内では実行しない）
mise use zig@0.15.2
```

### 2. ghosttyライブラリのビルド

```sh
cd ghostty
zig build -Dapp-runtime=none -Doptimize=ReleaseSafe
```

ビルド前に以下の2つのコマンドが必要だった。

```sh
# Metal Toolchain のインストール（要認証）
sudo xcodebuild -runFirstLaunch
xcodebuild -downloadComponent MetalToolchain
```

### 3. ライブラリの出力先

計画では `ghostty/zig-out/lib/libghostty-internal.a` を想定していたが、
macOS上では xcframework としてビルドされるため実際の出力先は異なる。

```
ghostty/macos/GhosttyKit.xcframework/macos-arm64_x86_64/ghostty-internal.a
```

fat binary（arm64 + x86_64 の universal）として生成される。

---

## 実装詳細

### cgoリンク設定（ghostty.go）

```go
// #cgo CFLAGS: -I../../ghostty/include -DGHOSTTY_STATIC
// #cgo darwin LDFLAGS: ${SRCDIR}/../../ghostty/macos/GhosttyKit.xcframework/macos-arm64_x86_64/ghostty-internal.a
// #cgo darwin LDFLAGS: -framework Foundation -framework AppKit -framework Carbon
// #cgo darwin LDFLAGS: -framework Metal -framework MetalKit -framework QuartzCore
// #cgo darwin LDFLAGS: -framework IOSurface -framework CoreText -framework CoreGraphics
// #cgo darwin LDFLAGS: -lc++
```

### コールバック登録（app.go）

`ghostty_runtime_config_s` へのコールバック登録はCの静的ヘルパー関数 `ghosttyMakeRuntimeConfig` を経由する。
これにより、`//export` を持つ `callbacks.go` からC関数ポインタを参照せずに済む。

```c
static ghostty_runtime_config_s ghosttyMakeRuntimeConfig(void* userdata) {
    ghostty_runtime_config_s cfg;
    cfg.userdata                     = userdata;
    cfg.supports_selection_clipboard = true;
    cfg.wakeup_cb                    = ghosttyGoWakeupCB;
    cfg.action_cb                    = ghosttyGoActionCB;
    cfg.read_clipboard_cb            = ghosttyGoReadClipboardCB;
    cfg.confirm_read_clipboard_cb    = ghosttyGoConfirmReadClipboardCB;
    cfg.write_clipboard_cb           = ghosttyGoWriteClipboardCB;
    cfg.close_surface_cb             = ghosttyGoCloseSurfaceCB;
    return cfg;
}
```

### cgo.Handle によるオブジェクト管理

GoオブジェクトへのポインタをC側に渡すため `runtime/cgo` パッケージの `cgo.Handle` を使用。

```go
// App生成時
app.handle = cgo.NewHandle(app)
rtcfg := C.ghosttyMakeRuntimeConfig(C.ghosttyHandleToPtr(C.uintptr_t(app.handle)))

// コールバック側での復元
func appFromHandle(userdata unsafe.Pointer) *App {
    h := cgo.Handle(uintptr(userdata))
    app, _ := h.Value().(*App)
    return app
}
```

### コールバックの userdata の出所

ghosttyの実装（`src/apprt/embedded.zig`）を読んだ結果、コールバック種別によって
渡される userdata の出所が異なることがわかった。

| コールバック | userdata の出所 |
|---|---|
| `wakeup_cb` | App の userdata（runtime_config で設定） |
| `action_cb` | 引数が `ghostty_app_t`。`ghostty_app_userdata()` で取得 |
| `read_clipboard_cb` | Surface の userdata（surface_config で設定） |
| `confirm_read_clipboard_cb` | Surface の userdata |
| `write_clipboard_cb` | Surface の userdata |
| `close_surface_cb` | Surface の userdata |

---

## トラブルシューティング

### 問題1: ライブラリ名の `lib` プレフィックス欠如

`-lghostty-internal` でリンクしようとしたが、実際のファイル名は
`ghostty-internal.a`（`lib` プレフィックスなし）のため `ld` が見つけられなかった。

**解決策:** `-l` フラグではなくアーカイブファイルのパスを直接指定。
cgoのセキュリティチェックを通すため `${SRCDIR}` を使って絶対パスに展開する。

```go
// #cgo darwin LDFLAGS: ${SRCDIR}/../../ghostty/macos/GhosttyKit.xcframework/macos-arm64_x86_64/ghostty-internal.a
```

### 問題2: `typedef void*` の型ミスマッチ

`ghostty_app_t`・`ghostty_surface_t` は `typedef void*` だが、
cgoは `unsafe.Pointer` と distinct な型として扱うため型ミスマッチが発生。

```
cannot use _cgo0 (variable of type unsafe.Pointer) as _Ctype_ghostty_app_t value
```

**解決策:** ポインタのバイト表現を直接読み替えてキャスト。

```go
appTyped := *(*C.ghostty_app_t)(unsafe.Pointer(&appPtr))
ud := C.ghostty_app_userdata(appTyped)
```

### 問題3: `go vet` の unsafeptr 警告

`unsafe.Pointer(uintptr(handle))` というパターンに対して `go vet` が警告を出す。
`cgo.Handle` は GCが管理するポインタではなく安定した整数インデックスであるため
安全だが、`go vet` はこれを区別できない。

**解決策:** Cのヘルパーでuintptrをvoid*に変換し、Goで `unsafe.Pointer` への変換をしない。

```c
// app.go / surface.go プリアンブルに定義
static void* ghosttyHandleToPtr(uintptr_t h) { return (void*)h; }
```

```go
// Go側での使用
C.ghosttyHandleToPtr(C.uintptr_t(app.handle))
```

### 問題4: `ghostty_init` の二重呼び出し

テスト間で `Init()` が複数回呼ばれると Zig の unreachable コードに到達してクラッシュ。

**解決策:** `sync.Once` で初期化を冪等化。

```go
var initOnce sync.Once

func Init() error {
    initOnce.Do(func() {
        if ret := C.ghostty_init(0, nil); ret != C.GHOSTTY_SUCCESS {
            initErr = errors.New("ghostty_init failed")
        }
    })
    return initErr
}
```

### 問題5: 不足フレームワーク

リンク時に `IOSurface`・`TIS*`（Carbon）・C++標準ライブラリのシンボルが未解決になった。

**解決策:** 以下のフレームワークとライブラリを追加。

```
-framework Carbon
-framework IOSurface
-framework CoreText
-framework CoreGraphics
-lc++
```

---

## テスト結果

```
=== RUN   TestInit
--- PASS: TestInit (0.00s)
=== RUN   TestGetInfo
    ghostty version="1.3.2-main-+ca7516bea" buildMode=1 (ReleaseSafe)
--- PASS: TestGetInfo (0.00s)
=== RUN   TestConfigLifecycle
--- PASS: TestConfigLifecycle (0.01s)
=== RUN   TestAppLifecycle
--- PASS: TestAppLifecycle (0.04s)
PASS
ok  github.com/rin2yh/gostty/pkg/ghostty  2.4s
```

`go build` および `go vet` もエラー・警告なし。
