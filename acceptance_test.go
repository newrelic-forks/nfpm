// +build acceptance

package nfpm_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goreleaser/nfpm/v2"
	_ "github.com/goreleaser/nfpm/v2/apk"
	_ "github.com/goreleaser/nfpm/v2/deb"
	_ "github.com/goreleaser/nfpm/v2/rpm"
)

// nolint: gochecknoglobals
var formatArchs = map[string][]string{
	"apk": []string{"amd64", "arm64", "386", "ppc64le"},
	"deb": []string{"amd64", "arm64", "ppc64le"},
	"rpm": []string{"amd64", "arm64", "ppc64le"},
}

func TestCore(t *testing.T) {
	var testNames = []string{
		"min",
		"simple",
		"no-glob",
		"complex",
		"env-var-version",
		"overrides",
		"meta",
		"withchangelog",
		"symlink",
		"signed",
	}
	for _, name := range testNames {
		name := name
		for format, architecture := range formatArchs {
			format := format
			for _, arch := range architecture {
				arch := arch
				t.Run(fmt.Sprintf("%s/%s/%s", format, arch, name), func(t *testing.T) {
					t.Parallel()
					if arch == "ppc64le" && os.Getenv("NO_TEST_PPC64LE") == "true" {
						t.Skip("ppc64le arch not supported in pipeline")
					}
					accept(t, acceptParms{
						Name:   fmt.Sprintf("%s_%s", name, arch),
						Conf:   fmt.Sprintf("core.%s.yaml", name),
						Format: format,
						Docker: dockerParams{
							File:   fmt.Sprintf("%s.dockerfile", format),
							Target: name,
							Arch:   arch,
						},
					})
				})
			}
		}
	}
}

// func TestConfigNoReplace(t *testing.T) {
// 	var target = "./testdata/acceptance/tmp/noreplace_old_rpm.rpm"
// 	require.NoError(t, os.MkdirAll("./testdata/acceptance/tmp", 0700))
//
// 	config, err := nfpm.ParseFile("./testdata/acceptance/rpm.config-noreplace-old.yaml")
// 	require.NoError(t, err)
//
// 	info, err := config.Get("rpm")
// 	require.NoError(t, err)
// 	require.NoError(t, nfpm.Validate(info))
//
// 	pkg, err := nfpm.Get("rpm")
// 	require.NoError(t, err)
//
// 	f, err := os.Create(target)
// 	require.NoError(t, err)
// 	info.Target = target
// 	require.NoError(t, pkg.Package(nfpm.WithDefaults(info), f))
//
// 	t.Run("rpm", func(t *testing.T) {
// 		accept(t, acceptParms{
// 			Name:       "noreplace_rpm",
// 			Conf:       "rpm.config-noreplace.yaml",
// 			Format:     "rpm",
// 			Dockerfile: "rpm.config-noreplace.dockerfile",
// 		})
// 	})
// }

func TestCompression(t *testing.T) {
	format := "rpm"
	compressFormats := []string{"gzip", "xz", "lzma"}
	for _, arch := range formatArchs[format] {
		arch := arch
		for _, compFormat := range compressFormats {
			compFormat := compFormat
			t.Run(fmt.Sprintf("%s/%s/%s", format, arch, compFormat), func(t *testing.T) {
				if arch == "ppc64le" && os.Getenv("NO_TEST_PPC64LE") == "true" {
					t.Skip("ppc64le arch not supported in pipeline")
				}
				accept(t, acceptParms{
					Name:   fmt.Sprintf("%s_compression_rpm", compFormat),
					Conf:   fmt.Sprintf("rpm.%s.compression.yaml", compFormat),
					Format: format,
					Docker: dockerParams{
						File: fmt.Sprintf("%s.dockerfile", format),
						Target: "compression",
						Arch: arch,
						BuildArgs: []string{fmt.Sprintf("compression=%s", compFormat)},
					},
				})
			})
		}
	}
}

// func TestRPMRelease(t *testing.T) {
// 	accept(t, acceptParms{
// 		Name:       "release_rpm",
// 		Conf:       "rpm.release.yaml",
// 		Format:     "rpm",
// 		Dockerfile: "rpm.release.dockerfile",
// 	})
// }

// func TestDebRules(t *testing.T) {
// 	accept(t, acceptParms{
// 		Name:       "rules.deb",
// 		Conf:       "deb.rules.yaml",
// 		Format:     "deb",
// 		Dockerfile: "deb.rules.dockerfile",
// 	})
// }

// func TestDebTriggers(t *testing.T) {
// 	t.Run("triggers-deb", func(t *testing.T) {
// 		accept(t, acceptParms{
// 			Name:       "triggers-deb",
// 			Conf:       "deb.triggers.yaml",
// 			Format:     "deb",
// 			Dockerfile: "deb.triggers.dockerfile",
// 		})
// 	})
// }

// func TestDebBreaks(t *testing.T) {
// 	t.Run("breaks-deb", func(t *testing.T) {
// 		accept(t, acceptParms{
// 			Name:       "breaks-deb",
// 			Conf:       "deb.breaks.yaml",
// 			Format:     "deb",
// 			Dockerfile: "deb.breaks.dockerfile",
// 		})
// 	})
// }

type acceptParms struct {
	Name   string
	Conf   string
	Format string
	Docker dockerParams
}

type dockerParams struct {
	File      string
	Target    string
	Arch      string
	BuildArgs []string
}

type testWriter struct {
	*testing.T
}

func (t testWriter) Write(p []byte) (n int, err error) {
	t.Log(string(p))
	return len(p), nil
}

func accept(t *testing.T, params acceptParms) {
	var configFile = filepath.Join("./testdata/acceptance/", params.Conf)
	tmp, err := filepath.Abs("./testdata/acceptance/tmp")
	require.NoError(t, err)
	var packageName = params.Name + "." + params.Format
	var target = filepath.Join(tmp, packageName)
	t.Log("package: " + target)

	require.NoError(t, os.MkdirAll(tmp, 0700))

	os.Setenv("SEMVER", "v1.0.0-0.1.b1+git.abcdefgh")
	os.Setenv("BUILD_ARCH", params.Docker.Arch)
	config, err := nfpm.ParseFile(configFile)
	require.NoError(t, err)

	info, err := config.Get(params.Format)
	require.NoError(t, err)
	require.NoError(t, nfpm.Validate(info))

	pkg, err := nfpm.Get(params.Format)
	require.NoError(t, err)

	cmdArgs := []string{
		"build", "--rm", "--force-rm",
		"--platform", fmt.Sprintf("linux/%s", params.Docker.Arch),
		"-f", params.Docker.File,
		"--target", params.Docker.Target,
		"--build-arg", "package=" + filepath.Join("tmp", packageName),
	}
	for _, arg := range params.Docker.BuildArgs {
		cmdArgs = append(cmdArgs, "--build-arg", arg)
	}
	cmdArgs = append(cmdArgs, ".")

	f, err := os.Create(target)
	require.NoError(t, err)
	info.Target = target
	require.NoError(t, pkg.Package(nfpm.WithDefaults(info), f))
	//nolint:gosec
	cmd := exec.Command("docker", cmdArgs...)
	cmd.Dir = "./testdata/acceptance"
	cmd.Stderr = testWriter{t}
	cmd.Stdout = cmd.Stderr

	t.Log("will exec:", cmd.Args)
	require.NoError(t, cmd.Run())
}
