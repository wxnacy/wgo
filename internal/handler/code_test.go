package handler

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSerializeCodeVarsBasic(t *testing.T) {
	input := `package main

func main() {
	var a = 1
	b := 2
}`

	c := GetCoder()
	got := c.SerializeCodeVars(input)

	callNames := serializeCallNamesFromCode(t, got)
	expect := []string{"a", "b"}
	if !reflect.DeepEqual(callNames, expect) {
		t.Fatalf("unexpected serialize call order: %v", callNames)
	}

	if !reflect.DeepEqual(c.VarNames, expect) {
		t.Fatalf("unexpected VarNames: %v", c.VarNames)
	}
}

func TestSerializeCodeVarsSkipsNilAndUnassigned(t *testing.T) {
	input := `package main

func main() {
	var a = 1
	var b int
	var c *int = nil
	_ = b
}`

	c := GetCoder()
	got := c.SerializeCodeVars(input)
	callNames := serializeCallNamesFromCode(t, got)
	expect := []string{"a"}
	if !reflect.DeepEqual(callNames, expect) {
		t.Fatalf("unexpected serialize calls: %v", callNames)
	}
	if !reflect.DeepEqual(c.VarNames, expect) {
		t.Fatalf("unexpected VarNames: %v", c.VarNames)
	}
}

func TestSerializeCodeVarsKeepsExisting(t *testing.T) {
	input := `package main

func main() {
	var a = 1
	_Serialize("var-a", a)
	b := 2
}`

	c := GetCoder()
	got := c.SerializeCodeVars(input)
	callNames := serializeCallNamesFromCode(t, got)
	expect := []string{"a", "b"}
	if !reflect.DeepEqual(callNames, expect) {
		t.Fatalf("expected calls %v, got %v", expect, callNames)
	}
	if !reflect.DeepEqual(c.VarNames, expect) {
		t.Fatalf("unexpected VarNames: %v", c.VarNames)
	}

	// ensure _Serialize("var-a", a) only appears once
	seen := 0
	for _, name := range callNames {
		if name == "a" {
			seen++
		}
	}
	if seen != 1 {
		t.Fatalf("expected single serialization for 'a', got %d", seen)
	}
}

func TestSerializeCodeVarsStoresFuncLiteral(t *testing.T) {
	input := `package main

func main() {
	test := func() string { return "wxnacy" }
}
`

	c := &Coder{}
	got := c.SerializeCodeVars(input)
	callNames := serializeCallNamesFromCode(t, got)
	expect := []string{"test"}
	if !reflect.DeepEqual(callNames, expect) {
		t.Fatalf("期望序列化顺序为 %v, 实际为 %v", expect, callNames)
	}
	if !reflect.DeepEqual(c.VarNames, expect) {
		t.Fatalf("VarNames 应为 %v, 实际为 %v", expect, c.VarNames)
	}
	code, ok := c.FuncCodeMap["test"]
	if !ok {
		t.Fatalf("FuncCodeMap 中缺少 test 键: %+v", c.FuncCodeMap)
	}
	expectedCode := `func() string { return "wxnacy" }`
	if code != expectedCode {
		t.Fatalf("函数代码序列化异常: 期望 %q, 实际 %q", expectedCode, code)
	}
}

func TestAfterRunCodeRemovesInvalidEntries(t *testing.T) {
	code := `package main

func main() {
	fmt.Println(test)
}
`

	c := &Coder{
		VarNames:    []string{"test", "keep"},
		FuncCodeMap: map[string]string{"test": "func() string { return \"wxnacy\" }", "keep": "func() {}"},
	}

	runErr := errors.New(`# command-line-arguments
/tmp/main.go:4:14: undefined: test`)
	runOut := "output"
	out, err := c.AfterRunCode(code, runOut, runErr)
	if out != runOut {
		t.Fatalf("输出应保持不变, 期望 %q, 实际 %q", runOut, out)
	}
	if err == nil {
		t.Fatalf("错误应原样返回, 不应为 nil")
	}
	expectErr := "fmt.Println(test): undefined: test"
	if err.Error() != expectErr {
		t.Fatalf("错误信息应格式化, 期望 %q, 实际 %q", expectErr, err.Error())
	}

	expectVars := []string{"keep"}
	if !reflect.DeepEqual(c.VarNames, expectVars) {
		t.Fatalf("VarNames 应更新为 %v, 实际 %v", expectVars, c.VarNames)
	}
	if _, exists := c.FuncCodeMap["test"]; exists {
		t.Fatalf("FuncCodeMap 中 test 应被移除, 当前: %+v", c.FuncCodeMap)
	}
	if _, exists := c.FuncCodeMap["keep"]; !exists {
		t.Fatalf("FuncCodeMap 应保留无关的键, 当前: %+v", c.FuncCodeMap)
	}
}

func TestAfterRunCodeIgnoresConfiguredErrors(t *testing.T) {
	code := `package main

func main() {
	a := 1
	a := 2
}
`

	c := &Coder{
		VarNames: []string{"a"},
	}

	runErr := errors.New(`# command-line-arguments
/tmp/main.go:5:5: no new variables on left side of :=`)
	out, err := c.AfterRunCode(code, "out", runErr)
	if out != "out" {
		t.Fatalf("输出应保持不变, 期望 %q, 实际 %q", "out", out)
	}
	if err == nil {
		t.Fatalf("忽略错误仍应返回错误信息")
	}
	expectErr := "a := 2: no new variables on left side of :="
	if err.Error() != expectErr {
		t.Fatalf("忽略错误应格式化, 期望 %q, 实际 %q", expectErr, err.Error())
	}
	if !reflect.DeepEqual(c.VarNames, []string{"a"}) {
		t.Fatalf("忽略错误不应清理 VarNames, 当前 %v", c.VarNames)
	}
	if c.FuncCodeMap != nil && len(c.FuncCodeMap) != 0 {
		t.Fatalf("忽略错误不应修改 FuncCodeMap, 当前 %+v", c.FuncCodeMap)
	}
}

