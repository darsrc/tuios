#!/bin/sh
# dartuios Claude Code agent-state shim, kept for settings that already point here.
#
# The reporter is built into dartuios now: `dartuios agent-hook claude-code` reads the
# hook payload on stdin, maps the event (see docs/AGENT_STATE.md), finds the
# pane, and reports. It needs no python3, and `dartuios integration install
# claude-code` wires it into settings.json without this file.
#
# This shim only hands the payload over. It exits 0 when dartuios is not on PATH,
# so it stays safe to leave wired up.

command -v dartuios >/dev/null 2>&1 || exit 0
exec dartuios agent-hook claude-code
