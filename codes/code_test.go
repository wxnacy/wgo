package codes

import (
    "testing"
    "github.com/wxnacy/wgo/arrays"
)

var VAR_SOURCES = []string{
    "var a int",
    "a = 1234",
    "var b = 1",
    "c := 12",
    "fmt.Println(a)",
    "c = 34",
    "b",
    "d:=1",
    "var f, e string",
    "f = 1",
    "e = 1",
    "var g, k=1,2",
    "var i,j=1,2",
    "var o = fmt.Sprintf()",
}
var VARS = []string{"a", "b", "c", "d", "f", "e", "g", "k", "i", "j", "o"}

func TestParseCodeVars(t *testing.T) {
    var vars = ParseVarnamesFromArray(VAR_SOURCES)

    if !arrays.EqualsString(vars, VARS) {
        t.Errorf("%v is error", vars)
    }
}

func BenchmarkParseCodeVars(b *testing.B) {
    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        ParseVarnamesFromArray(VAR_SOURCES)
    }
}
