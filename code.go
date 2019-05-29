package main

import (
    "fmt"
    "os"
    "github.com/wxnacy/wgo/file"
    "github.com/wxnacy/wgo/color"
    "github.com/wxnacy/wgo/arrays"
    "os/exec"
    "bytes"
    "strings"
    "errors"
)

var tempfile string
var code *Code

type CodeMode uint8

const (
    CodeImport CodeMode = iota
    CodeFunc
    CodeMain
)

type Import struct {
    Name string
    Aliasname string
}

func newImport(input string) Import {
    var name string
    var aliasname string
    if strings.HasPrefix(input, "import") {
        ipt := input[6:]
        ipt = strings.Trim(ipt, " ")
        ipt = strings.Trim(ipt, "\"")
        if strings.Contains(ipt, "\"") {
            na := strings.Split(ipt, "\"")
            aliasname = strings.Trim(na[0], " ")
            name = strings.Trim(na[1], " ")

        } else {
            name = ipt
            aliasname = name
            if strings.Contains(name, "/") {
                names := strings.Split(name, "/")
                aliasname = names[len(names) - 1]
            }
        }

    } else {
        name = input
        aliasname = name
        if strings.Contains(input, "/") {
            names := strings.Split(input, "/")
            aliasname = names[len(names) - 1]
        }
    }
    i := Import{Name:name, Aliasname: aliasname}
    return i
}

type Code struct {
    importMap map[string]Import
    mainFunc []string
    variables []string
    lastInput string
    lastInputUse bool
    lastInputMode CodeMode
    funcs map[string][]string
    codes []string
}

func Coder() *Code {
    if code == nil {
        code = &Code{
            importMap: make(map[string]Import),
            mainFunc: make([]string, 0),
            codes: make([]string, 0),
            funcs: make(map[string][]string),
        }
        code.importMap["fmt"] = newImport("fmt")
        code.importMap["os"] = newImport("os")
        code.importMap["time"] = newImport("time")
        code.importMap["runtime"] = newImport("runtime")
    }
    return code
}

func (this *Code) resetInput() {
    this.lastInput = ""
    this.lastInputUse = false
}

func (this *Code) clear() {
    this.codes = make([]string, 0)
}

func (this *Code) GetImportMap() map[string]Import {
    return this.importMap
}

func (this *Code) GetImports() []string {
    res := make([]string, 0)
    for k, _ := range this.importMap {
        res = append(res, k)
    }
    return res
}

func (this *Code) GetImportNames() []string {
    res := make([]string, 0)
    for _, v := range this.importMap {
        res = append(res, v.Aliasname)
    }
    return res
}

func (this *Code) GetVariables() []string {
    return this.variables
}

func (this *Code) GetMains() []string {
    return this.mainFunc
}

func (this *Code) Input(line string) {
    this.lastInput = line
    this.lastInputUse = false

    if strings.HasPrefix(line, "import") {
        this.lastInputMode = CodeImport
    } else {
        this.lastInputMode = CodeMain
    }

}

// 处理输入命令
func (this *Code) input() {
    switch this.lastInputMode {
        case CodeMain: {
            if arrays.StringsContains(this.mainFunc, this.lastInput) == -1 {
                this.mainFunc = append(this.mainFunc, this.lastInput)
            }
        }
        case CodeImport: {
            this.inputImport(this.lastInput)
        }
    }
    this.variables = parseCodeVars(this.mainFunc)
    this.resetInput()
}

// 输入 import
func (this *Code) inputImport(input string) {
    impot := newImport(input)
    this.importMap[impot.Name] = impot
}

func (this *Code) makePrintCode(input string) string {
    return fmt.Sprintf(
        "%s.Println(%s)", this.importMap["fmt"].Aliasname, input,
    )
}

func (this *Code) mainFormat() []string {
    var mains = make([]string, 0)
    var codes = make([]string, 0)

    if arrays.StringsContains(this.variables, this.lastInput) > -1 {
        this.lastInput = this.makePrintCode(this.lastInput)
    }
    if strings.Count(this.lastInput, ".") == 1 {
        index := strings.Index(this.lastInput, ".")
        imptName := this.lastInput[0:index]
        I, ok := this.importMap[imptName]

        // if (ok && I.Name != "fmt") ||
        // strings.HasPrefix(this.lastInput, this.importMap["fmt"].Aliasname + ".Sp") {
        if (ok && I.Name != "fmt") {
            this.lastInput = this.makePrintCode(this.lastInput)
        }
    }

    mains = this.mainFunc
    has := arrays.StringsContains(this.mainFunc, this.lastInput)
    if !strings.HasPrefix(this.lastInput, "import") && has == -1{
        mains = append(mains, this.lastInput)
    }
    if len(mains) == 0 {
        return codes
    }

    for _, m := range mains {
        codes = append(codes, m)
    }

    varList := parseCodeVars(mains)

    for _, v := range varList {
        codes = append(codes, "_ = " + v)
    }
    return codes
}

