/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/

package cmd

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gepaplexx/multena-rbac-collector/util"

	"github.com/gepaplexx/multena-rbac-collector/collector"
	progressbar "github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// Make sure to import the necessary packages for your logic
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Collects RBAC permissions and stores them in a ConfigMap",
	Long: `Collects RBAC permissions and stores them in a ConfigMap.
In cluster ConfigMap can be specified with the --cmName and --cmNamespace flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		logCommit()
		initializeKubernetesClient()
		log.Info().Msg("Starting RBAC analyzer...")

		bar := progressbar.NewOptions(8,
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionShowBytes(false),
			progressbar.OptionSetWidth(15),
			progressbar.OptionSetDescription("[cyan][rbac-collector][reset] Collecting Permissions..."),
		)

		start := time.Now()

		roles, err := clientset.RbacV1().Roles(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Error().Err(err).Msg("error getting roles")
			return
		}
		_ = bar.Add(1)

		clusterRoles, err := clientset.RbacV1().ClusterRoles().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Error().Err(err).Msg("error getting cluster roles")
			return
		}
		_ = bar.Add(1)

		rolesWithPerm, clusterRolesWithPerm := collector.GetRoles(*roles, *clusterRoles)
		_ = bar.Add(1)

		roleBindings, err := clientset.RbacV1().RoleBindings(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Error().Err(err).Msg("error getting role bindings")
			return
		}
		_ = bar.Add(1)

		clusterRoleBindings, err := clientset.RbacV1().ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Error().Err(err).Msg("error getting cluster role bindings")
			return
		}
		_ = bar.Add(1)

		permissions := collector.Collect(rolesWithPerm, clusterRolesWithPerm, roleBindings, clusterRoleBindings)
		_ = bar.Add(1)

		yamlBytes, err := yaml.Marshal(permissions)
		if err != nil {
			log.Error().Err(err).Msg("error marshalling permissions")
			return
		}
		_ = bar.Add(1)

		err = os.WriteFile("labels.yaml", yamlBytes, os.ModePerm)
		_ = bar.Add(1)
		if err != nil {
			log.Error().Err(err).Msg("error writing permissions to file")
			return
		}
		log.Info().Msg("Finished collecting permissions")
		if cmName != "" && cmNamespace != "" {
			log.Info().Msg("Updating ConfigMap...")
			err := util.WriteConfigmap(clientset, permissions, util.Config{CMName: cmName, CMNamespace: cmNamespace})
			if err != nil {
				log.Error().Err(err).Msg("Error writing configmap")
				return
			}
			log.Info().TimeDiff("duration", time.Now(), start).Msg("ConfigMap updated")
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
