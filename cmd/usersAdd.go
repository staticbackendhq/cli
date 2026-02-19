package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// usersAddCmd adds a new user
var usersAddCmd = &cobra.Command{
	Use:   "add <email> <password>",
	Short: "Add a new user to your application.",
	Long: fmt.Sprintf(`
%s

Creates a new user account with the given email and password.
	`,
		clbold("Add a new application user"),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if !setBackend() {
			return
		}

		if len(args) < 2 {
			printError("Arguments missing: please supply an email and password.")
			return
		}

		email := args[0]
		password := args[1]

		authToken, ok := getAuthToken()
		if !ok {
			return
		}

		user, err := backend.AddUser(authToken, email, password)
		if err != nil {
			printError("An error occurred: %v", err)
			return
		}

		fmt.Printf("User created: %s | %s\n", user.UserID, user.Email)
	},
}

func init() {
	usersCmd.AddCommand(usersAddCmd)
}
