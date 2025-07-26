# Makefile for httplog project

# 変数定義
APP_NAME := httplog
VERSION := v1.0.0
BUILD_DIR := build
DIST_DIR := dist

# Goビルド設定
GOOS_LIST := linux darwin windows
GOARCH_LIST := amd64 arm64
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

# デフォルトターゲット
.PHONY: all
all: clean test build

# ヘルプの表示
.PHONY: help
help:
	@echo "使用可能なコマンド:"
	@echo "  make build       - 現在のプラットフォーム用にビルド"
	@echo "  make build-all   - 全プラットフォーム用にビルド"
	@echo "  make test        - テスト実行"
	@echo "  make clean       - ビルド成果物を削除"
	@echo "  make run         - アプリケーション実行"
	@echo "  make fmt         - コードフォーマット"
	@echo "  make vet         - 静的解析"
	@echo "  make package     - 全プラットフォーム用パッケージ作成"
	@echo "  make install     - 依存関係のインストール"

# 依存関係のインストール
.PHONY: install
install:
	go mod download
	go mod tidy

# テスト実行
.PHONY: test
test:
	go test -v -race -cover ./...

# コードフォーマット
.PHONY: fmt
fmt:
	go fmt ./...

# 静的解析
.PHONY: vet
vet:
	go vet ./...

# 現在のプラットフォーム用ビルド
.PHONY: build
build: clean
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) main.go

# 全プラットフォーム用ビルド
.PHONY: build-all
build-all: clean
	@mkdir -p $(BUILD_DIR)
	@for os in $(GOOS_LIST); do \
		for arch in $(GOARCH_LIST); do \
			echo "Building for $$os/$$arch..."; \
			output_name=$(APP_NAME); \
			if [ "$$os" = "windows" ]; then \
				output_name="$$output_name.exe"; \
			fi; \
			GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) \
				-o $(BUILD_DIR)/$(APP_NAME)-$$os-$$arch/$$output_name main.go; \
		done; \
	done

# パッケージ作成
.PHONY: package
package: build-all
	@mkdir -p $(DIST_DIR)
	@for os in $(GOOS_LIST); do \
		for arch in $(GOARCH_LIST); do \
			echo "Packaging for $$os/$$arch..."; \
			package_name="$(APP_NAME)-$(VERSION)-$$os-$$arch"; \
			if [ "$$os" = "windows" ]; then \
				cd $(BUILD_DIR)/$(APP_NAME)-$$os-$$arch && \
				zip -r ../../$(DIST_DIR)/$$package_name.zip . && \
				cd ../..; \
			else \
				tar -czf $(DIST_DIR)/$$package_name.tar.gz \
					-C $(BUILD_DIR)/$(APP_NAME)-$$os-$$arch .; \
			fi; \
		done; \
	done

# アプリケーション実行
.PHONY: run
run:
	go run main.go

# 開発用ビルド（デバッグ情報付き）
.PHONY: build-dev
build-dev: clean
	@mkdir -p $(BUILD_DIR)
	go build -race -o $(BUILD_DIR)/$(APP_NAME)-dev main.go

# クリーンアップ
.PHONY: clean
clean:
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -rf logs
	@echo "クリーンアップ完了"

# 特定プラットフォーム用ビルド
.PHONY: build-linux
build-linux: clean
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 main.go

.PHONY: build-darwin
build-darwin: clean
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 main.go

.PHONY: build-windows
build-windows: clean
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe main.go

# リリース用（全体チェック付き）
.PHONY: release
release: clean fmt vet test build-all package
	@echo "リリース準備完了: $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/

# 継続的統合用
.PHONY: ci
ci: fmt vet test

# ファイル変更監視（developmentモード）
.PHONY: watch
watch:
	@which fswatch > /dev/null || (echo "fswatch が必要です: brew install fswatch"; exit 1)
	fswatch -o . --exclude='logs' --exclude='build' --exclude='dist' | xargs -n1 -I{} make build-dev

# Docker用ビルド（Linuxバイナリ）
.PHONY: docker-build
docker-build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux main.go