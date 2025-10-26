.PHONY: install release_builtin_func

GO ?= go

install: release_builtin_func
	$(GO) install ./cmd/wgo

release_builtin_func:
	$(GO) run ./scripts/release_builtin_func/main.go
