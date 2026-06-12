package cmd

import (
	"testing"

	"github.com/staticbackendhq/backend-go"
)

func TestArgsToQueryItemEmpty(t *testing.T) {
	filters, err := argsToQueryItem(nil)
	if err != nil {
		t.Fatalf("argsToQueryItem returned error: %v", err)
	}
	if len(filters) != 0 {
		t.Fatalf("argsToQueryItem returned %v, want no filters", filters)
	}
}

func TestArgsToQueryItemValidFilters(t *testing.T) {
	filters, err := argsToQueryItem([]string{"done", "==", "true", "access", ">=", "2"})
	if err != nil {
		t.Fatalf("argsToQueryItem returned error: %v", err)
	}

	want := []backend.QueryItem{
		{Field: "done", Op: backend.QueryEqual, Value: "true"},
		{Field: "access", Op: backend.QueryGreaterThanEqual, Value: "2"},
	}
	if len(filters) != len(want) {
		t.Fatalf("argsToQueryItem returned %v, want %v", filters, want)
	}

	for i := range want {
		if filters[i] != want[i] {
			t.Fatalf("argsToQueryItem returned %v, want %v", filters, want)
		}
	}
}

func TestArgsToQueryItemRejectsIncompleteFilter(t *testing.T) {
	if _, err := argsToQueryItem([]string{"done", "=="}); err == nil {
		t.Fatal("argsToQueryItem returned nil error for incomplete filter")
	}
}

func TestArgsToQueryItemRejectsUnknownOperator(t *testing.T) {
	if _, err := argsToQueryItem([]string{"done", "contains", "true"}); err == nil {
		t.Fatal("argsToQueryItem returned nil error for unknown operator")
	}
}
