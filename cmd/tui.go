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
		if ParseFlag(cmd.Flags().GetBool("mock-app")) {
			runMock()
			return
		}
		app := view.NewAppController(AppVersion)
		if err := app.Init(); err != nil {
			die("failed initializing app: %v\n", err)
		}
		if err := app.Run(); err != nil {
			die("failed running app: %v\n", err)
		}
	},
}

func runMock() {
	autocompleteTypes := []string{"profile"}
	awsTypes := awsu.NewCloudControlAPI(nil).ListSupportedResourceTypes()
	for _, rt := range awsTypes {
		autocompleteTypes = append(autocompleteTypes, rt.String())
	}
	app := view.NewAppMock(AppName, AppVersion, autocompleteTypes)
	if err := app.Init(); err != nil {
		panic(err)
	}
	if err := app.Run(); err != nil {
		panic(err)
	}
}
func init() {
	rootCmd.AddCommand(tuiCmd)
	tuiCmd.Flags().Bool("mock-app", false, "run the app in mock mode")
}
