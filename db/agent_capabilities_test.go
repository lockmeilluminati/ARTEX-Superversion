package db

import (
	"errors"
	"reflect"
	"testing"
)

func TestNormalizeProfileValues(t *testing.T) {
	got, err := normalizeProfileValues([]string{" terminal ", "browser", "terminal"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"browser", "terminal"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalized values = %#v, want %#v", got, want)
	}
}

func TestNormalizeProfileValuesRejectsBlank(t *testing.T) {
	if _, err := normalizeProfileValues([]string{"terminal", " "}); !errors.Is(err, ErrInvalidAgentCapabilityProfile) {
		t.Fatalf("error = %v, want ErrInvalidAgentCapabilityProfile", err)
	}
}
