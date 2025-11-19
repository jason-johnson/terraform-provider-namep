// The following directive is necessary to make the package coherent:

//go:build ignore
// +build ignore

// This program generates models_generated.go. It can be invoked by running
// go generate

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"sort"
	"terraform-provider-namep/internal/cloud/azure"
	"text/template"
	"time"
)

type templateData struct {
	ResourceStructures []azure.ResourceStructure
	GeneratedTime      time.Time
	SlugMap            map[string]string
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Panicln("No directory found")
	}
	fmt.Println()
	files, err := os.ReadDir(path.Join(wd, "tools/azure/templates"))
	if err != nil {
		log.Fatal(err)
	}
	var fileNames = make([]string, len(files))
	for i, file := range files {
		fileNames[i] = path.Join(wd, "tools/azure/templates", file.Name())
	}
	parsedTemplate, err := template.New("templates").Funcs(template.FuncMap{
		// Terraform not yet support lookahead in their regex function
		"cleanRegex": func(dirtyString string) string {
			var re = regexp.MustCompile(`(?m)\(\?=.{\d+,\d+}\$\)|\(\?!\.\*--\)`)
			return re.ReplaceAllString(dirtyString, "")
		},
	}).ParseFiles(fileNames...)
	if err != nil {
		log.Fatal(err)
	}

	sourceDefinitions, err := os.ReadFile(path.Join(wd, "tools/azure/data/resourceDefinition.json"))
	if err != nil {
		log.Fatal(err)
	}

	var data []azure.ResourceStructure
	err = json.Unmarshal(sourceDefinitions, &data)
	if err != nil {
		log.Fatal(err)
	}

	sort.SliceStable(data, func(i, j int) bool {
		return data[i].ResourceTypeName < data[j].ResourceTypeName
	})

	slugMap := make(map[string]string)
	for _, res := range data {
		if _, exists := slugMap[res.CafPrefix]; !exists {
			slugMap[res.CafPrefix] = res.ResourceTypeName
		}
	}

	modelsFile, err := os.OpenFile(path.Join(wd, "internal/cloud/azure/models_generated.go"), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	err = parsedTemplate.ExecuteTemplate(modelsFile, "model.tmpl", templateData{
		GeneratedTime:      time.Now(),
		ResourceStructures: data,
		SlugMap:            slugMap,
	})

	if err != nil {
		log.Fatalf("execution failed: %s", err)
	}
	log.Println("File generated")
}
