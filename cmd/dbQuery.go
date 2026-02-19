package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
)

// dbQueryCmd perform a query against a repository
var dbQueryCmd = &cobra.Command{
	Use:   "query filters",
	Short: "Perform a query against a repository using the provided filter(s)",
	Long: fmt.Sprintf(`
%s

You require a rootToken to execute a query.

The %s parameter is a list of filter separated by a comma (,).

field op value, field2 op value2

ex: name == "Dominic"

You may provide multiple filters separated by a comma.

name == "Dominic", access >= 2

Note: make sure you use double-quote for strings.

Supported operators:

	= or ==		For equality clause
	!= or <>	For inequality clause.
	> or >=		Greater than clause
	< or <=		Lower than clause
	`,
		clbold("Query a repository"),
		clbold("filters"),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if setBackend() == false {
			return
		}

		tok, ok := getRootToken()
		if !ok {
			return
		}

		if len(args) == 0 {
			printError("Argument missing: repository — please supply a table name.")
			return
		} else if len(args) == 1 {
			printError("Argument missing: filters — please provide filters.")
			return
		}

		repo := args[0]

		page, err := cmd.Flags().GetInt("page")
		if err != nil {
			return
		}

		size, err := cmd.Flags().GetInt("size")
		if err != nil {
			return
		}

		desc, err := cmd.Flags().GetBool("descending")
		if err != nil {
			return
		}

		lp := &backend.ListParams{
			Page:       page,
			Size:       size,
			Descending: desc,
		}

		filters := argsToQueryItem(args[1:])

		fmt.Println(filters)

		var results []map[string]interface{}
		meta, err := backend.SudoFind(tok, repo, filters, &results, lp)
		if err != nil {
			printError("An error occurred: %v", err)
			return
		}

		fmt.Printf("%s result(s)\n\n", clbold(meta.Total))
		for _, doc := range results {
			o := "{ "
			for k, v := range doc {
				o += fmt.Sprintf("%s: %v, ", k, v)
			}

			o = strings.TrimSuffix(o, ", ") + " }"

			fmt.Println(o)
		}
	},
}

func init() {
	dbCmd.AddCommand(dbQueryCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")

	dbQueryCmd.Flags().BoolP("descending", "d", false, "List in descending order of creation")
	dbQueryCmd.Flags().Int("page", 1, "Page index")
	dbQueryCmd.Flags().Int("size", 50, "Number of documents to retrieve")
}

func argsToQueryItem(args []string) []backend.QueryItem {
	var filters []backend.QueryItem

	for i := 0; i < len(args); i++ {
		filters = append(filters, backend.QueryItem{
			Field: args[i],
			Op:    stringToQueryOperator(args[i+1]),
			Value: stringToQueryValue(args[i+2]),
		})

		i += 2
	}

	return filters
}

func stringToQueryOperator(op string) backend.QueryOperator {
	var qop backend.QueryOperator

	switch op {
	case "=", "==":
		qop = backend.QueryEqual
	case "!=", "<>":
		qop = backend.QueryNotEqual
	case ">":
		qop = backend.QueryGreaterThan
	case ">=":
		qop = backend.QueryGreaterThanEqual
	case "<":
		qop = backend.QueryLowerThan
	case "<=":
		qop = backend.QueryLowerThanEqual
	}

	return qop
}

func stringToQueryValue(s string) interface{} {
	return s
}
