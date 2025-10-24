package cli

import (
	"github.com/spf13/cobra"
	"github.com/wxnacy/wgo/internal/handler"
)

var testCmd = &cobra.Command{
	Use: "test",
	Run: func(cmd *cobra.Command, args []string) {
		handler.InsertCodeAndRun("a := time.Now()")
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
