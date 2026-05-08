# ccpr

[![CI](https://github.com/hidetzu/ccpr/actions/workflows/ci.yml/badge.svg)](https://github.com/hidetzu/ccpr/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/hidetzu/ccpr)](https://github.com/hidetzu/ccpr/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

## CodeCommit の PR を "そのまま AI に渡せる形" にする CLI

**1コマンドで PR → AIレビューまで完結。**

[English](README.md)

---

## こんな課題ありませんか？

- CodeCommit の PR を AI でレビューしたい
- diff / コメント / メタ情報を毎回集めるのが面倒
- AWS CLI がバラバラでつらい
- コピペ地獄から抜けたい

ccpr で全部解決できます。

---

## 一番シンプルな使い方

```bash
ccpr review <PR_URL> --format json | claude -p "このPRをレビューして"
```

---

## Before / After

### ccpr なし（つらい）

```bash
aws codecommit get-pull-request --pull-request-id 123
aws codecommit get-comments-for-pull-request --pull-request-id 123
aws codecommit get-differences --repository-name my-repo \
  --before-commit-specifier main --after-commit-specifier feature/x

# + 手動で整形してAIに貼り付け...
```

### ccpr あり（1コマンド）

```bash
ccpr review <PR_URL> --format json | claude -p "このPRをレビューして"
```

PR情報・コメント・diffをまとめて取得 → そのままAIに渡せる。

---

## クイックスタート（30秒）

いずれかの方法でインストールします:

```bash
# Homebrew（macOS / Linux）— ccpr と ccpr-mcp が同時に入る
brew install --cask hidetzu/tap/ccpr

# Arch Linux（AUR）— ccpr と ccpr-mcp が同時に入る
yay -S ccpr-bin       # または: paru -S ccpr-bin

# プリビルドバイナリ
# https://github.com/hidetzu/ccpr/releases/latest からダウンロード

# Go ツールチェイン（Go 1.25 以上が必要）
go install github.com/hidetzu/ccpr/cmd/ccpr@latest
```

その後:

```bash
ccpr init
ccpr review <codecommit-pr-url>
```

---

## 主なユースケース

- AIコードレビュー（Claude / GPT）
- PRの要約生成
- CIでの自動レビュー
- CLIだけでPR操作（list / review / create）
- コメント自動投稿

---

## Claude Code との連携

### パイプで使う（最短）

```bash
ccpr review <PR_URL> --format json | claude -p "このPRをレビューして"
```

### Claude Code からインストール

Claude Code に以下を貼り付けると、インストールからセットアップまで実行できます。

```
https://raw.githubusercontent.com/hidetzu/ccpr/main/docs/claude-integration.md を見てインストールして
```

### Claude Code スキル（推奨）

```bash
mkdir -p .claude/skills/ccpr-review
cp /path/to/ccpr/examples/claude/ccpr-review/SKILL.md .claude/skills/ccpr-review/
```

```
/ccpr-review <codecommit-pr-url>
```

詳細は [docs/claude-integration.md](docs/claude-integration.md)（英語）を参照してください。

### MCPサーバー

ccpr は Claude Code などのMCPクライアント向けに、実験的なMCPサーバーバイナリも提供します。現在公開しているtool:

```text
ccpr_list      リポジトリのPR一覧を取得          (読み取り)
ccpr_review    指定PRのメタデータ・コメント・diffをまとめて取得  (読み取り)
ccpr_comment   PRにコメントを投稿                (書き込み)
ccpr_create    PRを作成                          (書き込み)
```

Homebrew またはプリビルドリリースで ccpr を入れた場合、`ccpr-mcp` は既に `PATH` 上にあります。それ以外はローカルでビルドします。

```bash
make build-mcp
```

Claude Codeに登録します。

```bash
# Homebrew / プリビルド経由
claude mcp add ccpr -- ccpr-mcp

# ローカルビルド
claude mcp add ccpr -- /path/to/ccpr/bin/ccpr-mcp
```

その後、Claude CodeからCodeCommitのPR一覧取得やレビュー対象PRの取得を依頼できます。

- `ccpr_list` は `repo`, `status`, `region`, `profile`, `config` を受け取り、`ccpr list --format json` と同じPRサマリースキーマを `pullRequests` キー配下で返します。
- `ccpr_review` は `url` または `repo` + `prId` のいずれかと、任意の `region`, `profile`, `config` を受け取り、`ccpr review --format json` と同じ `{metadata, comments, diff}` を返します。
- `ccpr_comment` は `body` と、`url` または `repo` + `prId`、任意の `region`, `profile`, `config` を受け取り、`ccpr comment --format json` と同じ `{commentId, pullRequestId, authorArn, creationDate}` を返します。**書き込み系**のため、呼び出すたびにCodeCommitへ実際にコメントが投稿されます。MCPホスト側の承認プロンプトで明示的に許可してください。
- `ccpr_create` は `repo`, `title`, `sourceBranch`, `destinationBranch` と、任意の `description`, `region`, `profile`, `config` を受け取り、`ccpr create --format json` と同じ `{prId, title, repository, sourceBranch, destinationBranch, url}` を返します。**書き込み系**のため、呼び出すたびにCodeCommitへ実際にPRが作成されます。MCPホスト側の承認プロンプトで明示的に許可してください。CLIと違い `sourceBranch` は必須です（MCPにはローカルGitの「カレントブランチ」がないため）。

---

## できること

- PR取得（summary / json / patch）
- PR一覧
- PR作成
- コメント投稿
- ブラウザでPRを開く
- CLIだけでCodeCommit操作完結

---

## 詳細リファレンス

<details>
<summary>コマンド一覧</summary>

```bash
# レビュー
ccpr review <PR_URL>                 # サマリー（デフォルト）
ccpr review <PR_URL> --format json   # JSON（AI ツール向け）
ccpr review <PR_URL> --format patch  # 差分のみ

# 一覧
ccpr list --repo <repo>                         # OPEN な PR（デフォルト）
ccpr list --repo <repo> --status closed         # CLOSED な PR
ccpr list --repo <repo> --format json           # JSON 出力

# 作成
ccpr create --repo <repo> --title "機能追加" --dest main
ccpr create --repo <repo> --title "機能追加" --dest main --source feature/x

# コメント
ccpr comment <PR_URL> --body "LGTM"
ccpr comment <PR_URL> --body-file review.md

# その他
ccpr open <PR_URL>          # ブラウザで PR を開く
ccpr init                   # 設定ファイル生成
ccpr doctor                 # 環境と設定を検証
```

</details>

<details>
<summary>出力例</summary>

#### サマリー（デフォルト）

```
$ ccpr review <codecommit-pr-url>
PR #883: Fix login bug
Author:   example-user
Status:   OPEN
Branch:   feature/login → main
Created:  2026-03-25

Comments: 3 (2 threads)
Files:    7 changed

## Description

Fix session timeout on login page
```

#### JSON（AI ツール向け）

```json
{
  "metadata": {
    "prId": "883",
    "title": "Fix login bug",
    "author": "example-user",
    "sourceBranch": "feature/login",
    "destinationBranch": "main",
    "status": "OPEN",
    "creationDate": "2026-03-25T10:30:00Z"
  },
  "comments": [...],
  "diff": "diff --git a/src/login.go ..."
}
```

</details>

<details>
<summary>JSON 出力の安定性</summary>

ccpr は v1.x リリース内で JSON 出力の後方互換性を保証しています。

- [JSON Output Reference](docs/json-schema.md) — 各コマンドのフィールド定義（英語）
- [Versioning Policy](docs/versioning.md) — SemVer ルールと互換性ポリシー（英語）

</details>

---

## 設定

`~/.config/ccpr/config.yaml`

```yaml
profile: your-aws-profile
region: ap-northeast-1
repoMappings:
  your-repo: ~/src/your-repo
```

**AWS プロファイル解決順序:** `--profile` フラグ > 設定ファイル > `AWS_PROFILE` 環境変数 > デフォルト

**AWS リージョン解決順序:** PR URL（自動検出）> `--region` フラグ > 設定ファイル > `AWS_REGION` 環境変数 > `AWS_DEFAULT_REGION` 環境変数

---

## トラブルシューティング

- `ccpr doctor` をまず実行
- AWS認証確認: `aws sts get-caller-identity --profile your-aws-profile`
- SSOの場合: `aws sso login --profile your-aws-profile`
- `no local path mapping` → `config.yaml` の `repoMappings` を追加
- `region is required` → `--region` フラグ、設定ファイル、または `AWS_REGION` / `AWS_DEFAULT_REGION` 環境変数で指定
- diff が空 → `git fetch origin` を実行してから再試行

---

## 開発

```bash
make build    # バイナリを bin/ccpr にビルド
make build-mcp # MCPサーバーバイナリを bin/ccpr-mcp にビルド
make test     # 全テスト実行（-v -race）
make lint     # golangci-lint
make vet      # go vet
make clean    # bin/ を削除
```

## ライセンス

MIT
