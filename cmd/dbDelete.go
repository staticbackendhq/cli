package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// dbDeleteCmd deletes a document in a repository.
var dbDeleteCmd = &cobra.Command{
	Use:   "delete repo-name id",
	Short: "Delete a document.",
	Long: fmt.Sprintf(`
%s

Permanently delete a specific document by its id from a repository.
	`,
		clbold("Delete document"),
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
			printError("Argument missing: repository — please supply a table name.")
			return
		} else if len(args) == 1 {
			printError("Argument missing: id — please supply a document id.")
			return
		}

		repo, id := args[0], args[1]
		if err := backend.SudoDelete(tok, repo, id); err != nil {
			printError("An error occurred: %v", err)
			return
		}

		printSuccess("the document %s has been deleted", id)
	},
}

func init() {
	dbCmd.AddCommand(dbDeleteCmd)
}
