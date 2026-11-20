package kali

// PackageMapping keeps distro package identity separate from upstream project identity.
// Values must be populated from authoritative package metadata and reviewed sources.
type PackageMapping struct {
	PackageName string `json:"package_name"`
	PackageVersion string `json:"package_version,omitempty"`
	UpstreamName string `json:"upstream_name,omitempty"`
	UpstreamRepository string `json:"upstream_repository,omitempty"`
	License string `json:"license,omitempty"`
	SourceConfidence string `json:"source_confidence"`
	Verified bool `json:"verified"`
}

type Inventory struct {
	Platform string `json:"platform"`
	Release string `json:"release,omitempty"`
	Architecture string `json:"architecture,omitempty"`
	Packages []PackageMapping `json:"packages"`
}
