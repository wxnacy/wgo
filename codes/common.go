package codes

import (

)

func GetKeywords() []string{
    return []string{
        "var", "bool", "int", "int8", "int16", "int32", "int64", "byte", "string",
        "float64", "float32", "uint", "uint8", "uint16", "uint32", "uint64",
        "uintptr", "rune", "complex64", "complex128", "import", "package", "main",
        "func", "return", "make", "for", "range", "if",
    }
}

