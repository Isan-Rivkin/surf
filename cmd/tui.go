/*
Copyright © 2023 Isan Rivkin isanrivkin@gmail.com
*/
package cmd

import (
	"github.com/isan-rivkin/surf/internal/view"
	"github.com/isan-rivkin/surf/lib/awsu"
	"github.com/spf13/cobra"
)

// cloudcontrolCmd represents the cloudcontrol command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Experimental GUI",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		autocompleteTypes := []string{"profile"}
		awsTypes := awsu.NewCloudControlAPI(nil).ListSupportedResourceTypes()
		for _, rt := range awsTypes {
			autocompleteTypes = append(autocompleteTypes, rt.String())
		}
		app := view.NewApp(AppName, AppVersion, autocompleteTypes)
		if err := app.Init(); err != nil {
			panic(err)
		}
		if err := app.Run(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
