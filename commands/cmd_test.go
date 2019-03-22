package commands

import (
    "fmt"
    "testing"
)

func TestHasCommand(t *testing.T) {
    path, ok := HasCommand("wc")
    fmt.Println("Hello World")
    if !ok || path != "/usr/bin/wc" {
        t.Errorf("go path is %s", path)
    }
}
