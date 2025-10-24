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
	"sync"

	log "github.com/wxnacy/wgo/internal/logger"
)

const DEFAULT_CODE_TPL = `package main

func main() {
	%s
}`

var (
	logger    = log.GetLogger()
	onceCoder sync.Once
	coder     *Coder
)

func WriteCode(code string, filePath string) error {
	return os.WriteFile(filePath, []byte(code), 0o644)
}

func GetCoder() *Coder {
	if coder == nil {
		onceCoder.Do(func() {
			coder = &Coder{}
		})
	}
	return coder
}

type Coder struct {
	VarFiles []string
}

// 插入代码并运行
// 功能需求:
// - 可以连续插入代码片段并成功运行并返回结果
// - 代码的格式化和处理在 ProcessCode 中进行
func (c *Coder) InsertCodeAndRun(input string) string {
	codePath := GetMainFile()

	// 确保 .wgo 目录存在
	if err := os.MkdirAll(filepath.Dir(codePath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		return ""
	}

	// 检查文件是否存在，如果存在则读取现有代码
	var existingCode string
	if _, err := os.Stat(codePath); err == nil {
		content, err := os.ReadFile(codePath)
		if err == nil {
			existingCode = string(content)
		}
	}

	var code string
	var mainContent string

	// 提取现有代码中的变量声明信息
	var existingVars map[string]bool = make(map[string]bool)

	if existingCode != "" {
		// 提取 main 函数内容
		mainStart := strings.Index(existingCode, "func main() {")
		mainEnd := strings.LastIndex(existingCode, "}")

		if mainStart != -1 && mainEnd != -1 && mainEnd > mainStart {
			// 提取 main 函数体内部的代码
			mainContent = existingCode[mainStart+len("func main() {") : mainEnd]
			mainContent = strings.TrimSpace(mainContent)

			// 分析现有代码，找出所有变量声明
			lines := strings.Split(mainContent, "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				// 匹配简单的变量声明
				if strings.Contains(trimmed, ":=") && !strings.HasPrefix(trimmed, "_") && !strings.HasPrefix(trimmed, "func") && !strings.Contains(trimmed, ",") {
					parts := strings.SplitN(trimmed, ":=", 2)
					if len(parts) > 0 {
						varName := strings.TrimSpace(parts[0])
						// 只处理简单变量名
						if !strings.Contains(varName, "(") && !strings.Contains(varName, "[") && varName != "" {
							existingVars[varName] = true
						}
					}
				}
			}
		}
	}

	// 处理输入代码，转换重复变量声明
	processedInput := processVariableDeclarations(input, existingVars)

	// 构建新的 main 函数内容
	var newMainContent strings.Builder
	if mainContent != "" {
		newMainContent.WriteString(mainContent)
		newMainContent.WriteString("\n")
	}
	newMainContent.WriteString(processedInput)

	// 使用默认模板构建完整代码
	code = fmt.Sprintf(DEFAULT_CODE_TPL, strings.TrimSpace(newMainContent.String()))

	// 处理代码
	processedCode, err := c.ProcessCode(code)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing code: %v\n", err)
		return code
	}

	// 写入文件
	err = WriteCode(processedCode, codePath)
	if err != nil {
		fmt.Printf("写入临时文件失败: %v\n", err)
		panic(err)
	}

	// 运行 goimports
	if _, err := Command("goimports", "-w", codePath); err != nil {
		logger.Errorf("goimports failed: %v", err)
		return err.Error()
	}

	// 运行代码
	out, err := Command(
		"go",
		"run",
		codePath,
		filepath.Join(GetMainDir(), "builtin_func.go"),
		filepath.Join(GetMainDir(), "request.go"),
	)
	if err != nil {
		logger.Errorf("go run failed: %v", err)
		// 即使执行失败，也返回输出
		return out
	}
	return out + "\n"
}

