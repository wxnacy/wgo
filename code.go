package main

import (
    "fmt"
    "os"
    "time"
    "github.com/wxnacy/wgo/file"
    "github.com/wxnacy/wgo/color"
    "github.com/wxnacy/wgo/util"
    "os/exec"
    "bytes"
    "strings"
    "errors"
)

var tempdir string
var tempfile string
var code *Code

type Code struct {
    imports []string
    mainFunc []string
    variables []string
    lastInput string
    lastInputUse bool
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
        code.imports = append(code.imports, "fmt", "time")
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

func (this *Code) Input(line string) {
    this.lastInput = line
    this.lastInputUse = false
}

func (this *Code) GetImport() []string {
    return this.imports
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

func (this *Code) input() {
    if strings.HasPrefix(this.lastInput, "import") {
        impt, ok := parseImport(this.lastInput)
        if ok {
            hasImport := util.ArrayContains(this.imports, impt)
            if hasImport == -1 {
                this.imports = append(this.imports, impt)
            }
        }
    } else if strings.HasPrefix(this.lastInput, "fmt.") {

    } else if strings.Contains(this.lastInput, "=") {
        if has := util.ArrayContains(this.mainFunc, this.lastInput); has == -1 {
            this.mainFunc = append(this.mainFunc, this.lastInput)
        }
    }

    this.variables = parseCodeVars(this.mainFunc)

    this.resetInput()
}

func (this *Code) mainFormat() {
    var mains = make([]string, 0)

    if util.ArrayContains(this.variables, this.lastInput) > -1 {
        this.lastInput = fmt.Sprintf("fmt.Println(%s)", this.lastInput)
    }

    mains = this.mainFunc
    has := util.ArrayContains(this.mainFunc, this.lastInput)
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
            variable := strings.Split(m, "=")[0]
            variable = strings.Trim(variable, ":")
            vars := strings.Split(variable, ",")
            for _, v := range vars {
                v = strings.Trim(v, " ")
                varList = append(varList, v)
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
            if strings.Contains(mainString, d + ".") {
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
    return strings.Join(this.codes, "\n")
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
        // Logger().Debug(out.String())
        fmt.Println(out.String())
        return out.String(), nil
    }
    errStr := outErr.String()
    if errStr != "" {
        fmt.Println(color.Red(errStr))
        this.resetInput()
        return "", errors.New(outErr.String())
    } else {
        this.input()
        // Logger().Debug(out.String())
        return "", nil
    }
}

func tempDir() string {
    if tempdir == "" {
        tempdir = fmt.Sprintf("%s%s-%d/", os.TempDir(), "wgo", time.Now().Unix())
        tempdir = fmt.Sprintf("%s%s-%d/", os.TempDir(), "wgo", 0)
    }
    return tempdir
}

func tempFile() string {
    if tempfile == "" {
        tempfile = fmt.Sprintf("%swgo_temp.go", tempDir())
    }
    return tempfile
}

func initTempDir() {
    if !file.Exists(tempDir()) {
        os.Mkdir(tempDir(), 0700)
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


// func Fmt() {
    // fmt.Println("Hello World ")
    // tmpdir := os.Getenv("TMPDIR")
    // fmt.Println(tmpdir)
    // os.Mkdir("test", 0700)
    // fmt.Println(os.TempDir())
    // // os.Create("test/ss")
    // // os.Chmod("test", 0700
    // fmt.Println(time.Now().Unix())
    // fmt.Println(tempDir())
    // fmt.Println(tempDir())
    // initTempDir()
    // // writeCode("package main;import \"fmt\";func main(){fmt.Print(\"hw\")}")
    // writeCode("package main;import \"fmt\";func main(){fmt.sPrintss(\"hw\")}")
    // // err := cmd.Start()
    // // handlerErr(err)
    // // bytes, err := cmd.Output()
    // // handlerErr(err)
    // // fmt.Println("ss", string(bytes))
// }

// func handlerErr(err error) {
    // if err != nil {
        // fmt.Errorf("%s", err)
    // }
// }

// func main() {
    // c := Coder()
    // c.Input("import \"fmt\"")
    // // fmt.Println(c.Format())
    // c.Run()
    // // fmt.Println(c.Format())

    // c.Input("import \"time\"")
    // // c.Print()
    // c.Run()
    // // c.Print()

    // c.Input("fmt.Println(\"Hello World \")")
    // // c.Print()
    // c.Run()
    // // c.Print()
    // c.Input("import os strings")
    // c.Print()
    // c.Run()
    // c.Print()
    // // c.Input("fmt.Println(\"Hello World\")")
    // // c.Run()
    // // fmt.Println(c.Format())

// }
