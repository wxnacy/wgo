package arrays

import (
     "testing"
)

func TestContains(t *testing.T) {
    var arr = []string{"wxnacy", "winn"}
    var s = "wxnacy"
    i := Contains(arr, s)
    if i != 0 {
        t.Error(i)
    }
}

func TestStringsContains(t *testing.T) {
    var arr = []string{"wxnacy", "winn"}
    var s = "wxnacy"
    i := StringsContains(arr, s)
    if i != 0 {
        t.Error(i)
    }
}

func TestIntsContains(t *testing.T) {
    var arr = []int64{1, 3, 4, 8, 12, 4, 9}
    var s = 12
    i := ContainsInt(arr, int64(s))
    if i != 4 {
        t.Error(i)
    }
    i = Contains(arr, int64(s))
    if i != 4 {
        t.Error(i)
    }
}

func TestContainsFloat64(t *testing.T) {
    var arr = []float64{1.2, 3.4, 5.6}
    var s = 3.4
    i := ContainsFloat(arr, s)
    if i != 1 {
        t.Error(i)
    }
}

// func TestContainsFloat32(t *testing.T) {
    // var arr = []float32{1.2, 3.4, 5.6}
    // var s float32
    // s = 3.4
    // i := ContainsFloat32(arr, s)
    // if i != 1 {
        // t.Error(i)
    // }
// }

func TestStringsDeduplicate(t *testing.T) {
    var arr = []string{"a", "b", "c", "a", "c"}
    var n = StringsDeduplicate(arr)
    if !StringsEqual(n, []string{"a", "b", "c"}) {
        t.Error("n is [a, b, c]")
    }
}

func TestIntsDeduplicate(t *testing.T) {
    var arr = []int{1, 2, 3, 2, 1}
    var n = IntsDeduplicate(arr)
    if !IntsEqual(n, []int{1, 2, 3}) {
        t.Error("n is [1, 2, 3]")
    }
}

func BenchmarkContains(b *testing.B) {
    sa := []string{"q", "w", "e", "r", "t"}
    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        Contains(sa, "r")
    }
}

func BenchmarkStringsContains(b *testing.B) {
    sa := []string{"q", "w", "e", "r", "t"}
    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        StringsContains(sa, "r")
    }
}


