package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// dbCountCmd counts documents in a repository.
var dbCountCmd = &cobra.Command{
	Use:   "count repo-name [field op value ...]",
	Short: "Count documents in a repository.",
	Long: fmt.Sprintf(`
%s

Count all documents in a repository, or count documents matching the provided filter(s).

$> backend db count tasks
$> backend db count tasks done == true
	`,
		clbold("Count documents"),
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
		}

		repo := args[0]
		filters, err := argsToQueryItem(args[1:])
		if err != nil {
			printError("An error occurred: %v", err)
			return
		}

		n, err := backend.Count(tok, repo, filters)
		if err != nil {
			printError("An error occurred: %v", err)
			return
		}

		fmt.Println(n)
	},
}

func init() {
	dbCmd.AddCommand(dbCountCmd)
}
