package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"os/signal"
	"syscall"
)

// main This function is the entry point of the application. It sets up a mechanism to handle the SIGINT (Ctrl+C) and
// SIGTERM system signals. If such a signal is received, it calls the mapsToYaml function to save the collected data
// to a YAML file, and then exits the program with a status code of 1. After setting up this mechanism, it calls the
// collectAll function to collect the RBAC data from the Kubernetes cluster. If this function call results in an error,
// it prints the error and terminates the program. If the function call succeeds, it then calls the mapsToYaml function
// to save the collected data to a YAML file.
func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		mapsToYaml()
		os.Exit(1)
	}()
	_, _, err := collectAll()
	if err != nil {
		fmt.Println(err)
	}
	mapsToYaml()
}

// mapsToYaml This function converts the userAllowedNamespaces and groupAllowedNamespaces maps into a YAML format and
// writes the YAML-formatted data into a file named labels.yaml. If the YAML conversion or file writing operations
// result in an error, the function will panic and cause the program to exit. The fs.ModePerm mode is used for the file,
// which gives read, write, and execute permissions to all.
func mapsToYaml() {
	tenants := map[string]map[string][]string{"users": userAllowedNamespaces, "groups": groupAllowedNamespaces}
	yamlBytes, err := yaml.Marshal(tenants)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("labels.yaml", yamlBytes, fs.ModePerm)
	if err != nil {
		panic(err)
	}
}
