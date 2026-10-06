package app

import (
	"strings"
	"time"

	"github.com/darsrc/tuios/internal/config"
	"github.com/darsrc/tuios/internal/overlay"
	"github.com/darsrc/tuios/internal/session"
	"github.com/darsrc/tuios/internal/terminal"
	"github.com/darsrc/tuios/internal/ui"
)

// Specimen renders the DAR ledger of standardized elements (DAR §45) through
// the ordinary renderer: one pane carrying the ledger, on the DAR default
// chrome, in a 100×30 frame. `dartuios specimen` prints it to stdout, and
// `dartuios shot` can save what it drew as PNG or SVG.
//
// Every glyph the ledger shows is read from the same source the chrome reads
// it from — the config getters, the agent mark table, the pulse and the
// filament table — so a ledger that drifts from the chrome is a ledger that
// lies.
func Specimen() string {
	cfg := config.DefaultConfig()
	// The process seed (config.Global) is not what a bare `specimen` run has
	// been configured with — only the TUI and tape paths fold the user config
	// in. Seed a Settings copy from the shipped user defaults so the chrome
	// here is the DAR default a fresh install draws: anchored window corners,
	// the DAR border weights, the thin scrollbar thumb.
	seed := config.DefaultSettings()
	config.ApplyAppearanceConfig(cfg, &seed)
	m := NewOS(OSOptions{UserConfig: cfg, Settings: &seed})
	if m.KeybindRegistry == nil {
		m.KeybindRegistry = config.NewKeybindRegistry(cfg)
	}
	m.Width, m.Height = 100, 30
	m.EffectiveWidth, m.EffectiveHeight = 100, 30

	// The pane takes the whole content region — two rows shorter than the
	// usable height, because the lone-window fast path only takes a pane that
	// fills the region, and keeping it off that path keeps the desktop and the
	// dock footer on screen. Width stops at the sidebar's reservation, so the
	// ledger rows land inside the pane the compositor shows.
	w := m.GetRenderWidth() - m.GetSidebarWidth()
	h := m.GetUsableHeight() - 2
	win := terminal.NewDaemonWindow("specimen", "specimen", 0, m.GetTopMargin(), w, h, 0, "pty-specimen", make(chan struct{}, 1), config.DefaultScrollbackLines)
	win.Workspace = 1
	win.MarkPositionDirty()
	win.WriteOutput([]byte(specimenLedger(&m.Settings)))
	m.Windows = []*terminal.Window{win}
	m.FocusedWindow = 0
	m.NumWorkspaces = 9
	m.CurrentWorkspace = 1

	frame, ok := m.safeComposeFrame()
	if !ok {
		return ""
	}
	return frame
}

// specimenLedger is the pane content: the standardized elements in DAR order,
// one row per element, labels in the first ten cells so the element column
// lines up down the page.
func specimenLedger(s *config.Settings) string {
	rows := []string{
		"dartuios specimen — DAR §45, the standardized elements",
		ledgerRow("panel", overlay.AnchorTL()+" anchored panel "+overlay.AnchorTR()),
		ledgerRow("", overlay.AnchorBL()+" "+overlay.AnchorBR()),
		ledgerRow("frame", "┌ hard ──────────────────────────────┐"),
		ledgerRow("weight", "──│──┼── light      ━━┃━━╋━━ heavy"),
		ledgerRow("rail", s.GetRailFocusMark()+" focus   "+s.GetRailHoverMark()+" hover   "+s.GetRailAttentionMark()+" attention   "+s.GetRailBullet()+" rest"),
		ledgerRow("matrix", "◊ ◆ ◇ ◈"),
		ledgerRow("state", agentStateIndicator(string(session.AgentStateWorking))+" "+agentStateIndicator(string(session.AgentStateNeedsInput))+" "+agentStateIndicator(string(session.AgentStateIdle))+" "+agentStateIndicator(string(session.AgentStateDone))+" "+agentStateIndicator(string(session.AgentStateErrored))+" "+agentStateIndicator(string(session.AgentStateUnknown))),
		ledgerRow("chip", config.DockPillLeftChar+" action "+config.DockPillRightChar+"    pulse "+weightPulseFrames()),
		ledgerRow("tabs", config.DockWorkspaceMoreLeft+" "+s.GetRailCollapseGlyph()+" "+s.GetRailExpandGlyph()+" "+config.DockWorkspaceMoreRight),
		ledgerRow("fold", s.GetRailFoldOpenGlyph()+" open   "+s.GetRailFoldShutGlyph()+" shut"),
		ledgerRow("scroll", s.GetScrollbarThumbChar()+" thin thumb, no track"),
		ledgerRow("filament", filamentFrames()),
		ledgerRow("footer", "fragments, no bar:  "+config.DockPillLeftChar+" 1 "+config.DockPillRightChar+"  2  "+config.DockWorkspaceMoreLeft+" "+config.DockWorkspaceMoreRight),
	}
	return strings.Join(rows, "\r\n")
}

// ledgerRow lays a label and a body out on the element column.
func ledgerRow(label, body string) string {
	if label == "" {
		label = " "
	}
	return label + strings.Repeat(" ", 10-len(label)) + "  " + body
}

// weightPulseFrames is the action chip's pulse, the four frames the side
// border burns in the 200 ms after a panel opens.
func weightPulseFrames() string {
	start := time.Unix(0, 0)
	frames := make([]string, 4)
	for i := range frames {
		frames[i] = string(overlay.WeightPulse(start, start.Add(time.Duration(i)*50*time.Millisecond)))
	}
	return strings.Join(frames, " ")
}

// filamentFrames is the Living Filament's first eight consecutive frames,
// side by side: the same glyphs the working-row mark burns, at the same rate.
func filamentFrames() string {
	frames := make([]string, 8)
	for i := range frames {
		frames[i] = string(ui.FilamentFrames[i])
	}
	return strings.Join(frames, " ")
}
