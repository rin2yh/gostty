# Gostty

Go言語で動くGUIフレームワーク guigui-gui/guiguiとghostty-org/ghosttyを繋ぎ込んでguiguiでターミナルを開発します。

## 技術スタック
- guigui：GUIフレームワーク
    - 内部でEbitengineを使用
- libghostty：ghosttyを外部から利用するための仕組み

## インストール（macOS Apple Silicon）

1. [Releases](https://github.com/rin2yh/gostty/releases) から最新の `gostty-darwin-arm64.tar.gz` をダウンロード
2. 展開して任意の場所に置く:
   ```sh
   tar -xzf gostty-darwin-arm64.tar.gz
   ```
3. 未署名バイナリなので Gatekeeper の quarantine 属性を外す:
   ```sh
   xattr -d com.apple.quarantine ./gostty
   ```
4. 実行:
   ```sh
   ./gostty
   ```
