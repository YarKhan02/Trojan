package modules

import (
	"fmt"
	"os"
)

func Environment(args ...interface{}) interface{}{
	fmt.Println("[*] In environment module.")
	envVars := os.Environ()
	for _, env := range envVars { 
        fmt.Println(env) 
    }
	return envVars
}