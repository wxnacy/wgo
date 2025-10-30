package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wxnacy/go-tools"
	"github.com/wxnacy/wgo/internal/handler"
)

var runCmd = &cobra.Command{
	Use: "run",
	Run: func(cmd *cobra.Command, args []string) {
		var out string
		var err error
		code := args[0]
		if tools.FileExists(code) {
			out, err = handler.RunCode(code)
		} else {
			out, err = handler.GetCoder().InputAndRun(code)
		}
		if err != nil {
			// 功能需求:
			// - 将 err 使用红色字体打印
			fmt.Fprintf(os.Stderr, "\033[31m%v\033[0m\n", err)
		} else {
			fmt.Println(out)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
