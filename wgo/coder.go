package wgo

import (
	"fmt"
	"strings"

	"github.com/wxnacy/go-tools"
	"github.com/wxnacy/wgo/commands"
)

const (
	simpleMainFormat = `package main
func main(){
	%s
}
	`
)

func NewCoder() *Coder {
	return &Coder{}
}

type Coder struct {
	DisableRunAll bool

	codeLines []string
	tmpPath   string
}

func (c *Coder) Run() error {
	c.init()
	res, err := commands.Command("go", "run", c.tmpPath)
	if err != nil {
		return err
	}
	fmt.Println(res)
	return nil
}

func (c *Coder) AddCode(code string) *Coder {
	c.codeLines = append(c.codeLines, code)
	return c
}

func (c *Coder) init() error {
	c.tmpPath = GetTempMainFile()
	err := c.joinCode()
	if err != nil {
		return err
	}
	err = c.formatCode()
	if err != nil {
		return err
	}
	return nil
}

// 拼接代码
func (c *Coder) joinCode() error {
	// 对最后一行代码进行结果输出
	lastCode := c.codeLines[len(c.codeLines)-1]
	if !strings.HasPrefix(lastCode, "fmt.Print") {
		lastCode = fmt.Sprintf("fmt.Println(%s)", lastCode)
	}

	var writeCode string
	if c.DisableRunAll {
		writeCode = lastCode
	} else {
		writeCode = strings.Join(c.codeLines[0:len(c.codeLines)-1], ";") + ";" + lastCode
	}
	// _ = []string{"a", "b", "c", "d"}[1:2]
	wholeCode := fmt.Sprintf(simpleMainFormat, writeCode)
	tools.FileWriteWithInterface(c.tmpPath, wholeCode)
	return nil
}

// 格式化代码
func (c *Coder) formatCode() error {
	// TODO: 检查 goimports
	_, err := commands.Command("goimports", "-w", c.tmpPath)
	return err
}

func RunSimpleCode(code string) error {
	coder := &Coder{DisableRunAll: true}
	return coder.AddCode(code).Run()
}
