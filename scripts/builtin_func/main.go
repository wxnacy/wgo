package main

import "fmt"

func main() {
	test := func() string {
		fmt.Println("wxnacy")
		return "wxnacy"
	}

	_Serialize("func_test", test)

	t, err := DeserializeFromFile[func() string]("/Users/wxnacy/Documents/Projects/wxnacy/wgo/scripts/builtin_func/.wgo/func_test")
	t()
	fmt.Println(err)
}

// const (
// builtinFuncFilePath = "./scripts/builtin_func/builtin_func.go"
// toBuiltinFuncFile   = "./internal/handler/builtin_func.go"
// toBuiltinFuncName   = "BuiltinFuncCode"
// )

// 功能需求:
// - 将 BUILTIN_FUNC_FILE 地址中的内容复制到 TO_BUILTIN_FUNC_FILE 文件的 TO_BUILTIN_FUNC_NAME 字段名中
// - 测试文件变化，如果内容发生变化，要对字段内容进行覆盖
//
// 项目发布前优先调用该方法
// func main() {
// src, err := os.ReadFile(builtinFuncFilePath)
// if err != nil {
// fatalf("读取源文件失败: %v", err)
// }

// target, err := os.ReadFile(toBuiltinFuncFile)
// if err != nil {
// fatalf("读取目标文件失败: %v", err)
// }

// updated, err := buildUpdatedContent(target, src)
// if err != nil {
// fatalf("更新内容失败: %v", err)
// }

// if bytes.Equal(updated, target) {
// fmt.Println("BuiltinFuncCode 无需更新")
// return
// }

// if err := os.WriteFile(toBuiltinFuncFile, updated, 0o644); err != nil {
// fatalf("写入目标文件失败: %v", err)
// }

// fmt.Println("BuiltinFuncCode 已更新")
// }

// func buildUpdatedContent(target, source []byte) ([]byte, error) {
// fset := token.NewFileSet()
// fileNode, err := parser.ParseFile(fset, toBuiltinFuncFile, target, parser.ParseComments)
// if err != nil {
// return nil, fmt.Errorf("解析目标文件失败: %w", err)
// }

// literal := makeStringLiteral(string(source))
// if err := replaceGlobalString(fileNode, toBuiltinFuncName, literal); err != nil {
// return nil, err
// }

// var buf bytes.Buffer
// if err := format.Node(&buf, fset, fileNode); err != nil {
// return nil, fmt.Errorf("格式化代码失败: %w", err)
// }

// return buf.Bytes(), nil
// }

// func replaceGlobalString(fileNode *ast.File, targetName, literal string) error {
// for _, decl := range fileNode.Decls {
// gen, ok := decl.(*ast.GenDecl)
// if !ok || gen.Tok != token.VAR {
// continue
// }

// for _, spec := range gen.Specs {
// valSpec, ok := spec.(*ast.ValueSpec)
// if !ok {
// continue
// }

// for i, name := range valSpec.Names {
// if name.Name != targetName {
// continue
// }

// if len(valSpec.Values) <= i {
// valSpec.Values = append(valSpec.Values, make([]ast.Expr, i-len(valSpec.Values)+1)...)
// }

// valSpec.Values[i] = &ast.BasicLit{Kind: token.STRING, Value: literal}
// return nil
// }
// }
// }

// return errors.New("未找到目标变量 BuiltinFuncCode")
// }

// func makeStringLiteral(content string) string {
// if !strings.Contains(content, "`") {
// return "`" + content + "`"
// }
// return strconv.Quote(content)
// }

// func fatalf(formatStr string, args ...any) {
// fmt.Fprintf(os.Stderr, formatStr+"\n", args...)
// os.Exit(1)
// }
