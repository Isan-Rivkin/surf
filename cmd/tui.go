/*
Copyright © 2023 Isan Rivkin isanrivkin@gmail.com
*/
package cmd

import (
	"github.com/isan-rivkin/surf/internal/view"
	"github.com/spf13/cobra"
)

// cloudcontrolCmd represents the cloudcontrol command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Experimental GUI",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		app := view.NewApp(AppName, AppVersion)
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
