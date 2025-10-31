//go:build tools
// +build tools

// 该文件用于将命令行工具（gopls）记录为模块依赖，便于团队协作与版本管理。
// 注意：这些工具不会参与项目编译，仅用于在需要时固定工具版本；当前项目在安装时使用 latest 安装，保持最新。
// 如需固定版本，可在 go.mod 中显式 require 相应版本，并将 Makefile 中的版本从 latest 改为具体版本。
package tools

import (
    // 引入 gopls（Go 语言服务器），用于 LSP 功能
    _ "golang.org/x/tools/gopls"
)
