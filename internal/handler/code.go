package handler

import (
	"bufio"
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
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	log "github.com/wxnacy/wgo/internal/logger"
	"github.com/wxnacy/wgo/pkg/utils"
)

const (
	VAR_PREFIX       = "var-"
	INPUT_SUFFIX     = "// :INPUT"
	DEFAULT_CODE_TPL = `package main

func main() {
	%s
}`
)

var (
	logger               = log.GetLogger()
	onceCoder            sync.Once
	coder                *Coder
	errLineNumberPattern = regexp.MustCompile(`:(\d+)(?::\d+)?:`)
	errLineInfoPattern   = regexp.MustCompile(`^(.+?):(\d+)(?::\d+)?:\s*(.*)$`)
)

var ignoredRunErrorSubstrings = []string{
	"no new variables on left side of :=",
}

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
	VarNames    []string          // 代码文件中 main 函数中出现的变量列表
	FuncCodeMap map[string]string // 代码文件中 main 函数中出现的函数代码
}

// 输入并运行代码
// 功能需求:
// - 调用 InsertOrJoinCode 插入并拼接代码
// - 调用 JoinPrintCode 拼接打印代码
// - 调用 SerializeCodeVars 收集并序列化参数列表
// - 调用 WriteAndRunCode 写入并运行代码
// - 调用 AfterRunCode 处理运行代码后的操作
func (c *Coder) InputAndRun(input string) (string, error) {
	code := c.InsertOrJoinCode(input)
	// 处理代码
	code, err := c.JoinPrintCode(code)
	code = c.SerializeCodeVars(code)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing code: %v\n", err)
		return "", err
	}
	codePath := GetMainFile()
	out, err := c.WriteAndRunCode(code, codePath)
	if err != nil {
		logger.Errorf("RunCode Err:\n%v", err)
	} else {
		logger.Infof("RunCode Out: %s", out)
	}
	// 重新读取最终代码
	if latest, readErr := os.ReadFile(codePath); readErr == nil {
		code = string(latest)
	}
	out, err = c.AfterRunCode(code, out, err)
	return out, err
}

// 运行代码后的操作
// 处理 WriteAndRunCode 命令后的操作
//   - code: 包含 main 函数的完成 go 文件内容
//   - runOut: 运行后的结果输出
//   - runErr: 运行后的错误输出
//
// 忽略错误如下
//   - no new variables on left side of :=
//
// 功能需求:
//
// - runErr 不为空，且不在忽略错误中，判断报错所在行，在 code 中是否包含变量，如果变量在 VerNames 中，需要删除这个条目，如果在 FuncCodeMap 中也要删除
// - runErr 不为空，作如下处理再进行返回
//   - 去掉第一行 # command-line-arguments
//   - 将每行错误的文件名和行号列号信息替换为 code 中对应行的代码
//
// 增加测试用例
func (c *Coder) AfterRunCode(code, runOut string, runErr error) (string, error) {
	if runErr == nil {
		return runOut, nil
	}

	errText := runErr.Error()
	ignored := isIgnoredRunError(errText)

	if !ignored {
		codeLines := strings.Split(code, "\n")
		contextLines := collectErrorContextLines(errText, codeLines)
		if len(contextLines) > 0 {
			removeSet := make(map[string]struct{})
			for _, line := range contextLines {
				if line == "" {
					continue
				}
				for _, name := range c.VarNames {
					if strings.Contains(line, name) {
						removeSet[name] = struct{}{}
					}
				}
				for name := range c.FuncCodeMap {
					if strings.Contains(line, name) {
						removeSet[name] = struct{}{}
					}
				}
			}
			if len(removeSet) > 0 {
				filtered := make([]string, 0, len(c.VarNames))
				for _, name := range c.VarNames {
					if _, ok := removeSet[name]; ok {
						continue
					}
					filtered = append(filtered, name)
				}
				c.VarNames = filtered

				for name := range removeSet {
					delete(c.FuncCodeMap, name)
				}
			}
		}
	}

	formatted := formatRunErrorMessage(code, errText)
	if formatted != errText {
		runErr = errors.New(formatted)
	}

	return runOut, runErr
}

