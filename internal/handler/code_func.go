package handler

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/imports"
)

// 格式化代码
func FormatCode(code string) (string, error) {
	fmtByte, err := format.Source([]byte(code))
	if err != nil {
		return "", err
	}
	code = string(fmtByte)
	logger.Debugf("FormatCode %s", code)
	return code, nil
}

// 格式化代码
func ImportsCode(code string) (string, error) {
	// 处理代码：自动添加/移除导入 + 格式化
	// 参数1: 虚拟文件名（用于错误提示）
	// 参数2: 原始代码字节切片
	// 参数3: 配置选项（nil表示默认配置）
	formattedCode, err := imports.Process("virtual.go", []byte(code), nil)
	if err != nil {
		return "", nil
	}
	code = string(formattedCode)
	logger.Debugf("ImportsCode %s", code)
	return code, nil
}

// ImportsInFile 处理指定的.go文件，自动补全缺失的导入并修改源文件
// 参数：filePath 目标文件路径
// 返回：是否修改成功，以及可能的错误
func ImportsInFile(filePath string) (bool, error) {
	// 1. 读取源文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("读取文件失败: %w", err)
	}

	// 2. 用 imports.Process 处理内容（补全导入+格式化）
	// 第一个参数：传入真实文件名（用于正确解析包路径和导入）
	// 第二个参数：原始文件内容
	// 第三个参数：配置项（nil 表示默认配置，可自定义本地包前缀等）
	fixedContent, err := imports.Process(filePath, content, nil)
	if err != nil {
		return false, fmt.Errorf("处理导入失败: %w", err)
	}

	// 3. 对比处理前后的内容，避免无意义的写入
	if string(fixedContent) == string(content) {
		return false, nil
	}

	// 4. 将处理后的内容写回源文件（覆盖原文件）
	err = os.WriteFile(filePath, fixedContent, 0o644) // 保持原文件权限（0644 是常见权限）
	if err != nil {
		return false, fmt.Errorf("写入文件失败: %w", err)
	}

	return true, nil
}

// 运行 main 文件
// 功能需求:
// - 对 codePath 进行 imports 操作
// - go run codePath 时，需要带上同目录下其他的 go 文件
func RunCode(codePath string) (string, error) {
	// 运行 imports
	if _, err := ImportsInFile(codePath); err != nil {
		logger.Errorf("imports failed: %v", err)
		return "", err
	}

	// 收集同目录下的其他 Go 文件
	dir := filepath.Dir(codePath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		logger.Errorf("读取目录失败: %v", err)
		return "", err
	}

	var goFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		// 过滤测试文件
		if strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		fullPath := filepath.Join(dir, entry.Name())
		if fullPath == codePath {
			continue
		}
		goFiles = append(goFiles, fullPath)
	}

	sort.Strings(goFiles)
	args := append([]string{"run", codePath}, goFiles...)

	// 运行代码
	return Command("go", args...)
}

func WriteAndRunCode(code, codePath string) (string, error) {
	// 写入文件
	err := WriteCode(code, codePath)
	if err != nil {
		logger.Errorf("写入临时文件失败: %v\n", err)
		return "", err
	}
	return RunCode(codePath)
}
