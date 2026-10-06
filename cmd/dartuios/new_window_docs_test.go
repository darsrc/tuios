package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// dockCommandPattern finds a dartuios command in a TOML string value, such as a
// dock component's on-click.
var dockCommandPattern = regexp.MustCompile(`"(dartuios [^"]*)"`)

// TestDockExampleCommandsResolve checks the dartuios commands the dock examples
// run. CONFIGURATION.md and examples/dock/dock.toml had
// "dartuios new-window log git log --oneline -20", which fails because dartuios reads
// --oneline as its own flag. The argv after the name needs a -- in front.
func TestDockExampleCommandsResolve(t *testing.T) {
	checked := 0
	for _, path := range []string{"../../docs/CONFIGURATION.md", "../../examples/dock/dock.toml"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, match := range dockCommandPattern.FindAllStringSubmatch(string(data), -1) {
			for _, words := range shellCalls(match[1]) {
				if len(words) < 2 || words[0] != "dartuios" {
					continue
				}
				checked++
				resolveDartuiosCommand(t, path, words[1:])
			}
		}
	}
	if checked == 0 {
		t.Fatal("found no dartuios commands in the dock examples")
	}
}

// resolveDartuiosCommand resolves one command line against the real command tree
// the way TestSkillCommandsResolve does.
func resolveDartuiosCommand(t *testing.T, where string, args []string) {
	t.Helper()
	line := strings.Join(args, " ")
	cmd, rest, err := newRootCommand().Find(args)
	if err != nil {
		t.Errorf("%s: dartuios %s: no such command: %v", where, line, err)
		return
	}
	if err := cmd.ParseFlags(rest); err != nil {
		t.Errorf("%s: dartuios %s: flags rejected: %v", where, line, err)
		return
	}
	if err := cmd.ValidateArgs(cmd.Flags().Args()); err != nil {
		t.Errorf("%s: dartuios %s: arguments rejected: %v", where, line, err)
	}
}
