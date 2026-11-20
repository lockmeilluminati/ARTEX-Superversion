package toolregistry

import (
	"sort"
	"strings"
)

// AgentProfile describes the interaction capabilities of one ARTEX agent.
// Capability declarations must describe actual runtime support, not desired support.
type AgentProfile struct {
	ID           string             `json:"id"`
	Capabilities []AgentCapability  `json:"capabilities"`
	Platforms    []string           `json:"platforms,omitempty"`
}

type CompatibilityStatus string

const (
	CompatibilityCompatible   CompatibilityStatus = "compatible"
	CompatibilityConditional  CompatibilityStatus = "conditional"
	CompatibilityIncompatible CompatibilityStatus = "incompatible"
)

// CompatibilityResult is a capability match only. It never grants permission
// to execute a tool and does not imply that the tool is installed or verified.
type CompatibilityResult struct {
	AgentID string               `json:"agent_id"`
	ToolID string                `json:"tool_id"`
	Status CompatibilityStatus  `json:"status"`
	Missing []AgentCapability    `json:"missing_capabilities,omitempty"`
	Unknown []string             `json:"unknown_requirements,omitempty"`
	Reasons []string             `json:"reasons,omitempty"`
}

// MatchAgent compares a tool's declared interaction requirements with an
// agent's declared capabilities. Unknown metadata produces a conditional
// result rather than an optimistic match.
func MatchAgent(agent AgentProfile, tool Manifest) CompatibilityResult {
	result := CompatibilityResult{
		AgentID: agent.ID,
		ToolID:  tool.ID,
		Status:  CompatibilityCompatible,
	}

	agentCaps := make(map[AgentCapability]bool, len(agent.Capabilities))
	for _, capability := range agent.Capabilities {
		agentCaps[capability] = true
	}

	required := make(map[AgentCapability]bool, len(tool.RequiredAgentCapabilities)+2)
	for _, capability := range tool.RequiredAgentCapabilities {
		required[capability] = true
	}

	switch tool.ExecutionMode {
	case ExecutionCLI:
		required[CapabilityTerminal] = true
	case ExecutionGUI:
		required[CapabilityScreenCapture] = true
		required[CapabilityVisualInteraction] = true
	case ExecutionAPI:
		required[CapabilityAPI] = true
	case ExecutionMCP:
		required[CapabilityMCP] = true
	case ExecutionHybrid:
		// Hybrid tools must declare their concrete requirements explicitly.
		if len(tool.RequiredAgentCapabilities) == 0 {
			result.Unknown = append(result.Unknown, "hybrid execution mode has no declared agent requirements")
		}
	case "", ExecutionUnknown:
		result.Unknown = append(result.Unknown, "execution mode is unknown")
	default:
		result.Unknown = append(result.Unknown, "execution mode is unsupported: "+string(tool.ExecutionMode))
	}

	for capability := range required {
		if !agentCaps[capability] {
			result.Missing = append(result.Missing, capability)
		}
	}
	sort.Slice(result.Missing, func(i, j int) bool { return result.Missing[i] < result.Missing[j] })

	if len(tool.Platforms) > 0 {
		if len(agent.Platforms) == 0 {
			result.Unknown = append(result.Unknown, "agent platform support is undeclared")
		} else if !hasPlatformOverlap(agent.Platforms, tool.Platforms) {
			result.Missing = append(result.Missing, AgentCapability("platform:"+joinPlatforms(tool.Platforms)))
		}
	}

	if len(result.Missing) > 0 {
		result.Status = CompatibilityIncompatible
		result.Reasons = append(result.Reasons, "agent lacks one or more required capabilities or platform support")
	} else if len(result.Unknown) > 0 {
		result.Status = CompatibilityConditional
		result.Reasons = append(result.Reasons, "compatibility cannot be fully established until missing metadata is declared")
	}
	return result
}

func hasPlatformOverlap(agentPlatforms, toolPlatforms []string) bool {
	available := make(map[string]bool, len(agentPlatforms))
	for _, platform := range agentPlatforms {
		available[platformKey(platform)] = true
	}
	for _, platform := range toolPlatforms {
		if available[platformKey(platform)] {
			return true
		}
	}
	return false
}

func platformKey(platform string) string {
	// Platform identifiers are metadata labels, so normalize case and surrounding whitespace.
	return strings.ToLower(strings.TrimSpace(platform))
}

func joinPlatforms(platforms []string) string {
	copyOf := append([]string(nil), platforms...)
	sort.Strings(copyOf)
	result := ""
	for i, platform := range copyOf {
		if i > 0 {
			result += ","
		}
		result += platform
	}
	return result
}
