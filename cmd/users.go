package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// usersCmd represents the users command
var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage your application users.",
	Long: fmt.Sprintf(`
%s

You may list, add, and delete users for your application.
	`,
		clbold("Manage your application users"),
	),

}

func init() {
	rootCmd.AddCommand(usersCmd)
}
