# terminal native keycode fix

## 症状 / 背景

`./terminal` を起動するとターミナル画面は表示され、英数字の表示も行われる一方で
Enter キーを押しても改行されず、コマンドが実行できない。Escape や矢印キー、
Ctrl+C なども同様に無反応。マウス操作と printable 文字の表示だけが動く状態だった。

再現: `go build -o terminal ./cmd/terminal && ./terminal` → 何か文字を入力した後 Enter。
shell プロンプトが次行に進まない。

## 根本原因

`cmd/terminal/terminal_widget.go` のキー送出処理が、`ghostty_input_key_s.keycode`
に渡す値を **`ghostty.Key` 列挙値** (例: `ghostty.KeyEnter` ≒ 60) にしていた。

Ghostty 側の embedded apprt は `keycode` を **プラットフォーム native 仮想キーコード**
として解釈する実装になっている (`ghostty/src/apprt/embedded.zig:94-117`):

```zig
const physical_key = keycode: for (input.keycodes.entries) |entry| {
    if (entry.native == self.keycode) break :keycode entry.key;
} else .unidentified;
```

native キーコードテーブル (`ghostty/src/input/keycodes.zig:182-` の 5 列目 = Mac)
では Enter は `0x24`、Escape は `0x35`、Tab は `0x30` など。
enum 値を投げていたので全キーが `.unidentified` に落ち、`encodeKey` で何も
書き込まれず PTY にバイトが届かなかった。

printable 文字だけ「表示されている」ように見えたのは、`ebiten.AppendInputChars` →
`surface.SendText` (`ghostty_surface_text`) が並走しており、こちらは paste 扱いで
キー解釈を経由しないため。Enter や制御キーは `AppendInputChars` に乗らないので
完全に死んでいた。

確度: 確定 (ghostty ソースと keycodes.zig テーブルの照合で一致)。

## 対処

### `cmd/terminal/terminal_widget.go`

- `ebitenToGhosttyKey` (`map[ebiten.Key]ghostty.Key`) を削除。
- 代わりに `ebitenToMacKeycode` (`map[ebiten.Key]uint32`) を追加。値は
  `ghostty/src/input/keycodes.zig` の `raw_entries` Mac 列から抜粋。
  例: `KeyEnter: 0x24`, `KeyEscape: 0x35`, `KeyTab: 0x30`, `KeyBackspace: 0x33`,
  `KeyArrowUp: 0x7e`, `KeyF1: 0x7a`...
- `HandleButtonInput` の送出箇所:

```go
if kc, ok := ebitenToMacKeycode[key]; ok {
    t.surface.Key(ghostty.KeyEvent{
        Action:  b.action,
        Mods:    mods,
        Keycode: kc,
    })
}
```

### 意図的に変えなかった箇所

- `ebiten.AppendInputChars` → `surface.SendText` のパスは残した。printable 文字の
  表示は現状 paste 経由で既に動いているため、ここを `KeyEvent.Text` に寄せる
  改善は本修正のスコープ外とした (リスクを広げないため)。
- `pkg/ghostty/types.go` の `ghostty.Key` enum は公開 API として残置。今回の
  widget 実装では使わなくなったが、将来の他クライアント向けに残す。
- macOS 以外のプラットフォーム対応も対象外。`terminal_widget.go` は `//go:build darwin`
  なので問題にならない。
- `Pause` / `ScrollLock` / `PrintScreen` は macOS に対応する native キーコードが
  無い (`keycodes.zig` で `0xffff`) ため map から除外。

## デバッグ / 検証方法

1. ビルドと CI:
   ```bash
   go fmt ./...
   go vet ./...
   go test ./...
   go build -o terminal ./cmd/terminal
   ```
2. 手動確認 (起動後ターミナル領域をクリックしてフォーカス):
   - `ls` + Enter → ディレクトリ一覧が表示される
   - `echo hello` + Enter → `hello` 出力
   - Ctrl+C → 現在行キャンセル
   - 矢印 ↑/↓ → shell 履歴移動
   - `vi` 内で Escape → normal mode
3. 再発時の切り分け: `GOSTTY_DEBUG=1 ./terminal` で `debugf` ログを確認し、
   `surface.Key` に渡っている `Keycode` が native 値 (例: Enter=0x24) か
   enum 値 (60 付近) かを判定する。

## 参照

- `ghostty/src/apprt/embedded.zig:82-117` KeyEvent.core()
- `ghostty/src/input/keycodes.zig:182-` raw_entries Mac 列
- `ghostty/src/Surface.zig:2625-2769` keyCallback / encodeKey
- 関連ログ: [terminal-render-reliability.md](terminal-render-reliability.md)
