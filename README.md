# HTTPログサーバー

HTTPリクエストの詳細をログファイルに記録するGoアプリケーションです。

## 機能

- 全てのHTTPメソッド（GET、POST、PUT、DELETE等）に対応
- リクエストヘッダー、クエリパラメータ、ボディデータを記録
- タイムスタンプ付きのディレクトリ構造でログを整理
- CORS対応により任意のオリジンからのアクセスが可能
- JSONフォーマットでの構造化ログ出力

## インストールと実行

### 前提条件
- Go 1.21以上

### 実行方法

#### 開発・テスト用
```bash
# リポジトリをクローン
git clone <repository-url>
cd httplog

# 依存関係を解決
make install

# アプリケーションを実行
make run

# テスト実行
make test
```

#### ビルド・デプロイ用
```bash
# 現在のプラットフォーム用にビルド
make build

# 実行（build/フォルダからバイナリを実行）
./build/httplog

# 全プラットフォーム用にビルド
make build-all

# 配布用パッケージ作成
make package
```

### コマンドライン引数

```bash
# 基本的な使用方法
./httplog [オプション]

# オプション
-port string        サーバーのポート番号 (デフォルト: "8080")
-logdir string      ログファイルの保存ディレクトリ (デフォルト: "logs")
-help              ヘルプを表示
-version           バージョン情報を表示

# 使用例
./httplog -port 9000 -logdir /var/log/httplog
```

### 環境変数
- `PORT`: サーバーのポート番号（引数が指定されていない場合のみ有効、デフォルト: 8080）

## 使用方法

サーバーが起動したら、任意のHTTPクライアントでリクエストを送信できます。

### 例

```bash
# デフォルトポート（8080）での例
curl "http://localhost:8080/api/test?param1=value1&param2=value2"

# カスタムポート（9000）での例
curl "http://localhost:9000/api/test?param1=value1&param2=value2"

# POSTリクエスト（フォームデータ）
curl -X POST -d "key1=value1&key2=value2" http://localhost:8080/api/test

# POSTリクエスト（JSON）
curl -X POST -H "Content-Type: application/json" -d '{"test":"data","number":123}' http://localhost:8080/api/test
```

## ログファイル

ログは以下の構造で保存されます：

```
logs/                           # デフォルトのログディレクトリ
└── 20240126-143052-001/        # 日時-シーケンス番号
    ├── log.txt                 # リクエスト情報（JSON形式）
    └── data.txt               # リクエストボディ（存在する場合）

# カスタムログディレクトリを指定した場合
/var/log/httplog/              # 指定したログディレクトリ
└── 20240126-143052-001/
    ├── log.txt
    └── data.txt
```

### log.txtの例

```json
{
  "method": "POST",
  "url": "/api/test?param=value",
  "header": {
    "Content-Type": "application/json",
    "User-Agent": "curl/7.68.0"
  },
  "get": {
    "param": "value"
  },
  "post": {
    "key": "value"
  },
  "data": "{\"test\":\"data\"}"
}
```

## ビルドとデプロイ

このプロジェクトでは、Makefileを使用してマルチプラットフォーム対応のビルドを簡単に行えます。

### Makeコマンド一覧

```bash
# ヘルプの表示
make help

# 開発用コマンド
make install      # 依存関係のインストール
make run          # アプリケーション実行
make test         # テスト実行
make fmt          # コードフォーマット
make vet          # 静的解析

# ビルドコマンド
make build        # 現在のプラットフォーム用ビルド
make build-all    # 全プラットフォーム用ビルド
make build-dev    # 開発用ビルド（デバッグ情報付き）

# 特定プラットフォーム用ビルド
make build-linux
make build-darwin
make build-windows

# パッケージング・リリース
make package      # 配布用パッケージ作成
make release      # 全チェック＋ビルド＋パッケージング
make clean        # ビルド成果物削除
```

### 対応プラットフォーム

- **Linux**: amd64, arm64
- **macOS (Darwin)**: amd64, arm64 (Intel & Apple Silicon)
- **Windows**: amd64, arm64

### ビルド成果物

```
project/
├── build/           # ビルドされたバイナリ
│   ├── httplog                    # 現在のプラットフォーム用
│   ├── httplog-linux-amd64/       # Linux用
│   ├── httplog-darwin-amd64/      # macOS Intel用
│   ├── httplog-darwin-arm64/      # macOS Apple Silicon用
│   └── httplog-windows-amd64/     # Windows用
└── dist/            # 配布用パッケージ
    ├── httplog-v1.0.0-linux-amd64.tar.gz
    ├── httplog-v1.0.0-darwin-amd64.tar.gz
    ├── httplog-v1.0.0-darwin-arm64.tar.gz
    └── httplog-v1.0.0-windows-amd64.zip
```

## テスト

```bash
# テストを実行
make test

# 手動でテスト実行
go test -v -race -cover ./...
```

## 開発

```bash
# 全体チェック（フォーマット + 静的解析 + テスト）
make ci

# 個別実行
make fmt         # コードフォーマット
make vet         # 静的解析
make install     # 依存関係の更新
```
