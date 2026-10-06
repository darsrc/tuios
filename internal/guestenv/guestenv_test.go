package guestenv

import (
	"slices"
	"testing"
)

// TestWithoutHostMultiplexerDropsOuterDartuiosPane covers a daemon started from a
// dartuios pane: its panes must not inherit the outer pane's session, socket and
// ids, which would place them in the outer session.
func TestWithoutHostMultiplexerDropsOuterDartuiosPane(t *testing.T) {
	env := []string{
		"HOME=/home/u",
		"DARTUIOS_SESSION=outer", "DARTUIOS_SOCKET=/run/outer.sock", "DARTUIOS_PANE_ID=w1",
		"DARTUIOS_WINDOW_ID=w1", "DARTUIOS_PANE_TOKEN=t", "DARTUIOS_PANE_GRANTS=g",
		"DARTUIOS_RESTORED=1", "DARTUIOS_PANE_TTY=/dev/pts/9", "DARTUIOS_SESSION_REMOTE=r", "DARTUIOS_PANE_HOSTED=1",
		"DARTUIOS_ENV=1", "TMUX=/tmp/tmux",
	}
	got := WithoutHostMultiplexer(slices.Clone(env))
	want := []string{"HOME=/home/u", "DARTUIOS_ENV=1"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
