package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// usersDeleteCmd removes a user
var usersDeleteCmd = &cobra.Command{
	Use:   "delete <userID>",
	Short: "Delete a user from your application.",
	Long: fmt.Sprintf(`
%s

Permanently removes the user with the given ID from your application.
	`,
		clbold("Delete an application user"),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if !setBackend() {
			return
		}

		if len(args) == 0 {
			printError("Argument missing: userID — please supply a user ID.")
			return
		}

		userID := args[0]

		authToken, ok := getAuthToken()
		if !ok {
			return
		}

		if err := backend.RemoveUser(authToken, userID); err != nil {
			printError("An error occurred: %v", err)
			return
		}

		fmt.Printf("User %s has been deleted.\n", userID)
	},
}

func init() {
	usersCmd.AddCommand(usersDeleteCmd)
}
