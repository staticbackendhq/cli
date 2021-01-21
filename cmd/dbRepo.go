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
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// dbRepo list all database repositories
var dbRepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "List all database repositories.",
	Long: fmt.Sprintf(`
%s

Retrieve a list of all database repositories.
	`,
		clbold(clsecondary("List repositories")),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if setBackend() == false {
			return
		}

		tok, ok := getRootToken()
		if !ok {
			return
		}

		names, err := backend.SudoListRepositories(tok)
		if err != nil {
			fmt.Printf("%s: %v\n", cldanger("An error occured"), err)
			return
		}

		o := fmt.Sprintf("%d repositories, repo using %s are reserved repositories\n\n",
			len(names),
			clsecondary("this format"),
		)

		o += "[\n"
		for _, name := range names {
			if strings.HasPrefix(name, "sb_") {
				o += fmt.Sprintf("\t%s, \n", clsecondary(name))
			} else {
				o += fmt.Sprintf("\t%s, \n", name)
			}
		}

		o += "]"

		fmt.Println(o)
	},
}

func init() {
	dbCmd.AddCommand(dbRepoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")
}
