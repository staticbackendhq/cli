package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// functionList lists all server-side functions
var functionDeleteCmd = &cobra.Command{
	Use:   "delete name",
	Short: "Delete the function by its name",
	Long: fmt.Sprintf(`
%s

This will delete the function permanently, including all of its run history.
	`,
		clbold(clsecondary("Delete a functions")),
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
			fmt.Printf("%s: only a name should be specified\n", cldanger("argument missmatch"))
			return
		}

		if err := backend.DeleteFunction(tok, args[0]); err != nil {
			fmt.Printf("%s: %v\n", cldanger("error deleting your function"), err)
			return
		}

		fmt.Printf("%s: the function %s has been deleted\n", clsuccess("success"))
	},
}

func init() {
	functionCmd.AddCommand(functionDeleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")
}
