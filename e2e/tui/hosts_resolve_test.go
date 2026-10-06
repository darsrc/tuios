package tuie2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gaurav-Gosain/tuitest"
)

// Finding dartuios on a host where it is not on the PATH ssh gives a command.
//
// This is the failure the maintainer hit on the first host he added: dartuios was
// installed by the project's own install script, in ~/.local/bin, and the
// link said "command not found" because a non-interactive ssh shell does not
// read the profile that adds that directory. Everything here is real: the CLI
// binary, the config file, the daemon, the subprocess transport, the probe
// script parsed by a real sh, and the proxy. The only stand-in is ssh, and it
// is the one from sidebar_hosts_test.go with a far side of this test's own: a
// HOME under the temp dir, and a PATH holding sh and nothing else, so nothing
// on the developer's machine can be found by name. Nothing reads the
// developer's ssh configuration, and no network connection is made.

// writeFakeSSHWithFarHome puts an ssh stand-in in dir whose far side has the
// given HOME and a PATH with only sh on it.
func writeFakeSSHWithFarHome(t *testing.T, dir, home string) string {
	t.Helper()
	shOnly := filepath.Join(dir, "sh-only-path")
	if err := os.MkdirAll(shOnly, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Symlink("/bin/sh", filepath.Join(shOnly, "sh")); err != nil {
		t.Fatalf("symlink sh: %v", err)
	}
	path := filepath.Join(dir, "fake-ssh-far")
	script := "#!/bin/sh\n" +
		"while [ $# -gt 0 ]; do case \"$1\" in -o) shift 2 ;; -T|-t) shift ;; *) break ;; esac; done\n" +
		"shift\n" + // the address
		"export HOME=" + home + " PATH=" + shOnly + " SHELL=/bin/sh\n" +
		"exec /bin/sh -c \"$*\"\n"
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil { //nolint:gosec // an ssh stand-in this test runs
		t.Fatalf("write the ssh stand-in: %v", err)
	}
	return path
}

// TestHostLinkFindsDartuiosOutsideThePath adds a host with no --command whose
// dartuios is only in ~/.local/bin, and proves the add, the running daemon's own
// link and 'dartuios hosts test' all find it and say where.
func TestHostLinkFindsDartuiosOutsideThePath(t *testing.T) {
	base := t.TempDir()
	farHome := filepath.Join(base, "far-home")
	installed := filepath.Join(farHome, ".local", "bin", "dartuios")
	if err := os.MkdirAll(filepath.Dir(installed), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// The far dartuios is this build. It reaches this test's own daemon because
	// the XDG directories are inherited through the stand-in.
	if err := os.Symlink(dartuiosBin, installed); err != nil {
		t.Fatalf("install the far dartuios: %v", err)
	}
	ssh := writeFakeSSHWithFarHome(t, base, farHome)
	env := []string{"DARTUIOS_SSH=" + ssh}

	term := startIn(t, base, startOpts{args: []string{"new", "fed-resolve"}, env: env})
	waitBoot(t, term)

	// No --command. This is the command that failed before.
	out, err := dartuiosCLIEnv(t, base, env, "hosts", "add", "far", "someone@farbox", "--connect-timeout", "5")
	if err != nil {
		t.Fatalf("ASSERTION: 'dartuios hosts add' failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "The link runs "+installed+" on the host.") {
		t.Errorf("ASSERTION: the add did not say which dartuios the link found:\n%s", out)
	}
	t.Logf("dartuios hosts add:\n%s", out)

	// The daemon's own link, over its own ssh, has to find it too.
	listing := waitForHostListing(t, base, func(s string) bool {
		return strings.Contains(s, "far") && strings.Contains(s, "up")
	}, "the running daemon's link never found the dartuios in ~/.local/bin")
	t.Logf("dartuios hosts:\n%s", listing)

	out, err = dartuiosCLIEnv(t, base, env, "hosts", "test", "far")
	if err != nil {
		t.Fatalf("ASSERTION: 'dartuios hosts test' failed against a host whose dartuios is in ~/.local/bin: %v\n%s", err, out)
	}
	if !strings.Contains(out, "The link runs "+installed+" on the host.") {
		t.Errorf("ASSERTION: 'dartuios hosts test' did not report the path it found:\n%s", out)
	}
	t.Logf("dartuios hosts test:\n%s", out)
}

// TestHostTestSaysWhereItLookedForDartuios is the message that replaces "command
// not found": a host with no dartuios anywhere is reported as that, with every
// place the link looked and the flag that points it somewhere else.
func TestHostTestSaysWhereItLookedForDartuios(t *testing.T) {
	// The far HOME and PATH are this test's own, but the fixed list also
	// names system directories, and a dartuios in one of them would be found.
	for _, p := range []string{"/usr/local/bin/dartuios", "/usr/bin/dartuios", "/opt/homebrew/bin/dartuios",
		"/home/linuxbrew/.linuxbrew/bin/dartuios", "/nix/var/nix/profiles/default/bin/dartuios", "/run/current-system/sw/bin/dartuios"} {
		if _, err := os.Stat(p); err == nil {
			t.Skipf("dartuios is installed at %s on this machine, so the probe would find it", p)
		}
	}
	base := t.TempDir()
	farHome := filepath.Join(base, "far-home")
	if err := os.MkdirAll(farHome, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	env := []string{"DARTUIOS_SSH=" + writeFakeSSHWithFarHome(t, base, farHome)}

	out, err := dartuiosCLIEnv(t, base, env, "hosts", "add", "bare", "someone@barebox", "--connect-timeout", "5")
	if err != nil {
		t.Fatalf("ASSERTION: 'dartuios hosts add' failed; a host with no dartuios must still be added: %v\n%s", err, out)
	}
	t.Logf("dartuios hosts add:\n%s", out)

	out, err = dartuiosCLIEnv(t, base, env, "hosts", "test", "bare")
	if err == nil {
		t.Errorf("ASSERTION: 'dartuios hosts test' succeeded against a host with no dartuios:\n%s", out)
	}
	for _, want := range []string{
		"no_dartuios",
		"The link cannot find dartuios on the host.",
		"~/.local/bin/dartuios",
		"~/go/bin/dartuios",
		"/usr/local/bin/dartuios",
		"--command PATH",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("ASSERTION: 'dartuios hosts test' output lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "command not found") {
		t.Errorf("ASSERTION: the shell's own 'command not found' is still what the user reads:\n%s", out)
	}
	t.Logf("dartuios hosts test against a host with no dartuios:\n%s", out)
}

// TestAttachOnAHostWithSSHFindsDartuiosOutsideThePath keeps the ssh fallback on
// the same footing as the link: with no --command and dartuios only in
// ~/.local/bin, 'dartuios attach --host NAME SESSION --ssh' finds it and runs it.
func TestAttachOnAHostWithSSHFindsDartuiosOutsideThePath(t *testing.T) {
	base := t.TempDir()
	farHome := filepath.Join(base, "far-home")
	installed := filepath.Join(farHome, ".local", "bin", "dartuios")
	if err := os.MkdirAll(filepath.Dir(installed), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Symlink(dartuiosBin, installed); err != nil {
		t.Fatalf("install the far dartuios: %v", err)
	}
	env := []string{"DARTUIOS_SSH=" + writeFakeSSHWithFarHome(t, base, farHome)}

	dir := filepath.Join(base, "XDG_CONFIG_HOME", "dartuios")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	body := "[hosts.far]\naddr = \"someone@farbox\"\nconnect_timeout = 5\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if out, err := dartuiosCLIEnv(t, base, env, "new", "ssh-far-target", "--detach"); err != nil {
		t.Fatalf("create the session to open: %v\n%s", err, out)
	}

	term := startIn(t, base, startOpts{args: []string{"attach", "--host", "far", "ssh-far-target", "--ssh"}, env: env})
	if err := term.WaitFor(func(s tuitest.Screen) bool {
		return strings.Contains(s.Text(), "╰──")
	}, bootTimeout); err != nil {
		t.Fatalf("ASSERTION: the far dartuios in ~/.local/bin never drew the session over --ssh: %v\n%s", err, term.Snapshot())
	}

	// The proof that the probe ran the dartuios it found: a plain attach process
	// whose program is the far install, not this test's binary by name.
	lines := commandLinesContaining(t, "attach ssh-far-target")
	found := false
	for _, line := range lines {
		if strings.Contains(line, installed+" attach ssh-far-target") {
			found = true
		}
	}
	if !found {
		t.Fatalf("ASSERTION: --ssh did not run the dartuios found at %s:\n%s", installed, strings.Join(lines, "\n"))
	}
	alive(t, term, "after opening a session over ssh with a found dartuios")
}
