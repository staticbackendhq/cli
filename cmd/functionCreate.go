package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// functionCreate creates a new server-side function
var functionCreateCmd = &cobra.Command{
	Use:   "add",
	Short: "Create a new server-side function",
	Long: fmt.Sprintf(`
%s

Create a function that executes on a recurring schedule:

backend funciton add --name trial_expire --trigger daily_task_trial_expire --source ./functions/trial_expire.js

You may create server-side functions that execute based on those triggers:

%s: invoke via a URL
%s: a topic/event that will run your function when published
	`,
		clbold("Create a function"),
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

		fn := backend.Function{
			FunctionName: name,
			TriggerTopic: trigger,
			Code:         string(b),
		}

		if err := backend.AddFunction(tok, fn); err != nil {
			printError("error adding your function: %v", err)
			return
		}

		printSuccess("Function %s created successfully", clbold(name))
		if trigger == "web" {
			fmt.Printf("Function URL: %s\n", clbold("[your_domain]/fn/exec/"+name))
		} else {
			fmt.Printf("Function will trigger on topic: %s\n", clbold(trigger))
		}
	},
}

func init() {
	functionCmd.AddCommand(functionCreateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")

	functionCreateCmd.Flags().String("name", "", "function name")
	functionCreateCmd.Flags().String("trigger", "", "execution trigger either web or topic")
	functionCreateCmd.Flags().String("source", "", "path of the JavaScript file")
}
