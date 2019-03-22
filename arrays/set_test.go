package arrays

import (
     "testing"
)

// func TestSetInt(t *testing.T) {
    // s := MakeSet()
    // if !s.IsEmpty() {
        // t.Error("s is empty")
    // }
    // s.Add(2)
    // if !s.Has(2) {
        // t.Error("s has 2")
    // }
    // s.Add(5)
    // s.Add(9)
    // if s.Len() != 3 {
        // t.Error("s len is 3")
    // }

    // // if s.SortList() != []int{2, 5, 9} {
        // // t.Error("s SortList is [2, 3, 9]")
    // // }

    // s.Clear()

    // if !s.IsEmpty() {
        // t.Error("s is empty")
    // }
// }

func TestSetString(t *testing.T) {
    s := MakeSet()
    if !s.IsEmpty() {
        t.Error("s is empty")
    }
    s.Add("b")
    if !s.Has("b") {
        t.Error("s has b")
    }
    s.Add("u")
    s.Add("h")
    if s.Len() != 3 {
        t.Error("s len is 3")
    }

    // if s.SortList() != []int{2, 5, 9} {
        // t.Error("s SortList is [2, 3, 9]")
    // }

    s.Clear()

    if !s.IsEmpty() {
        t.Error("s is empty")
    }
    // t.Log(s.List())
    t.Log(s.Strings())
    // t.Log(s.SortList())
}
