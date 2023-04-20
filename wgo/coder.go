package wgo

import (
	"fmt"
	"strings"

	"github.com/wxnacy/gotool"
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
	res, err := commands.Command("go", "run", c.tmpPath)
	if err != nil {
		return err
	}
	fmt.Println(res)
	return nil
}

func (c *Coder) AddCode(code string) *Coder {
	if !strings.HasPrefix(code, "fmt.Print") {
		code = fmt.Sprintf("fmt.Println(%s)", code)
	}
	c.codeLines = append(c.codeLines, code)
	c.init()
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
	wholeCode := fmt.Sprintf(simpleMainFormat, strings.Join(c.codeLines, ";"))
	gotool.FileWriteWithInterface(c.tmpPath, wholeCode)
	return nil
}

// 格式化代码
func (c *Coder) formatCode() error {
	// TODO: 检查 goimports
	_, err := commands.Command("goimports", "-w", c.tmpPath)
	return err
}

func RunSimpleCode(code string) error {
	return NewCoder().AddCode(code).Run()
}
