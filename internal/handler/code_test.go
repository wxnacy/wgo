package handler

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
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
