package main

import (
     "testing"
     "fmt"
)


func TestComplete(t *testing.T) {
    prompts := Complete("fmt.e")
    fmt.Print(prompts)

}
