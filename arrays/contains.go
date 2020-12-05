package arrays

import (
	"reflect"
)

// Contains Returns if or not the val in array
func Contains(array interface{}, val interface{}) bool {
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		{
			s := reflect.ValueOf(array)
			for i := 0; i < s.Len(); i++ {
				if reflect.DeepEqual(val, s.Index(i).Interface()) {
					return true
				}
			}
		}
	}
	return false
}

// ContainsString Returns if or not the string val in array
func ContainsString(array []string, val string) bool {
	for i := 0; i < len(array); i++ {
		if array[i] == val {
			return true
		}
	}
	return false
}

// ContainsInt Returns if or not the int64 val in array
func ContainsInt(array []int64, val int64) bool {
	for i := 0; i < len(array); i++ {
		if array[i] == val {
			return true
		}
	}
	return false
}

// ContainsUint Returns if or not the uint64 val in array
func ContainsUint(array []uint64, val uint64) bool {
	for i := 0; i < len(array); i++ {
		if array[i] == val {
			return true
		}
	}
	return false
}

// ContainsBool Returns if or not the bool val in array
func ContainsBool(array []bool, val bool) bool {
	for i := 0; i < len(array); i++ {
		if array[i] == val {
			return true
		}
	}
	return false
}

// ContainsFloat Returns if or not the float64 val in array
func ContainsFloat(array []float64, val float64) bool {
	for i := 0; i < len(array); i++ {
		if array[i] == val {
			return true
		}
	}
	return false
}

// ContainsComplex Returns if or not the complex128 val in array
func ContainsComplex(array []complex128, val complex128) bool {
	for i := 0; i < len(array); i++ {
		if array[i] == val {
			return true
		}
	}
	return false
}
