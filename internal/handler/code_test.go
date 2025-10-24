package handler

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"testing"
)

func TestSerializeCodeVarsBasic(t *testing.T) {
	input := `package main

func main() {
	var a = 1
	b := 2
}`

	c := &Coder{}
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

	c := &Coder{}
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

	c := &Coder{}
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
