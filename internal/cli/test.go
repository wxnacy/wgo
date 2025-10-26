package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/wxnacy/wgo/internal/handler"
)

var testCmd = &cobra.Command{
	Use: "test",
	Run: func(cmd *cobra.Command, args []string) {
		os.Remove(handler.GetMainFile())
		InsertCodeAndRun := handler.GetCoder().InsertCodeAndRun
		var out string
		action := args[0]
		switch action {
		case "1":
			InsertCodeAndRun("a := time.Now()")
		case "2":
			out = InsertCodeAndRun("a := time.Now()")
			fmt.Println(out)
			out = InsertCodeAndRun("time.Now()")
			fmt.Println(out)
			out = InsertCodeAndRun("a := time.Now()")
			fmt.Println(out)
			out = InsertCodeAndRun("a += 1")
			fmt.Println(out)
			out = InsertCodeAndRun("time.Now()")
			fmt.Println(out)
		case "3":
			handler.SerializeVar("var-a", time.Now())
			handler.GetCoder().VarNames = []string{
				"a",
			}
			out := handler.GetCoder().InsertOrJoinCode("fmt.Println(a)")
			fmt.Println(out)
			codePath := handler.GetMainFile()
			handler.WriteCode(out, handler.GetMainFile())
			// 运行 goimports
			if _, err := handler.Command("goimports", "-w", codePath); err != nil {
				logger.Errorf("goimports failed: %v", err)
			}
			// 运行代码
			out, err := handler.Command(
				"go",
				"run",
				codePath,
				filepath.Join(handler.GetMainDir(), "builtin_func.go"),
				filepath.Join(handler.GetMainDir(), "request.go"),
			)
			fmt.Println(out)
			fmt.Println(err)
		case "4":
			code := `
package main

import "time"

func main() {
	t, _ := _Deserialize[time.Time]("var-t")
	t
	_Serialize("var-t", t)

}
			`
			code, err := handler.GetCoder().ProcessCode(code)
			fmt.Println(code)
			fmt.Println(err)

		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
