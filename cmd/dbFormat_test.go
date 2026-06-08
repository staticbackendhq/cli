package cmd

import "testing"

func TestFormatDBDocumentCompact(t *testing.T) {
	doc := map[string]interface{}{
		"name":   "Dominic",
		"access": 2,
	}

	want := "{ access: 2, name: Dominic }"
	if got := formatDBDocument(doc, dbDocumentFormatOptions{}); got != want {
		t.Fatalf("formatDBDocument compact returned %q, want %q", got, want)
	}
}

func TestFormatDBDocumentPretty(t *testing.T) {
	doc := map[string]interface{}{
		"name":   "Dominic",
		"access": 2,
	}

	want := "{\n\taccess: 2\n\tname: Dominic\n}"
	if got := formatDBDocument(doc, dbDocumentFormatOptions{pretty: true}); got != want {
		t.Fatalf("formatDBDocument pretty returned %q, want %q", got, want)
	}
}

func TestFormatDBDocumentWithFields(t *testing.T) {
	doc := map[string]interface{}{
		"id":     "doc-1",
		"name":   "Dominic",
		"access": 2,
	}

	opts := dbDocumentFormatOptions{fields: []string{"name", "access"}}
	want := "{ name: Dominic, access: 2 }"
	if got := formatDBDocument(doc, opts); got != want {
		t.Fatalf("formatDBDocument with fields returned %q, want %q", got, want)
	}
}

func TestFormatDBDocumentWithFieldsOmitsMissingFields(t *testing.T) {
	doc := map[string]interface{}{
		"name": "Dominic",
	}

	opts := dbDocumentFormatOptions{fields: []string{"id", "name"}}
	want := "{ name: Dominic }"
	if got := formatDBDocument(doc, opts); got != want {
		t.Fatalf("formatDBDocument with missing fields returned %q, want %q", got, want)
	}
}

func TestNormalizeDBDocumentFields(t *testing.T) {
	fields := []string{" name ", "", "access", "name"}
	want := []string{"name", "access"}

	got := normalizeDBDocumentFields(fields)
	if len(got) != len(want) {
		t.Fatalf("normalizeDBDocumentFields returned %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("normalizeDBDocumentFields returned %v, want %v", got, want)
		}
	}
}
