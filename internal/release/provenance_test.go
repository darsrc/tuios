package release

import "testing"

// Detect decides whether a binary is ours to overwrite, and getting it wrong in
// the permissive direction corrupts a package manager's view of the world
// silently. Every row here is a real install path taken from the project's own
// packaging.
func TestDetect(t *testing.T) {
	cases := []struct {
		name        string
		facts       Facts
		want        Origin
		replaceable bool
	}{
		{
			name:        "curl script into /usr/local/bin",
			facts:       Facts{Path: "/usr/local/bin/dartuios", BuiltBy: "goreleaser", Version: "v0.7.0"},
			want:        OriginRelease,
			replaceable: true,
		},
		{
			name:        "curl script into a home directory",
			facts:       Facts{Path: "/home/x/.local/bin/dartuios", BuiltBy: "goreleaser", Version: "v0.7.0"},
			want:        OriginRelease,
			replaceable: true,
		},
		{
			// The AUR package installs the goreleaser binary into /usr/bin, so
			// its build stamp is identical to the curl script's and only the
			// path says a package manager owns it. This is the row that fails
			// if the build stamp is consulted before the path.
			name:  "AUR package in /usr/bin",
			facts: Facts{Path: "/usr/bin/dartuios", BuiltBy: "goreleaser", Version: "v0.7.0"},
			want:  OriginSystemPackage,
		},
		{
			name:  "nix store",
			facts: Facts{Path: "/nix/store/abc123-dartuios-0.7.0/bin/dartuios", BuiltBy: "unknown", Version: "v0.7.0"},
			want:  OriginNixStore,
		},
		{
			name:  "homebrew cellar on apple silicon",
			facts: Facts{Path: "/opt/homebrew/Cellar/dartuios/0.7.0/bin/dartuios", BuiltBy: "goreleaser", GOOS: "darwin"},
			want:  OriginHomebrew,
		},
		{
			name:  "homebrew cask",
			facts: Facts{Path: "/opt/homebrew/Caskroom/dartuios/0.7.0/dartuios", BuiltBy: "goreleaser", GOOS: "darwin"},
			want:  OriginHomebrew,
		},
		{
			// Intel macOS puts the Homebrew prefix at /usr/local, which is also
			// the curl script's first choice. Only the environment separates
			// them, and only for <prefix>/bin.
			name: "homebrew prefix on intel",
			facts: Facts{
				Path: "/usr/local/bin/dartuios", BuiltBy: "goreleaser",
				GOOS: "darwin", BrewPrefix: "/usr/local",
			},
			want: OriginHomebrew,
		},
		{
			name:  "linuxbrew",
			facts: Facts{Path: "/home/linuxbrew/.linuxbrew/bin/dartuios", BuiltBy: "goreleaser"},
			want:  OriginHomebrew,
		},
		{
			name:  "built from source by scripts/install.sh",
			facts: Facts{Path: "/home/x/.local/bin/dartuios", BuiltBy: "install.sh", Version: "dev+abc123def456"},
			want:  OriginSourceScript,
		},
		{
			name: "go install",
			facts: Facts{
				Path: "/home/x/go/bin/dartuios", BuiltBy: "unknown",
				Version: "dev", ModuleVersion: "v0.7.0",
			},
			want: OriginGoInstall,
		},
		{
			// A local `go build` stamps the module version "(devel)", which is
			// not a `go install` and not a release either.
			name: "local go build",
			facts: Facts{
				Path: "/home/x/dartuios/dartuios", BuiltBy: "unknown",
				Version: "dev", ModuleVersion: "(devel)",
			},
			want: OriginUnknown,
		},
		{
			name:  "distroless container",
			facts: Facts{Path: "/dartuios", BuiltBy: "unknown", Version: "dev"},
			want:  OriginUnknown,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Detect(tc.facts)
			if got.Origin != tc.want {
				t.Errorf("Origin = %v, want %v", got.Origin, tc.want)
			}
			if got.Replaceable != tc.replaceable {
				t.Errorf("Replaceable = %v, want %v", got.Replaceable, tc.replaceable)
			}
		})
	}
}

// TestOnlyAReleaseBuildIsReplaceable, stated as its own claim so a new Origin
// added later cannot become replaceable by accident.
//
// Negative control: set Replaceable on any other branch and this fails.
func TestOnlyAReleaseBuildIsReplaceable(t *testing.T) {
	for _, f := range []Facts{
		{Path: "/usr/local/bin/dartuios", BuiltBy: "goreleaser"},
		{Path: "/usr/bin/dartuios", BuiltBy: "goreleaser"},
		{Path: "/nix/store/x/bin/dartuios", BuiltBy: "goreleaser"},
		{Path: "/home/x/.local/bin/dartuios", BuiltBy: "install.sh"},
		{Path: "/tmp/dartuios", BuiltBy: "unknown"},
	} {
		p := Detect(f)
		if p.Replaceable != (p.Origin == OriginRelease) {
			t.Errorf("%s: Origin=%v but Replaceable=%v", f.Path, p.Origin, p.Replaceable)
		}
	}
}

// TestHomebrewFixNamesFormulaOrCask. `brew install dartuios` is the homebrew-core
// formula, which lives in the Cellar; the tap ships a cask, which lives in the
// Caskroom. `brew upgrade --cask dartuios` on a formula install fails, so the
// command printed has to follow the path.
//
// Negative control: return "brew upgrade --cask dartuios" for every Homebrew path
// and the two Cellar rows fail.
func TestHomebrewFixNamesFormulaOrCask(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"/opt/homebrew/Cellar/dartuios/0.8.0/bin/dartuios", "brew upgrade dartuios"},
		{"/home/linuxbrew/.linuxbrew/Cellar/dartuios/0.8.0/bin/dartuios", "brew upgrade dartuios"},
		{"/opt/homebrew/Caskroom/dartuios/0.8.0/dartuios", "brew upgrade --cask dartuios"},
	}
	for _, tc := range cases {
		p := Detect(Facts{Path: tc.path, BuiltBy: "goreleaser", GOOS: "darwin"})
		if p.Origin != OriginHomebrew {
			t.Fatalf("%s: Origin = %v, want Homebrew", tc.path, p.Origin)
		}
		if p.Fix != tc.want {
			t.Errorf("%s: Fix = %q, want %q", tc.path, p.Fix, tc.want)
		}
	}
}
