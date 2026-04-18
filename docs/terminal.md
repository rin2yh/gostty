# cmd/terminal 実装記録

## 概要

`cmd/terminal` は guigui の `TerminalWidget` として ghostty のターミナルを表示するアプリケーション。
NSView を作成して ghostty の Surface をレンダリングする。

---

## プロジェクト構成

```
cmd/terminal/
├── main.go              # エントリポイント・Ghostty 事前初期化
├── root.go              # guigui ルートウィジェット
├── terminal_widget.go   # TerminalWidget（キー/マウス入力・描画）
└── nsview_darwin.go     # NSView 操作（purego/objc）
```

---

## スレッドモデルと初期化

### 問題：非メインスレッドからの ghostty 初期化による SIGILL

`ghostty_app_new` → `App.init` → `input.Keymap.init()` → `TISCopyCurrentKeyboardLayoutInputSource()` は
macOS の TIS API を呼び出す。TIS API はメインスレッドからの呼び出しが必要。

ebitengine のゲームループ構造：

```
goroutine 1  (メインスレッド): OSThread.Loop() でブロック待機
goroutine 53: LoopRenderThread()
goroutine 54 (ゲームループ): loopGame() → Update() → Build() → initialize()
```

`Build()` 内で `ghostty_app_new` を呼ぶと非メインスレッドで実行されるため、
Objective-C 例外が Zig/C フレームを越えて `objc_terminate` → `ud2` → SIGILL が発生する。

### 解決策：初期化を main() に移動

`ghostty.Init()` / `ghostty.NewApp()` を `guigui.Run()` 呼び出し前（メインスレッド上）で実行する。
NSView が必要な `ghostty.NewSurface()` は引き続き `Build()` 内（`surfaceOnce` で一度だけ）で実行する。

```
main() [メインスレッド]
  ├── ghostty.Init()
  ├── ghostty.NewConfig() / cfg.Finalize()
  ├── ghostty.NewApp(callbacks, cfg)   ← TIS API はここで呼ばれる
  ├── root.terminal.app = app
  └── guigui.Run(root, op)
        └── Build() [ゲームループスレッド]
              └── surfaceOnce.Do → initSurface()
                    ├── createNSView()
                    └── ghostty.NewSurface(t.app, ...)
```

---

## NSView 操作（nsview_darwin.go）

CGO を使わず `purego/objc` で Objective-C ランタイムを呼び出す。

- `mainWindowContentView()` — メインウィンドウの contentView を取得
- `createNSView(x, y, w, h)` — contentView の子 NSView を生成して返す
- `updateNSViewFrame(view, x, y, w, h)` — フレーム座標を更新
- `removeNSView(view)` — スーパービューから除去

ghostty が Metal で NSView へ直接描画するため、`TerminalWidget.Draw()` は何もしない。

---

## TerminalWidget のライフサイクル

| タイミング | 処理 |
|---|---|
| `main()` | `ghostty.Init` / `NewApp` をメインスレッドで実行 |
| `Build()` 初回 | `surfaceOnce.Do` → `initSurface()` で NSView + Surface 生成 |
| `Tick()` 毎フレーム | `wakeupCh` を受け取ったら `app.Tick()`、NSView フレーム更新、`surface.Draw()` |
| 終了時 | `defer app.Free()` / `defer root.terminal.Dispose()` |

---

## ビルドと起動

```sh
go build -o ./terminal ./cmd/terminal/ && ./terminal
```
