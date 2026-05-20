package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/staticbackendhq/backend-go"
)

// functionRun executes a server-side function by its name.
var functionRunCmd = &cobra.Command{
	Use:   "run name",
	Short: "Run a server-side function",
	Long: fmt.Sprintf(`
%s

Run a function by name using your authenticated user token.

Use %s to execute with your root token instead.

The optional %s flag accepts a JSON value that is passed as the first argument
to your function handler.

backend function run fn_name --data '{"from":"cli"}'
backend function run fn_name --use-root-token --data '{"from":"cli"}'
	`,
		clbold("Run a function"),
		clbold("--use-root-token"),
		clbold("--data"),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if setBackend() == false {
			return
		}

		tok, usingRoot, ok := functionRunToken(cmd)
		if !ok {
			return
		}

		if len(args) != 1 {
			printError("argument mismatch: only a name should be specified")
			return
		}

		data, ok := functionRunData(cmd)
		if !ok {
			return
		}

		name := args[0]
		started := time.Now()
		if err := backend.Post(tok, functionRunPath(name, usingRoot), data, nil); err != nil {
			printError("error running your function: %v", err)
			return
		}

		printSuccess("Function %s executed successfully", clbold(name))
		functionRunPrintOutput(cmd, name, tok, usingRoot, started)
	},
}

func init() {
	functionCmd.AddCommand(functionRunCmd)

	functionRunCmd.Flags().String("data", "{}", "JSON value to send to the function")
	functionRunCmd.Flags().String("data-file", "", "path of a JSON file to send to the function")
	functionRunCmd.Flags().Bool("output", true, "display the latest run output when rootToken is configured")
	functionRunCmd.Flags().Bool("use-root-token", false, "run the function with rootToken instead of authToken")
}

func functionRunToken(cmd *cobra.Command) (token string, usingRoot, ok bool) {
	useRoot, err := cmd.Flags().GetBool("use-root-token")
	if err != nil {
		printError("unable to read --use-root-token option: %v", err)
		return "", false, false
	}

	if useRoot {
		tok, ok := getRootToken()
		return tok, true, ok
	}

	tok, ok := getAuthToken()
	return tok, false, ok
}

func functionRunPath(name string, usingRoot bool) string {
	if usingRoot {
		return "/fn/sudoexec/" + url.PathEscape(name)
	}

	return "/fn/exec/" + url.PathEscape(name)
}

func functionRunData(cmd *cobra.Command) (any, bool) {
	dataFile, err := cmd.Flags().GetString("data-file")
	if err != nil {
		printError("unable to read --data-file option: %v", err)
		return nil, false
	}

	raw, err := cmd.Flags().GetString("data")
	if err != nil {
		printError("unable to read --data option: %v", err)
		return nil, false
	}

	if len(dataFile) > 0 {
		b, err := os.ReadFile(dataFile)
		if err != nil {
			printError("error reading data file: %v", err)
			return nil, false
		}
		raw = string(b)
	}

	var data any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		printError("invalid JSON data: %v", err)
		return nil, false
	}

	return data, true
}

func functionRunPrintOutput(cmd *cobra.Command, name, token string, usingRoot bool, started time.Time) {
	showOutput, err := cmd.Flags().GetBool("output")
	if err != nil || !showOutput {
		return
	}

	if !usingRoot {
		token = cleanConfigValue(viper.GetString("rootToken"))
		if len(token) == 0 {
			return
		}
	}

	fn, ok := functionRunLatestInfo(token, name, started)
	if !ok || len(fn.History) == 0 {
		return
	}

	run := fn.History[len(fn.History)-1]
	fmt.Printf("\n==== %s ====\n\n", clbold("RUN OUTPUT"))
	for _, o := range run.Output {
		fmt.Println(o)
	}
}

func functionRunLatestInfo(token, name string, started time.Time) (backend.Function, bool) {
	var fn backend.Function
	for i := 0; i < 10; i++ {
		current, err := backend.FunctionInfo(token, name)
		if err == nil {
			fn = current
			if len(fn.History) > 0 && !fn.History[len(fn.History)-1].Started.Before(started) {
				return fn, true
			}
		}

		time.Sleep(200 * time.Millisecond)
	}

	return fn, false
}
