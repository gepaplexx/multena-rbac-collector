package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
)

func main() {
	uan, gan, err := collectAll()
	users := map[string]map[string][]string{"users": uan}
	groups := map[string]map[string][]string{"groups": gan}

	if err != nil {
		fmt.Println(err)
	}
	err = mapToYAML(users, "users.yaml")
	if err != nil {
		fmt.Println(err)
	}
	err = mapToYAML(groups, "groups.yaml")
	if err != nil {
		fmt.Println(err)
	}

}

func mapToYAML(data map[string]map[string][]string, filename string) error {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	// Write the YAML data to a file
	err = os.WriteFile(filename, yamlBytes, fs.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
