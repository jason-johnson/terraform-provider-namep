// The following directive is necessary to make the package coherent:

// +build ignore

// This program generates model_locations_generated.go. It can be invoked by running
// go generate

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"text/template"
	"time"
)

type LocationStructure struct {
	RegionName     string `json:"region"`
	DefinitionName string `json:"name"`
	// Name used in Azure
	AzureName    string `json:"azName"`
	ShortName    string `json:"short_name_1"`
	AltShortName string `json:"short_name_2,omitempty"`
}

type templateData struct {
	LocationStructures []LocationStructure
	GeneratedTime      time.Time
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Panicln("No directory found")
	}
	fmt.Println()

	templateBytes, err := ioutil.ReadFile(path.Join(wd, "tools/templates/locationsModel.tmpl"))
	if err != nil {
		log.Fatal(err)
	}

	templateText := string(templateBytes)

	parsedTemplate, err := template.New("templates").Parse(templateText)
	if err != nil {
		log.Fatal(err)
	}

	sourceDefinitions, err := ioutil.ReadFile(path.Join(wd, "tools/data/locationDefinitions.json"))
	if err != nil {
		log.Fatal(err)
	}

	var data []LocationStructure
	err = json.Unmarshal(sourceDefinitions, &data)
	if err != nil {
		log.Fatal(err)
	}

	sort.SliceStable(data, func(i, j int) bool {
		return data[i].DefinitionName < data[j].DefinitionName
	})

	modelsFile, err := os.OpenFile(path.Join(wd, "internal/provider/model_locations_generated.go"), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
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
