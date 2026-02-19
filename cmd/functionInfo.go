package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// functionInfo returns all run history for a specific function
var functionInfoCmd = &cobra.Command{
	Use:   "info name",
	Short: "Display run history for a function",
	Long: fmt.Sprintf(`
%s

This command displays the last 100 run history and function output.

This is useful to diagnose if you're using the %s runtime function inside your
function code.
	`,
		clbold("Display function run history"),
		clbold("log"),
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

		fn, err := backend.FunctionInfo(tok, args[0])
		if err != nil {
			printError("error while retrieving the function: %v", err)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', tabwriter.DiscardEmptyColumns)

		fmt.Fprintf(w, "NAME\tVERSION\tTRIGGER\tLAST RUN\n")

		fmt.Fprintf(w, "%s\t%d\t%s\t%s\n",
			fn.FunctionName,
			fn.Version,
			fn.TriggerTopic,
			fn.LastRun.Format("2006/01/02 15:04"),
		)

		w.Flush()

		fmt.Printf("\n==== %s ====\n\n", clbold("RUN HISTORY"))

		start := len(fn.History)
		end := start - 100
		if end < 0 {
			end = 0
		}
		if start > 0 {
			for i := start; i > end; i-- {
				run := fn.History[i-1]
				fmt.Printf("version:%d | start:%s | execution time:%v\n",
					run.Version,
					run.Started.Format("2006/01/02 15:04"),
					run.Completed.Sub(run.Started),
				)
				for _, o := range run.Output {
					fmt.Println("\t", o)
				}

				fmt.Println("------------------")
			}
		}
	},
}

func init() {
	functionCmd.AddCommand(functionInfoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")
}
