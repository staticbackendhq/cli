package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// funcitonCmd operates on server-side functions
var functionCmd = &cobra.Command{
	Use:   "function",
	Short: "Manage your server-side functions.",
	Long: fmt.Sprintf(`
%s

You can create, update, list and view run history for all your functions.

You'll need a rootToken in your 
config file.
	`,
		clbold(clsecondary("Manage functions")),
	),
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(functionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// accountCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
