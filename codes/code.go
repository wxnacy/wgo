package codes

import (

    // "github.com/"
    "strings"
    // "github.com/wxnacy/wgo/strs"
    "github.com/wxnacy/wgo/arrays"
    // "fmt"
)

// var VAR_FILTERS = map[string]int{
    // "var": 0,
    // "bool": 0,
    // "int": 0,
    // "int8": 0,
    // "int16": 0,
    // "int32": 0,
    // "int64": 0,
    // "byte": 0,
    // "string": 0,
    // "float64": 0,
    // "float32": 0,
    // "uint": 0,
    // "uint8": 0,
    // "uint16": 0,
    // "uint32": 0,
    // "uint64": 0,
    // "uintptr": 0,
    // "rune": 0,
    // "complex64": 0,
    // "complex128": 0,
    // "": 0,
// }

var VAR_FILTERS = []string{
    "var", "bool", "int", "int8", "int16", "int32", "int64", "byte", "string",
    "float64", "float32", "uint", "uint8", "uint16", "uint32", "uint64",
    "uintptr", "rune", "complex64", "complex128", "",
}

// 从字符串中解析变量名集合
func ParseVarnamesFromString(s string) []string {
    var vars = make([]string, 0)
    index := strings.Index(s, "=")
    if index > -1 {
        s = s[0:index]
    }
    s = strings.TrimRight(s, ":")

    arr := strings.Split(s, "")
    for _, d := range arr {
        if d == "." {
            return vars
        }
    }

    s = strings.Replace(s, ",", " ", -1)

    names := strings.Split(s, " ")
    for _, d := range names {
        v := strings.Trim(d, " ")
        // _, exists := VAR_FILTERS[v]

        if arrays.ContainsString(VAR_FILTERS, v) == -1  {
            vars = append(vars, v)
        }
    }

    return vars
}

// 从数组冲解析变量名集合
func ParseVarnamesFromArray(arr []string) []string {
    var res = make([]string, 0)
    var varMap = make(map[string]int, 0)

    for _, d := range arr {
        vars := ParseVarnamesFromString(d)
        for _, v := range vars {
            _, ok := varMap[v]
            if !ok {
                res = append(res, v)
                varMap[v] = 0
            }
        }
    }
    return res
}
