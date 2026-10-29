# HTTP2SMTP

**アプリケーションごとに SMTP 認証情報を埋め込まず、HTTP でメールを送信。**

[![Release](https://img.shields.io/github/v/release/chuccp/http2smtp)](https://github.com/chuccp/http2smtp/releases)
[![Actions](https://img.shields.io/github/actions/workflow/status/chuccp/http2smtp/release.yml)](https://github.com/chuccp/http2smtp/actions/workflows/release.yml)
[![Docker Pulls](https://img.shields.io/docker/pulls/cooge123/http2smtp)](https://hub.docker.com/r/cooge123/http2smtp)
[![Go](https://img.shields.io/github/go-mod/go-version/chuccp/http2smtp)](./go.mod)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](./LICENSE)
[![Website](https://img.shields.io/badge/website-http2smtp-blue)](https://chuccp.github.io/http2smtp/)

[English](./README.md) | [简体中文](./README_zh.md) | [繁體中文](./README.zh-TW.md) | **日本語**

HTTP2SMTP はセルフホスト可能な HTTP-to-SMTP ゲートウェイです。管理画面で SMTP
サーバー、宛先、API トークン、スケジュールタスクを管理し、アプリケーションからは
シンプルな HTTP リクエストだけでメールを送信できます。

> [!IMPORTANT]
> このバージョンは **v0.2.x とデータ互換性がありません**。旧バージョンまたはデータ
> バージョンの不一致を検出すると、管理画面だけを起動する互換性モードに入ります。
> 旧データベースは開かず、変更もしません。画面にバックアップとリセット手順を表示します。
> アップグレード前に[「v0.2.x からのアップグレード」](#v02x-からのアップグレード)を
> 確認してください。

## 主な機能

- JSON、フォーム、URL パラメーターによるメール送信
- 複数の SMTP サーバー、宛先、API トークン
- 一般ユーザーごとのデータ分離と管理者による全体管理
- Cron スケジュール、レスポンステンプレート、失敗時のみ送信
- HTML メールテンプレートと複数添付ファイル
- メール履歴、添付ファイルのダウンロード、操作監査ログ
- 回転しても判別しやすい図形を使ったクリック式 CAPTCHA
- サーバーサイドテンプレートと静的ファイルを内蔵した単一バイナリ
- 简体中文、繁體中文、English、日本語の UI
- `/data` だけで永続化できる Docker イメージ
- GitHub による埋め込みリソース検証、ビルド、バージョン付きリリース

## スクリーンショット

| ダッシュボード | SMTP サーバー追加 |
|---|---|
| <img src="./docs/screenshots/dashboard.webp" alt="ダッシュボード" width="100%"> | <img src="./docs/screenshots/smtp.webp" alt="SMTP サーバー追加" width="100%"> |

| API トークン作成 | スケジュールタスク追加 |
|---|---|
| <img src="./docs/screenshots/token.webp" alt="API トークン作成" width="100%"> | <img src="./docs/screenshots/schedule.webp" alt="スケジュールタスク追加" width="100%"> |

## クイックスタート

### Docker Compose

`docker/docker-compose.yml` は `docker/data` を `/data` にマウントします。設定、
SQLite、ログ、添付ファイルはコンテナを置き換えても保持されます。

```bash
cd docker
docker compose up -d
```

<http://127.0.0.1:12566> を開き、セットアップウィザードを完了してください。

### Docker Run

Windows PowerShell：

```powershell
docker pull cooge123/http2smtp

docker run -d --name http2smtp -p 12566:12566 -p 12567:12567 -v "${PWD}\data:/data" cooge123/http2smtp
```

Linux または macOS：

```bash
docker run -d --name http2smtp -p 12566:12566 -p 12567:12567 -v "$PWD/data:/data" cooge123/http2smtp
```

コンテナとポートを確認：

```bash
docker inspect --format '{{.State.Health.Status}}' http2smtp
docker exec http2smtp nc -z -w 2 127.0.0.1 12566
docker exec http2smtp nc -z -w 2 127.0.0.1 12567
```

内蔵ヘルスチェックは常に管理ポートを確認します。互換性モードでは公開 API を意図的に
停止するため、API ポートの確認はバックアップとリセットが完了するまで失敗します。

### バイナリ

[GitHub Releases](https://github.com/chuccp/http2smtp/releases/latest) から対象プラット
フォームのアーカイブをダウンロードします。

Windows PowerShell：

```powershell
Invoke-WebRequest `
  -Uri "https://github.com/chuccp/http2smtp/releases/latest/download/http2smtp-windows-amd64.tar.gz" `
  -OutFile "http2smtp-windows-amd64.tar.gz"
tar -xzf http2smtp-windows-amd64.tar.gz
.\http2smtp.exe
```

Linux：

```bash
wget https://github.com/chuccp/http2smtp/releases/latest/download/http2smtp-linux-amd64.tar.gz
tar -xzf http2smtp-linux-amd64.tar.gz
chmod +x http2smtp
./http2smtp
```

リリースアーカイブにはバイナリだけが含まれます。テンプレート、CSS、JavaScript、
アイコン、言語ファイルはすべてバイナリに埋め込まれています。

### バージョン

```bash
./http2smtp -v
./http2smtp --version
```

出力例：

```text
http2smtp version v0.3.0 (linux/amd64)
```

GitHub のリリース処理はビルド前に Git タグを `version.go` に書き込むため、配布
バイナリには正確なリリースバージョンが含まれます。通常のローカルビルドでは `dev` と
表示されます。

## 初回セットアップ

1. <http://127.0.0.1:12566> の管理画面を開きます。
2. SQLite または MySQL を選択し、接続をテストします。
3. 管理者アカウントを作成します。
4. SMTP サーバー、宛先、API トークンを追加します。
5. `12567` ポートの公開 API からメールを送信します。

管理画面と管理 API は `12566` を共有します。公開メール API は `12567` を別途使用する
ため、管理画面を外部公開せずにメール送信 API だけを公開できます。

## 永続データ

Docker コンテナはすべての可変データを `/data` に保存します。

```text
/data/
  config.ini
  d-mail.db                 # 既定の SQLite データベース
  logs/
    http2smtp.log           # 内蔵ロガーによるローテーション
  YYYY-MM-DD/               # 添付ファイルキャッシュ
```

MySQL を使用する場合、データベースは外部に残り、設定、ログ、添付ファイルは
`/data` に保存されます。

Docker 以外では `-storage_root` または `storage_root` 環境変数で添付ファイルの
ルートを指定できます。カスタム保存先では、アプリケーションログも
`<storage_root>/logs/http2smtp.log` に書き込まれます。

## v0.2.x からのアップグレード

新しいデータ配置への変更が大きいため、インプレースでのスムーズなアップグレードは
できません。HTTP2SMTP はデータベース接続と自動テーブル作成の前に
`core.dataVersion` を確認します。

旧設定、バージョン情報のないデータベース、バージョン不一致を検出した場合：

- 既存の管理ポートで互換性案内だけを表示します
- データベース、モデル、スケジュールタスク、公開 API は起動しません
- 旧データベースを読み取りまたは変更しません
- バックアップ対象、バックアップ手順、リセット方法を画面に表示します

リセット前に：

1. 旧バージョンを停止します。
2. `config.ini`、SQLite データベース、添付ディレクトリ、アプリケーションログを
   バックアップします。MySQL の場合はデータベース全体をダンプします。
3. `config.ini` をアプリケーションディレクトリの外へ移動します。
4. SQLite データベースをバックアップ先へ移動するか、空の MySQL データベースを作成します。
5. 新しいバージョンを起動し、セットアップウィザードを完了します。

HTTP2SMTP は旧設定、データベース、添付ファイルを自動削除しません。Docker では
`/data` にマウントしたホストディレクトリ全体をバックアップしてください。

## 公開 API

### メール送信

`POST /sendMail`、既定では公開 API ポート `12567` を使用します。

| フィールド | 型 | 必須 | 説明 |
|---|---|---:|---|
| `token` | string | はい | 管理画面で作成した API トークン |
| `subject` | string | いいえ | 省略時はトークンの既定件名 |
| `content` | string | はい | プレーンテキストまたは HTML 本文 |
| `recipients` | string[] | いいえ | トークン設定の宛先に追加 |
| `files` | object[] | いいえ | `name` と Base64 の `data` を持つ添付ファイル |

トークンに設定した宛先とリクエストの宛先は統合され、重複アドレスには一度だけ送信
されます。不正なアドレスはスキップされ、有効な宛先が 1 件もない場合だけ失敗します。

JSON の例：

```bash
curl -X POST 'http://127.0.0.1:12567/sendMail' \
  -H 'Content-Type: application/json' \
  -d '{
    "token": "YOUR_TOKEN",
    "subject": "デプロイ完了",
    "content": "<b>ビルド 42 が完了しました。</b>",
    "recipients": ["ops@example.com"],
    "files": [{
      "name": "result.txt",
      "data": "BASE64_CONTENT"
    }]
  }'
```

フォームまたは multipart の例：

```bash
curl -X POST 'http://127.0.0.1:12567/sendMail' \
  --form 'token=YOUR_TOKEN' \
  --form 'subject=日次レポート' \
  --form 'content=レポートを添付します' \
  --form 'recipients=ops@example.com' \
  --form 'files=@./report.pdf'
```

GET の例：

```bash
curl 'http://127.0.0.1:12567/sendMail?token=YOUR_TOKEN&subject=test&content=hello'
```

リクエスト本文の上限は `core.maxbodysize` で制御します。既定値は `10M` で、`K/M/G`
接尾辞を利用でき、`-1` は無制限です。JSON、フォーム、multipart のすべてに適用され、
エンコード後のサイズで判定されるため、Base64 添付には追加の容量が必要です。

## 設定

設定の優先順位：

`既定値` → `config.ini` → コマンドラインフラグ → 環境変数

主な起動フラグ：

```bash
./http2smtp \
  -web_port 12566 \
  -api_port 12567 \
  -storage_root /path/to/storage
```

フラグまたは `web_port` / `api_port` 環境変数で指定したポートは固定扱いになります。
Docker はこの仕組みで、設定画面からポートマッピングを変更できないようにしています。

一般的な `config.ini`：

```ini
[core]
init = true
dbType = sqlite
dataVersion = 1
maxBodySize = 10M

[sqlite]
filename = d-mail.db

[manage]
port = 12566

[api]
port = 12567
```

MySQL を使用する場合は、`[mysql]` セクションに `host`、`port`、`dbname`、
`username`、`password`、`charset` を設定します。

`core.dataVersion` は手動で変更しないでください。ディスク上のデータ形式を示し、
互換性チェックで使用されます。

## ソースからビルド

Go のサーバーサイドテンプレートを使用しており、Node/pnpm のビルド手順は不要です。

```bash
go test ./...
go build -tags release -o http2smtp ./
./http2smtp -v
```

通常の `go build` は開発ビルドで、作業ディレクトリからテンプレートと静的ファイルを
読み込みます。リリースビルドは `-tags release` を使用し、埋め込みコピーを参照するため
任意のディレクトリから実行できます。

`webui/embed.go` が UI 全体をバイナリに含めます。

```go
//go:embed templates
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS
```

4 言語の JSON も埋め込まれます。GitHub リリース処理は：

1. `go test -tags release ./webui` でテンプレートと静的ファイルを検証
2. プッシュされた Git タグを `version.go` に書き込み
3. 全プラットフォームを `-tags release` でビルド
4. Actions artifact を集約して 1 つの GitHub Release を作成
5. 同じ Release タグから Docker イメージをビルド

最新ビルドは [GitHub Releases](https://github.com/chuccp/http2smtp/releases) と
[Docker Hub](https://hub.docker.com/r/cooge123/http2smtp) から取得できます。

## コミュニティ

Telegram：<https://t.me/+JClG9_DojaM0ZGE1>

問題報告や機能提案は [GitHub Issues](https://github.com/chuccp/http2smtp/issues) へ
お寄せください。

## ライセンス

Apache License 2.0。詳細は [LICENSE](./LICENSE) を参照してください。