// 写入代码并运行
// 功能需求:
// - 写入到指定地址 codePath
// - 调用 RunCode 运行代码
//
// 增加测试用例
func (c *Coder) WriteAndRunCode(code, codePath string) (string, error) {
	// 写入文件
	err := WriteCode(code, codePath)
	if err != nil {
		logger.Errorf("写入临时文件失败: %v\n", err)
		panic(err)
	}
	return c.RunCode(codePath)
}

// 运行 main 文件
// 功能需求:
// - 对 codePath 进行 goimports 操作
// - go run codePath 时，需要带上同目录下其他的 go 文件
func (c *Coder) RunCode(codePath string) (string, error) {
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

// 插入或者拼接代码
// 功能需求：
// - input 是需要插入的代码片段
// - 如果 VarNames 有数据使用，参数列表使用 _Deserialize 拼接代码, 获取变量最终值
//   - 如果变量是函数，直接拼接函数代码
//
// - 新输入的代码放在最后
// - 最后将拼接好的代码拼接到魔板 DEFAULT_CODE_TPL 中
func (c *Coder) InsertOrJoinCode(input string) string {
	codes := make([]string, 0)
	for _, v := range c.VarNames {
		var line string
		if funcCode, exist := c.FuncCodeMap[v]; exist {
			// 函数变量直接填充代码
			line = fmt.Sprintf("%s := %s", v, funcCode)
		} else {
			// 普通变量通过反序列化函数获取最终值
			name := VAR_PREFIX + v
			typePath := filepath.Join(GetTempDir(), name+".type")
			typeName, err := ReadCode(typePath)
			if err != nil {
				continue
			}
			line = fmt.Sprintf("%s, _ := _Deserialize[%s](\"%s\")", v, typeName, name)
		}
		codes = append(codes, line)
	}
	// 如果已经拼接过 INPUT_SUFFIX 不在拼接
	if strings.Index(input, INPUT_SUFFIX) == -1 {
		input += INPUT_SUFFIX
	}
	codes = append(codes, input)
	code := strings.Join(codes, "\n")
	return fmt.Sprintf(DEFAULT_CODE_TPL, code)
}

// 序列化代码中的变量
// 功能需求:
// - code 会是一个含有 main 函数的完整 go 文件内容
// - 将 main 函数中出现的函数中出现的变量名都包装一个函数名 _Serialize() 放到 main 函数结尾
//   - 举例 `var a = 1` => `_Serialize("var-a", a)`
//   - 举例 `b := 1` => `_Serialize("var-b", b)`
//
// - 如果变量是个函数，则将变量名和对应的函数代码内容插入到 FuncCodeMap 中
//   - 举例 `test := func () { }` => FuncCodeMap["test"] = "func() {}"
//
// - 如果某个参数没有赋值，或者为 nil 则不要做这个操作
// - 如果代码中已经有 _Serialize 包装的代码，则迁移到 main 函数结尾
// - 将 _Serialize 包装过的参数名列表保存到 VarNames 中
// - VarNames 顺序要严格按照 main 中出现的顺序，不可以做其他排序
//
// 增加测试用例
func (c *Coder) SerializeCodeVars(code string) string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		return code
	}

	mainFunc := findMainFunc(file)
	if mainFunc == nil || mainFunc.Body == nil {
		return code
	}

	originalSerialize, serializedNames := extractSerializeCalls(mainFunc.Body)
	mainFunc.Body.List = removeSerializeStmts(mainFunc.Body.List)

	orderedVars := collectSerializableVars(mainFunc.Body)
	newSerialize := makeSerializeCalls(orderedVars, serializedNames)

	mainFunc.Body.List = append(mainFunc.Body.List, append(originalSerialize, newSerialize...)...)

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return code
	}

	c.VarNames = gatherVarNames(orderedVars, serializedNames, originalSerialize)
	nameSet := make(map[string]struct{}, len(c.VarNames))
	for _, name := range c.VarNames {
		nameSet[name] = struct{}{}
	}
	c.captureFuncLiterals(fset, orderedVars, nameSet)
	logger.Debugf("VarNames %v", c.VarNames)
	logger.Debugf("FuncCodeMap %#v", c.FuncCodeMap)

	return buf.String()
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
	processedCode = c.SerializeCodeVars(processedCode)
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

