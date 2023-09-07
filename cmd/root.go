/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/spf13/cobra"
)

var (
	Commit         string
	cfgFile        string
	kubeconfigPath string
	clientset      *kubernetes.Clientset
	cmName         string
	cmNamespace    string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "multena-rbac-collector",
	Short: "rbac collector collects RBAC data from a Kubernetes cluster and stores it in a ConfigMap",
	Long: `rbac collector collects RBAC data from a Kubernetes cluster and offers two modes:
- run: one shot collection of RBAC data
- serve: continues collection of RBAC data

With the --cmName and --cmNamespace flags you can specify the ConfigMap (inside the cluster) to store the RBAC data in.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.multena-rbac-collector.yaml)")
	rootCmd.PersistentFlags().StringVar(&kubeconfigPath, "kubeconfig", "", "path to the kubeconfig file (default is $HOME/.kube/config for local development)")
	rootCmd.PersistentFlags().StringVar(&cmName, "cmName", "", "in cluster name of the ConfigMap to store the RBAC data")
	rootCmd.PersistentFlags().StringVar(&cmNamespace, "cmNamespace", "", "cluster namespace of the ConfigMap to store the RBAC data")
	rootCmd.MarkFlagsRequiredTogether("cmName", "cmNamespace")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initializeKubernetesClient() {
	var config *rest.Config
	var err error

	config, err = rest.InClusterConfig()
	if err != nil {
		// Not in-cluster, use local kubeconfig
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal().Err(err).Msgf("Could not get user home directory")
		}
		kubeconfig := filepath.Join(home, ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatal().Err(err).Msgf("Could not build Kubernetes config from local kubeconfig")
		}
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msgf("Could not create Kubernetes clientset")
	}
}

func logCommit() {
	log.Info().Msgf("Commit: %s", Commit)
}
