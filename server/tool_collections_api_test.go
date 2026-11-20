package server

import (
	"reflect"
	"strings"
	"testing"
)

func TestValidateToolCollectionIDs(t *testing.T) {
	available := map[string]struct{}{
		"runtime.nmap": {},
		"external.yakit": {},
	}
	got, err := validateToolCollectionIDs(available, []string{
		" runtime.nmap ", "external.yakit", "runtime.nmap",
	})
	if err != nil {
		t.Fatalf("valid IDs rejected: %v", err)
	}
	want := []string{"runtime.nmap", "external.yakit"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("validated IDs = %#v, want %#v", got, want)
	}
}

func TestValidateToolCollectionIDsRejectsUnknownAndEmpty(t *testing.T) {
	available := map[string]struct{}{"runtime.nmap": {}}
	for _, ids := range [][]string{
		{"runtime.nmap", "runtime.not_registered"},
		{"runtime.nmap", "   "},
	} {
		if _, err := validateToolCollectionIDs(available, ids); err == nil {
			t.Fatalf("expected invalid IDs to be rejected: %#v", ids)
		}
	}
}

func TestCollectionValidationErrorDoesNotSuggestExecution(t *testing.T) {
	_, err := validateToolCollectionIDs(map[string]struct{}{}, []string{"external.unknown"})
	if err == nil || !strings.Contains(err.Error(), "unknown tool ID") {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if strings.Contains(strings.ToLower(err.Error()), "execute") {
		t.Fatalf("validation error implies execution behavior: %v", err)
	}
}
