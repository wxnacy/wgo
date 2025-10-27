.PHONY: install release_builtin_func release tools

GO ?= go

# 工具安装的默认版本。按照需求使用 latest 保持最新（如需稳定，可改为具体版本号）
GOPLS_VERSION ?= latest
GOIMPORTS_VERSION ?= latest

# 需要安装的通用命令行工具列表。安装到 GOBIN 或 GOPATH/bin 中。
TOOLS := \
	golang.org/x/tools/gopls@$(GOPLS_VERSION) \
	golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)

# 安装项目所需的通用工具。使用 latest，避免固定在 go.mod；版本由 Makefile 控制。
tools:
	@echo "Installing tools..."
	@for t in $(TOOLS); do \
		echo "  -> $$t"; \
		$(GO) install $$t || exit 1; \
	done
	@echo "Tools installed."

# 注意：install 增加对 tools 的依赖，确保在安装项目可执行文件前安装公共工具
install: tools release_builtin_func
	$(GO) install ./cmd/wgo

release_builtin_func:
	$(GO) run ./scripts/release_builtin_func/main.go

# 代码发布
# 功能需求:
# - 创建 scripts/release 脚本完成具体操作，然后 `make release` 调用该脚本
# - 使用 `go run ./cmd/wgo/main.go --version` 获取具体版本号
# - 从 `git tag` 中获取对应标签 tag，前面可能带 v
# - 从 HISTORY 中获取对应版本的更新信息，版本前面可能带 v
# - 构建 MacOS、LInux、Windows 三个平台的运行脚本，压缩成 zip 包作为附件（包名版本号前边加v）
# - 将附件和更新日志信息一起发布到 `git release` 上
release:
	./scripts/release
