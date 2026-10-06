# Using dartuios as a Library

dartuios can be imported and used as a library in your own Go applications. This allows you to embed a full-featured terminal window manager in your Bubble Tea applications.

## Installation

```bash
go get github.com/darsrc/tuios/pkg/dartuios
```

## Quick Start

### Basic Usage

```go
package main

import (
    "log"

    "github.com/darsrc/tuios/pkg/dartuios"
    tea "charm.land/bubbletea/v2"
)

func main() {
    // Create a new dartuios instance with default options
    model := dartuios.New()

    // Create the Bubble Tea program with recommended options
    p := tea.NewProgram(model, dartuios.ProgramOptions()...)

    // Run the program
    if _, err := p.Run(); err != nil {
        log.Fatal(err)
    }
}
```

### With Custom Options

```go
model := dartuios.New(
    dartuios.WithTheme("dracula"),
    dartuios.WithShowKeys(true),
    dartuios.WithAnimations(false),
    dartuios.WithWorkspaces(9),
    dartuios.WithBorderStyle("rounded"),
    dartuios.WithDockbarPosition("bottom"),
    dartuios.WithScrollbackLines(20000),
)
```

### With Mouse Motion Filtering

For better performance, use the provided mouse motion filter:

```go
p := tea.NewProgram(
    model,
    tea.WithFPS(60),
    tea.WithFilter(dartuios.FilterMouseMotion),
)
```

## Options Reference

### WithTheme(name string)

Set the color theme. Available themes include "dracula", "nord", "tokyonight", and 300+ others from bubbletint.

```go
dartuios.WithTheme("dracula")
```

### WithShowKeys(enabled bool)

Enable the showkeys overlay to display pressed keys (useful for demos).

```go
dartuios.WithShowKeys(true)
```

### WithAnimations(enabled bool)

Enable or disable window animations. When disabled, windows snap instantly.

```go
dartuios.WithAnimations(false)
```

### WithASCIIOnly(enabled bool)

Use ASCII characters instead of Nerd Font icons for compatibility.

```go
dartuios.WithASCIIOnly(true)
```

### WithWorkspaces(n int)

Set the number of workspaces (1-9).

```go
dartuios.WithWorkspaces(4)
```

### WithBorderStyle(style string)

Set the window border style. Valid values:
- `"rounded"` (default)
- `"normal"`
- `"thick"`
- `"double"`
- `"hidden"`
- `"block"`
- `"outer-half-block"`
- `"inner-half-block"`
- `"ascii"`

```go
dartuios.WithBorderStyle("thick")
```

### WithDockbarPosition(position string)

Set the dockbar position. Valid values:
- `"top"` (default)
- `"bottom"`
- `"hidden"`

```go
dartuios.WithDockbarPosition("bottom")
```

### WithHideWindowButtons(hide bool)

Hide the minimize/maximize/close buttons in window title bars.

```go
dartuios.WithHideWindowButtons(true)
```

### WithWindowButtonStyle(style string)

