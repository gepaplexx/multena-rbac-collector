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
