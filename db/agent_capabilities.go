package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"
)

var ErrInvalidAgentCapabilityProfile = errors.New("invalid agent capability profile")

// AgentCapabilityProfile is an observed snapshot of the interfaces actually
// assembled for an agent. It is descriptive metadata, never an execution grant.
type AgentCapabilityProfile struct {
	AgentKey    string    `json:"agent_key"`
	Capabilities []string `json:"capabilities"`
	Platforms   []string  `json:"platforms"`
	Evidence    []string  `json:"evidence"`
	ObservedAt  time.Time `json:"observed_at"`
}

func normalizeProfileValues(values []string) ([]string, error) {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, ErrInvalidAgentCapabilityProfile
		}
		set[value] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for value := range set {
		out = append(out, value)
	}
	sort.Strings(out)
	return out, nil
}

// UpsertAgentCapabilityProfile stores the latest observed runtime profile.
func (d *DB) UpsertAgentCapabilityProfile(profile AgentCapabilityProfile) error {
	profile.AgentKey = strings.TrimSpace(profile.AgentKey)
	if profile.AgentKey == "" {
		return ErrInvalidAgentCapabilityProfile
	}
	var err error
	if profile.Capabilities, err = normalizeProfileValues(profile.Capabilities); err != nil {
		return err
	}
	if profile.Platforms, err = normalizeProfileValues(profile.Platforms); err != nil {
		return err
	}
	if profile.Evidence, err = normalizeProfileValues(profile.Evidence); err != nil {
		return err
	}
	capabilities, _ := json.Marshal(profile.Capabilities)
	platforms, _ := json.Marshal(profile.Platforms)
	evidence, _ := json.Marshal(profile.Evidence)
	_, err = d.Exec(`
INSERT INTO agent_capability_profiles(agent_key, capabilities, platforms, evidence, observed_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (agent_key) DO UPDATE SET
    capabilities=EXCLUDED.capabilities,
    platforms=EXCLUDED.platforms,
    evidence=EXCLUDED.evidence,
    observed_at=now()`, profile.AgentKey, capabilities, platforms, evidence)
	return err
}

// GetAgentCapabilityProfile returns the latest runtime-observed profile.
func (d *DB) GetAgentCapabilityProfile(agentKey string) (*AgentCapabilityProfile, error) {
	agentKey = strings.TrimSpace(agentKey)
	if agentKey == "" {
		return nil, ErrInvalidAgentCapabilityProfile
	}
	var profile AgentCapabilityProfile
	var capabilities, platforms, evidence []byte
	err := d.QueryRow(`
SELECT agent_key, capabilities, platforms, evidence, observed_at
FROM agent_capability_profiles WHERE agent_key=$1`, agentKey).Scan(
		&profile.AgentKey, &capabilities, &platforms, &evidence, &profile.ObservedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(capabilities, &profile.Capabilities); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(platforms, &profile.Platforms); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(evidence, &profile.Evidence); err != nil {
		return nil, err
	}
	return &profile, nil
}

// ListAgentCapabilityProfiles returns the latest observed profiles by agent key.
func (d *DB) ListAgentCapabilityProfiles() ([]AgentCapabilityProfile, error) {
	rows, err := d.Query(`
SELECT agent_key, capabilities, platforms, evidence, observed_at
FROM agent_capability_profiles ORDER BY agent_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AgentCapabilityProfile, 0)
	for rows.Next() {
		var profile AgentCapabilityProfile
		var capabilities, platforms, evidence []byte
		if err := rows.Scan(&profile.AgentKey, &capabilities, &platforms, &evidence, &profile.ObservedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(capabilities, &profile.Capabilities); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(platforms, &profile.Platforms); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(evidence, &profile.Evidence); err != nil {
			return nil, err
		}
		out = append(out, profile)
	}
	return out, rows.Err()
}
