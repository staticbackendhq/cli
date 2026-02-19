package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to your StaticBackend account.",
	Long: fmt.Sprintf(`
%s

You have to authenticate to manipulate your StaticBackend data.

We're saving your root token in the .backend.yml file, make sure to add it to your .gitignore file.
	`, clbold("Login to your account")),
	Run: func(cmd *cobra.Command, args []string) {
		dev, err := cmd.Flags().GetBool("dev")
		if err != nil {
			fmt.Println(err)
			return
		}
		pk := "dev-memory-pk"
		region := "dev"
		rtoken := "safe-to-use-in-dev-root-token"

		if !dev {
			var err error

			reader := bufio.NewReader(os.Stdin)
			fmt.Print("enter your Public Key: ")
			pk, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("error: ", err)
				return
			}

			pk = strings.Replace(pk, "\n", "", -1)

			fmt.Print("enter host URL: ")
			region, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("error: ", err)
				return
			}

			region = strings.Replace(region, "\n", "", -1)

			fmt.Print("enter your Root Token: ")
			rtoken, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("error: ", err)
				return
			}

			rtoken = strings.Replace(rtoken, "\n", "", -1)

		}

		backend.PublicKey = pk
		backend.Region = region

		// we use the SudoListRepositories as a root token validator
		if _, err := backend.SudoListRepositories(rtoken); err != nil {
			fmt.Println("invalid root token: ", err)
			return
		}

		s := fmt.Sprintf("pubKey: %s\nregion: %s\nrootToken: %s", pk, region, rtoken)
		if err := os.WriteFile(".backend.yml", []byte(s), 0660); err != nil {
			fmt.Println("unable to save your credentials: ", err)
			return
		}

		fmt.Println("Your .backend.yml file has been setup.\n\nYou're ready to use the CLI.")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	loginCmd.Flags().Bool("dev", false, "Setup for local development credentials")
}
