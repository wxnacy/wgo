package main

import (
    "github.com/c-bata/go-prompt"
    "github.com/wxnacy/wgo/commands"

    "strings"
    "os"
)


func completer(d prompt.Document) []prompt.Suggest {
    line := d.GetWordBeforeCursor()
    var s = make([]prompt.Suggest, 0)
    // filterString := d.GetWordBeforeCursor()
    var prompts = make([]Prompt, 0)

    if strings.Contains(line, ".") {
        _, ok := commands.HasCommand("gocode")
        if ok {
            prompts = Complete(line)
        }
        prefix := line[0:strings.Index(line, ".") + 1]

        for _, p := range prompts {
            s = append(s, prompt.Suggest{
                Text: prefix + p.Name, Description: p.Class + " " + p.Type,
            },)
        }
    } else  {

        prompts = GetPromptBySpace()
        for _, p := range prompts {
            s = append(s, prompt.Suggest{
                Text: p.Name, Description: p.Class,
            },)
        }

    }
    // Logger().Debugf("s %v", s)
    return prompt.FilterContains(s, d.GetWordBeforeCursor(), true)
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
    initLogger()
    p := prompt.New(
        executor,
        completer,
        prompt.OptionPrefix(">>> "),
        prompt.OptionLivePrefix(changeLivePrefix),
    )
    p.Run()
}
