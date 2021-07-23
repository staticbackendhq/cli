package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
	"golang.org/x/crypto/ssh/terminal"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to your StaticBackend account.",
	Long: fmt.Sprintf(`
%s
	
You have to authenticate to manipulate your StaticBackend data.
	
We're saving your email/password in the .backend file, make sure to add it to your .gitignore file.
	`, clbold(clsecondary("Login to your account"))),
	Run: func(cmd *cobra.Command, args []string) {
		if ok := setBackend(); !ok {
			return
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("%s\n", clsecondary("enter your email: "))
		email, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("error: ", err)
			return
		}

		email = strings.Replace(email, "\n", "", -1)

		fmt.Printf("%s\n", clsecondary("enter your password: "))
		pw, err := terminal.ReadPassword(0)
		if err != nil {
			fmt.Println("error: ", err)
			return
		}

		tok, err := backend.Login(email, string(pw))
		if err != nil {
			fmt.Printf("%s: %v\n", cldanger("an error occured"), err)
			return
		}

		fmt.Println("token", tok)
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
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
