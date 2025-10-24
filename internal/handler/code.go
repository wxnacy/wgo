package handler

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/wxnacy/wgo/internal/logger"
)

const DEFAULT_CODE_TPL = `package main

func main() {
	%s
}`

var logger = log.GetLogger()

func GetWorkspace() string {
	workspace, _ := os.Getwd()
	logger.Infof("workspace %s", workspace)
	return workspace
}

func GetMainFile() string {
	return filepath.Join(GetWorkspace(), ".wgo", "main.go")
}

func WriteCode(code string, filePath string) error {
	return os.WriteFile(filePath, []byte(code), 0o644)
}

// 插入代码并运行
// 可以连续插入代码片段并成功运行并返回结果
// 代码的格式化和处理在 ProcessCode 中进行
// TODO: 还不支持连续输入，需要可以读取文件
func InsertCodeAndRun(input string) string {
	tpl := DEFAULT_CODE_TPL
	code := fmt.Sprintf(tpl, input)
	code, err := ProcessCode(code)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing code: %v\n", err)
	}

	codePath := GetMainFile()
	err = WriteCode(code, codePath)
	if err != nil {
		fmt.Printf("写入临时文件失败: %v\n", err)
		panic(err)
	}

	if _, err := Command("goimports", "-w", codePath); err != nil {
		logger.Errorf("goimports failed: %v", err)
		return err.Error()
	}
	out, err := Command("go", "run", codePath)
	if err != nil {
		logger.Errorf("go run failed: %v", err)
		// 即使执行失败，也返回 out
		return out
	}
	return out + "\n"
}

// ProcessCode finds unused variables in the main function of the provided Go code
// and adds assignments to the blank identifier (_) to make the code compile.
// 功能满足：
// 1、a := 1 有没有用到的，需要变成 _ = a 插入 main 函数最后
// 2、time.Now 这种没有用到的代码，套一层 fmt.Println 起到打印效果，避免代码报错
// 3、在完成前面操作后只保留最后一个 fmt.Print 前缀函数代码
// TODO: 需要实现 2、3
func ProcessCode(code string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "go-process-")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		return "", fmt.Errorf("writing temp file: %w", err)
	}

	// Initialize a temporary go module.
	modCmd := exec.Command("go", "mod", "init", "tmpmodule")
	modCmd.Dir = tmpDir
	var modErr bytes.Buffer
	modCmd.Stderr = &modErr
	if err := modCmd.Run(); err != nil {
		return "", fmt.Errorf("go mod init failed: %s", modErr.String())
	}

	// Run 'go build' and capture stderr. We expect it to fail if there are unused vars.
	cmd := exec.Command("go", "build")
	cmd.Dir = tmpDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Run()

	// Regex to find "declared and not used: var" errors.
	re := regexp.MustCompile(`(?m)^.*: declared and not used: (\w+)$`)
	matches := re.FindAllStringSubmatch(stderr.String(), -1)

	var unusedVars []string
	for _, match := range matches {
		if len(match) > 1 {
			unusedVars = append(unusedVars, match[1])
		}
	}

	if len(unusedVars) == 0 {
		// No unused variables found, or a different build error occurred.
		// For this task, we assume other errors are not present and return the original code.
		return code, nil
	}

	// Use AST to find the position of the main function's closing brace.
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", code, 0)
	if err != nil {
		return "", fmt.Errorf("parsing code: %w", err)
	}

	var mainFuncEnd token.Pos = -1
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
			mainFuncEnd = fn.Body.Rbrace
			return false // Stop searching
		}
		return true
	})

	if mainFuncEnd == -1 {
		return "", fmt.Errorf("main function not found")
	}

	// The position is a 1-based offset from the beginning of the file.
	offset := fset.File(mainFuncEnd).Offset(mainFuncEnd)

	var assignments strings.Builder
	for _, v := range unusedVars {
		assignments.WriteString(fmt.Sprintf("\n\t_ = %s", v))
	}

	// Insert the assignments before the closing brace.
	newCode := code[:offset] + assignments.String() + "\n" + code[offset:]

	// Format the resulting code for proper indentation.
	formatted, err := format.Source([]byte(newCode))
	if err != nil {
		// If formatting fails, return the unformatted version.
		return newCode, nil
	}

	return string(formatted), nil
}

func Command(name string, args ...string) (string, error) {
	c := exec.Command(name, args...)
	var out bytes.Buffer
	var outErr bytes.Buffer
	c.Stdout = &out
	c.Stderr = &outErr
	err := c.Run()

	outStr := strings.TrimSpace(out.String())
	errStr := strings.TrimSpace(outErr.String())

	if err != nil {
		if errStr != "" {
			return outStr, errors.New(errStr)
		}
		return outStr, err
	}

	if errStr != "" {
		// 即使成功，但是 err 有输出，也认为是错误
		return outStr, errors.New(errStr)
	}

	return outStr, nil
}
