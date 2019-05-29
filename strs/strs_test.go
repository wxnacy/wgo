package strs

import (
    "testing"
    // "github.com/wxnacy/wgo/arrays"
)

func TestIndexLast(t *testing.T) {
    sources := map[string]int{
        "var a int": 5,
        " var": 0,
        "var": -1,
    }

    for i, d := range sources {
        index := IndexLast(i, " ")
        if index != d {
            t.Errorf("%v is error", index)
        }
    }

}

// func BenchmarkParseCodeVars(b *testing.B) {
    // b.ResetTimer()
    // for n := 0; n < b.N; n++ {
        // ParseVarnamesFromArray(VAR_SOURCES)
    // }
// }