func TestWriteAndRunCodeIncludesSiblingFiles(t *testing.T) {
	c := GetCoder()

	dir := t.TempDir()
	helperPath := filepath.Join(dir, "helper.go")
	helperCode := `package main

func greet() string {
    return "hello"
}
`
	if err := os.WriteFile(helperPath, []byte(helperCode), 0o644); err != nil {
		t.Fatalf("写入辅助文件失败: %v", err)
	}

	codePath := filepath.Join(dir, "main.go")
	mainCode := `package main

func main() {
    fmt.Println(greet())
}
`

	out, err := c.WriteAndRunCode(mainCode, codePath)
	if err != nil {
		t.Fatalf("WriteAndRunCode 执行失败: %v", err)
	}

	if out != "hello" {
		t.Fatalf("期望输出 hello, 实际为 %q", out)
	}
}

func TestJoinPrintCodeWrapsFunctionCall(t *testing.T) {
	c := &Coder{}
	input := `package main

func main() {
	time.Now()// :INPUT
}
`

	got, err := c.JoinPrintCode(input)
	if err != nil {
		t.Fatalf("JoinPrintCode 返回错误: %v", err)
	}

	if strings.Contains(got, INPUT_SUFFIX) {
		t.Fatalf("JoinPrintCode 未移除输入标记: %s", got)
	}

	if !strings.Contains(got, "fmt.Println(time.Now())") {
		t.Fatalf("期望输出包含 fmt.Println 包装: %s", got)
	}
}

func TestJoinPrintCodeWrapsIdentifier(t *testing.T) {
	c := &Coder{}
	input := `package main

func main() {
	var t int
	t// :INPUT
}
`

	got, err := c.JoinPrintCode(input)
	if err != nil {
		t.Fatalf("JoinPrintCode 返回错误: %v", err)
	}

	if !strings.Contains(got, "fmt.Println(t)") {
		t.Fatalf("期望标识符被打印: %s", got)
	}
}

func TestJoinPrintCodeKeepsOnlyLastPrint(t *testing.T) {
	c := &Coder{}
	input := `package main

func main() {
	fmt.Println("a")
	fmt.Println("b")// :INPUT
}
`

	got, err := c.JoinPrintCode(input)
	if err != nil {
		t.Fatalf("JoinPrintCode 返回错误: %v", err)
	}

	if strings.Contains(got, "fmt.Println(\"a\")") {
		t.Fatalf("应当移除非最后的 fmt.Println: %s", got)
	}
	if !strings.Contains(got, "fmt.Println(\"b\")") {
		t.Fatalf("缺少最后的 fmt.Println: %s", got)
	}
}

func TestJoinPrintCodeKeepsAssignments(t *testing.T) {
	c := &Coder{}
	input := `package main

func main() {
	a := 1// :INPUT
}
`

	got, err := c.JoinPrintCode(input)
	if err != nil {
		t.Fatalf("JoinPrintCode 返回错误: %v", err)
	}

	if !strings.Contains(got, "a := 1") {
		t.Fatalf("赋值语句应当保留: %s", got)
	}
	if strings.Contains(got, "fmt.Println(a := 1)") {
		t.Fatalf("赋值语句不应被包装: %s", got)
	}
}

func TestJoinPrintCodeSplitsSemicolonExpression(t *testing.T) {
	c := &Coder{}
	input := `package main

import "time"

func main() {
	a := time.Now(); a// :INPUT
}
`

	got, err := c.JoinPrintCode(input)
	if err != nil {
		t.Fatalf("JoinPrintCode 返回错误: %v", err)
	}

	if !strings.Contains(got, "a := time.Now()") {
		t.Fatalf("应保留赋值语句: %s", got)
	}
	if !strings.Contains(got, "fmt.Println(a)") {
		t.Fatalf("应打印最后的表达式: %s", got)
	}
}

func TestJoinPrintCodeSplitsFuncLiteral(t *testing.T) {
	c := &Coder{}
	input := `package main

func main() {
	test := func() string { return "wxnacy" }; test// :INPUT
}
`

	got, err := c.JoinPrintCode(input)
	if err != nil {
		t.Fatalf("JoinPrintCode 返回错误: %v", err)
	}

	if !strings.Contains(got, "test := func() string { return \"wxnacy\" }") {
		t.Fatalf("应保留函数定义: %s", got)
	}
	if !strings.Contains(got, "fmt.Println(test)") {
		t.Fatalf("应打印函数变量: %s", got)
	}
}

func serializeCallNamesFromCode(t *testing.T, code string) []string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", code, 0)
	if err != nil {
		t.Fatalf("parse result error: %v", err)
	}
	mainFunc := findMainFunc(file)
	if mainFunc == nil {
		t.Fatalf("main function not found")
	}
	var names []string
	for _, stmt := range mainFunc.Body.List {
		exprStmt := collectSerializeCall(stmt)
		if exprStmt == nil {
			continue
		}
		call, _ := exprStmt.X.(*ast.CallExpr)
		if call == nil {
			continue
		}
		name := serializeCallVarName(call)
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}
