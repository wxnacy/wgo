package main

import (
    "github.com/c-bata/go-prompt"

    "strings"
    "os"
)


func completer(d prompt.Document) []prompt.Suggest {
    var s = make([]prompt.Suggest, 0)

    for _, d := range Coder().GetImport() {
        s = append(s, prompt.Suggest{Text: d, Description: ""})
    }
    // s := []prompt.Suggest{
        // {Text: "fmt", Description: ""},
        // {Text: "time", Description: ""},
    // }

    return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func executor(t string) {
    cmd := strings.Split(t, " ")[0]

    switch cmd {
        case "exit": {
            os.Exit(0)
        }
        default: {
            Coder().Input(t)
            Coder().Run()

        }
    }
    return
}

var LivePrefixState struct {
    LivePrefix string
    IsEnable   bool
}

func changeLivePrefix() (string, bool) {
    return LivePrefixState.LivePrefix, LivePrefixState.IsEnable
}

func main() {
    p := prompt.New(
        executor,
        completer,
        prompt.OptionPrefix(">>> "),
        prompt.OptionLivePrefix(changeLivePrefix),
    )
    p.Run()
}
