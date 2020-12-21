/*
Copyright © 2020 Focus Centric inc. <dominicstpierre@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"staticbackendhq/cli/core"
	"strings"

	"github.com/spf13/cobra"
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
		pubKey, ok := getPublicKey()
		if !ok {
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

		tok, err := core.Login(pubKey, email, string(pw))
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
