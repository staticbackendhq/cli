package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// usersListCmd lists all users
var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users for your application.",
	Long: fmt.Sprintf(`
%s

Lists all users registered in your application.
	`,
		clbold("List application users"),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if !setBackend() {
			return
		}

		authToken, ok := getAuthToken()
		if !ok {
			return
		}

		users, err := backend.Users(authToken)
		if err != nil {
			printError("An error occurred: %v", err)
			return
		}

		fmt.Printf("%s user(s)\n\n", clbold(len(users)))
		for _, u := range users {
			fmt.Printf("%s | %s | %d\n", u.UserID, u.Email, u.Role)
		}
	},
}

func init() {
	usersCmd.AddCommand(usersListCmd)
}
