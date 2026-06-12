package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// dbUpdateCmd updates a document in a repository.
var dbUpdateCmd = &cobra.Command{
	Use:   `update repo-name id json-content`,
	Short: "Update a document.",
	Long: fmt.Sprintf(`
%s

To update a document you name the repository, pass the document id, and pass a JSON object in a string.

$> backend db update tasks doc-id "{ \"done\": true }"
	`,
		clbold("Update document"),
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
		} else if len(args) == 2 {
			printError("Argument missing: json object — please supply a document json object.")
			return
		}

		repo, id, raw := args[0], args[1], args[2]

		var doc map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &doc); err != nil {
			printError("An error occurred: %v", err)
			return
		}

		var result map[string]interface{}
		if err := backend.SudoUpdate(tok, repo, id, doc, &result); err != nil {
			printError("An error occurred: %v", err)
			return
		}

		formatOpts, err := getDBDocumentFormatOptions(cmd)
		if err != nil {
			return
		}

		fmt.Println(formatDBDocument(result, formatOpts))
	},
}

func init() {
	dbCmd.AddCommand(dbUpdateCmd)
	addDBDocumentFormatFlag(dbUpdateCmd)
}
