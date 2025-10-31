package utils

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// 存储函数（含匿名函数）的返回值信息：key 是函数名或变量名，value 是是否有返回值
type funcReturnInfo map[string]bool

// HasFunctionReturnByCode 检查代码文本中指定函数（含匿名函数绑定的变量）是否有返回值
// 参数：
//   - code: 待解析的Go代码文本（需包含完整的包声明和函数定义）
//   - funcName: 要检查的函数名称（区分大小写）
//
// 返回：
//   - bool: 函数是否有返回值（true=有，false=无）
//   - error: 错误信息（如代码解析失败、函数未找到等）
func HasFunctionReturnByCode(code, name string) (bool, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "virtual.go", []byte(code), parser.ParseComments)
	if err != nil {
		return false, fmt.Errorf("解析代码失败: %w", err)
	}

	info := make(funcReturnInfo)

	// 遍历AST收集函数信息
	ast.Inspect(file, func(n ast.Node) bool {
		// 处理命名函数
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			hasReturn := funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0
			// 若自身无返回值，但形参中存在“带返回值的函数类型参数”，也视为 true
			if !hasReturn && funcDecl.Type.Params != nil {
				for _, field := range funcDecl.Type.Params.List {
					if ft, ok := field.Type.(*ast.FuncType); ok {
						if ft.Results != nil && len(ft.Results.List) > 0 {
							hasReturn = true
							break
						}
					}
				}
			}
			info[funcDecl.Name.Name] = hasReturn
			return true
		}

		// 处理匿名函数（绑定到变量的情况）
		if assignStmt, ok := n.(*ast.AssignStmt); ok {
			for i, rhs := range assignStmt.Rhs {
				funcLit, ok := rhs.(*ast.FuncLit)
				if !ok {
					continue
				}

				// 处理单个变量绑定
				if ident, ok := assignStmt.Lhs[i].(*ast.Ident); ok {
					if ident.Name == "_" { // 忽略空白标识符
						continue
					}
					hasReturn := funcLit.Type.Results != nil && len(funcLit.Type.Results.List) > 0
					info[ident.Name] = hasReturn
				}
			}
		}

		return true
	})

	if hasReturn, ok := info[name]; ok {
		return hasReturn, nil
	}

	return false, fmt.Errorf("未找到名称为 %q 的函数或绑定匿名函数的变量", name)
}
