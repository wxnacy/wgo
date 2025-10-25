package handler

import (
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
