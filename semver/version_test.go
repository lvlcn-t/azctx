package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionString(t *testing.T) {
	tests := []struct {
		name string
		v    Version
		want string
	}{
		{
			name: "plain semver",
			v:    Version("v1.2.3"),
			want: "v1.2.3",
		},
		{
			name: "group version",
			v:    Version("azctx.lvlcn-t.dev/v1alpha1"),
			want: "azctx.lvlcn-t.dev/v1alpha1",
		},
		{
			name: "empty",
			v:    Version(""),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.v.String())
		})
	}
}

func TestVersionIsValid(t *testing.T) {
	tests := []struct {
		name  string
		v     Version
		valid bool
	}{
		{
			name:  "valid semver with v prefix",
			v:     Version("v1.2.3"),
			valid: true,
		},
		{
			name:  "valid semver without v prefix",
			v:     Version("1.2.3"),
			valid: true,
		},
		{
			name:  "valid semver prerelease rc dot zero",
			v:     Version("v1.2.3-rc.0"),
			valid: true,
		},
		{
			name:  "valid semver prerelease rc zero",
			v:     Version("v1.2.3-rc0"),
			valid: true,
		},
		{
			name:  "valid semver with build metadata",
			v:     Version("v1.2.3+build.1"),
			valid: true,
		},
		{
			name:  "valid semver prerelease and build metadata",
			v:     Version("v1.2.3-rc.0+build.1"),
			valid: true,
		},
		{
			name:  "valid group semver",
			v:     Version("azctx.lvlcn-t.dev/v1.2.3"),
			valid: true,
		},
		{
			name:  "valid group semver without v prefix",
			v:     Version("azctx.lvlcn-t.dev/1.2.3"),
			valid: true,
		},
		{
			name:  "valid kubernetes stable version",
			v:     Version("v1"),
			valid: true,
		},
		{
			name:  "valid kubernetes alpha version",
			v:     Version("v1alpha1"),
			valid: true,
		},
		{
			name:  "valid kubernetes alpha version without v prefix",
			v:     Version("1alpha1"),
			valid: true,
		},
		{
			name:  "valid kubernetes beta version",
			v:     Version("v1beta1"),
			valid: true,
		},
		{
			name:  "valid group kubernetes alpha version",
			v:     Version("azctx.lvlcn-t.dev/v1alpha1"),
			valid: true,
		},
		{
			name:  "valid group kubernetes beta version",
			v:     Version("azctx.lvlcn-t.dev/v1beta1"),
			valid: true,
		},
		{
			name:  "invalid empty version",
			v:     Version(""),
			valid: false,
		},
		{
			name:  "invalid v only",
			v:     Version("v"),
			valid: false,
		},
		{
			name:  "invalid kubernetes alpha without number",
			v:     Version("v1alpha"),
			valid: false,
		},
		{
			name:  "invalid kubernetes beta without number",
			v:     Version("v1beta"),
			valid: false,
		},
		{
			name:  "invalid kubernetes rc shorthand",
			v:     Version("v1rc0"),
			valid: false,
		},
		{
			name:  "invalid kubernetes gamma shorthand",
			v:     Version("v1gamma1"),
			valid: false,
		},
		{
			name:  "invalid suffix after alpha number",
			v:     Version("v1alpha1foo"),
			valid: false,
		},
		{
			name:  "invalid suffix after beta number",
			v:     Version("v1beta1foo"),
			valid: false,
		},
		{
			name:  "invalid semver bad prerelease",
			v:     Version("v1.2.3-"),
			valid: false,
		},
		{
			name:  "invalid group with empty version",
			v:     Version("azctx.lvlcn-t.dev/"),
			valid: false,
		},
		{
			name:  "valid version after last slash",
			v:     Version("example.com/apis/v1beta1"),
			valid: true,
		},
		{
			name:  "invalid version after last slash",
			v:     Version("example.com/apis/not-a-version"),
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.v.IsValid())
		})
	}
}

