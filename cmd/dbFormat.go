package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func addDBDocumentFormatFlag(cmd *cobra.Command) {
	cmd.Flags().BoolP("pretty", "p", false, "Print document fields one per line")
	cmd.Flags().StringSlice("fields", nil, "Only print the provided comma-separated document fields")
}

type dbDocumentFormatOptions struct {
	pretty bool
	fields []string
}

func getDBDocumentFormatOptions(cmd *cobra.Command) (dbDocumentFormatOptions, error) {
	pretty, err := cmd.Flags().GetBool("pretty")
	if err != nil {
		return dbDocumentFormatOptions{}, err
	}

	fields, err := cmd.Flags().GetStringSlice("fields")
	if err != nil {
		return dbDocumentFormatOptions{}, err
	}

	return dbDocumentFormatOptions{
		pretty: pretty,
		fields: normalizeDBDocumentFields(fields),
	}, nil
}

func normalizeDBDocumentFields(fields []string) []string {
	normalized := make([]string, 0, len(fields))
	seen := map[string]struct{}{}
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		normalized = append(normalized, field)
	}
	return normalized
}

func formatDBDocument(doc map[string]interface{}, opts dbDocumentFormatOptions) string {
	keys := dbDocumentFields(doc, opts.fields)

	if !opts.pretty {
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s: %v", k, doc[k]))
		}
		return "{ " + strings.Join(parts, ", ") + " }"
	}

	lines := []string{"{"}
	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("\t%s: %v", k, doc[k]))
	}
	lines = append(lines, "}")

	return strings.Join(lines, "\n")
}

func dbDocumentFields(doc map[string]interface{}, fields []string) []string {
	if len(fields) > 0 {
		keys := make([]string, 0, len(fields))
		for _, field := range fields {
			if _, ok := doc[field]; ok {
				keys = append(keys, field)
			}
		}
		return keys
	}

	keys := make([]string, 0, len(doc))
	for k := range doc {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func printDBDocuments(docs []map[string]interface{}, opts dbDocumentFormatOptions) {
	for i, doc := range docs {
		if opts.pretty && i > 0 {
			fmt.Println()
		}
		fmt.Println(formatDBDocument(doc, opts))
	}
}
