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
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// dbListCmd list document in a repository
var dbCreateCmd = &cobra.Command{
	Use:   `create repo-name json-content`,
	Short: "Create a document.",
	Long: fmt.Sprintf(`
%s

To create a document you name the repository and pass a JSON object in a string.

$> backend db create tasks "{ name: \"task 1\", assign: \"dominic\", done: false }"
	`,
		clbold(clsecondary("Create document")),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if setBackend() == false {
			return
		}

		tok, ok := getRookToken()
		if !ok {
			return
		}

		if len(args) == 0 {
			fmt.Printf("%s %s %s\n", cldanger("Argument missing"), clerror("repository"), cldanger("please supply a table name."))
			return
		} else if len(args) == 1 {
			fmt.Printf("%s %s %s\n", cldanger("Argument missing"), clerror("json object"), cldanger("please supply a document json object."))
			return
		}

		repo, raw := args[0], args[1]

		var doc map[string]interface{}

		if err := json.Unmarshal([]byte(raw), &raw); err != nil {
			fmt.Printf("%s: %v\n", cldanger("An error occured"), err)
			return
		}

		var result map[string]interface{}
		if err := backend.SudoCreate(tok, repo, doc, &result); err != nil {
			fmt.Printf("%s: %v\n", cldanger("An error occured"), err)
			return
		}

		o := "{\n"
		for k, v := range result {
			o += fmt.Sprintf("\t%s: %v, \n", clsecondary(k), v)
		}

		o += "}"

		fmt.Println(o)
	},
}

func init() {
	dbCmd.AddCommand(dbCreateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")
}
