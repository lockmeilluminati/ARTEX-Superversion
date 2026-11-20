package toolregistry

import "testing"

func TestRegisterAndLookup(t *testing.T) {
	r := New()
	m := Manifest{ID:"nmap", Name:"Nmap", Lifecycle:LifecycleCandidate}
	if err := r.Register(m); err != nil { t.Fatal(err) }
	if _, ok := r.Get("nmap"); !ok { t.Fatal("registered tool not found") }
	if err := r.Register(m); err != ErrDuplicate { t.Fatalf("got %v, want duplicate", err) }
}
