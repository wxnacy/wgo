package commands

import (
    "os/exec"
    "bytes"
    "strings"
)

// Determine if a command exists
func HasCommand(name string) (path string, flag bool) {
    flag = false
    c := exec.Command("command", "-v", name)
    var out bytes.Buffer
    c.Stdout = &out
    c.Run()
    var outStr = out.String()
    if outStr != "" {
        outStr = strings.Trim(outStr, "\n")
        path = outStr
        flag = true
    }
    return
}