func TestVersionAtLeast(t *testing.T) {
	tests := []struct {
		name  string
		v     Version
		other Version
		want  bool
	}{
		{
			name:  "equal semver",
			v:     Version("v1.2.3"),
			other: Version("v1.2.3"),
			want:  true,
		},
		{
			name:  "greater semver",
			v:     Version("v1.2.4"),
			other: Version("v1.2.3"),
			want:  true,
		},
		{
			name:  "less semver",
			v:     Version("v1.2.2"),
			other: Version("v1.2.3"),
			want:  false,
		},
		{
			name:  "semver without v prefix compares",
			v:     Version("1.2.4"),
			other: Version("v1.2.3"),
			want:  true,
		},
		{
			name:  "different group prefix when comparing",
			v:     Version("foo.example/v1.2.3"),
			other: Version("bar.example/v1.2.3"),
			want:  false,
		},
		{
			name:  "different group prefix with same version is not at least",
			v:     Version("azctx.lvlcn-t.dev/v1.2.4"),
			other: Version("v1.2.3"),
			want:  false,
		},
		{
			name:  "alpha one at least alpha one",
			v:     Version("v1alpha1"),
			other: Version("v1alpha1"),
			want:  true,
		},
		{
			name:  "alpha two at least alpha one",
			v:     Version("v1alpha2"),
			other: Version("v1alpha1"),
			want:  true,
		},
		{
			name:  "alpha one not at least alpha two",
			v:     Version("v1alpha1"),
			other: Version("v1alpha2"),
			want:  false,
		},
		{
			name:  "beta one at least alpha one",
			v:     Version("v1beta1"),
			other: Version("v1alpha1"),
			want:  true,
		},
		{
			name:  "alpha one not at least beta one",
			v:     Version("v1alpha1"),
			other: Version("v1beta1"),
			want:  false,
		},
		{
			name:  "stable at least beta",
			v:     Version("v1"),
			other: Version("v1beta1"),
			want:  true,
		},
		{
			name:  "beta not at least stable",
			v:     Version("v1beta1"),
			other: Version("v1"),
			want:  false,
		},
		{
			name:  "major two stable at least major one stable",
			v:     Version("v2"),
			other: Version("v1"),
			want:  true,
		},
		{
			name:  "major one stable not at least major two alpha",
			v:     Version("v1"),
			other: Version("v2alpha1"),
			want:  false,
		},
		{
			name:  "full rc semver at least beta semver",
			v:     Version("v1.0.0-rc.0"),
			other: Version("v1.0.0-beta.1"),
			want:  true,
		},
		{
			name:  "stable full semver at least rc",
			v:     Version("v1.0.0"),
			other: Version("v1.0.0-rc.0"),
			want:  true,
		},
		{
			name:  "rc not at least stable",
			v:     Version("v1.0.0-rc.0"),
			other: Version("v1.0.0"),
			want:  false,
		},
		{
			name:  "invalid receiver returns false",
			v:     Version("not-a-version"),
			other: Version("v1.0.0"),
			want:  false,
		},
		{
			name:  "invalid other returns false",
			v:     Version("v1.0.0"),
			other: Version("not-a-version"),
			want:  false,
		},
		{
			name:  "both invalid returns false",
			v:     Version("not-a-version"),
			other: Version("also-not-a-version"),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.v.AtLeast(tt.other))
		})
	}
}

func TestNormalizeGroupVersion(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "stable with v prefix",
			in:   "v1",
			want: "v1.0.0",
		},
		{
			name: "stable without v prefix",
			in:   "1",
			want: "v1.0.0",
		},
		{
			name: "alpha with v prefix",
			in:   "v1alpha1",
			want: "v1.0.0-alpha.1",
		},
		{
			name: "alpha without v prefix",
			in:   "1alpha1",
			want: "v1.0.0-alpha.1",
		},
		{
			name: "alpha multi digit",
			in:   "v1alpha12",
			want: "v1.0.0-alpha.12",
		},
		{
			name: "beta with v prefix",
			in:   "v1beta1",
			want: "v1.0.0-beta.1",
		},
		{
			name: "beta without v prefix",
			in:   "1beta1",
			want: "v1.0.0-beta.1",
		},
		{
			name: "beta multi digit",
			in:   "v1beta12",
			want: "v1.0.0-beta.12",
		},
		{
			name: "empty",
			in:   "",
			want: "",
		},
		{
			name: "v only",
			in:   "v",
			want: "",
		},
		{
			name: "no leading digits",
			in:   "alpha1",
			want: "",
		},
		{
			name: "alpha missing number",
			in:   "v1alpha",
			want: "",
		},
		{
			name: "beta missing number",
			in:   "v1beta",
			want: "",
		},
		{
			name: "alpha with trailing suffix",
			in:   "v1alpha1foo",
			want: "",
		},
		{
			name: "beta with trailing suffix",
			in:   "v1beta1foo",
			want: "",
		},
		{
			name: "unsupported rc",
			in:   "v1rc1",
			want: "",
		},
		{
			name: "unsupported gamma",
			in:   "v1gamma1",
			want: "",
		},
		{
			name: "full semver is not normalized here",
			in:   "v1.2.3",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeGroupVersion(tt.in))
		})
	}
}

