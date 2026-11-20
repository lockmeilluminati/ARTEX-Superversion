package toolregistry

import (
	"encoding/json"
	"errors"

	"github.com/Autumn-27/artex/db"
)

const YakitID = "external.yakit"

// YakitManifest describes Yakit as an external security-testing platform.
// Capabilities are descriptive catalogue metadata only; no Yakit operation is
// executable through this manifest.
func YakitManifest() Manifest {
	return Manifest{
		ID:          YakitID,
		Name:        "Yakit",
		Description: "External cyber security all-in-one platform. ARTEX connectivity and supported automation operations must be verified against the installed Yakit version before use.",
		License: "AGPL-3.0",
		Maintainer: "yaklang",
		CanonicalRepository: "https://github.com/yaklang/yakit",
		Sources: []Source{
			{Kind: "upstream_project", URL: "https://github.com/yaklang/yakit"},
			{Kind: "operator_installed_application"},
		},
		Capabilities:  []string{"traffic_inspection", "request_replay", "web_security_testing", "plugin_workflows"},
		Permissions:   []Permission{PermissionDiscover, PermissionReadOnly},
		ExecutionKind: "external_adapter",
		Enabled:       false,
		Availability: AvailabilityUnknown,
		Trust:         TrustUnknown,
		ApprovalRequired: true,
		Lifecycle:     LifecycleCandidate,
		Verified:      false,
	}
}

// SeedYakit persists the initial Yakit catalogue record idempotently. It does
// not configure an endpoint, claim connectivity, enable execution, or start scans.
func SeedYakit(pg *db.DB) error {
	if pg == nil {
		return errors.New("database is required")
	}
	manifest := YakitManifest()
	if err := manifest.Validate(); err != nil {
		return err
	}
	raw, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	return pg.UpsertRegistryEntry(db.RegistryEntry{
		ID: manifest.ID, Name: manifest.Name, Manifest: raw,
	})
}
