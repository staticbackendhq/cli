package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// dbListCmd list document in a repository
var formListCmd = &cobra.Command{
	Use:   "list [form-name]",
	Short: "View form submissions",
	Long: fmt.Sprintf(`
%s

If you specify a %s only submissions for this form will be listed.

Otherwise, the latest 100 submissions across all your forms will be displayed.
	`,
		clbold(clsecondary("List form submissions")),
		clbold("form-name"),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if setBackend() == false {
			return
		}

		tok, ok := getRootToken()
		if !ok {
			return
		}

		var name string
		if len(args) == 1 {
			name = args[0]
		}

		results, err := backend.ListForm(tok, name)
		if err != nil {
			fmt.Printf("%s: %v\n", cldanger("An error occured"), err)
			return
		}

		fmt.Printf("%s result(s)\n\n", clbold(len(results)))
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
	formCmd.AddCommand(formListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")
}
