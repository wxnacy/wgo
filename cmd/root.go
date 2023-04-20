/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/wxnacy/wgo/wgo"
)

var (
	rootCommand = &RootCommand{}
	globalArg   = &GlobalArg{}
)

type GlobalArg struct {
	IsVerbose bool
}

type RootCommand struct {
	ShowVersion bool // 展示版本信息
	Cmd         string
}

func (r RootCommand) Run() error {
	if r.Cmd != "" {
		return wgo.RunSimpleCode(r.Cmd)
	}
	if r.ShowVersion {
		fmt.Println(wgo.GetVersion())
		return nil
	}
	return r.RunPrompt()
}

// 运行交互终端
func (r RootCommand) RunPrompt() error {
	// 输出完整版本信息
	wholeVer, err := wgo.GetWholeVersion()
	if err != nil {
		return err
	}
	fmt.Println(wholeVer)
	// 运行交互终端
	p := prompt.New(
		r.Executor,
		r.Completer,
	)
	p.Run()
	return nil
}

// 交互命令执行器
func (r RootCommand) Executor(t string) {
	fmt.Println(t)
}

// 交互界面提示
func (r RootCommand) Completer(d prompt.Document) []prompt.Suggest {
	// d.GetWordBeforeCursor()
	s := make([]prompt.Suggest, 0)
	s = append(s, prompt.Suggest{Text: "name", Description: "des"})

	return prompt.FilterContains(s, d.GetWordBeforeCursor(), true)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wgo",
	Short: "类似 Python 命令的脚本化运行工具",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		return rootCommand.Run()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&globalArg.IsVerbose, "verbose", "v", false, "输出更多详细信息")
	rootCmd.Flags().BoolVarP(&rootCommand.ShowVersion, "version", "V", false, "输出简要版本信息")
	rootCmd.Flags().StringVarP(&rootCommand.Cmd, "cmd", "c", "", "作为字符串传入的程序")
}