// 格式化代码
// - code 是个包含 main 函数的 go 代码
// - 对 code 完成一下一些列操作后返回
//
// 功能需求:
// - 对未使用的表达式或参数使用 fmt.Println 进行包装
//   - 比如 time.Now() => fmt.Println(time.Now())
//   - 比如 t => fmt.Println(t) ，其中 t 是参数
//   - 比如 a := time.Now(); a => a := time.Now() 换行 fmt.Println(a)
//   - 比如 test := func ()  { return "wxnacy" }; test 方法格式化以后 换行 fmt.Println(test)
//
// - 以下情况不要进行 fmt.Println 封装
//   - var 定义变量，比如 `var name string`
//
// - 最后只保留最后一个 fmt.Print 开头的代码
//
// 增加测试用例
func (c *Coder) JoinPrintCode(code string) (string, error) {
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		idx := strings.Index(line, INPUT_SUFFIX)
		if idx == -1 {
			continue
		}

		content := line[:idx]
		indent := content[:len(content)-len(strings.TrimLeft(content, "\t "))]
		segments := strings.Split(content, ";")

		var tokens []string
		for _, segment := range segments {
			trimmed := strings.TrimSpace(segment)
			if trimmed == "" {
				continue
			}
			tokens = append(tokens, trimmed)
		}

		if len(tokens) == 0 {
			lines[i] = indent
			continue
		}

		var newLines []string
		if len(tokens) > 1 {
			for _, stmt := range tokens[:len(tokens)-1] {
				newLines = append(newLines, indent+stmt)
			}
		}

		last := tokens[len(tokens)-1]
		plain := strings.TrimSpace(last)
		replacement := indent + plain
		if plain == "" {
			replacement = indent
		} else if !strings.HasPrefix(plain, "fmt.Print") &&
			!strings.Contains(plain, ":=") && !strings.Contains(plain, "=") &&
			!strings.HasPrefix(plain, "if ") && !strings.HasPrefix(plain, "for ") &&
			!strings.HasPrefix(plain, "switch ") && !strings.HasPrefix(plain, "select ") &&
			!strings.HasPrefix(plain, "var ") {
			replacement = indent + fmt.Sprintf("fmt.Println(%s)", plain)
		}
		newLines = append(newLines, replacement)

		lines[i] = strings.Join(newLines, "\n")
	}

	joined := strings.Join(lines, "\n")

	wrapped, err := wrapUnusedExpressions(joined)
	if err != nil {
		return code, err
	}

	processed := processFmtPrintStatements(wrapped)
	return processed, nil
}

// ProcessCode 处理代码，满足以下功能：
// - 包装未使用的表达式（如 time.Now()）为 fmt.Println 语句
// - 只保留最后一个 fmt.Print 相关的语句
// - 处理未使用的变量，添加 _ = var 到 main 函数末尾
func (c *Coder) ProcessCode(code string) (string, error) {
	processedCode, err := c.JoinPrintCode(code)
	if err != nil {
		return code, err
	}

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

type varEntry struct {
	name string
	pos  token.Pos
	expr ast.Expr
}

func findMainFunc(file *ast.File) *ast.FuncDecl {
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
			return fn
		}
	}
	return nil
}

func extractSerializeCalls(body *ast.BlockStmt) ([]ast.Stmt, map[string]struct{}) {
	var calls []ast.Stmt
	names := make(map[string]struct{})
	for _, stmt := range body.List {
		exprStmt := collectSerializeCall(stmt)
		if exprStmt == nil {
			continue
		}
		call, _ := exprStmt.X.(*ast.CallExpr)
		if call == nil || len(call.Args) < 2 || isNilExpr(call.Args[1]) {
			continue
		}
		if name := serializeCallVarName(call); name != "" {
			names[name] = struct{}{}
		}
		calls = append(calls, exprStmt)
	}
	return calls, names
}

