package server

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRegistryCatalogRecordDoesNotExposeExecutionSpec(t *testing.T) {
	record := registryCatalogRecord{
		ID: "runtime.custom_probe",
		Name: "custom_probe",
		Source: "custom",
		Description: "Example custom tool",
		Enabled: true,
		Kind: "command",
		Agents: []string{"worker"},
	}
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal catalogue record: %v", err)
	}
	encoded := string(raw)
	if strings.Contains(encoded, "\"exec\"") || strings.Contains(encoded, "secret-command") {
		t.Fatalf("catalogue projection unexpectedly exposed execution data: %s", encoded)
	}
	if !strings.Contains(encoded, "\"source\":\"custom\"") {
		t.Fatalf("catalogue projection omitted source: %s", encoded)
	}
}
