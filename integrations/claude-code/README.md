# Claude Code agent-state integration

The Claude Code reporter is built into dartuios. Install it with:

```sh
dartuios integration install claude-code
dartuios integration status claude-code   # installed, current (v1)
```

That writes one managed hook entry per event into `~/.claude/settings.json` (or
`$CLAUDE_CONFIG_DIR/settings.json`), each running
`dartuios agent-hook claude-code --integration 1`. Your own settings and hooks are
kept where they are, the file is replaced atomically (through a symlink, if
`settings.json` is one), and the file as it was before dartuios first touched it
is kept as `settings.json.dartuios.bak`. `dartuios integration uninstall claude-code`
removes exactly the entries dartuios wrote.

The event map, the pane lookup and the nested-session guard are documented in
[Agent state](../../docs/AGENT_STATE.md#harness-integrations).

## The old shim

`dartuios-agent-state.sh` in this directory used to be the integration: a shell
script that parsed the payload with python3 and mapped every `Notification` to
`needs_input`. It is now a two-line wrapper that runs `dartuios agent-hook
claude-code`, so settings that still point at it keep working and get the
corrected map.

Do not wire both. With the shim and an installed integration on the same events,
every event is reported twice. `dartuios integration status` and `dartuios doctor
agents` say so when they find the shim in `settings.json`. Remove the shim's
entries and keep the installed ones.

## Verifying by hand

You do not need Claude Code to check the wiring. From inside a dartuios pane:

```sh
echo '{"hook_event_name":"PermissionRequest","session_id":"s1","tool_name":"Bash","tool_input":{"command":"make"}}' \
  | dartuios agent-hook claude-code --explain
dartuios get-agent-state --json     # needs_input, blocked_by approval, agent_session_id s1
echo '{"hook_event_name":"Stop","session_id":"s1"}' | dartuios agent-hook claude-code
dartuios get-agent-state            # done
dartuios set-agent-state none       # clear it
```
