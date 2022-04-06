package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

func main() {
	resourceListInput, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	m, err := readResourceList(resourceListInput)
	if err != nil {
		panic(err)
	}

	items := m["items"].([]interface{})
	for _, i := range items {
		item := i.(map[interface{}]interface{})
		metadata := item["metadata"].(map[interface{}]interface{})
		labelsInterface := metadata["labels"]
		if labelsInterface != nil {
			labels := metadata["labels"].(map[interface{}]interface{})
			labels["custom-composition-label"] = "poc-label"
		} else {
			metadata["labels"] = map[interface{}]interface{}{
				"custom-composition-label": "poc-label",
			}
		}
	}
	b, err := yaml.Marshal(m)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func readResourceList(content []byte) (map[interface{}]interface{}, error) {
	m := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(content, m); err != nil {
		return nil, err
	}
	return m, nil
}
