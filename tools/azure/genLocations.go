// The following directive is necessary to make the package coherent:

//go:build ignore
// +build ignore

// This program generates model_locations_generated.go. It can be invoked by running
// go generate

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"terraform-provider-namep/internal/cloud/azure"
	"text/template"
	"time"
)

type templateData struct {
	LocationStructures []azure.LocationRecord
	GeneratedTime      time.Time
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Panicln("No directory found")
	}
	fmt.Println()

	templateBytes, err := os.ReadFile(path.Join(wd, "tools/azure/templates/locationsModel.tmpl"))
	if err != nil {
		log.Fatal(err)
	}

	templateText := string(templateBytes)

	parsedTemplate, err := template.New("templates").Parse(templateText)
	if err != nil {
		log.Fatal(err)
	}

	sourceDefinitions, err := os.ReadFile(path.Join(wd, "tools/azure/data/locationDefinitions.json"))
	if err != nil {
		log.Fatal(err)
	}

	var data []azure.LocationRecord
	err = json.Unmarshal(sourceDefinitions, &data)
	if err != nil {
		log.Fatal(err)
	}

	sort.SliceStable(data, func(i, j int) bool {
		return data[i].DefinitionName < data[j].DefinitionName
	})

	modelsFile, err := os.OpenFile(path.Join(wd, "internal/cloud/azure/model_locations_generated.go"), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	err = parsedTemplate.Execute(modelsFile, templateData{
		GeneratedTime:      time.Now(),
		LocationStructures: data,
	})

	if err != nil {
		log.Fatalf("execution failed: %s", err)
	}
	log.Println("File generated")
}
