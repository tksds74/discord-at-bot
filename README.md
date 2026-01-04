# Discord AT Bot

## 機能

- `/at`: 募集を開始
- `/dice`: 6面ダイスの結果を返却

## セットアップ

### 前提条件

- Go 1.21以上（ローカル実行の場合）
- Docker & Docker Compose（Docker実行の場合）
- Discord Bot Token

### Discord Botの作成

1. [Discord Developer Portal](https://discord.com/developers/applications)でアプリケーションを作成
2. Botタブから「Add Bot」をクリック
3. Bot Tokenを取得
4. OAuth2タブで以下のスコープを選択
   - `bot`
   - `applications.commands`
5. Bot Permissionsで必要な権限を付与
   - Send Messages
   - Use Slash Commands
   - Mention Everyone

### 環境変数の設定

```bash
cp .env.example .env
```

`.env`ファイルを編集し、Discord Bot Tokenを設定

```
DISCORD_BOT_TOKEN=your_discord_bot_token_here
```

### 起動方法

#### Dockerで起動

```bash
docker compose up -d
```

#### ローカルで起動

```bash
go run ./cmd/at-bot-main
```
