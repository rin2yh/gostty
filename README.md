# Gostty

Go言語で動くGUIフレームワーク guigui-gui/guiguiとghostty-org/ghosttyを繋ぎ込んでguiguiでターミナルを開発します。

## 技術スタック
- guigui：GUIフレームワーク
    - 内部でEbitengineを使用
- libghostty：ghosttyを外部から利用するための仕組み

## ローカルビルドと実行（macOS Apple Silicon）

### 前提

- macOS (Apple Silicon)
- [mise](https://mise.jdx.dev/) — `mise.toml` 経由で `zig 0.15.2` を解決
- Go（`go.mod` の `go 1.25.0` 以上）
- Xcode Command Line Tools と Metal Toolchain

### 手順

1. submodule を取得

   ```sh
   git submodule update --init --recursive
   ```

2. zig をセットアップ（プロジェクトルートで）

   ```sh
   mise install
   ```

3. Xcode の初回起動セットアップ（必要な場合のみ）

   ```sh
   sudo xcodebuild -runFirstLaunch
   ```

4. ghostty 静的ライブラリをビルド

   ```sh
   cd ghostty
   zig build -Dapp-runtime=none -Doptimize=ReleaseSafe
   cd ..
   ```

   成果物: `ghostty/macos/GhosttyKit.xcframework/macos-arm64_x86_64/ghostty-internal.a`

5. gostty バイナリをビルドして実行

   ```sh
   go build -o gostty ./cmd/terminal
   ./gostty
   ```
