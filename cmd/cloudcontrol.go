/*
Copyright © 2023 Isan Rivkin isanrivkin@gmail.com
*/
package cmd

import (
	"github.com/isan-rivkin/surf/lib/awsu"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// cloudcontrolCmd represents the cloudcontrol command
var cloudcontrolCmd = &cobra.Command{
	Use:   "cloudcontrol",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("cloudcontrol called")
		awsu.NewCloudControl(awsu.NewSessionInput("default", "us-east-1"))
	},
}

func init() {
	rootCmd.AddCommand(cloudcontrolCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cloudcontrolCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cloudcontrolCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
