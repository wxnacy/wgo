package arrays

import (
)

// EqualsInt returns a bool value indicating whether the int64[] arr1 and arr2 are equal
func EqualsInt(arr1, arr2 []int64) bool {
    if len(arr1) != len(arr2) {
        return false
    }
    if (arr1 == nil) != (arr2 == nil) {
        return false
    }
    arr2 = arr2[:len(arr1)]
    for i, v := range arr1 {
        if v != arr2[i] {
            return false
        }
    }
    return true
}

// EqualsString returns a bool value indicating whether the string[] arr1 and arr2 are equal
func EqualsString(arr1, arr2 []string) bool {
    if len(arr1) != len(arr2) {
        return false
    }
    if (arr1 == nil) != (arr2 == nil) {
        return false
    }
    arr2 = arr2[:len(arr1)]
    for i, v := range arr1 {
        if v != arr2[i] {
            return false
        }
    }
    return true
}
