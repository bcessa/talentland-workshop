package internal

import (
	"fmt"
	"net/http"
	"path"
	"runtime/debug"
	"strings"
	"time"
)

// VersionInfo collects information about the application at build time.
// This information can be used to distinguish different app releases.
type VersionInfo struct {
	// Main application identifier.
	Name string `json:"name,omitempty"`

	// Semantic version of the build. If not available a short version of
	// the commit hash will be used.
	Version string `json:"version,omitempty"`

	// Provides the commit identifier used to build the binary.
	BuildCode string `json:"build_code,omitempty"`

	// Provides the UNIX timestamp of the build.
	BuildDate string `json:"build_date,omitempty"`

	// Where to look for documentation and further information.
	Home string `json:"home,omitempty"`

	// Release identifier for the application. A release identifier is of the
	// form: `app-name@version+commit_hash`. If `version` or `commit_hash`
	// are not available will be omitted.
	Release string `json:"release,omitempty"`

	// Operating system target: one of darwin, freebsd, linux, and so on. To
	// view possible combinations of GOOS and GOARCH, run "go tool dist list".
	OS string `json:"os"`

	// Architecture target: one of 386, amd64, arm, s390x, and so on.
	Arch string `json:"arch"`

	// Go version used to build the application.
	GoVersion string `json:"go"`
}

// BuildDetails returns the version information for the application.
func BuildDetails() *VersionInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}
	settings := toMap(info.Settings)
	vi := &VersionInfo{
		Home:      info.Main.Path,
		Version:   info.Main.Version,
		BuildCode: settings["vcs.revision"],
		BuildDate: settings["vcs.time"],
		OS:        settings["GOOS"],
		Arch:      settings["GOARCH"],
		GoVersion: info.GoVersion,
	}
	if vi.Home != "" {
		vi.Name = path.Base(vi.Home)
	}
	if vi.Name != "" {
		vi.Release = releaseCode(vi)
	}
	return vi
}

// Values returns a map with the version information for the application.
func (vi *VersionInfo) Values() map[string]string {
	kv := map[string]string{
		"Name":       vi.Name,
		"Home":       vi.Home,
		"Version":    vi.Version,
		"Build Code": vi.BuildCode,
		"OS/Arch":    fmt.Sprintf("%s/%s", vi.OS, vi.Arch),
		"Go Version": vi.GoVersion,
		"Release":    vi.Release,
	}
	if vi.BuildDate != "" {
		st, err := time.Parse(time.RFC3339, vi.BuildDate)
		if err == nil {
			kv["Release Date"] = st.Format(time.RFC822)
		}
	}
	return kv
}

// Middleware returns a standard HTTP middleware that injects the version
// information as headers in the response. The following headers will be
// added to the response:
//   - x-build-code: commit hash used to build the binary
//   - x-app-version: semantic version of the build
func (vi *VersionInfo) Middleware() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("x-build-code", vi.BuildCode)
			w.Header().Set("x-app-version", vi.Version)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func releaseCode(vi *VersionInfo) string {
	// service name
	release := vi.Name

	// attach version tag
	if strings.Count(vi.Version, ".") >= 2 {
		release = fmt.Sprintf("%s@%s", release, vi.Version)
	}

	// attach commit hash if available
	if vi.BuildCode != "" {
		release = fmt.Sprintf("%s+%s", release, vi.BuildCode)
	}
	return release
}

func toMap(s []debug.BuildSetting) map[string]string {
	m := make(map[string]string)
	for _, k := range s {
		m[k.Key] = k.Value
	}
	return m
}
