package main

import (
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func main() {
	// create a struct matching the structure of ResourceList.FunctionConfig to hold its data
	var config struct {
		Data map[string]string `yaml:"data"`
	}
	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		for i := range items {
			for k, v := range config.Data {
				err := items[i].PipeE(yaml.SetAnnotation(k, v))
				if err != nil {
					return nil, err
				}
			}
		}
		return items, nil
	}
	p := framework.SimpleProcessor{Filter: kio.FilterFunc(fn), Config: &config}
	cmd := command.Build(p, command.StandaloneDisabled, false)
	// Adds a "gen" subcommand to create a Dockerfile for building the function into a container image.
	command.AddGenerateDockerfile(cmd)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
