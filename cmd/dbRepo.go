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
		clbold("List repositories"),
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
			printError("An error occurred: %v", err)
			return
		}

		o := fmt.Sprintf("%d repositories, repos using this format are reserved repositories\n\n",
			len(names),
		)

		o += "[\n"
		for _, name := range names {
			if strings.HasPrefix(name, "sb_") {
				o += fmt.Sprintf("\t%s (reserved), \n", name)
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
