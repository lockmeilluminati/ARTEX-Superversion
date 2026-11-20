// Package normalizer translates collector output into ARTEX-neutral records.
// Normalized observations are a contract, not a parallel persistence store.
package normalizer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ObservationType string

const (
	ObservationDomain     ObservationType = "domain"
	ObservationHost       ObservationType = "host"
	ObservationIP         ObservationType = "ip"
	ObservationPort       ObservationType = "port"
	ObservationService    ObservationType = "service"
	ObservationURL        ObservationType = "url"
	ObservationTechnology ObservationType = "technology"
	ObservationFinding    ObservationType = "finding"
)

type CollectionStatus string

const (
	CollectionUnknown        CollectionStatus = "unknown"
	CollectionNotTested       CollectionStatus = "not_tested"
	CollectionTestedPositive  CollectionStatus = "tested_positive"
	CollectionTestedNegative  CollectionStatus = "tested_negative"
	CollectionPartial         CollectionStatus = "partial"
	CollectionFailed          CollectionStatus = "failed"
)

type Confidence string

const (
	ConfidenceUnknown Confidence = "unknown"
	ConfidenceInferred Confidence = "inferred"
	ConfidenceLow      Confidence = "low"
	ConfidenceMedium   Confidence = "medium"
	ConfidenceHigh     Confidence = "high"
)

type VerificationStatus string

const (
	VerificationNotAssessed  VerificationStatus = "not_assessed"
	VerificationUnverified   VerificationStatus = "unverified"
	VerificationVerified     VerificationStatus = "verified"
	VerificationFalsePositive VerificationStatus = "false_positive"
	VerificationBlocked      VerificationStatus = "blocked"
)

type ScopeStatus string

const (
	ScopeUnknown   ScopeStatus = "unknown"
	ScopeInScope   ScopeStatus = "in_scope"
	ScopeOutOfScope ScopeStatus = "out_of_scope"
)

// EvidenceReference points to retained original collector evidence. URI may be
// an ARTEX-managed evidence identifier or a confined artifact reference; it is
// never assumed to be a publicly fetchable URL.
type EvidenceReference struct {
	Kind       string    `json:"kind"`
	URI        string    `json:"uri"`
	Digest     string    `json:"digest,omitempty"`
	MediaType  string    `json:"media_type,omitempty"`
	ObservedAt time.Time `json:"observed_at"`
}

// Observation is one source-attributed claim about an entity at a point in
// time. IdentityKey groups claims about the same entity; ObservationID
// distinguishes separate claims so contradictory values are retained.
type Observation struct {
	ObservationID string             `json:"observation_id"`
	IdentityKey   string             `json:"identity_key"`
	Type          ObservationType    `json:"type"`
	Target        string             `json:"target"`
	Port          *uint16            `json:"port,omitempty"`
	Protocol      string             `json:"protocol,omitempty"`
	Name          string             `json:"name,omitempty"`
	Value         string             `json:"value,omitempty"`
	Attributes    map[string]string  `json:"attributes,omitempty"`

	SourceToolID      string    `json:"source_tool_id"`
	SourceToolVersion string    `json:"source_tool_version,omitempty"`
	SourceRunID       string    `json:"source_run_id"`
	TaskID            int64     `json:"task_id"`
	ObservedAt        time.Time `json:"observed_at"`
	Evidence          []EvidenceReference `json:"evidence"`

	Scope             ScopeStatus        `json:"scope"`
	Confidence        Confidence         `json:"confidence"`
	Verification      VerificationStatus `json:"verification"`
	Collection        CollectionStatus   `json:"collection_status"`
}

// Validate checks required provenance and type-specific identity fields.
// Unknown is a valid explicit state; empty enum values are not.
func (o Observation) Validate() error {
	if !validObservationType(o.Type) { return fmt.Errorf("invalid observation type %q", o.Type) }
	if strings.TrimSpace(o.Target) == "" { return errors.New("target is required") }
	if strings.TrimSpace(o.SourceToolID) == "" { return errors.New("source_tool_id is required") }
	if strings.TrimSpace(o.SourceRunID) == "" { return errors.New("source_run_id is required") }
	if o.TaskID <= 0 { return errors.New("task_id must be positive") }
	if o.ObservedAt.IsZero() { return errors.New("observed_at is required") }
	if len(o.Evidence) == 0 { return errors.New("at least one evidence reference is required") }
	if !validScope(o.Scope) { return fmt.Errorf("invalid scope %q", o.Scope) }
	if !validConfidence(o.Confidence) { return fmt.Errorf("invalid confidence %q", o.Confidence) }
	if !validVerification(o.Verification) { return fmt.Errorf("invalid verification %q", o.Verification) }
	if !validCollection(o.Collection) { return fmt.Errorf("invalid collection_status %q", o.Collection) }
	switch o.Type {
	case ObservationIP:
		if net.ParseIP(o.Target) == nil { return errors.New("ip observation target must be an IP address") }
	case ObservationPort:
		if net.ParseIP(o.Target) == nil { return errors.New("port observation target must be an IP address") }
		if o.Port == nil || *o.Port == 0 { return errors.New("port observation requires a port from 1 to 65535") }
	case ObservationURL:
		u, err := url.Parse(o.Target)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
			return errors.New("url observation target must be an absolute http or https URL")
		}
	case ObservationDomain, ObservationHost:
		if strings.ContainsAny(o.Target, " /\t\r\n") { return errors.New("domain/host target must not contain whitespace or a path") }
	case ObservationService, ObservationTechnology, ObservationFinding:
		if strings.TrimSpace(o.Name) == "" { return fmt.Errorf("%s observation requires name", o.Type) }
	}
	for i, e := range o.Evidence {
		if strings.TrimSpace(e.Kind) == "" || strings.TrimSpace(e.URI) == "" { return fmt.Errorf("evidence[%d] requires kind and uri", i) }
		if !e.ObservedAt.IsZero() && e.ObservedAt.After(o.ObservedAt.Add(time.Minute)) {
			return fmt.Errorf("evidence[%d] timestamp is later than observation", i)
		}
	}
	return nil
}

