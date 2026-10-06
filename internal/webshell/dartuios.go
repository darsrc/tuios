package webshell

import (
	"path"
	"sort"
	"strings"
)

// cmdDartuios is the dartuios command inside the demo. The one subcommand that does
// something is tape play, which hands a tape to the dartuios this shell runs in.
// The rest say what they would do on a real machine.
func cmdDartuios(s *shell, args []string, _ string) int {
	t := s.t
	sub := argOr(args, 1, "")
	switch sub {
	case "", "help", "--help", "-h":
		t.Print("You are already in dartuios. Press " + bold + "Ctrl+B" + reset + " then " + bold + "?" + reset + " to see every key.\r\n\r\n")
		t.Print("In this shell:\r\n")
		t.Print("  " + green + "dartuios tape play demo.tape" + reset + "  watch dartuios drive itself\r\n")
		t.Print("  " + green + "dartuios tape list" + reset + "            the tapes here\r\n")
		t.Print("  " + green + "dartuios version" + reset + "\r\n")
		t.Print("\r\n" + dim + "These show what they print on a real machine:" + reset + "\r\n")
		t.Print("  " + green + "dartuios ls" + reset + ", " + green + "dartuios fan" + reset + ", " + green + "dartuios worktree" + reset + ", " +
			green + "dartuios list-agents" + reset + ", " + green + "dartuios list-verbs" + reset + ", " + green + "dartuios list-hooks" + reset + "\r\n")
		return 0
	case "version", "--version", "-v":
		t.Print("dartuios (browser demo, the real thing compiled to WebAssembly)\r\n")
		return 0
	case "tape":
		return dartuiosTape(s, args[2:])
	}
	if printSample(t, sub) {
		return 0
	}
	t.Print(yellow + "dartuios " + sub + reset + " needs a real machine. Install dartuios to try it: " + bold + "dartuios.dev" + reset + "\r\n")
	return 0
}

func dartuiosTape(s *shell, args []string) int {
	t := s.t
	sub := argOr(args, 0, "")
	switch sub {
	case "play", "run":
		if len(args) < 2 {
			return s.fail("Which tape? Try " + bold + "dartuios tape play demo.tape" + reset)
		}
		p := resolve(s.cwd, args[1])
		script, ok := readFile(p)
		if !ok {
			return s.fail("dartuios: no tape at " + args[1] + ". Try " + bold + "dartuios tape list" + reset)
		}
		name := path.Base(p)
		t.Print(green + ">" + reset + " Playing " + bold + name + reset + ". Sit back.\r\n")
		t.Emit(EventTapePlay, map[string]any{"name": name, "script": script})
		return 0
	case "list", "ls", "":
		names := tapeFiles()
		if len(names) == 0 {
			t.Print("No tapes here.\r\n")
			return 0
		}
		for _, n := range names {
			t.Print("  " + green + prettyPath(n) + reset + "\r\n")
		}
		t.Print(dim + "Play one with: dartuios tape play <file>" + reset + "\r\n")
		return 0
	}
	t.Print(yellow + "dartuios tape " + sub + reset + " needs a real machine. In the demo, try " + bold + "dartuios tape play demo.tape" + reset + "\r\n")
	return 0
}

// tapeFiles lists every .tape file in the filesystem.
func tapeFiles() []string {
	fsMu.RLock()
	defer fsMu.RUnlock()
	var out []string
	for p := range files {
		if strings.HasSuffix(p, ".tape") {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}
