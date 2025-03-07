package modules

import (
    "fmt"
    "os"
)

func Dirlister(args ...interface{}) interface{} {
    fmt.Println("[*] In dirlister module.")
    files, err := os.ReadDir(".")
    if err != nil {
        fmt.Println(err)
    }

    var fileList []string
    for _, file := range files {
        fileList = append(fileList, file.Name())
    }
    return fileList
}