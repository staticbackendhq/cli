package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// functionList lists all server-side functions
var functionListCmd = &cobra.Command{
	Use:   "list [trigger]",
	Short: "List server-side functions",
	Long: fmt.Sprintf(`
%s

If you specify a %s only functions for that trigger will be included.

Otherwise, all functions are displayed.
	`,
		clbold(clsecondary("List functions")),
		clbold("trigger"),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if setBackend() == false {
			return
		}

		tok, ok := getRootToken()
		if !ok {
			return
		}

		var trigger string
		if len(args) == 1 {
			trigger = args[0]
		}

		results, err := backend.ListFunctions(tok)
		if err != nil {
			fmt.Printf("%s: %v\n", cldanger("An error occured"), err)
			return
		}

		// filter for trigger if supplied
		var filtered []backend.Function
		if len(trigger) > 0 {
			for _, f := range results {
				if strings.EqualFold(trigger, f.TriggerTopic) {
					filtered = append(filtered, f)
				}
			}
		} else {
			filtered = append(filtered, results...)
		}

		fmt.Printf("%s result(s)\n\n", clbold(len(filtered)))
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', tabwriter.DiscardEmptyColumns)

		fmt.Fprintf(w, "NAME\tVERSION\tTRIGGER\tLAST RUN\n")

		for _, f := range filtered {
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\n",
				f.FunctionName,
				f.Version,
				f.TriggerTopic,
				f.LastRun.Format("2006/01/02 15:04"),
			)
		}

		w.Flush()
	},
}

func init() {
	functionCmd.AddCommand(functionListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")
}