func collectSerializeCall(stmt ast.Stmt) *ast.ExprStmt {
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return nil
	}
	call, ok := exprStmt.X.(*ast.CallExpr)
	if !ok {
		return nil
	}
	if fun, ok := call.Fun.(*ast.Ident); !ok || fun.Name != "_Serialize" {
		return nil
	}
	return exprStmt
}

func removeSerializeStmts(stmts []ast.Stmt) []ast.Stmt {
	var filtered []ast.Stmt
	for _, stmt := range stmts {
		if collectSerializeCall(stmt) != nil {
			continue
		}
		filtered = append(filtered, stmt)
	}
	return filtered
}

func collectSerializableVars(body *ast.BlockStmt) []varEntry {
	seen := make(map[string]struct{})
	var entries []varEntry
	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			if node.Tok != token.DEFINE && node.Tok != token.ASSIGN {
				return true
			}
			for i, lhs := range node.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if !ok || ident.Name == "_" {
					continue
				}
				if _, exists := seen[ident.Name]; exists {
					continue
				}
				if i >= len(node.Rhs) {
					continue
				}
				rhs := node.Rhs[i]
				if isNilExpr(rhs) {
					continue
				}
				seen[ident.Name] = struct{}{}
				entries = append(entries, varEntry{name: ident.Name, pos: ident.NamePos, expr: rhs})
			}
		case *ast.ValueSpec:
			valueCount := len(node.Values)
			for i, name := range node.Names {
				if name.Name == "_" {
					continue
				}
				if _, exists := seen[name.Name]; exists {
					continue
				}
				if i >= valueCount {
					continue
				}
				rhs := node.Values[i]
				if isNilExpr(rhs) {
					continue
				}
				seen[name.Name] = struct{}{}
				entries = append(entries, varEntry{name: name.Name, pos: name.NamePos, expr: rhs})
			}
		}
		return true
	})

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].pos < entries[j].pos
	})

	return entries
}

func makeSerializeCalls(entries []varEntry, existing map[string]struct{}) []ast.Stmt {
	var stmts []ast.Stmt
	for _, entry := range entries {
		if _, ok := existing[entry.name]; ok {
			continue
		}
		lit := &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", VAR_PREFIX+entry.name)}
		call := &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun:  ast.NewIdent("_Serialize"),
				Args: []ast.Expr{lit, ast.NewIdent(entry.name)},
			},
		}
		stmts = append(stmts, call)
		existing[entry.name] = struct{}{}
	}
	return stmts
}

func gatherVarNames(entries []varEntry, existing map[string]struct{}, original []ast.Stmt) []string {
	var names []string
	seen := make(map[string]struct{})
	for _, entry := range entries {
		if _, ok := existing[entry.name]; !ok {
			continue
		}
		if _, added := seen[entry.name]; added {
			continue
		}
		names = append(names, entry.name)
		seen[entry.name] = struct{}{}
	}
	for _, stmt := range original {
		exprStmt := collectSerializeCall(stmt)
		if exprStmt == nil {
			continue
		}
		call, _ := exprStmt.X.(*ast.CallExpr)
		if call == nil {
			continue
		}
		name := serializeCallVarName(call)
		if name == "" {
			continue
		}
		if _, added := seen[name]; added {
			continue
		}
		names = append(names, name)
		seen[name] = struct{}{}
	}
	return names
}

func (c *Coder) captureFuncLiterals(fset *token.FileSet, entries []varEntry, validNames map[string]struct{}) {
	if len(validNames) == 0 {
		return
	}
	if c.FuncCodeMap == nil {
		c.FuncCodeMap = make(map[string]string)
	}
	updated := make(map[string]struct{})
	for _, entry := range entries {
		if _, ok := validNames[entry.name]; !ok {
			continue
		}
		if entry.expr == nil {
			continue
		}
		funcLit, ok := entry.expr.(*ast.FuncLit)
		if !ok {
			continue
		}
		var buf bytes.Buffer
		if err := format.Node(&buf, fset, funcLit); err != nil {
			continue
		}
		c.FuncCodeMap[entry.name] = buf.String()
		updated[entry.name] = struct{}{}
	}
	for name := range validNames {
		if _, ok := updated[name]; ok {
			continue
		}
		delete(c.FuncCodeMap, name)
	}
}

