package handler

import (
	"errors"
	"fmt"
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

// CanPrintFunction: 代码内定义的函数有返回值 => true
func TestCanPrintFunction_CodeLocal_WithReturn(t *testing.T) {
	c := &Coder{}
	code := `package main

func F() int { return 1 }

func main() {}
`
	if !c.CanPrintFunction(code, "F") {
		t.Fatalf("期望 F 有返回值，CanPrintFunction 返回 false")
	}
}

// CanPrintFunction: 代码内定义的函数无返回值 => false
func TestCanPrintFunction_CodeLocal_NoReturn(t *testing.T) {
	c := &Coder{}
	code := `package main

func G() {}

func main() {}
`
	if c.CanPrintFunction(code, "G") {
		t.Fatalf("期望 G 无返回值，CanPrintFunction 返回 true")
	}
}

// CanPrintFunction: 外部标准库有返回值（time.Now）=> true（通过反射探测）
func TestCanPrintFunction_External_TimeNow(t *testing.T) {
	c := &Coder{}
	code := `package main
func main() {}
`
	if !c.CanPrintFunction(code, "time.Now") {
		t.Fatalf("期望 time.Now 有返回值，CanPrintFunction 返回 false")
	}
}

// CanPrintFunction: 外部标准库无返回值（time.Sleep）=> false（通过反射探测）
func TestCanPrintFunction_External_TimeSleep(t *testing.T) {
	c := &Coder{}
	code := `package main
func main() {}
`
	if c.CanPrintFunction(code, "time.Sleep") {
		t.Fatalf("期望 time.Sleep 无返回值，CanPrintFunction 返回 true")
	}
}

// 无返回值（标准库）不应被自动打印
func TestJoinPrintCode_NoReturnCall_TimeSleep_NotWrapped(t *testing.T) {
	c := &Coder{}
	input := `package main

import "time"

func main() {
    time.Sleep(0)// :INPUT
}
`
	got, err := c.JoinPrintCode(input)
	if err != nil {
		t.Fatalf("JoinPrintCode 返回错误: %v", err)
	}
	if strings.Contains(got, "fmt.Println(time.Sleep(0))") {
		t.Fatalf("无返回值的 time.Sleep 不应被打印: %s", got)
	}
}

// 链式调用（有返回值）应被打印
func TestJoinPrintCode_ChainedCall_WithOut_Wrapped(t *testing.T) {
	c := &Coder{}
	input := `package main

type T struct{}
func GetT() T { return T{} }
func (T) Val() int { return 42 }

func main() {
    GetT().Val()// :INPUT
}
`
	got, err := c.JoinPrintCode(input)
	if err != nil {
		t.Fatalf("JoinPrintCode 返回错误: %v", err)
	}
	if !strings.Contains(got, "fmt.Println(GetT().Val())") {
		t.Fatalf("链式有返回值调用应被打印: %s", got)
	}
}

// 无返回值调用不应被自动打印
func TestJoinPrintCode_NoReturnCall_NotWrapped(t *testing.T) {
	c := &Coder{}
	input := `package main

func InitX() {}

func main() {
    InitX()// :INPUT
}
`

	got, err := c.JoinPrintCode(input)
	if err != nil {
		t.Fatalf("JoinPrintCode 返回错误: %v", err)
	}
	if strings.Contains(got, "fmt.Println(InitX())") {
		t.Fatalf("无返回值调用不应被 fmt.Println 包裹: %s", got)
	}
	if !strings.Contains(got, "InitX()") {
		t.Fatalf("应当保留原始调用以顺利执行: %s", got)
	}
}

func TestJoinPrintCodeSkipsVarDefinition(t *testing.T) {
	c := &Coder{}
	input := `package main

func main() {
    var name string// :INPUT
}
`

	got, err := c.JoinPrintCode(input)
	if err != nil {
		t.Fatalf("JoinPrintCode 返回错误: %v", err)
	}

	if !strings.Contains(got, "var name string") {
		t.Fatalf("应保留 var 定义: %s", got)
	}
	if strings.Contains(got, "fmt.Println(") {
		t.Fatalf("var 定义不应被 fmt.Println 包装: %s", got)
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
	originalCode := `package main

func main() {
	fmt.Println(test)
}
`

	formattedFile := `package main

import "fmt"

func main() {
	fmt.Println(test)
}
`

	dir := t.TempDir()
	filePath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(filePath, []byte(formattedFile), 0o644); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	c := &Coder{
		VarNames:    []string{"test", "keep"},
		FuncCodeMap: map[string]string{"test": "func() string { return \"wxnacy\" }", "keep": "func() {}"},
	}

	runErr := errors.New(fmt.Sprintf(`# command-line-arguments
%s:6:14: undefined: test`, filePath))
	runOut := "output"
	out, err := c.AfterRunCode(originalCode, runOut, runErr)
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
	originalCode := `package main

func main() {
	a := 1
	a := 2
}
`

	formattedFile := `package main

func main() {
	a := 1
	a := 2
}
`

	dir := t.TempDir()
	filePath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(filePath, []byte(formattedFile), 0o644); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	c := &Coder{
		VarNames: []string{"a"},
	}

	runErr := errors.New(fmt.Sprintf(`# command-line-arguments
%s:5:5: no new variables on left side of :=`, filePath))
	out, err := c.AfterRunCode(originalCode, "out", runErr)
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

	out, err := WriteAndRunCode(mainCode, codePath)
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