// ProcessCode 处理代码，满足以下功能：
// - 包装未使用的表达式（如 time.Now()）为 fmt.Println 语句
// - 只保留最后一个 fmt.Print 相关的语句
// - 处理未使用的变量，添加 _ = var 到 main 函数末尾
func (c *Coder) ProcessCode(code string) (string, error) {
	// 首先处理未使用的表达式，如 time.Now()
	processedCode, err := wrapUnusedExpressions(code)
	if err != nil {
		return code, err
	}

	// 在处理未使用变量之前，先处理 fmt.Print 语句，只保留最后一个
	processedCode = processFmtPrintStatements(processedCode)

	// 创建临时目录和文件来编译检查未使用变量
	tmpDir, err := os.MkdirTemp("", "go-process-")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(tmpFile, []byte(processedCode), 0o644); err != nil {
		return "", fmt.Errorf("writing temp file: %w", err)
	}

	// 初始化临时 go module
	modCmd := exec.Command("go", "mod", "init", "tmpmodule")
	modCmd.Dir = tmpDir
	var modErr bytes.Buffer
	modCmd.Stderr = &modErr
	if err := modCmd.Run(); err != nil {
		return "", fmt.Errorf("go mod init failed: %s", modErr.String())
	}

	// 运行 go build 并捕获 stderr，我们期望它在有未使用变量时失败
	cmd := exec.Command("go", "build")
	cmd.Dir = tmpDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Run()

	// 正则表达式查找 "declared and not used: var" 错误
	re := regexp.MustCompile(`(?m)^.*: declared and not used: (\w+)$`)
	matches := re.FindAllStringSubmatch(stderr.String(), -1)

	var unusedVars []string
	for _, match := range matches {
		if len(match) > 1 {
			unusedVars = append(unusedVars, match[1])
		}
	}

	// 使用 AST 分析代码，找到 main 函数的结束位置
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", processedCode, 0)
	if err != nil {
		return "", fmt.Errorf("parsing code: %w", err)
	}

	var mainFuncEnd token.Pos = -1

	// 分析 AST 找出 main 函数的右花括号位置
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
			mainFuncEnd = fn.Body.Rbrace
			return false // 找到 main 函数后停止搜索
		}
		return true
	})

	if mainFuncEnd == -1 {
		return "", fmt.Errorf("main function not found")
	}

	// 获取 main 函数右花括号的偏移量
	offset := fset.File(mainFuncEnd).Offset(mainFuncEnd)

	// 构建未使用变量的赋值语句
	var assignments strings.Builder
	for _, v := range unusedVars {
		assignments.WriteString(fmt.Sprintf("\n\t_ = %s", v))
	}

	// 在右花括号前插入赋值语句
	newCode := processedCode[:offset] + assignments.String() + "\n" + processedCode[offset:]

	// 格式化代码以确保正确的缩进
	formatted, err := format.Source([]byte(newCode))
	if err != nil {
		// 如果格式化失败，返回未格式化的版本
		return newCode, nil
	}

	return string(formatted), nil
}

// processFmtPrintStatements 处理代码中的 fmt.Print 语句，只保留最后一个
func processFmtPrintStatements(code string) string {
	// 按行分割代码
	lines := strings.Split(code, "\n")
	var newLines []string
	var lastFmtPrintIndex int = -1

	// 第一次遍历：找出最后一个 fmt.Print 语句的索引
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// 检查是否是 fmt.Print 系列函数调用
		if strings.HasPrefix(trimmed, "fmt.Print") && strings.Contains(trimmed, "(") && strings.Contains(trimmed, ")") {
			lastFmtPrintIndex = i
		}
	}

	// 第二次遍历：构建新代码，只保留最后一个 fmt.Print 语句
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		isFmtPrint := strings.HasPrefix(trimmed, "fmt.Print") && strings.Contains(trimmed, "(") && strings.Contains(trimmed, ")")

		// 如果不是 fmt.Print 语句，或者是最后一个 fmt.Print 语句，则保留
		if !isFmtPrint || (lastFmtPrintIndex != -1 && i == lastFmtPrintIndex) {
			newLines = append(newLines, line)
		}
	}

	// 重新合并代码
	return strings.Join(newLines, "\n")
}

// wrapUnusedExpressions 检测并包装未使用的表达式，如 time.Now()
func wrapUnusedExpressions(code string) (string, error) {
	// 简单的正则表达式来检测可能的未使用表达式
	re := regexp.MustCompile(`(?m)^\s*(\w+(\.\w+)*\([^)]*\))\s*$`)

	lines := strings.Split(code, "\n")
	for i, line := range lines {
		// 跳过注释行和已有的赋值语句
		if !strings.HasPrefix(strings.TrimSpace(line), "//") &&
			!strings.Contains(line, ":=") &&
			!strings.Contains(line, "=") &&
			!strings.Contains(line, "fmt.Print") {

			// 检查是否是独立的函数调用表达式
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				// 检查是否已经是有效的语句
				trimmed := strings.TrimSpace(line)
				if !strings.Contains(trimmed, ":=") && !strings.Contains(trimmed, "=") {
					// 包装表达式到 fmt.Println
					indent := strings.Repeat("\t", strings.Count(line, "\t"))
					lines[i] = indent + "fmt.Println(" + trimmed + ")"
				}
			}
		}
	}

	return strings.Join(lines, "\n"), nil
}

// processVariableDeclarations 处理输入代码中的变量声明，将重复声明转换为赋值
func processVariableDeclarations(input string, existingVars map[string]bool) string {
	lines := strings.Split(input, "\n")
	result := make([]string, 0, len(lines))

	// 处理每一行代码
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 检查是否是简单的变量声明语句
		if strings.Contains(trimmed, ":=") && !strings.HasPrefix(trimmed, "_") &&
			!strings.HasPrefix(trimmed, "func") && !strings.Contains(trimmed, ",") {

			parts := strings.SplitN(trimmed, ":=", 2)
			if len(parts) == 2 {
				varName := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// 检查是否是简单变量名
				if !strings.Contains(varName, "(") && !strings.Contains(varName, "[") && varName != "" {
					// 如果变量已存在，将 := 替换为 =
					if existingVars[varName] {
						// 保留缩进
						indent := ""
						for _, r := range line {
							if r == '\t' {
								indent += "\t"
							} else {
								break
							}
						}
						result = append(result, indent+varName+" = "+value)
					} else {
						// 新变量，保留原样
						result = append(result, line)
						// 添加到已存在变量列表
						existingVars[varName] = true
					}
					continue
				}
			}
		}

		// 非变量声明或复杂声明，保留原样
		result = append(result, line)
	}

	return strings.Join(result, "\n")
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
