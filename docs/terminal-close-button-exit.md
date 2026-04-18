# Mac close ボタン押下でプロセスが残る問題の修正

## 症状 / 背景

gostty (`cmd/terminal`) を macOS で起動し、ウィンドウ左上の赤い × (close) ボタンを押してもバイナリのプロセスがハングしたまま残り続ける。ウィンドウ自体は閉じるが、`pgrep -f ./terminal` にヒットし続ける。

## 根本原因

確定。

`cmd/terminal/terminal_widget.go` の `Dispose()` が `ebiten.RunOnMainThread(...)` を使って ghostty surface と NSView を解放している。`Dispose()` は `cmd/terminal/main.go` の `defer root.terminal.Dispose()` から、`guigui.Run()` が return した**後**に呼ばれる。

ebiten の `RunOnMainThread` は `ebiten/v2/internal/thread/thread.go:54-83` の `OSThread.Call` 経由で unbuffered channel にブロッキング送信する実装（`t.funcs <- queueItem{...}` → `<-t.done`）。しかし ebiten の main thread loop は `ebiten.RunGameWithOptions` から抜けた時点で停止しているため、送信先が受信しない → `Dispose()` が永久にブロック → `main` が return できず Go ランタイムが `os.Exit` を呼ばず、プロセスが残る。

build tag `ebitenginesinglethread` が付いていない通常ビルドでは `OSThread`（deadlock 経路）が使われる。single thread モードなら `NoopThread` で即実行されるが、このビルドではそのタグが無い。

## 対処

close イベントを「ebiten がまだ動いているうち」に検知し、そのタイミングで `Dispose()` を呼んでから `ebiten.Termination` を返す。`main` の defer に残る `Dispose()` は二重呼び出し防止の既存ガード（`viewID == 0 && surface == nil`）によって no-op になるので deadlock は起こらない。

### `cmd/terminal/main.go`

`guigui.Run` 前に `ebiten.SetWindowClosingHandled(true)` を呼んで ebiten の自動終了をオフにし、アプリ側で終了制御する。

```go
ebiten.SetWindowClosingHandled(true)
if err := guigui.Run(root, op); err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
}
```

### `cmd/terminal/terminal_widget.go`

`Tick` 冒頭で `ebiten.IsWindowBeingClosed()` を確認し、true なら `Dispose()` を呼んで `ebiten.Termination` を返す。

```go
func (t *TerminalWidget) Tick(ctx *guigui.Context, wb *guigui.WidgetBounds) error {
    if ebiten.IsWindowBeingClosed() {
        t.Dispose()
        return ebiten.Termination
    }
    // ...既存処理
}
```

`Tick` は ebiten の Update ループ中（= ebiten main thread loop が生きている）に呼ばれるため、`Dispose()` 内の `ebiten.RunOnMainThread` は正常に drain される。

### 意図的に変えなかった箇所

- 既存の `Dispose()`（terminal_widget.go:368-386）の同期クローズ処理はそのまま再利用。`surface.Free()` と `removeNSView()` を同じ main thread closure で順序保証する設計は変えない。
- `main.go` の `defer root.terminal.Dispose()` も残す。既存 nil ガードで no-op になるため、フォールバック（異常終了パス）として価値がある。

### 検討した代替案

- **`Dispose()` を `RunOnMainThread` なしで実装する**: 不可。IOSurfaceLayer の dispatch_async と view 解放の順序保証ができなくなる。
- **`ebitenginesinglethread` build tag を付ける**: NoopThread で即実行になるため deadlock は解消するが、ebiten の他のスレッド前提が崩れる可能性がある。close イベント検知で解決できるなら build tag 追加は不要。

## デバッグ / 検証方法

1. ビルド
   ```bash
   go build -o terminal ./cmd/terminal
   ```
2. `./terminal` を起動
3. 赤い × をクリック → ウィンドウが閉じる
4. 別ターミナルで `pgrep -f ./terminal` を実行し、何もヒットしないことを確認
5. CI
   ```bash
   go fmt ./...
   go vet ./...
   go test ./...
   ```

## 参照

- `cmd/terminal/main.go` — `ebiten.SetWindowClosingHandled` 呼び出し箇所
- `cmd/terminal/terminal_widget.go` — `Tick` / `Dispose`
- `guigui/app.go:344` — `app.Update` (widget.Tick のエラーを透過して ebiten に返す)
- `ebiten/v2/internal/thread/thread.go:54-83` — `OSThread.Call` の unbuffered channel 実装
- `ebiten/v2/internal/glfw/cocoa_window_darwin.go:207` — NSWindow `windowShouldClose` の委譲先
- 関連ログ: [terminal-render-reliability.md](terminal-render-reliability.md) — 同じ `RunOnMainThread` 経路に関する先行調査
