# ビルドステージ
FROM golang:1.25-bookworm AS builder

WORKDIR /build

# 依存関係のダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピーとビルド
COPY . .

ARG VERSION
ARG COMMIT_ID

RUN CGO_ENABLED=1 GOOS=linux \
    BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
    GO_BUILD=$(go env GOVERSION) \
    go build -ldflags "\
    -X at-bot/internal/meta.version=${VERSION} \
    -X at-bot/internal/meta.commitID=${COMMIT_ID} \
    -X at-bot/internal/meta.buildTime=${BUILD_TIME} \
    -X at-bot/internal/meta.goBuild=${GO_BUILD} \
    " \
    -o bot ./cmd/at-bot

# 実行ステージ (最小構成)
FROM debian:bookworm-slim

# 必要な最小限のパッケージをインストール
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# ビルド済みバイナリをコピー
COPY --from=builder /build/bot .

# データディレクトリを作成
RUN mkdir -p /app/data

# バイナリを実行
CMD ["./bot"]
