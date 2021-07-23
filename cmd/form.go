package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// dbCmd represents the db command
var formCmd = &cobra.Command{
	Use:   "form",
	Short: "Manage your posted form data.",
	Long: fmt.Sprintf(`
%s

You can view and delete form submissions. You'll need a rootToken in your 
config file.
	`,
		clbold(clsecondary("Manage forms")),
	),
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(formCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// accountCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
