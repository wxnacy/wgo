package utils

import (
	"testing"
)

// 测试用的代码文本（包含多种函数类型）
const testCode = `
package demo

import "fmt"

// 1. 无返回值的命名函数
func noReturnFunc() {
	fmt.Println("无返回值命名函数")
}

// 2. 单返回值的命名函数
func singleReturnFunc() int {
	return 100
}

// 3. 多返回值的命名函数
func multiReturnFunc() (string, error) {
	return "", nil
}

// 4. 带参数的命名函数（验证参数不影响返回值判断）
func withParamsFunc(a int, b string) bool {
	return a > 0 && b != ""
}

func main() {
	// 5. 无返回值的匿名函数（绑定到变量）
	anonNoReturn := func() {
		fmt.Println("无返回值匿名函数")
	}

	// 6. 单返回值的匿名函数（绑定到变量）
	anonSingleReturn := func(x, y int) int {
		return x + y
	}

	// 7. 多返回值的匿名函数（绑定到变量）
	anonMultiReturn := func() (bool, error) {
		return true, nil
	}

	// 8. 未绑定变量的匿名函数（作为参数，本工具无法识别）
	callFunc(func() string {
		return "临时匿名函数"
	})

	// 9. 空白标识符绑定的匿名函数（无法通过名称查询）
	_ = func() int {
		return 999
	}
}

// 辅助函数：用于接收匿名函数参数
func callFunc(f func() string) {
	fmt.Println(f())
}
`

// 测试用例结构体
type testCase struct {
	name        string
	expected    bool
	expectError bool
}

// TestHasFunctionReturnByCode 单元测试主函数
func TestHasFunctionReturnByCode(t *testing.T) {
	// 定义所有测试用例
	testCases := []testCase{
		// 命名函数测试
		{"noReturnFunc", false, false},
		{"singleReturnFunc", true, false},
		{"multiReturnFunc", true, false},
		{"withParamsFunc", true, false},

		// 匿名函数测试（绑定到变量）
		{"anonNoReturn", false, false},
		{"anonSingleReturn", true, false},
		{"anonMultiReturn", true, false},

		// 边缘情况测试
		{"callFunc", true, false},
		{"undefinedFunc", false, true},
		{"_", false, true},
	}

	// 执行测试用例
	for _, tc := range testCases {
		// 使用 t.Run 进行子测试（方便单独运行某个用例）
		t.Run(tc.name, func(t *testing.T) {
			hasReturn, err := HasFunctionReturnByCode(testCode, tc.name)

			// 验证错误情况
			if tc.expectError {
				if err == nil {
					t.Error("预期错误但未发生错误")
				}
				return // 错误用例无需继续检查返回值
			}

			// 非错误用例必须无错误
			if err != nil {
				t.Errorf("意外错误: %v", err)
				return
			}

			// 验证返回值是否符合预期
			if hasReturn != tc.expected {
				t.Errorf("返回值错误: 预期 %v, 实际 %v", tc.expected, hasReturn)
			}
		})
	}
}
