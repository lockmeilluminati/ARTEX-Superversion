package server

import (
	"testing"

	"github.com/Autumn-27/artex/toolregistry"
	actool "github.com/Autumn-27/norma/tool"
)

func TestDeriveRuntimeAgentProfileUsesObservedTools(t *testing.T) {
	tools := []actool.CoreTool{actool.Build(actool.Spec{Name: "Bash"})}
	profile, evidence := deriveRuntimeAgentProfile("worker", tools, nil)
	if profile.ID != "worker" {
		t.Fatalf("agent id = %q", profile.ID)
	}
	if len(profile.Capabilities) != 1 || profile.Capabilities[0] != toolregistry.CapabilityTerminal {
		t.Fatalf("capabilities = %#v, want terminal only", profile.Capabilities)
	}
	if len(profile.Platforms) != 1 {
		t.Fatalf("platforms = %#v, want runtime platform", profile.Platforms)
	}
	if len(evidence) == 0 {
		t.Fatal("expected evidence for observed Bash tool")
	}
}

func TestDeriveRuntimeAgentProfileRequiresObservedBrowserTools(t *testing.T) {
	tools := []actool.CoreTool{actool.Build(actool.Spec{Name: "Bash"})}
	profile, _ := deriveRuntimeAgentProfile("worker", tools, map[string][]string{
		"browser": {"browser_navigate", "browser_click", "browser_take_screenshot"},
	})
	want := map[toolregistry.AgentCapability]bool{
		toolregistry.CapabilityTerminal: true,
		toolregistry.CapabilityMCP: true,
		toolregistry.CapabilityBrowser: true,
		toolregistry.CapabilityVisualInteraction: true,
		toolregistry.CapabilityScreenCapture: true,
	}
	for _, capability := range profile.Capabilities {
		delete(want, capability)
	}
	if len(want) != 0 {
		t.Fatalf("missing observed capabilities: %#v; got %#v", want, profile.Capabilities)
	}
}

func TestDeriveRuntimeAgentProfileDoesNotAssumeGUI(t *testing.T) {
	tools := []actool.CoreTool{actool.Build(actool.Spec{Name: "Bash"})}
	profile, _ := deriveRuntimeAgentProfile("worker", tools, map[string][]string{"browser": {"browser_get_text"}})
	for _, capability := range profile.Capabilities {
		if capability == toolregistry.CapabilityScreenCapture || capability == toolregistry.CapabilityVisualInteraction || capability == toolregistry.CapabilityBrowser {
			t.Fatalf("unrecognized browser tool must not imply GUI capability: %#v", profile.Capabilities)
		}
	}
}
