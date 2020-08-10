package main

import (
    "os/exec"
    "os"
    "fmt"
    "strings"
    "bytes"
    "strconv"
    "encoding/json"
    "github.com/wxnacy/wgo/codes"
)

func writeCompleteCode(code string) {
    initTempDir()
    f, err := os.OpenFile(tempCompleteFile(), os.O_CREATE|os.O_WRONLY, 0600)
    handlerErr(err)
    f.WriteString(code)
    f.Close()
}

type Prompt struct {
    Class string `json:"class"`         // eg. func
    Package string `json:"package"`     // eg. fmt
    Type string `json:"type"`           // eg. func(format string, a ...interface{}) error
    Name string `json:"name"`           // eg. Errorf
}

func GetPromptBySpace() []Prompt {
    var prompts = make([]Prompt, 0)

    for _, impt := range Coder().GetImportNames() {
        prompts = append(prompts, Prompt{Name: impt, Class: "package"})
    }

    for _, impt := range Coder().GetVariables() {
        prompts = append(prompts, Prompt{Name: impt, Class: "variable"})
    }

    for _, impt := range codes.GetKeywords() {
        prompts = append(prompts, Prompt{Name: impt, Class: "keyword"})
    }

    return prompts
}

var tmpPrompts []Prompt

func GetPrompts() []Prompt {
    return tmpPrompts
}

func clearPrintPrompts() {
    tmpPrompts = make([]Prompt, 0)
}

func Complete(s string) []Prompt {
    var codes = make([]string, 0)
    offset := 0                         // 补全 offset
    c := Coder()
    // imports := c.GetImports()
    p := "package main"
    codes = append(codes, p)
    for k, v := range c.GetImportMap() {
        impt := fmt.Sprintf("import %s \"%s\"", v.Aliasname, k)
        codes = append(codes, impt)
        offset += len(impt) +1
    }
    mains := c.GetMains()
    m := "func main(){"
    codes = append(codes, m)
    for _, d := range mains {
        codes = append(codes, d)
        offset += len(d) + 1
    }
    codes = append(codes, s)
    codes = append(codes, "}")
    writeCompleteCode("")
    writeCompleteCode(strings.Join(codes, "\n"))
    offset += len(p) + len(m) + 2 + len(s)
    // Logger().Debugf("offset %d", offset)

    cmd := exec.Command(
        "gocode", "-in=" + tempCompleteFile(), "-f=json", "autocomplete",
        tempCompleteFile(), strconv.Itoa(offset),
    )
    var out bytes.Buffer
    cmd.Stdout = &out
    cmd.Run()
    cmp := out.String()
    // Logger().Debugf(cmp)
    cmp = cmp[3:len(cmp)-2]
    var prompts []Prompt
    json.Unmarshal([]byte(cmp), &prompts)
    // Logger().Debug(len(tmpPrompts))
    if strings.HasSuffix(s, ".") && len(tmpPrompts) == 0 {
        Logger().Debug("input prompt")
        // tmpPrompts = make([]Prompt, 0)
        tmpPrompts = prompts
        // Logger().Debug(len(tmpPrompts))

    }

    return prompts
}
