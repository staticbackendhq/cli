package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// functionCreate creates a new server-side function
var functionUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a server-side function",
	Long: fmt.Sprintf(`
%s

You may update a function, we'll auto-increment its version for you.

backend function update --name fn_name --trigger web --source ./functions/web.js

We support two type of triggers at this moment:

%s: invoke via a URL
%s: a topic/event that will run your function when published
	`,
		clbold("Update a function"),
		clbold("web"),
		clbold("any_topic_here"),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if setBackend() == false {
			return
		}

		tok, ok := getRootToken()
		if !ok {
			return
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil || len(name) == 0 {
			printError("missing parameter: the --name option is required")
			return
		}

		trigger, err := cmd.Flags().GetString("trigger")
		if err != nil || len(trigger) == 0 {
			printError("missing parameter: the --trigger option is required")
			return
		}

		source, err := cmd.Flags().GetString("source")
		if err != nil || len(trigger) == 0 {
			printError("missing parameter: the --source option is required")
			return
		}

		b, err := os.ReadFile(source)
		if err != nil {
			printError("error reading source file: %v", err)
			return
		}

		fn, err := backend.FunctionInfo(tok, name)
		if err != nil {
			printError("function info error: %v", err)
			return
		}

		upfn := backend.Function{
			ID:           fn.ID,
			FunctionName: name,
			TriggerTopic: trigger,
			Code:         string(b),
		}

		if err := backend.UpdateFunction(tok, upfn); err != nil {
			printError("error updating your function: %v", err)
			return
		}

		printSuccess("Function %s updated successfully", clbold(name))
		if trigger == "web" {
			fmt.Printf("Function URL: %s\n", clbold("[your_domain]/fn/exec/"+name))
		} else {
			fmt.Printf("Function will trigger on topic: %s\n", clbold(trigger))
		}
	},
}

func init() {
	functionCmd.AddCommand(functionUpdateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")

	functionUpdateCmd.Flags().String("name", "", "function name")
	functionUpdateCmd.Flags().String("trigger", "", "execution trigger either web or topic")
	functionUpdateCmd.Flags().String("source", "", "path of the JavaScript file")
}
