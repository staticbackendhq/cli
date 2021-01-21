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

// dbListCmd list document in a repository
var dbListCmd = &cobra.Command{
	Use:   "list repo-name",
	Short: "List document in a repository.",
	Long: fmt.Sprintf(`
%s

You may view documents from first to last or last to first.
	`,
		clbold(clsecondary("List documents from a repository")),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if setBackend() == false {
			return
		}

		tok, ok := getRootToken()
		if !ok {
			return
		}

		if len(args) == 0 {
			fmt.Printf("%s %s %s\n", cldanger("Argument missing"), clerror("repository"), cldanger("please supply a table name."))
			return
		}

		repo := args[0]

		page, err := cmd.Flags().GetInt("page")
		if err != nil {
			return
		}

		size, err := cmd.Flags().GetInt("size")
		if err != nil {
			return
		}

		desc, err := cmd.Flags().GetBool("descending")
		if err != nil {
			return
		}

		lp := &backend.ListParams{
			Page:       page,
			Size:       size,
			Descending: desc,
		}

		var results []map[string]interface{}
		meta, err := backend.SudoList(tok, repo, &results, lp)
		if err != nil {
			fmt.Printf("%s: %v\n", cldanger("An error occured"), err)
			return
		}

		fmt.Printf("%s result(s)\n\n", clbold(meta.Total))
		for _, doc := range results {
			o := "{ "
			for k, v := range doc {
				o += fmt.Sprintf("%s: %v, ", clsecondary(k), v)
			}

			o = strings.TrimSuffix(o, ", ") + " }"

			fmt.Println(o)
		}
	},
}

func init() {
	dbCmd.AddCommand(dbListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")

	dbListCmd.Flags().BoolP("descending", "d", false, "List in descending order of creation")
	dbListCmd.Flags().Int("page", 1, "Page index")
	dbListCmd.Flags().Int("size", 50, "Number of documents to retrieve")
}
