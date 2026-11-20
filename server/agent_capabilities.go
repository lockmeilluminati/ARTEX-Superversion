package server

import (
	"runtime"
	"sort"
	"strings"

	"github.com/Autumn-27/artex/toolregistry"
	actool "github.com/Autumn-27/norma/tool"
)

// deriveRuntimeAgentProfile records only capabilities evidenced by tools in
// the assembled agent toolset and by explicitly connected MCP servers.
func deriveRuntimeAgentProfile(agentKey string, tools []actool.CoreTool, mcpTools map[string][]string) (toolregistry.AgentProfile, []string) {
	profile := toolregistry.AgentProfile{ID: agentKey}
	capabilities := map[toolregistry.AgentCapability]bool{}
	evidence := map[string]bool{}
	toolNames := make(map[string]bool, len(tools))
	for _, tool := range tools {
		if tool != nil {
			toolNames[strings.ToLower(strings.TrimSpace(tool.Name()))] = true
		}
	}
	if toolNames["bash"] {
		capabilities[toolregistry.CapabilityTerminal] = true
		capabilitiesListEvidence := "assembled-tool:Bash"
		evidence[capabilitiesListEvidence] = true
		profile.Platforms = append(profile.Platforms, runtime.GOOS)
	}
	for serverName, names := range mcpTools {
		if len(names) == 0 {
			continue
		}
		capabilities[toolregistry.CapabilityMCP] = true
		evidence["connected-mcp-server:"+strings.ToLower(serverName)] = true
		if strings.EqualFold(serverName, "browser") {
			for _, rawName := range names {
				name := strings.ToLower(strings.TrimSpace(rawName))
				evidence["browser-mcp-tool:"+name] = true
				switch name {
				case "browser_take_screenshot", "browser_screenshot", "take_screenshot", "screenshot":
					capabilities[toolregistry.CapabilityScreenCapture] = true
				case "browser_click", "browser_type", "browser_fill_form", "browser_navigate",
					"browser_press_key", "browser_select_option", "browser_hover":
					capabilities[toolregistry.CapabilityVisualInteraction] = true
					capabilities[toolregistry.CapabilityBrowser] = true
				}
			}
		}
	}
	for capability := range capabilities {
		profile.Capabilities = append(profile.Capabilities, capability)
	}
	sort.Slice(profile.Capabilities, func(i, j int) bool { return profile.Capabilities[i] < profile.Capabilities[j] })
	profile.Platforms = uniqueSortedStrings(profile.Platforms)
	evidenceList := make([]string, 0, len(evidence))
	for item := range evidence {
		evidenceList = append(evidenceList, item)
	}
	sort.Strings(evidenceList)
	return profile, evidenceList
}

func uniqueSortedStrings(values []string) []string {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		if value != "" {
			set[value] = true
		}
	}
	out := make([]string, 0, len(set))
	for value := range set {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
