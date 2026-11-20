package toolregistry

import (
	"encoding/json"
	"testing"

	"github.com/Autumn-27/artex/db"
)

func TestFromDBToolPreservesNativeIdentityAndBindings(t *testing.T) {
	row := &db.Tool{Key:"report_finding", System:true, Description:"Record a finding", Schema:json.RawMessage(`{"type":"object"}`), Agents:[]string{"worker","pentest"}, Enabled:true}
	got := FromDBTool(row)
	if got.ID!="artex.native.report_finding" || got.LegacyKey!=row.Key { t.Fatalf("unexpected identity: %#v",got) }
	if got.ExecutionKind!="native_handler" || got.Trust!=TrustCodeDefined || !got.Verified { t.Fatalf("unexpected native classification: %#v",got) }
	if len(got.AgentBindings)!=2 || got.AgentBindings[0]!="worker" { t.Fatalf("bindings not preserved: %#v",got.AgentBindings) }
	if !got.ApprovalRequired || got.Availability!=AvailabilityRegistered { t.Fatalf("unsafe default state: %#v",got) }
}

func TestFromDBToolClassifiesCustomShellAsDeclaration(t *testing.T) {
	row:=&db.Tool{Key:"nmap_env",System:false,Kind:"shell",Enabled:true}
	got:=FromDBTool(row)
	if got.ID!="artex.custom.nmap_env" || got.ExecutionKind!="shell_declaration" { t.Fatalf("unexpected custom shell mapping: %#v",got) }
	if got.Trust!=TrustOperatorConfigured || got.Verified { t.Fatalf("custom tool must not be marked verified: %#v",got) }
}

func TestImportDBToolsSkipsNilAndEmptyKeys(t *testing.T) {
	got:=ImportDBTools([]*db.Tool{nil,{}, {Key:"ping",System:true}})
	if len(got)!=1 || got[0].LegacyKey!="ping" { t.Fatalf("unexpected import result: %#v",got) }
}
