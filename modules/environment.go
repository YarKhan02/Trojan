package modules

import (
	"fmt"
	"os"
)

func Environment(args ...interface{}) {
	fmt.Println("[*] In environment module.")
	for _, env := range os.Environ() { 
        fmt.Println(env) 
    }
}