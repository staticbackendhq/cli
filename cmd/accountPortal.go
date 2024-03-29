package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// accountCreateCmd represents the accountCreate command
var accountPortalCmd = &cobra.Command{
	Use:   "portal",
	Short: "Retrieve a URL to manage your subscription and credit card.",
	Long: fmt.Sprintf(`
%s

Let you manage your billing account, change plan, update credit card and cancel.
		`,
		clbold(clsecondary("Access your billing portal")),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if !setBackend() {
			return
		}

		tok, ok := getRootToken()
		if !ok {
			return
		}

		var link string
		if err := backend.Get(tok, "/account/portal", &link); err != nil {
			fmt.Printf("%s: %v\n", cldanger("An error occured"), err)
			return
		}

		fmt.Printf("%s\n", clsecondary("You may access your billing portal via this URL:"))
		fmt.Println(link)
	},
}

func init() {
	accountCmd.AddCommand(accountPortalCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCreateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// accountCreateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
