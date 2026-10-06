package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/darsrc/tuios/internal/mcp"
	"github.com/darsrc/tuios/internal/session"
	"github.com/spf13/cobra"
)

func newMCPCommand() *cobra.Command {
	var write bool
	var scope string
	var integrationVersion int
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Serve dartuios to an agent harness as an MCP server over stdio",
		Long: `Serve the dartuios control surface as Model Context Protocol tools over stdin and
stdout, for an agent harness to load: list and read panes, wait for agent
states, follow the event stream, report the agent's own state and meta, and
send and read mail.

Each tool is a daemon verb, with its input schema generated from the verb
table. Every call opens its own connection and restricts it with
restrict-connection before anything else, so the daemon holds the call to
what this server was started with:

  default        read-only, and only the session of the pane the server runs
                 in, its fan group and the sessions a fan from it started
  --write        also list send_text, send_keys, ask_agent, respond and fan,
                 still inside that session and fan group
  --scope all    reach every session, for a harness that runs outside dartuios

The pane is found from the kernel's record of this process's pid, and where
the kernel cannot say, from DARTUIOS_PANE_ID and DARTUIOS_PANE_TOKEN. A server
restricted to its own session that runs in no pane reaches nothing.

dartuios integration install <harness> --mcp registers it with a harness.`,
		Example: `  dartuios mcp
  dartuios mcp --write
  claude mcp add dartuios -- dartuios mcp`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if scope != session.ScopeOwn && scope != session.ScopeAll {
				return fmt.Errorf("--scope must be own or all, not %q", scope)
			}
			_ = integrationVersion // written by dartuios integration install to mark its entry
			srv := mcp.New(mcp.Options{
				Name:      "dartuios",
				Version:   version,
				Write:     write,
				ScopeAll:  scope == session.ScopeAll,
				PaneID:    os.Getenv("DARTUIOS_PANE_ID"),
				PaneToken: os.Getenv("DARTUIOS_PANE_TOKEN"),
				Dial:      dialMCPConn,
				Verbs:     mcpVerbDocs(),
				Log: func(format string, args ...any) {
					fmt.Fprintf(os.Stderr, "dartuios mcp: "+format+"\n", args...)
				},
			})
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return srv.Serve(ctx, os.Stdin, os.Stdout)
		},
	}
	cmd.Flags().BoolVar(&write, "write", false, "List the tools that type into panes: send_text, send_keys, ask_agent, respond and fan")
	cmd.Flags().StringVar(&scope, "scope", session.ScopeOwn, "own: only this pane's session and fan group. all: every session")
	cmd.Flags().IntVar(&integrationVersion, "integration", 0, "Marks an entry dartuios integration install wrote")
	_ = cmd.Flags().MarkHidden("integration")
	return cmd
}

// mcpVerbDocs is the verb table, in the shape the MCP server reads.
func mcpVerbDocs() []mcp.VerbDoc {
	raw, err := json.Marshal(session.VerbDocs())
	if err != nil {
		return nil
	}
	var out []mcp.VerbDoc
	_ = json.Unmarshal(raw, &out)
	return out
}

// mcpSocketPath finds the daemon's socket. A harness may start its MCP servers
// with a pared-down environment, and XDG_RUNTIME_DIR is one of the variables
// such a harness drops, which would send the server to the /tmp fallback while
// the daemon listens under /run/user. So DARTUIOS_SOCKET, which every pane
// exports, comes first, and when XDG_RUNTIME_DIR is unset the systemd default
// is tried before the fallback.
func mcpSocketPath() (string, error) {
	if s := os.Getenv("DARTUIOS_SOCKET"); s != "" {
		return s, nil
	}
	path, err := session.GetSocketPath()
	if err != nil {
		return "", err
	}
	if _, statErr := os.Stat(path); statErr == nil || os.Getenv("XDG_RUNTIME_DIR") != "" || runtime.GOOS != "linux" {
		return path, nil
	}
	alt := filepath.Join("/run/user", strconv.Itoa(os.Getuid()), "dartuios", "dartuios.sock")
	if _, err := os.Stat(alt); err == nil {
		return alt, nil
	}
	return path, nil
}

func dialMCPConn() (mcp.Conn, error) {
	path, err := mcpSocketPath()
	if err != nil {
		return nil, err
	}
	c, err := session.DialVerbClientAt(path, version)
	if err != nil {
		return nil, fmt.Errorf("no dartuios daemon answered on %s: %w. Start dartuios, or run dartuios new --detach", path, err)
	}
	return mcpConn{c}, nil
}

// mcpConn adapts a verb client to what the MCP server calls.
type mcpConn struct{ c *session.VerbClient }

func (m mcpConn) Call(verb string, params any, timeout time.Duration) (json.RawMessage, error) {
	raw, err := m.c.CallWithTimeout(verb, params, timeout)
	if ce, ok := errors.AsType[*session.VerbCallError](err); ok {
		var hint any
		if ce.Hint != nil {
			hint = ce.Hint
		}
		return nil, &mcp.CallError{Code: ce.Code, Message: ce.Message, Hint: hint}
	}
	return raw, err
}

func (m mcpConn) ReadEventLine(timeout time.Duration) ([]byte, error) {
	return m.c.ReadEventLine(timeout)
}

func (m mcpConn) Close() error { return m.c.Close() }
