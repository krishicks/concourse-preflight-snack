package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	"github.com/concourse/atc"
	yaml "github.com/ghodss/yaml"
)

type LayoutType string

const (
	LayoutTypeJob      LayoutType = "job"
	LayoutTypeResource LayoutType = "resource"
)

type Layout struct {
	Name             string             `json:"name" yaml:"name"`
	Tags             []string           `json:"tags" yaml:"tags"`
	Groups           []string           `json:"groups" yaml:"groups"`
	Type             LayoutType         `json:"type" yaml:"type"`
	Template         json.RawMessage    `json:"template" yaml:"template"`
	CompiledTemplate *template.Template `json:"-" yaml:"-"`
}

type Resource struct {
	atc.ResourceConfig
	PPTags []string `json:"pp_tags,omitempty" yaml:"pp_tags,omitempty"`
}

type Config struct {
	PPLayouts     []Layout          `json:"pp_layouts,omitempty" yaml:"pp_layouts,omitempty"`
	Resources     []Resource        `json:"resources" yaml:"resources"`
	ResourceTypes atc.ResourceTypes `json:"resource_types,omitempty" yaml:"resource_types,omitempty"`
	Groups        atc.GroupConfigs  `json:"groups,omitempty" yaml:"groups,omitempty"`
	Jobs          []json.RawMessage `json:"jobs" yaml:"jobs"`
}

func main() {
	pipelineConfigBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("could not read pipeline config: %s", err)
	}

	var config Config
	err = yaml.Unmarshal(pipelineConfigBytes, &config)
	if err != nil {
		log.Fatalf("could not convert yaml to json: %s", err)
	}

	layoutsByTag := map[string][]*Layout{}
	for _, layout := range config.PPLayouts {
		layout := layout
		layout.CompiledTemplate = template.Must(template.New(layout.Name).Parse(string(layout.Template)))

		for _, tag := range layout.Tags {
			layoutsByTag[tag] = append(layoutsByTag[tag], &layout)
		}
	}

	for _, resource := range config.Resources {
		for _, tag := range resource.PPTags {
			if layouts, ok := layoutsByTag[tag]; ok {
				for _, layout := range layouts {
					buf := bytes.NewBuffer([]byte{})
					err = layout.CompiledTemplate.Execute(buf, resource)
					if err != nil {
						log.Fatalf("could not execute template: %s", err)
					}

					bs := buf.Bytes()

					switch layout.Type {
					case LayoutTypeJob:
						config.Jobs = append(config.Jobs, bs)

						var job atc.JobConfig
						err = yaml.Unmarshal(bs, &job)
						if err != nil {
							log.Fatalf("could not unmarshal job template: %s", err)
						}

						if len(layout.Groups) > 0 {
							newGroups := make(atc.GroupConfigs, len(config.Groups))
							for _, group := range layout.Groups {
								for i, cGroup := range config.Groups {
									if cGroup.Name == group {
										cGroup.Jobs = append(cGroup.Jobs, job.Name)
									}
									newGroups[i] = cGroup
								}
								config.Groups = newGroups
							}
						}

					case LayoutTypeResource:
						var resource Resource
						err = yaml.Unmarshal(bs, &resource)
						if err != nil {
							log.Fatalf("could not unmarshal resource template: %s", err)
						}
						config.Resources = append(config.Resources, resource)
					}

				}
			}
		}

		resource.PPTags = nil
	}

	config.PPLayouts = nil

	bs, err := yaml.Marshal(config)
	if err != nil {
		log.Fatalf("failed to marshal yaml: %s", err)
	}

	fmt.Fprint(os.Stdout, string(bs))
}
