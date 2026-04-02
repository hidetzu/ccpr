# ccpr

[![CI](https://github.com/hidetzu/ccpr/actions/workflows/ci.yml/badge.svg)](https://github.com/hidetzu/ccpr/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/hidetzu/ccpr)](https://github.com/hidetzu/ccpr/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

CodeCommit の PR を、AI レビューできる形式に 1 コマンドで変換する CLI ツールです。

```bash
ccpr review <PR_URL> --format json | claude -p "このPRをレビューして"
```

[English](README.md)

## ccpr でできること

- CodeCommit の PR メタデータ・コメント・差分を **1 コマンドで取得**
- JSON / Patch 形式で出力 — Claude Code や他の AI ツールにそのまま渡せる
- PR の作成・一覧・コメント投稿まで CLI で完結

## なぜ ccpr が必要か

CodeCommit には GitHub の `gh` に相当するツールがありません。PR のメタデータ、コメント、差分を取得するには複数の AWS CLI コマンドを組み合わせる必要があり、それを AI に渡すにはさらに加工が必要です。

ccpr を使えば、1 コマンドで構造化された JSON を取得し、Claude Code に直接渡せます。

## クイックスタート

### 1. インストール

[Go](https://go.dev/dl/) (1.23 以上) が必要です。

```bash
go install github.com/hidetzu/ccpr/cmd/ccpr@latest
```

### 2. 初期設定

```bash
ccpr init --profile your-aws-profile --region ap-northeast-1
```

設定ファイルが `~/.config/ccpr/config.yaml` に作成されます。

### 3. リポジトリのマッピングを追加

`~/.config/ccpr/config.yaml` を編集して、CodeCommit リポジトリとローカルクローンのパスを対応付けます。

```yaml
profile: your-aws-profile
region: ap-northeast-1
repoMappings:
  my-repo: ~/src/my-repo
```

### 4. 動作確認

```bash
ccpr doctor    # 設定と環境をチェック
```

すべて通ったら準備完了です。

### 5. PR をレビュー

```bash
ccpr review <codecommit-pr-url>
```

## 使い方

### PR レビュー

```bash
ccpr review <PR_URL>                 # サマリー表示（デフォルト）
ccpr review <PR_URL> --format json   # JSON 出力（AI ツール向け）
ccpr review <PR_URL> --format patch  # 差分のみ
```

### PR 一覧

```bash
ccpr list --repo <repo>                  # OPEN な PR（デフォルト）
ccpr list --repo <repo> --status closed  # CLOSED な PR
ccpr list --repo <repo> --format json    # JSON 出力
```

### PR 作成

```bash
ccpr create --repo <repo> --title "機能追加" --dest main
ccpr create --repo <repo> --title "機能追加" --dest main --source feature/x
ccpr create --repo <repo> --title "機能追加" --dest main --description-file desc.md
```

### コメント投稿

```bash
ccpr comment <PR_URL> --body "LGTM"
ccpr comment <PR_URL> --body-file review.md
echo "問題なし" | ccpr comment <PR_URL> --body -
```

### その他

```bash
ccpr open <PR_URL>     # ブラウザで PR を開く
ccpr init              # 設定ファイルを生成
ccpr doctor            # 環境と設定を検証
```

## Claude Code との連携

### Claude Code からインストール

Claude Code に以下を貼り付けると、インストールからセットアップまで実行できます。

```
https://raw.githubusercontent.com/hidetzu/ccpr/main/docs/claude-integration.md を見てインストールして
```

### パイプで渡す

```bash
ccpr review <PR_URL> --format json | claude -p "このPRをレビューして"
```

### Claude Code スキル（推奨）

ccpr にはレビュー用の Claude Code スキルが付属しています。

```bash
mkdir -p .claude/skills/ccpr-review
cp /path/to/ccpr/examples/claude/ccpr-review/SKILL.md .claude/skills/ccpr-review/
```

Claude Code で以下を実行するだけでレビューできます。

```
/ccpr-review <codecommit-pr-url>
```

詳細は [docs/claude-integration.md](docs/claude-integration.md)（英語）を参照してください。

## よくあるトラブル

### `ccpr doctor` で AWS credentials が失敗する

AWS の認証情報が正しく設定されているか確認してください。

```bash
aws sts get-caller-identity --profile your-aws-profile
```

このコマンドが成功すれば、ccpr でも認証が通ります。SSO を使っている場合は先に `aws sso login --profile your-aws-profile` を実行してください。

### `no local path mapping for repository` と表示される

`~/.config/ccpr/config.yaml` の `repoMappings` に、CodeCommit リポジトリ名とローカルクローンのパスを追加してください。

```yaml
repoMappings:
  my-repo: ~/src/my-repo
```

### `region is required` と表示される

リージョンが設定されていません。以下のいずれかで指定してください。

1. `--region ap-northeast-1` フラグ
2. `~/.config/ccpr/config.yaml` の `region` フィールド
3. PR URL から自動検出（URL を使う場合）

### diff が空になる

ローカルリポジトリが最新でない可能性があります。`git fetch origin` を実行してから再試行してください。

## 設定

`~/.config/ccpr/config.yaml`

```yaml
profile: your-aws-profile
region: ap-northeast-1
repoMappings:
  my-repo: ~/src/my-repo
  another-repo: ~/src/another-repo
```

### AWS プロファイルの解決順序

1. `--profile` フラグ
2. 設定ファイルの `profile`
3. `AWS_PROFILE` 環境変数
4. デフォルト

### AWS リージョンの解決順序

1. PR URL（自動検出）
2. `--region` フラグ
3. 設定ファイルの `region`

## JSON 出力の安定性

ccpr は v1.x リリース内で JSON 出力の後方互換性を保証しています。詳細は以下を参照してください（英語）。

- [JSON Output Reference](docs/json-schema.md) — 各コマンドのフィールド定義
- [Versioning Policy](docs/versioning.md) — SemVer ルールと互換性ポリシー

## ライセンス

MIT