// 解析代码中的变量
func parseCodeVars(codes []string) []string {
    var varList = make([]string, 0)
    for _, m := range codes {
        if strings.Contains(m, "=") {
            if strings.HasPrefix(m, "var ") {
                m = m[4:]
            }
            variable := strings.Split(m, "=")[0]
            variable = strings.Trim(variable, ":")
            vars := strings.Split(variable, ",")
            for _, v := range vars {
                v = strings.Trim(v, " ")
                if arrays.StringsContains(varList, v) == -1 {
                    varList = append(varList, v)
                }
            }
        } else if strings.HasPrefix(m, "var ") {
            vars := strings.Split(m, " ")
            if len(vars) > 1 && vars[1] != "" && arrays.StringsContains(varList, vars[1]) == -1{
                varList = append(varList, vars[1])
            }
        }
    }
    return varList
}

func (this *Code) Format() string {
    this.clear()
    var isLastInput bool
    mains := this.mainFormat()
    var mainString = strings.Join(mains, "\n")
    mainString = "\n" + mainString
    mainString += this.lastInput
    Logger().Debug(mainString)
    this.codes = append(this.codes, "package main")
    if !isLastInput && strings.HasPrefix(this.lastInput, "import") {
        im := newImport(this.lastInput)
        impt := fmt.Sprintf("import _ \"%s\"", im.Name)
        this.codes = append(this.codes, impt)
        isLastInput = true
        this.lastInputUse = true
    }

    this.codes = append(this.codes, "import (")
    if len(this.importMap) > 0 {
        for k, v := range this.importMap {
            Logger().Debugf("import %v", v)
            ifmt := ""
            importName := v.Aliasname
            if strings.Contains(mainString, "(" + importName + ".") ||
                strings.Contains(mainString, ")" + importName + ".") ||
                strings.Contains(mainString, " " + importName + ".") ||
                strings.Contains(mainString, "\n" + importName + ".") {
                ifmt = "\t%s \"%s\""
                ifmt = fmt.Sprintf(ifmt, v.Aliasname, k)
            } else {
                ifmt = "\t_ \"%s\""
                ifmt = fmt.Sprintf(ifmt, k)

            }
            this.codes = append(this.codes, ifmt)

        }
    }
    this.codes = append(this.codes, ")")
    this.codes = append(this.codes, "func main() {")
    this.codes = append(this.codes, mains...)
    this.codes = append(this.codes, "}")
    res := strings.Join(this.codes, "\n")
    Logger().Debugf("run code\n%s", res)
    return res
}

func (this *Code) Print() {
    fmt.Println(this.Format())
}

func (this *Code) Run() (string, error){
    os.Remove(tempFile())
    writeCode(this.Format())
    cmd := exec.Command("go", "run", tempFile())
    var out bytes.Buffer
    var outErr bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &outErr
    err := cmd.Run()
    if err != nil {
        fmt.Println(err)
    }
    if out.String() != "" {
        fmt.Println(out.String())
        return out.String(), nil
    }
    errStr := outErr.String()
    if errStr != "" {
        errStr = strings.Replace(errStr, tempFile() + ":", "", 2)
        fmt.Println(color.Red(errStr))
        this.resetInput()
        return "", errors.New(outErr.String())
    } else {
        this.input()
        return "", nil
    }
}

func tempFile() string {
    if tempfile == "" {
        tempfile = fmt.Sprintf("%swgo_temp.go", tempDir())
    }
    return tempfile
}

func initTempDir() {
    if !file.Exists(tempDir()) {
        err := os.Mkdir(tempDir(), 0700)
        handlerErr(err)
    }
}

func writeCode(code string) {
    initTempDir()
    f, err := os.OpenFile(tempFile(), os.O_CREATE|os.O_WRONLY, 0600)
    handlerErr(err)
    f.WriteString(code)
    f.Close()
}