How the window controls are drawn: `"pill"` (glyphs on a filled pill) or
`"dots"` (macOS traffic lights, which name themselves on hover). On the left,
the pill puts close at the outer corner: close, zoom, minimize. See
[the configuration reference](https://dartuios.dev/docs/configuration).

```go
dartuios.WithWindowButtonStyle("dots")
```

### WithWindowButtonPosition(position string)

Which end of the title bar the window controls sit on: `"left"` (default, the
way macOS does it) or `"right"`. See
[the configuration reference](https://dartuios.dev/docs/configuration).

```go
dartuios.WithWindowButtonPosition("left")
```

### WithScrollbackLines(lines int)

Set the scrollback buffer size (100-1000000).

```go
dartuios.WithScrollbackLines(50000)
```

### WithSize(width, height int)

Set the initial terminal size. Usually not needed as dartuios auto-detects.

```go
dartuios.WithSize(120, 40)
```

### WithSSHMode(enabled bool)

Enable SSH mode for running over SSH connections.

```go
dartuios.WithSSHMode(true)
```

### NewForPTY

`dartuios.NewForPTY` builds an instance bound to an existing PTY rather than the
process's own terminal; see the doc comment in `pkg/dartuios/dartuios.go` for the
contract.

### WithUserConfig(cfg *config.UserConfig)

Provide a custom user configuration instead of loading from file.

```go
cfg := dartuios.Config.DefaultConfig()
cfg.Keybindings.LeaderKey = "ctrl+a"
dartuios.WithUserConfig(cfg)
```

## Web Terminal Integration

dartuios can be served through the browser using the [sip library](https://github.com/Gaurav-Gosain/sip):

```go
package main

import (
    "context"
    "log"

    "github.com/Gaurav-Gosain/sip"
    "github.com/darsrc/tuios/pkg/dartuios"
    tea "charm.land/bubbletea/v2"
)

func main() {
    server := sip.NewServer(sip.DefaultConfig())

    err := server.ServeWithProgram(context.Background(), func(sess sip.Session) *tea.Program {
        pty := sess.Pty()

        // Create dartuios for the web session
        model := dartuios.New(
            dartuios.WithSize(pty.Width, pty.Height),
            dartuios.WithTheme("dracula"),
        )

        // sip's options first, then dartuios's.
        return tea.NewProgram(model, append(sip.MakeOptions(sess), dartuios.ProgramOptions()...)...)
    })

    if err != nil {
        log.Fatal(err)
    }
}
```

Build the program yourself with `ServeWithProgram` rather than returning the
options from `Serve`. Both `sip.MakeOptions` and `dartuios.ProgramOptions` carry a
`tea.WithFilter`, and the last one set wins. `Serve` appends sip's options after
yours, which drops the dartuios mouse motion filter. `dartuios-web` builds its
program the same way.

## SSH Server Integration

For SSH server integration, use the Wish library:

```go
package main

import (
    "github.com/darsrc/tuios/pkg/dartuios"
    tea "charm.land/bubbletea/v2"
    "charm.land/ssh"
    "charm.land/wish/v2"
    "charm.land/wish/v2/bubbletea"
)

func main() {
    s, _ := wish.NewServer(
        wish.WithAddress(":2222"),
        wish.WithMiddleware(
            bubbletea.MiddlewareWithProgramHandler(func(sess ssh.Session) *tea.Program {
                pty, _, _ := sess.Pty()

                model := dartuios.New(
                    dartuios.WithSize(pty.Window.Width, pty.Window.Height),
                    dartuios.WithSSHMode(true),
                )

                // wish's options first, then dartuios's.
                return tea.NewProgram(model, append(bubbletea.MakeOptions(sess), dartuios.ProgramOptions()...)...)
            }),
        ),
    )

    s.ListenAndServe()
}
```

`bubbletea.Middleware` has the same ordering problem as sip's `Serve`: it
appends `MakeOptions` after the options you return, so its `tea.WithFilter`
replaces the dartuios one. `MiddlewareWithProgramHandler` lets you put them in the
right order. The `dartuios ssh` server in `internal/server` does the same.

`WithSSHMode` makes the model an SSH client: the settings page does not write
the operator's config file and desktop actions do not run on the server. It
does not give the model the SSH session, a graphics output or the client's
terminal capabilities, so images are not forwarded to the client. A model
served without it, including one behind sip, is treated as a local client.

## Configuration Access

The `dartuios.Config` struct provides access to configuration utilities:

```go
// Load user config from file
cfg, err := dartuios.Config.LoadUserConfig()

// Get default config
cfg := dartuios.Config.DefaultConfig()

// Get config file path
path, err := dartuios.Config.GetConfigPath()
```

## Model Methods

The dartuios model provides several public methods:

### Window Management

- `AddWindow(title string)`: Create a new terminal window
- `DeleteWindow(i int)`: Close window at index
- `FocusWindow(i int)`: Focus window at index
- `GetFocusedWindow()`: Get the currently focused window

### Workspace Management

- `SwitchWorkspace(n int)`: Switch to workspace n (1-9)
- `MoveWindowToWorkspace(windowIndex, workspace int)`: Move a window

### Layout

- `ToggleTiling()`: Toggle automatic tiling mode
- `TileAllWindows()`: Retile all windows

### Cleanup

- `Cleanup()`: Clean up resources (call when done)

## Example: Custom Wrapper

You can wrap dartuios in your own model for additional functionality:

```go
type MyApp struct {
    dartuios *dartuios.Model
    // your additional state
}

func (m *MyApp) Init() tea.Cmd {
    return m.dartuios.Init()
}

func (m *MyApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle your custom messages first
    switch msg := msg.(type) {
    case myCustomMsg:
        // handle it
        return m, nil
    }

    // Delegate to dartuios
    updated, cmd := m.dartuios.Update(msg)
    m.dartuios = updated.(*dartuios.Model)
    return m, cmd
}

func (m *MyApp) View() string {
    return m.dartuios.View()
}
```

## Related Documentation

- [Architecture](ARCHITECTURE.md): Technical architecture
- [Keybindings](KEYBINDINGS.md): Keyboard shortcuts
- [Configuration](CONFIGURATION.md): Config file options
- [Web Terminal](WEB.md): Browser-based access
- [sip](https://github.com/Gaurav-Gosain/sip): the web serving library behind `dartuios-web`
