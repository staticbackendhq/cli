package cmd

import (
	"fmt"

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

		formatOpts, err := getDBDocumentFormatOptions(cmd)
		if err != nil {
			return
		}

		lp := &backend.ListParams{
			Page:       page,
			Size:       size,
			Descending: desc,
		}

		filters, err := argsToQueryItem(args[1:])
		if err != nil {
			printError("An error occurred: %v", err)
			return
		}

		var results []map[string]interface{}
		meta, err := backend.SudoFind(tok, repo, filters, &results, lp)
		if err != nil {
			printError("An error occurred: %v", err)
			return
		}

		fmt.Printf("%s result(s)\n\n", clbold(meta.Total))
		printDBDocuments(results, formatOpts)
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
	addDBDocumentFormatFlag(dbQueryCmd)
}

func argsToQueryItem(args []string) ([]backend.QueryItem, error) {
	if len(args)%3 != 0 {
		return nil, fmt.Errorf("filters must be provided as field/operator/value triples")
	}

	var filters []backend.QueryItem

	for i := 0; i < len(args); i++ {
		op, err := stringToQueryOperator(args[i+1])
		if err != nil {
			return nil, err
		}

		filters = append(filters, backend.QueryItem{
			Field: args[i],
			Op:    op,
			Value: stringToQueryValue(args[i+2]),
		})

		i += 2
	}

	return filters, nil
}

func stringToQueryOperator(op string) (backend.QueryOperator, error) {
	switch op {
	case "=", "==":
		return backend.QueryEqual, nil
	case "!=", "<>":
		return backend.QueryNotEqual, nil
	case ">":
		return backend.QueryGreaterThan, nil
	case ">=":
		return backend.QueryGreaterThanEqual, nil
	case "<":
		return backend.QueryLowerThan, nil
	case "<=":
		return backend.QueryLowerThanEqual, nil
	}

	return "", fmt.Errorf("unsupported query operator %q", op)
}

func stringToQueryValue(s string) interface{} {
	return s
}