// NormalizeIdentity returns a deterministic entity grouping key. It excludes
// mutable claims (service banner, technology value, finding status and source)
// so contradictory observations for the same entity can coexist.
func NormalizeIdentity(kind ObservationType, target string, port *uint16, protocol, name string) (string, error) {
	if !validObservationType(kind) { return "", fmt.Errorf("invalid observation type %q", kind) }
	t := strings.TrimSpace(target)
	switch kind {
	case ObservationDomain, ObservationHost:
		t = strings.ToLower(strings.TrimSuffix(t, "."))
	case ObservationIP, ObservationPort:
		ip := net.ParseIP(t)
		if ip == nil { return "", errors.New("target must be an IP address") }
		t = ip.String()
	case ObservationURL:
		u, err := url.Parse(t)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" { return "", errors.New("target must be an absolute http or https URL") }
		u.Scheme = strings.ToLower(u.Scheme)
		u.Host = strings.ToLower(u.Host)
		u.User = nil
		t = u.String()
	}
	if t == "" { return "", errors.New("target is required") }
	identity := struct {
		Type ObservationType `json:"type"`
		Target string `json:"target"`
		Port *uint16 `json:"port,omitempty"`
		Protocol string `json:"protocol,omitempty"`
		Name string `json:"name,omitempty"`
	}{kind, t, port, strings.ToLower(strings.TrimSpace(protocol)), strings.ToLower(strings.TrimSpace(name))}
	b, err := json.Marshal(identity)
	if err != nil { return "", err }
	sum := sha256.Sum256(b)
	return "obsent:" + hex.EncodeToString(sum[:]), nil
}

// AssignIDs fills stable identifiers. Replaying the same source/run/claim at
// the same timestamp yields the same ObservationID; a different claim or run
// yields a distinct ID while retaining the shared IdentityKey.
func (o *Observation) AssignIDs() error {
	if o == nil { return errors.New("observation is nil") }
	key, err := NormalizeIdentity(o.Type, o.Target, o.Port, o.Protocol, o.Name)
	if err != nil { return err }
	o.IdentityKey = key
	canonical := struct {
		IdentityKey string `json:"identity_key"`
		SourceToolID string `json:"source_tool_id"`
		SourceToolVersion string `json:"source_tool_version,omitempty"`
		SourceRunID string `json:"source_run_id"`
		TaskID int64 `json:"task_id"`
		ObservedAt string `json:"observed_at"`
		Value string `json:"value,omitempty"`
		Attributes map[string]string `json:"attributes,omitempty"`
		Scope ScopeStatus `json:"scope"`
		Confidence Confidence `json:"confidence"`
		Verification VerificationStatus `json:"verification"`
		Collection CollectionStatus `json:"collection_status"`
		Evidence []EvidenceReference `json:"evidence"`
	}{key, o.SourceToolID, o.SourceToolVersion, o.SourceRunID, o.TaskID, o.ObservedAt.UTC().Format(time.RFC3339Nano), o.Value, o.Attributes, o.Scope, o.Confidence, o.Verification, o.Collection, o.Evidence}
	b, err := json.Marshal(canonical)
	if err != nil { return err }
	sum := sha256.Sum256(b)
	o.ObservationID = "observation:" + hex.EncodeToString(sum[:])
	return nil
}

func validObservationType(v ObservationType) bool {
	switch v { case ObservationDomain, ObservationHost, ObservationIP, ObservationPort, ObservationService, ObservationURL, ObservationTechnology, ObservationFinding: return true }; return false
}
func validScope(v ScopeStatus) bool { return v == ScopeUnknown || v == ScopeInScope || v == ScopeOutOfScope }
func validConfidence(v Confidence) bool { return v == ConfidenceUnknown || v == ConfidenceInferred || v == ConfidenceLow || v == ConfidenceMedium || v == ConfidenceHigh }
func validVerification(v VerificationStatus) bool { return v == VerificationNotAssessed || v == VerificationUnverified || v == VerificationVerified || v == VerificationFalsePositive || v == VerificationBlocked }
func validCollection(v CollectionStatus) bool { return v == CollectionUnknown || v == CollectionNotTested || v == CollectionTestedPositive || v == CollectionTestedNegative || v == CollectionPartial || v == CollectionFailed }

// ParsePort is shared by adapters that receive port numbers as text.
func ParsePort(value string) (*uint16, error) {
	n, err := strconv.ParseUint(strings.TrimSpace(value), 10, 16)
	if err != nil || n == 0 { return nil, errors.New("port must be an integer from 1 to 65535") }
	p := uint16(n)
	return &p, nil
}
