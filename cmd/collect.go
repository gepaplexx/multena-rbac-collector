/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/gepaplexx/multena-rbac-collector/pkg/collector"

	"github.com/spf13/cobra"
)

// collectCmd represents the collect command
var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		collector.Collect()
		fmt.Println("collect called")
		fmt.Println(cmd.Flags().GetStringSlice("subjects"))
		fmt.Println(cmd.Flags().GetStringSlice("namespaces"))
		fmt.Println(cmd.Flags().GetString("verb"))
		fmt.Println(cmd.Flags().GetString("resource"))
		fmt.Println(cmd.Flags().GetBool("optimization"))
		fmt.Println(cmd.Flags().GetString("output"))
	},
}

func init() {
	rootCmd.AddCommand(collectCmd)
	collectCmd.Flags().StringSliceP("subjects", "s", nil, "List of subjects to collect RBAC for")
	collectCmd.Flags().StringSliceP("namespaces", "n", nil, "List of namespaces to collect RBAC for")
	collectCmd.Flags().StringP("verb", "v", "get", "Verb to collect RBAC for")
	collectCmd.Flags().StringP("resource", "r", "pods", "Resource to collect RBAC for")
	collectCmd.Flags().Bool("optimization", false, "Optimization to use for RBAC generation (skips internal namespaces)")
	collectCmd.Flags().StringP("output", "o", "", "Output file to write RBAC to")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// collectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// collectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
