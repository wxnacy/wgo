package arrays

import (
    "reflect"
)

func Contains(array interface{}, val interface{}) (index int) {
    index = -1
    switch reflect.TypeOf(array).Kind() {
        case reflect.Slice: {
            s := reflect.ValueOf(array)
            for i := 0; i < s.Len(); i++ {
                if reflect.DeepEqual(val, s.Index(i).Interface()) {
                    index = i
                    return
                }
            }
        }
    }
    return
}

func StringsContains(array []string, val string) (index int) {
    index = -1
    for i := 0; i < len(array); i++ {
        if array[i] == val {
            index = i
            return
        }
    }
    return
}

func IntsContains(array []int, val int) (index int) {
    index = -1
    for i := 0; i < len(array); i++ {
        if array[i] == val {
            index = i
            return
        }
    }
    return
}

func FloatsContains(array []float64, val float64) (index int) {
    index = -1
    for i := 0; i < len(array); i++ {
        if array[i] == val {
            index = i
            return
        }
    }
    return
}

// []string deduplicate
func StringsDeduplicate(array []string) []string {
    var arr = make([]string, 0)
    var m = make(map[string]bool)
    for _, d := range array {
        _, ok := m[d]
        if !ok {
            m[d] = true
            arr = append(arr, d)
        }
    }
    return arr
}

// []string equal
func StringsEqual(a, b []string) bool {
    if len(a) != len(b) {
        return false
    }
    if (a == nil) != (b == nil) {
        return false
    }
    b = b[:len(a)]
    for i, v := range a {
        if v != b[i] {
            return false
        }
    }
    return true
}

// []int deduplicate
func IntsDeduplicate(array []int) []int {
    var arr = make([]int, 0)
    var m = make(map[int]bool)
    for _, d := range array {
        _, ok := m[d]
        if !ok {
            m[d] = true
            arr = append(arr, d)
        }
    }
    return arr
}

// []int equal
func IntsEqual(a, b []int) bool {
    if len(a) != len(b) {
        return false
    }
    if (a == nil) != (b == nil) {
        return false
    }
    b = b[:len(a)]
    for i, v := range a {
        if v != b[i] {
            return false
        }
    }
    return true
}
