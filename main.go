package main

import (
	"flag"
	"fmt"
	"github.com/gepaplexx/multena-rbac-collector/cmd"
	"github.com/openshift/client-go/user/clientset/versioned"
	yaml "gopkg.in/yaml.v3"
	"io/fs"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
)

func main() {
	cmd.Execute()
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

var (
	clientset     *kubernetes.Clientset
	userClientset *versioned.Clientset
)

func init() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	err := error(nil)
	Config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Println(err)
	}

	clientset, err = kubernetes.NewForConfig(Config)
	if err != nil {
		fmt.Println(err)
	}

	userClientset, err = versioned.NewForConfig(Config)
	if err != nil {
		fmt.Println(err)
	}
}
