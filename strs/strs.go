package strs

import (
    "strings"
)

func IndexLast(s, substr string) int {
    sarr := strings.Split(s, "")
    for i := len(sarr) - 1; i >= 0 ; i-- {
        if sarr[i] == substr {
            return i
        }
    }
    return -1
}
