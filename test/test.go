package main

import "fmt"
import  "time"
import (
    "reflect"
)

var ss bool

func main() {
    // fmt.Println("Hello World ")
    // fmt.Println(time.Now())
    time.Now()
    // fmt.Println(ss)
    now := time.Now()
    // fmt.Println(now)
    _  = now
    value := reflect.ValueOf("fmt")
    t := reflect.TypeOf("fmt")
    for i := 0; i < t.NumField(); i++{
        fmt.Println(t.Field(i).Name, value.Field(i).Interface())
    }
}
