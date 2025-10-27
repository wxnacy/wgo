.PHONY: install release_builtin_func release

GO ?= go

install: release_builtin_func
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
