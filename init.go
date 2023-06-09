package main

import (
	"flag"
	"fmt"
	"github.com/openshift/client-go/user/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

// Declaring global variables for Kubernetes Clientset and UserClientset from OpenShift.
var (
	clientset     *kubernetes.Clientset
	userClientset *versioned.Clientset
)

// init is a special function in Go that gets called upon the package initialization.
// This function handles the initialization and configuration of the Kubernetes and OpenShift clientsets.
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
		return
	}

	clientset, err = kubernetes.NewForConfig(Config)
	if err != nil {
		fmt.Println(err)
		return
	}

	userClientset, err = versioned.NewForConfig(Config)
	if err != nil {
		fmt.Println(err)
		return
	}
}
