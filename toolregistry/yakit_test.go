package toolregistry

import "testing"

func TestYakitManifestIsCandidateAndNotExecutable(t *testing.T) {
	m := YakitManifest()
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	if m.ID != YakitID {
		t.Fatalf("ID = %q, want %q", m.ID, YakitID)
	}
	if m.Lifecycle != LifecycleCandidate {
		t.Fatalf("lifecycle = %q, want candidate", m.Lifecycle)
	}
	if m.Enabled || m.Verified || m.Availability != AvailabilityUnknown {
		t.Fatalf("Yakit must start disabled, unverified and unknown: %#v", m)
	}
	if m.ExecutionKind != "external_adapter" || !m.ApprovalRequired {
		t.Fatalf("unexpected execution policy: %#v", m)
	}
	for _, permission := range m.Permissions {
		if permission == PermissionExecute || permission == PermissionInstall {
			t.Fatalf("Yakit must not receive execute/install permissions: %#v", m.Permissions)
		}
	}
}
