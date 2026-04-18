# 描画が 4 回に 1 回しか成功しない問題の修正ログ

## 症状

`./terminal` を起動すると 4 回に 1 回程度しかターミナルが描画されず、残りは真っ黒画面。描画成功時はキー入力も動作するため、PTY 周辺ではなく NSView / Metal レイヤー初期化のタイミング依存が疑われた。

## 根本原因

### 1. layer-hosting の設定順序違反

`ghostty/src/renderer/Metal.zig` は NSView を layer-hosting view にするため、次の順序で設定する仕様。コメントにも `assigning it to the view's 'layer' property BEFORE setting 'wantsLayer' to 'true'` と明記されている。

```zig
info.view.setProperty("layer", layer.layer.value);   // IOSurfaceLayer を割り当て
info.view.setProperty("wantsLayer", true);           // 後から wantsLayer=YES
```

ところが `cmd/terminal/nsview_darwin.go` の `createNSView` は `ghostty_surface_new` より前に `setWantsLayer:YES` を呼んでいた。この時点で AppKit は「layer-backed view」として内部バッキングレイヤーを生成する。後から ghostty が `view.layer` を `IOSurfaceLayer` に差し替えても、内部状態が残っているためフレームによって反映されたりされなかったりする非決定挙動になる。

### 2. AppKit 呼び出しが非メインスレッド

`Build()` は ebitengine のゲームループ goroutine 上で走る。そこから `createNSView` / `ghostty_surface_new` / `updateNSViewFrame` を直接呼ぶのは AppKit の規約違反で、起動タイミング依存の描画失敗を引き起こす。前コミット `ad6eab3` の `ghostty_app_new` と同じ種類の問題。

## 対処

### `cmd/terminal/nsview_darwin.go`

- `createNSView` から `setWantsLayer:YES` の呼び出しと `sel_setWantsLayer` セレクタ定義を削除。wantsLayer は `ghostty_surface_new` 内部で正しい順序で設定される。

### `cmd/terminal/terminal_widget.go`

- `initSurface` / `Tick` の bounds 変化時 / `Dispose` 内の AppKit 呼び出しを `ebiten.RunOnMainThread` で包む。
- `surface.Draw()` はラップしない。ghostty の `IOSurfaceLayer.setSurface`（`ghostty/src/renderer/metal/IOSurfaceLayer.zig:49-77`）が内部で必要なときだけ main queue に `dispatch_async` するため、毎フレーム同期ディスパッチすると UI スレッドが詰まる。
- `Dispose` では `surface.Free` と `removeNSView` を同じ main-thread クロージャで順序保証する。`IOSurfaceLayer` 側の非同期 dispatch が view 解放後に走るのを防ぐため。

### デバッグログ

`GOSTTY_DEBUG=1 ./terminal` で stderr に `[gostty] ...` のログが出る。切り分け箇所:

- contentView サイズ / スケール
- NSView 生成結果（id / 失敗）
- `ghostty.NewSurface` の成否
- bounds 変化のサイズ

## 参照

- `ghostty/src/renderer/Metal.zig:114-126` — layer → wantsLayer の順序仕様
- `ghostty/src/renderer/metal/IOSurfaceLayer.zig:49-77` — `setSurface` が main thread を要求する根拠
- ebiten v2 `run.go:772` — `ebiten.RunOnMainThread(func())`
