package main

import (
     "testing"
)

var out string
var err error

func TestInputRun(t *testing.T) {
    c := Coder()
    c.Input("import \"time\"")
    _, err = c.Run()
    if err != nil {
        t.Error(err)
    }

    c.Input("import \"fmt\"")
    _, err = c.Run()
    if err != nil {
        t.Error(err)
    }

    c.Input("fmt.Println(\"Hello World\")")
    out, err = c.Run()
    if err != nil {
        t.Error(err)
    }
    if out != "Hello World\n" {
        t.Error(out)
    }

}

func TestInputRun2(t *testing.T) {
    c := Coder()

    c.Input("import \"fmt\"")
    _, err = c.Run()
    if err != nil {
        t.Error(err)
    }

    c.Input("s := \"Hello World\"")
    out, err = c.Run()
    if err != nil {
        t.Error(err)
    }

    c.Input("fmt.Println(s)")
    out, err = c.Run()
    if err != nil {
        t.Error(err)
    }
    if out != "Hello World\n" {
        t.Error(out)
    }

}

func TestInputRun3(t *testing.T) {
    c := Coder()
    c.Input("s := \"Hello World\"")
    out, err = c.Run()
    if err != nil {
        t.Error(err)
    }

    c.Input("fmt.Print(s)")
    out, err = c.Run()
    if err != nil {
        t.Error(err)
    }
    if out != "Hello World" {
        t.Error(out)
    }

}
