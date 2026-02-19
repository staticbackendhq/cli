package cmd

import (
	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
	"fmt"
)

// functionList lists all server-side functions
var functionDeleteCmd = &cobra.Command{
	Use:   "delete name",
	Short: "Delete the function by its name",
	Long: fmt.Sprintf(`
%s

This will delete the function permanently, including all of its run history.
	`,
		clbold("Delete a function"),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if setBackend() == false {
			return
		}

		tok, ok := getRootToken()
		if !ok {
			return
		}

		if len(args) != 1 {
			printError("argument mismatch: only a name should be specified")
			return
		}

		if err := backend.DeleteFunction(tok, args[0]); err != nil {
			printError("error deleting your function: %v", err)
			return
		}

		printSuccess("the function %s has been deleted", args[0])
	},
}

func init() {
	functionCmd.AddCommand(functionDeleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")
}
