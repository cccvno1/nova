.PHONY: run build clean test dev prod install-web build-web dev-web

# ============ 后端命令 ============
run:
	go run cmd/server/main.go

dev:
	go run cmd/server/main.go -config=configs/config.yaml

prod:
	go run cmd/server/main.go -config=configs/config.prod.yaml

build:
	go build -o bin/server cmd/server/main.go

test:
	go test -v ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run

tidy:
	go mod tidy

# ============ 前端命令 ============
install-web:
	cd web && npm install

dev-web:
	cd web && npm run dev

build-web:
	cd web && npm run build

# ============ 全栈命令 ============
# 安装所有依赖（前后端）
install: tidy install-web

# 构建全栈项目（前端 + 后端）
build-all: build-web build
	@echo "✅ 前后端构建完成！"
	@echo "📦 前端产物：dist/"
	@echo "📦 后端产物：bin/server"

# 清理所有构建产物
clean:
	rm -rf bin/ logs/ dist/ web/dist/ web/node_modules/.vite

# 开发模式（同时启动前后端）
# 注意：需要两个终端分别运行
dev-all:
	@echo "请在两个终端分别运行："
	@echo "  终端1: make dev"
	@echo "  终端2: make dev-web"

