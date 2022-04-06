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
		annotationsInterface := metadata["annotations"]
		if annotationsInterface != nil {
			annotations := metadata["annotations"].(map[interface{}]interface{})
			annotations["custom-composition"] = "poc"
		} else {
			metadata["annotations"] = map[interface{}]interface{}{
				"custom-composition": "poc",
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
