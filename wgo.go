package main

import (
	"github.com/c-bata/go-prompt"
	"github.com/wxnacy/wgo/commands"

	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// 输入提示
func completer(d prompt.Document) []prompt.Suggest {
	line := d.GetWordBeforeCursor()
	var s = make([]prompt.Suggest, 0)
	// filterString := d.GetWordBeforeCursor()
	var prompts = make([]Prompt, 0)

	if strings.Contains(line, ".") {
		_, ok := commands.HasCommand("gocode")
		if ok {
			prompts = Complete(line)
			// for _, d := range prompts {
			// Logger().Debug(d)

			// }
		}
		prefix := line[0 : strings.Index(line, ".")+1]

		for _, p := range prompts {
			s = append(s, prompt.Suggest{
				Text: prefix + p.Name, Description: p.Class + " " + p.Type,
			})
		}
	} else {

		prompts = GetPromptBySpace()
		for _, p := range prompts {
			s = append(s, prompt.Suggest{
				Text: p.Name, Description: p.Class,
			})
		}

	}
	// Logger().Debugf("s %v", s)
	return prompt.FilterContains(s, d.GetWordBeforeCursor(), true)
}

// 执行
func executor(t string) {
	b := time.Now()
	cmd := strings.Split(t, " ")[0]

	switch cmd {
	case "exit":
		{
			os.Exit(0)
		}
	default:
		{
			Coder().Input(t)
			Coder().Run()

		}
	}
	Logger().Debugf("once command time used: %v", time.Since(b))
	return
}

var LivePrefixState struct {
	LivePrefix string
	IsEnable   bool
}

func changeLivePrefix() (string, bool) {
	return LivePrefixState.LivePrefix, LivePrefixState.IsEnable
}

var VER = `%s
Wgo version %s
Copyright (C) 2019 wxnacy
`

var args []string

func initArgs() {
	flag.Parse()
	args = flag.Args()
	// fmt.Println(args)
	// fmt.Println(os.Args)
}

func commandArgs() {
	if len(args) > 0 {
		arg := args[0]
		switch arg {
		case "version":
			{
				fmt.Println(VERSION)
				os.Exit(0)
			}
		case "logs":
			{
				// t, _ := tail.TailFile(tempDir() + "wgo.log", tail.Config{Follow: true})
				// for {

				// for line := range t.Lines {
				// fmt.Println(line.Text)
				// }
				// }
			}
		}
	}
}

func oldMain() {
	initArgs()
	commandArgs()
	initLogger()
	goVer, _ := commands.Command("go", "version")
	fmt.Println(fmt.Sprintf(VER, goVer, VERSION))
	if len(args) == 0 {

		p := prompt.New(
			executor,
			completer,
			prompt.OptionPrefix(">>> "),
			prompt.OptionLivePrefix(changeLivePrefix),
		)
		p.Run()
	}
	// destroyTempDir()
}
