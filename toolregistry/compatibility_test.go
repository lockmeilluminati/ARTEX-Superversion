package toolregistry

import "testing"

func TestMatchAgentCLIRequiresTerminal(t *testing.T) {
	tool := Manifest{ID: "external.nmap", ExecutionMode: ExecutionCLI}
	got := MatchAgent(AgentProfile{ID: "worker", Capabilities: []AgentCapability{CapabilityTerminal}}, tool)
	if got.Status != CompatibilityCompatible {
		t.Fatalf("status = %q, want compatible: %#v", got.Status, got)
	}
	got = MatchAgent(AgentProfile{ID: "visual-worker", Capabilities: []AgentCapability{CapabilityScreenCapture, CapabilityVisualInteraction}}, tool)
	if got.Status != CompatibilityIncompatible || len(got.Missing) != 1 || got.Missing[0] != CapabilityTerminal {
		t.Fatalf("GUI-only agent should not match CLI tool: %#v", got)
	}
}

func TestMatchAgentGUIRequiresScreenAndVisualInteraction(t *testing.T) {
	tool := Manifest{ID: "external.yakit", ExecutionMode: ExecutionGUI}
	got := MatchAgent(AgentProfile{ID: "worker", Capabilities: []AgentCapability{CapabilityTerminal}}, tool)
	if got.Status != CompatibilityIncompatible || len(got.Missing) != 2 {
		t.Fatalf("terminal-only agent should not match GUI tool: %#v", got)
	}
	got = MatchAgent(AgentProfile{ID: "visual-worker", Capabilities: []AgentCapability{CapabilityScreenCapture, CapabilityVisualInteraction}}, tool)
	if got.Status != CompatibilityCompatible {
		t.Fatalf("screen-capable agent should match GUI tool: %#v", got)
	}
}

func TestMatchAgentUnknownMetadataIsConditional(t *testing.T) {
	got := MatchAgent(AgentProfile{ID: "worker"}, Manifest{ID: "external.unknown"})
	if got.Status != CompatibilityConditional || len(got.Unknown) == 0 {
		t.Fatalf("unknown execution mode must not be treated as compatible: %#v", got)
	}
}

func TestMatchAgentChecksPlatformAndExplicitRequirements(t *testing.T) {
	tool := Manifest{
		ID: "external.tool",
		ExecutionMode: ExecutionAPI,
		Platforms: []string{"linux"},
		RequiredAgentCapabilities: []AgentCapability{CapabilityBrowser},
	}
	got := MatchAgent(AgentProfile{
		ID: "windows-agent",
		Capabilities: []AgentCapability{CapabilityAPI, CapabilityBrowser},
		Platforms: []string{"Windows"},
	}, tool)
	if got.Status != CompatibilityIncompatible || len(got.MissingPlatforms) != 1 || got.MissingPlatforms[0] != "linux" {
		t.Fatalf("platform mismatch should be reported separately: %#v", got)
	}
	got = MatchAgent(AgentProfile{
		ID: "linux-agent",
		Capabilities: []AgentCapability{CapabilityAPI, CapabilityBrowser},
		Platforms: []string{"Linux"},
	}, tool)
	if got.Status != CompatibilityCompatible {
		t.Fatalf("matching platform and capabilities should be compatible: %#v", got)
	}
}

func TestCompatibilityDoesNotGrantExecution(t *testing.T) {
	tool := Manifest{ID: "external.tool", ExecutionMode: ExecutionCLI, Enabled: false}
	got := MatchAgent(AgentProfile{ID: "worker", Capabilities: []AgentCapability{CapabilityTerminal}}, tool)
	if got.Status != CompatibilityCompatible || tool.Enabled {
		t.Fatalf("compatibility must not enable the tool: result=%#v enabled=%v", got, tool.Enabled)
	}
}
