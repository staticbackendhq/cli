package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// dbListCmd list document in a repository
var dbGetCmd = &cobra.Command{
	Use:   "get repo-name id",
	Short: "Get a specific document by its id.",
	Long: fmt.Sprintf(`
%s

Retrieve a specific document by its id from a repository.
	`,
		clbold(clsecondary("Get a document by id")),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if !setBackend() {
			return
		}

		tok, ok := getRootToken()
		if !ok {
			return
		}

		if len(args) == 0 {
			fmt.Printf("%s %s %s\n", cldanger("Argument missing"), clerror("repository"), cldanger("please supply a table name."))
			return
		} else if len(args) == 1 {
			fmt.Printf("%s %s %s\n", cldanger("Argument missing"), clerror("id"), cldanger("please supply a document id."))
			return
		}

		repo, id := args[0], args[1]

		var result map[string]interface{}
		if err := backend.SudoGetByID(tok, repo, id, &result); err != nil {
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
	dbCmd.AddCommand(dbGetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")
}
