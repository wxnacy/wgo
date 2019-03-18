package util

import (
     "testing"
)

func TestInputRun(t *testing.T) {
    var arr = []string{"wxnacy", "winn"}
    var s = "wxnacy"

    i := ArrayContains(arr, s)
    if i != 0 {
        t.Error(i)
    }
}
