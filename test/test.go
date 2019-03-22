package main

import (
    "fmt"
)


func main() {

    var a interface{}
    var b interface{}

    a = "a"
    b = "b"

    fmt.Println(a.type == b.type)

}