func TestCutDigits(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		wantDigits string
		wantRest   string
		wantOK     bool
	}{
		{
			name:       "digits only",
			in:         "123",
			wantDigits: "123",
			wantRest:   "",
			wantOK:     true,
		},
		{
			name:       "digits then letters",
			in:         "123alpha1",
			wantDigits: "123",
			wantRest:   "alpha1",
			wantOK:     true,
		},
		{
			name:       "digits then symbols",
			in:         "123.4",
			wantDigits: "123",
			wantRest:   ".4",
			wantOK:     true,
		},
		{
			name:       "starts with letter",
			in:         "alpha1",
			wantDigits: "",
			wantRest:   "alpha1",
			wantOK:     false,
		},
		{
			name:       "empty",
			in:         "",
			wantDigits: "",
			wantRest:   "",
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			digits, rest, ok := cutDigits(tt.in)

			assert.Equal(t, tt.wantDigits, digits)
			assert.Equal(t, tt.wantRest, rest)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

func TestVersionCompatible(t *testing.T) {
	tests := []struct {
		name       string
		v          Version
		other      Version
		compatible bool
	}{
		{
			name:       "same stable kubernetes API version",
			v:          Version("v1"),
			other:      Version("v1"),
			compatible: true,
		},
		{
			name:       "stable v1 compatible with full semver v1.0.0",
			v:          Version("v1"),
			other:      Version("v1.0.0"),
			compatible: true,
		},
		{
			name:       "full semver v1.0.0 compatible with stable v1",
			v:          Version("v1.0.0"),
			other:      Version("v1"),
			compatible: true,
		},
		{
			name:       "stable v1 compatible with minor v1.1.0",
			v:          Version("v1"),
			other:      Version("v1.1.0"),
			compatible: true,
		},
		{
			name:       "minor v1.1.0 compatible with stable v1",
			v:          Version("v1.1.0"),
			other:      Version("v1"),
			compatible: true,
		},
		{
			name:       "same stable major with patch difference",
			v:          Version("v1.2.3"),
			other:      Version("v1.9.9"),
			compatible: true,
		},
		{
			name:       "stable version without v prefix compatible",
			v:          Version("1.2.3"),
			other:      Version("v1"),
			compatible: true,
		},
		{
			name:       "different stable majors are incompatible",
			v:          Version("v2"),
			other:      Version("v1"),
			compatible: false,
		},
		{
			name:       "same alpha API version",
			v:          Version("v1alpha1"),
			other:      Version("v1alpha1"),
			compatible: true,
		},
		{
			name:       "alpha shorthand compatible with normalized prerelease",
			v:          Version("v1alpha1"),
			other:      Version("v1.0.0-alpha.1"),
			compatible: true,
		},
		{
			name:       "normalized prerelease compatible with alpha shorthand",
			v:          Version("v1.0.0-alpha.1"),
			other:      Version("v1alpha1"),
			compatible: true,
		},
		{
			name:       "different alpha API versions are incompatible",
			v:          Version("v1alpha2"),
			other:      Version("v1alpha1"),
			compatible: false,
		},
		{
			name:       "same beta API version",
			v:          Version("v1beta1"),
			other:      Version("v1beta1"),
			compatible: true,
		},
		{
			name:       "beta shorthand compatible with normalized prerelease",
			v:          Version("v1beta1"),
			other:      Version("v1.0.0-beta.1"),
			compatible: true,
		},
		{
			name:       "different beta API versions are incompatible",
			v:          Version("v1beta2"),
			other:      Version("v1beta1"),
			compatible: false,
		},
		{
			name:       "alpha and beta are incompatible",
			v:          Version("v1beta1"),
			other:      Version("v1alpha1"),
			compatible: false,
		},
		{
			name:       "beta and stable are incompatible",
			v:          Version("v1beta1"),
			other:      Version("v1"),
			compatible: false,
		},
		{
			name:       "stable and beta are incompatible",
			v:          Version("v1"),
			other:      Version("v1beta1"),
			compatible: false,
		},
		{
			name:       "different group prefix for same stable API",
			v:          Version("azctx.lvlcn-t.dev/v1"),
			other:      Version("example.com/v1.2.3"),
			compatible: false,
		},
		{
			name:       "different group prefix for same alpha API",
			v:          Version("azctx.lvlcn-t.dev/v1alpha1"),
			other:      Version("example.com/v1.0.0-alpha.1"),
			compatible: false,
		},
		{
			name:       "same group prefix for different API versions incompatible",
			v:          Version("azctx.lvlcn-t.dev/v1alpha1"),
			other:      Version("azctx.lvlcn-t.dev/v1beta1"),
			compatible: false,
		},
		{
			name:       "same group domain but different group path incompatible",
			v:          Version("azctx.lvlcn-t.dev/apis/v1"),
			other:      Version("azctx.lvlcn-t.dev/v1.1.0"),
			compatible: false,
		},
		{
			name:       "invalid receiver is incompatible",
			v:          Version("not-a-version"),
			other:      Version("v1"),
			compatible: false,
		},
		{
			name:       "invalid other is incompatible",
			v:          Version("v1"),
			other:      Version("not-a-version"),
			compatible: false,
		},
		{
			name:       "both invalid are incompatible",
			v:          Version("not-a-version"),
			other:      Version("also-not-a-version"),
			compatible: false,
		},
		{
			name:       "empty receiver is incompatible",
			v:          Version(""),
			other:      Version("v1"),
			compatible: false,
		},
		{
			name:       "empty other is incompatible",
			v:          Version("v1"),
			other:      Version(""),
			compatible: false,
		},
		{
			name:       "rc prerelease compatible with same rc prerelease",
			v:          Version("v1.0.0-rc.0"),
			other:      Version("v1.0.0-rc.0"),
			compatible: true,
		},
		{
			name:       "rc prerelease incompatible with stable",
			v:          Version("v1.0.0-rc.0"),
			other:      Version("v1"),
			compatible: false,
		},
		{
			name:       "different rc prereleases are incompatible",
			v:          Version("v1.0.0-rc.1"),
			other:      Version("v1.0.0-rc.0"),
			compatible: false,
		},
		{
			name:       "v0 stable shorthand compatible with v0.0.0",
			v:          Version("v0"),
			other:      Version("v0.0.0"),
			compatible: true,
		},
		{
			name:       "v0 stable shorthand incompatible with v0.1.0",
			v:          Version("v0"),
			other:      Version("v0.1.0"),
			compatible: false,
		},
		{
			name:       "v0 minor versions are incompatible",
			v:          Version("v0.1.0"),
			other:      Version("v0.2.0"),
			compatible: false,
		},
		{
			name:       "same v0 minor patch versions are incompatible unless exact",
			v:          Version("v0.1.1"),
			other:      Version("v0.1.0"),
			compatible: false,
		},
		{
			name:       "same v0 full version compatible",
			v:          Version("v0.1.0"),
			other:      Version("v0.1.0"),
			compatible: true,
		},
		{
			name:       "v0 alpha shorthand compatible with normalized prerelease",
			v:          Version("v0alpha1"),
			other:      Version("v0.0.0-alpha.1"),
			compatible: true,
		},
		{
			name:       "v0 alpha incompatible with v0 stable",
			v:          Version("v0alpha1"),
			other:      Version("v0"),
			compatible: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.compatible, tt.v.Compatible(tt.other))
		})
	}
}
