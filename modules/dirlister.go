package modules

import (
    "fmt"
    "os"
)

func Dirlister(args ...interface{}) {
    fmt.Println("[*] In dirlister module.")
    files, err := os.ReadDir(".")
    if err != nil {
        fmt.Println(err)
    }
    for _, file := range files {
        fmt.Println(file.Name())
    }
}