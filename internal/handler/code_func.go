package handler

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

// 运行 main 文件
// 功能需求:
// - 对 codePath 进行 goimports 操作
// - go run codePath 时，需要带上同目录下其他的 go 文件
func RunCode(codePath string) (string, error) {
	// 运行 goimports
	if _, err := Command("goimports", "-w", codePath); err != nil {
		logger.Errorf("goimports failed: %v", err)
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

// HasFunctionReturnByCode 检查代码文本中指定名称的函数是否有返回值
// 参数：
//   - code: 待解析的Go代码文本（需包含完整的包声明和函数定义）
//   - funcName: 要检查的函数名称（区分大小写）
//
// 返回：
//   - bool: 函数是否有返回值（true=有，false=无）
//   - error: 错误信息（如代码解析失败、函数未找到等）
func HasFunctionReturnByCode(code, funcName string) (bool, error) {
	// 1. 解析代码文本（虚拟一个文件名，如 "virtual.go"，不影响解析）
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "virtual.go", []byte(code), parser.ParseComments)
	if err != nil {
		return false, fmt.Errorf("解析代码文本失败: %w", err) // 包装语法错误等信息
	}

	// 2. 遍历AST，查找目标函数并检查返回值（逻辑与之前一致）
	var hasReturn bool
	found := false

	ast.Inspect(file, func(n ast.Node) bool {
		if found {
			return false // 找到后提前退出
		}

		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if funcDecl.Name.Name == funcName {
			found = true
			// 检查返回值列表
			hasReturn = funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0
			return false
		}

		return true
	})

	// 3. 处理未找到函数的情况
	if !found {
		return false, fmt.Errorf("代码文本中未找到函数 %q", funcName)
	}

	return hasReturn, nil
}
