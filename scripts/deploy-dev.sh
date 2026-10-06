#!/bin/sh
# Build this checkout and put it on this machine and on every host in the
# [hosts] config table, then restart each daemon so the new code is running.
#
# Sessions come back through resurrection: the daemon writes its state before
# it stops and rebuilds the layout on start. The shells under them are new, so
# anything still running in a pane ends. That is the cost of a restart and it
# is why this is a script you run rather than something that happens on a save.
#
# It also installs a second binary, dartuios-ghostty, built against the
# libghostty-vt emulator instead of the pure Go one, so the same checkout can be
# tried on either. That one needs zig and is skipped where zig is missing, which
# is every host that only receives a cross-compiled binary: the emulator is a
# static archive built for the machine it runs on, and there is no cross build
# of it here.
#
# The two do not share a daemon. A pane's emulator lives in the daemon that owns
# the pty, so a ghostty client against the pure daemon tests the pure emulator
# and proves nothing. Run it standalone instead, which is a session with no
# daemon at all and therefore its own emulator:
#
#   dartuios-ghostty --standalone
#
# Usage: scripts/deploy-dev.sh [host ...]
#   With no arguments it does this machine and the hosts listed below.
set -e

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
INSTALL_DIR="${DARTUIOS_INSTALL_DIR:-$HOME/.local/bin}"
BIN="$INSTALL_DIR/dartuios"
GHOSTTY_BIN="$INSTALL_DIR/dartuios-ghostty"
HOSTS="${*:-ente forgejo}"

echo "building"
go build -o /tmp/dartuios-local ./cmd/dartuios
GOOS=linux GOARCH=amd64 go build -o /tmp/dartuios-linux ./cmd/dartuios

# The ghostty backend, when this machine can build it. Built before anything is
# taken down, so a failure here costs nothing that is already running.
ghostty_built=no
if command -v zig >/dev/null 2>&1; then
	echo "building the ghostty backend"
	"$ROOT/scripts/ghostty-lib.sh" native >/dev/null
	PKG_CONFIG_PATH="$ROOT/.ghostty-vt/native/pkgconfig" \
		go build -tags ghostty -o /tmp/dartuios-ghostty ./cmd/dartuios
	ghostty_built=yes
else
	echo "skipping the ghostty backend: zig not found"
fi

# The far side first, so a failure there happens before this machine's daemon
# is taken down.
for h in $HOSTS; do
	echo "deploying to $h"
	scp -q /tmp/dartuios-linux "$h:/tmp/dartuios-new"
	# The bracket keeps the remote shell's own command line from matching the
	# pattern, which would make pkill kill the shell running it.
	ssh -o BatchMode=yes "$h" '
		set -e
		install -m 755 /tmp/dartuios-new "$HOME/.local/bin/dartuios"
		rm -f /tmp/dartuios-new
		pkill -f "dartuios[ ]daemon" || true
		sleep 1
		setsid "$HOME/.local/bin/dartuios" daemon >/tmp/dartuios-daemon.log 2>&1 </dev/null &
		sleep 1
	'
	echo "  $(ssh -o BatchMode=yes "$h" '~/.local/bin/dartuios --version | head -1')"
done

echo "installing here"
install -m 755 /tmp/dartuios-local "$BIN"
pkill -f "dartuios[ ]daemon" || true
sleep 1
"$BIN" daemon >/tmp/dartuios-daemon.log 2>&1 </dev/null &
sleep 1
echo "  $("$BIN" --version | head -1)"

if [ "$ghostty_built" = yes ]; then
	# No daemon is started for it and none is stopped: it is the same code with
	# a different emulator under it, and the one daemon on this machine belongs
	# to the binary above.
	install -m 755 /tmp/dartuios-ghostty "$GHOSTTY_BIN"
	echo "  $("$GHOSTTY_BIN" --version | head -1)  (run: dartuios-ghostty --standalone)"
fi
echo "done"
