package semver

import (
	"strings"

	"golang.org/x/mod/semver"
)

// Version represents a semantic version that may be prefixed with a domain,
// such as azctx.lvlcn-t.dev/v1alpha1.
//
// It accepts ordinary semver forms such as v1.2.3, 1.2.3,
// v1.2.3-rc.0, and v1.2.3+build.1. It also accepts Kubernetes-style
// API versions v1, v1alphaN, and v1betaN, normalizing them to semver
// equivalents for comparison.
type Version string

// String returns the [Version] as a string.
func (v Version) String() string {
	return string(v)
}

// IsValid reports whether the [Version] is a valid semantic version.
func (v Version) IsValid() bool {
	return semver.IsValid(v.semver())
}

// Compatible reports whether the [Version] is compatible with another [Version].
// Stable versions are compatible when they have the same major version.
// Prerelease API versions, such as v1alpha1 and v1beta1, are compatible only
// when they normalize to the same API version.
// Versions with different groups are always incompatible.
// If either version is invalid, it returns false.
func (v Version) Compatible(other Version) bool {
	if !v.IsValid() || !other.IsValid() {
		return false
	}

	vg, _ := v.groupVersion()
	og, _ := other.groupVersion()
	if vg != og {
		return false
	}

	return v.apiVersion() == other.apiVersion()
}

// AtLeast reports whether the [Version] is greater than or equal to another [Version].
// If either version is invalid, it returns false.
// Versions with different groups are always incomparable, and AtLeast returns false.
func (v Version) AtLeast(other Version) bool {
	if !v.IsValid() || !other.IsValid() {
		return false
	}

	vg, _ := v.groupVersion()
	og, _ := other.groupVersion()
	if vg != og {
		return false
	}

	return semver.Compare(v.semver(), other.semver()) >= 0
}

// semver returns the version in semver format, ensuring it has a "v" prefix.
func (v Version) semver() string {
	_, version := v.groupVersion()

	// Preserve normal semver behavior first, allowing both "v1.2.3" and "1.2.3" formats.
	if strings.HasPrefix(version, "v") && semver.IsValid(version) {
		return version
	}
	if !strings.HasPrefix(version, "v") && semver.IsValid("v"+version) {
		return "v" + version
	}

	// Then fallback, to allow Kubernetes-style versions like v1, v1alpha1, v1beta1.
	if normalized := normalizeGroupVersion(version); normalized != "" {
		return normalized
	}

	return ""
}

// groupVersion returns the group and version parts of the [Version],
// if there is a group attached to the version.
func (v Version) groupVersion() (group, version string) {
	s := string(v)

	if i := strings.LastIndexByte(s, '/'); i >= 0 {
		return s[:i], s[i+1:]
	}

	return "", s
}

// apiVersion returns the compatibility key for the [Version].
// Stable v1+ versions are compatible by major version. Prerelease versions
// and non-zero v0 versions are compatible only when they match exactly.
func (v Version) apiVersion() string {
	s := v.semver()
	if s == "" {
		return ""
	}

	if semver.Prerelease(s) != "" {
		return s
	}

	major := semver.Major(s)
	if major == "" {
		return ""
	}

	if major == "v0" {
		// v0 normalizes to v0.0.0, so keep v0 and v0.0.0 compatible
		// while treating other v0 versions as exact, unstable API versions.
		if s == "v0.0.0" {
			return "v0"
		}
		return s
	}

	return major
}

// normalizeGroupVersion returns a semantic version equivalent for a
// Kubernetes-style group version, such as v1alpha1 or v1beta1.
func normalizeGroupVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	if v == "" {
		return ""
	}

	major, rest, ok := cutDigits(v)
	if !ok {
		return ""
	}

	switch {
	case rest == "":
		return "v" + major + ".0.0"

	case strings.HasPrefix(rest, "alpha"):
		n, tail, ok := cutDigits(strings.TrimPrefix(rest, "alpha"))
		if !ok || tail != "" {
			return ""
		}
		return "v" + major + ".0.0-alpha." + n

	case strings.HasPrefix(rest, "beta"):
		n, tail, ok := cutDigits(strings.TrimPrefix(rest, "beta"))
		if !ok || tail != "" {
			return ""
		}
		return "v" + major + ".0.0-beta." + n

	default:
		return ""
	}
}

// cutDigits returns the leading digits in s and the remaining suffix.
// It reports whether s starts with at least one digit.
func cutDigits(s string) (digits, rest string, ok bool) {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 {
		return "", s, false
	}
	return s[:i], s[i:], true
}
