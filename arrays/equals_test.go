package arrays

import (
     "testing"
)

func TestEqualsString(t *testing.T) {
    var arr1 = []string{"wxnacy", "wen", "go"}
    var arr2 = []string{"wxnacy", "wen", "go"}
    flag := EqualsString(arr1, arr2)
    if !flag {
        t.Errorf("%v is error", flag)
    }
    arr1 = []string{"wxnacy",  "go"}
    arr2 = []string{"wxnacy", "wen", "go"}
    flag = EqualsString(arr1, arr2)
    if flag {
        t.Errorf("%v is error", flag)
    }
}

func TestEqualsInt(t *testing.T) {
    var arr1 = []int64{1, 2, 3}
    var arr2 = []int64{1, 2, 3}
    flag := EqualsInt(arr1, arr2)
    if !flag {
        t.Errorf("%v is error", flag)
    }
    arr1 = []int64{1, 2}
    arr2 = []int64{3, 4}
    flag = EqualsInt(arr1, arr2)
    if flag {
        t.Errorf("%v is error", flag)
    }
}
