package commands

import (
    "os/exec"
    "bytes"
    "strings"
    "errors"
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

func Command(name string, args ...string) (string, error) {
    c := exec.Command(name, args...)
    var out bytes.Buffer
    var outErr bytes.Buffer
    c.Stdout = &out
    c.Stderr = &outErr
    err := c.Run()
    if err != nil {
        return "", err
    }
    var outStr = out.String()
    if outStr != "" {
        outStr = strings.Trim(outStr, "\n")
        return outStr, nil
    }
    var outErrStr = outErr.String()
    if outErrStr != "" {
        outErrStr = strings.Trim(outErrStr, "\n")
        return "", errors.New(outErrStr)
    }
    return "", nil
}
