package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/cli/llm"
)

var llmCmd = &cobra.Command{
	Use:   "llm [js|node|go]",
	Short: "Write StaticBackend client library docs for LLM context.",
	Long: fmt.Sprintf(`
%s

Writes a Markdown reference file for the StaticBackend client library into
the current directory. Drop it into your project so AI coding assistants
have full context about the API.

Available libraries:

  js   writes sb-js.md  (JavaScript / TypeScript client)
  node writes sb-node.md  (Node.js / TypeScript server client)
  go   writes sb-go.md  (Go client)

Example:

  backend llm js
  backend llm node
  backend llm go
	`, clbold("StaticBackend LLM context files")),
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		lib := args[0]

		var data []byte
		var dest string
		switch lib {
		case "js":
			data = llm.JS
			dest = "sb-js.md"
		case "node":
			data = llm.Node
			dest = "sb-node.md"
		case "go":
			data = llm.Go
			dest = "sb-go.md"
		default:
			printError("unknown library %q — use \"js\", \"node\", or \"go\"", lib)
			return
		}

		if err := os.WriteFile(dest, data, 0644); err != nil {
			printError("could not write %s: %v", dest, err)
			return
		}

		printSuccess("written %s", dest)
	},
}

func init() {
	rootCmd.AddCommand(llmCmd)
}
