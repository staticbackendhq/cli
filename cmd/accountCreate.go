package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// accountCreateCmd represents the accountCreate command
var accountCreateCmd = &cobra.Command{
	Use:   "create email",
	Short: "Create a new account.",
	Long: fmt.Sprintf(`
%s

We require a credit card to create new account.

No charges or subscription will be assigned to the account creation.

You may pick a paid plan later when you're ready.
		`,
		clbold(clsecondary("Create your account")),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Printf("%s %s %s\n", cldanger("Argument missing"), clerror("email"), cldanger("please supply an email."))
			return
		}

		email := args[0]
		stripeURL, err := backend.NewSystemAccount(email)
		if err != nil {
			fmt.Printf("%s: %v\n", cldanger("An error occured"), err)
			return
		}

		fmt.Printf("%s\n", clbold("Your account has been created and your 14-day trial is almost ready."))
		fmt.Println("To complete your registration follow this link:")
		fmt.Printf("%s\n", clbold(stripeURL))
		fmt.Printf("\n\n%s\n", clsecondary("Your account will unlock once you add a payment method."))
	},
}

func init() {
	accountCmd.AddCommand(accountCreateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCreateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// accountCreateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
