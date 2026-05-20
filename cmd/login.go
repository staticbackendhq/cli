package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
	"golang.org/x/term"
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
		pk := "dev_memory_pk"
		region := "dev"
		rtoken := "safe-to-use-in-dev-root-token"
		email := "admin@dev.com"
		password := "devpw1234"

		if dev {
			fmt.Println("In development, an admin user is already available: admin@dev.com / devpw1234")
		} else {
			var err error

			reader := bufio.NewReader(os.Stdin)
			fmt.Print("enter your Public Key: ")
			pk, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("error: ", err)
				return
			}

			pk = cleanConfigValue(pk)

			fmt.Print("enter host URL: ")
			region, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("error: ", err)
				return
			}

			region = normalizeBackendRegion(region)

			fmt.Print("enter your Root Token: ")
			rtoken, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("error: ", err)
				return
			}

			rtoken = cleanConfigValue(rtoken)

			fmt.Print("enter your email: ")
			email, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("error: ", err)
				return
			}

			email = cleanConfigValue(email)

			password, err = readPassword(reader, "enter your password: ")
			if err != nil {
				fmt.Println("error: ", err)
				return
			}

			password = cleanConfigValue(password)
		}

		backend.PublicKey = pk
		backend.Region = normalizeBackendRegion(region)

		// we use the SudoListRepositories as a root token validator
		if _, err := backend.SudoListRepositories(rtoken); err != nil {
			fmt.Println("invalid root token: ", err)
			return
		}

		authToken, err := backend.Login(email, password)
		if err != nil {
			fmt.Println("error logging in with email/password: ", err)
			authToken = ""
		}

		s := fmt.Sprintf("pubKey: %s\nregion: %s\nrootToken: %s\nemail: %s\npassword: %s\nauthToken: %s", pk, region, rtoken, email, password, authToken)
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

func readPassword(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	return reader.ReadString('\n')
}