func extractErrorLineNumbers(errText string) []int {
	matches := errLineNumberPattern.FindAllStringSubmatch(errText, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[int]struct{})
	var lines []int
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		num, err := strconv.Atoi(match[1])
		if err != nil {
			continue
		}
		if _, ok := seen[num]; ok {
			continue
		}
		seen[num] = struct{}{}
		lines = append(lines, num)
	}
	return lines
}

func collectErrorContextLines(errText string, codeLines []string) []string {
	var contexts []string
	lines := strings.Split(errText, "\n")
	for _, line := range lines {
		matches := errLineInfoPattern.FindStringSubmatch(line)
		if matches == nil || len(matches) < 4 {
			continue
		}
		lineNumber, err := strconv.Atoi(matches[2])
		if err != nil {
			continue
		}
		codeLine := lookupCodeLine(codeLines, lineNumber)
		if codeLine == "" {
			codeLine = lookupFileLine(matches[1], lineNumber)
		}
		contexts = append(contexts, codeLine)
	}
	return contexts
}

func isIgnoredRunError(errText string) bool {
	for _, substr := range ignoredRunErrorSubstrings {
		if strings.Contains(errText, substr) {
			return true
		}
	}
	return false
}

func formatRunErrorMessage(code, errText string) string {
	if errText == "" {
		return errText
	}
	lines := strings.Split(errText, "\n")
	if len(lines) == 0 {
		return errText
	}
	if strings.TrimSpace(lines[0]) == "# command-line-arguments" {
		lines = lines[1:]
	}
	if len(lines) == 0 {
		return ""
	}
	codeLines := strings.Split(code, "\n")
	for i, line := range lines {
		lines[i] = replaceErrorLineWithCode(line, codeLines)
	}
	return strings.Join(lines, "\n")
}

func replaceErrorLineWithCode(line string, codeLines []string) string {
	matches := errLineInfoPattern.FindStringSubmatch(line)
	if matches == nil || len(matches) < 4 {
		return line
	}
	lineNumber, err := strconv.Atoi(matches[2])
	if err != nil {
		return line
	}
	codeLine := lookupCodeLine(codeLines, lineNumber)
	if codeLine == "" {
		codeLine = lookupFileLine(matches[1], lineNumber)
	}
	if codeLine == "" {
		codeLine = fmt.Sprintf("line %d", lineNumber)
	}
	message := strings.TrimSpace(matches[3])
	if message == "" {
		return codeLine
	}
	return fmt.Sprintf("%s: %s", codeLine, message)
}

func lookupCodeLine(codeLines []string, lineNumber int) string {
	idx := lineNumber - 1
	if idx < 0 || idx >= len(codeLines) {
		return ""
	}
	return strings.TrimSpace(codeLines[idx])
}

func lookupFileLine(path string, lineNumber int) string {
	if path == "" {
		return ""
	}
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	current := 1
	for scanner.Scan() {
		if current == lineNumber {
			return strings.TrimSpace(scanner.Text())
		}
		current++
	}
	return ""
}

func isNilExpr(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "nil"
}

func serializeCallVarName(call *ast.CallExpr) string {
	if len(call.Args) < 2 {
		return ""
	}
	if ident, ok := call.Args[1].(*ast.Ident); ok {
		return ident.Name
	}
	if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
		if unquoted, err := strconv.Unquote(lit.Value); err == nil {
			if strings.HasPrefix(unquoted, VAR_PREFIX) {
				return unquoted[len(VAR_PREFIX):]
			}
		}
	}
	return ""
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

func SerializeVar[T any](name string, value T) error {
	dir := GetTempDir()
	filePath := filepath.Join(dir, name)
	typePath := filepath.Join(dir, name+".type")

	if err := utils.SerializeToFile(filePath, value); err != nil {
		return err
	}

	t := reflect.TypeOf(value)
	if t == nil {
		return errors.New("无法获取对象类型")
	}
	return WriteCode(t.String(), typePath)
}

func ReadCode(p string) (string, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
