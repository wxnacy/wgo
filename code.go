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

type Code struct {
    imports []string
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
            imports: make([]string, 0),
            mainFunc: make([]string, 0),
            codes: make([]string, 0),
            funcs: make(map[string][]string),
        }
        code.imports = append(code.imports, "fmt", "time", "os", "strings")
        code.mainFunc = append(code.mainFunc, "fmt.Sprintf(\"%s\", \"Hello World\")")
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

func (this *Code) GetImport() []string {
    return this.imports
}

func (this *Code) GetVariables() []string {
    return this.variables
}

func (this *Code) GetMains() []string {
    return this.mainFunc
}

// 解析 import
func parseImport(impt string) (string, bool) {
    if strings.HasPrefix(impt, "import") {
        ipt := impt[6:]
        ipt = strings.Trim(ipt, " ")
        ipt = strings.Trim(ipt, "\"")
        return ipt, true
    }
    return impt, false
}

func (this *Code) Input(line string) {
    this.lastInput = line
    this.lastInputUse = false

    if strings.HasPrefix(line, "var ") || strings.Contains(line, "=") {
        this.lastInputMode = CodeMain
    } else if strings.HasPrefix(line, "import") {
        this.lastInputMode = CodeImport
    }
}

func (this *Code) input() {
    switch this.lastInputMode {
        case CodeMain: {
            if arrays.StringsContains(this.mainFunc, this.lastInput) == -1 {
                this.mainFunc = append(this.mainFunc, this.lastInput)
            }
        }
        case CodeImport: {
            impt, ok := parseImport(this.lastInput)
            if ok {
                hasImport := arrays.StringsContains(this.imports, impt)
                if hasImport == -1 {
                    this.imports = append(this.imports, impt)
                }
            }
        }
    }


    this.variables = parseCodeVars(this.mainFunc)

    this.resetInput()
}

func (this *Code) mainFormat() {
    var mains = make([]string, 0)

    if arrays.StringsContains(this.variables, this.lastInput) > -1 {
        this.lastInput = fmt.Sprintf("fmt.Println(%s)", this.lastInput)
    }

    mains = this.mainFunc
    has := arrays.StringsContains(this.mainFunc, this.lastInput)
    if ! this.lastInputUse && has == -1{
        mains = append(mains, this.lastInput)
    }
    if len(mains) == 0 {
        return
    }

    for _, m := range mains {
        this.codes = append(this.codes, m)
    }

    varList := parseCodeVars(mains)

    for _, v := range varList {
        this.codes = append(this.codes, "_ = " + v)
    }
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
    var mainString = strings.Join(this.mainFunc, "\n")
    mainString += this.lastInput
    this.codes = append(this.codes, "package main")
    if !isLastInput && strings.HasPrefix(this.lastInput, "import") {
        impt := "import _ " + this.lastInput[6:]
        this.codes = append(this.codes, impt)
        isLastInput = true
        this.lastInputUse = true
    }

    this.codes = append(this.codes, "import (")
    if len(this.imports) > 0 {
        for _, d := range this.imports {
            ifmt := ""
            importName := d
            if strings.Contains(d, "/") {
                ims := strings.Split(d, "/")
                importName = ims[len(ims) - 1]
            }
            if strings.Contains(mainString, importName + ".") {
                ifmt = "\t\"%s\""
            } else {
                ifmt = "\t_ \"%s\""
            }
            this.codes = append(this.codes, fmt.Sprintf(ifmt, d))

        }
    }
    this.codes = append(this.codes, ")")
    this.codes = append(this.codes, "func main() {")
    this.mainFormat()
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
    // this.Print()
    // Logger().Debug(this.Format())
    if err != nil {
        fmt.Println(err)
    }
    if out.String() != "" {
        // Logger().Debug(out.String())
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
        // Logger().Debug(out.String())
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

// func destroyTempDir() {
    // err := os.RemoveAll(tempDir())
    // handlerErr(err)
// }

func writeCode(code string) {
    initTempDir()
    f, err := os.OpenFile(tempFile(), os.O_CREATE|os.O_WRONLY, 0600)
    handlerErr(err)
    f.WriteString(code)
    f.Close()
}



