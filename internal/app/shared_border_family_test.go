package app

import (
	"fmt"
	"strings"
	"testing"

	"github.com/darsrc/tuios/internal/config"
)

func sidebarSides() []string { return []string{"left", "right", ""} }

// TestDividerCellsStayInTheStylesOwnGlyphs walks every style the settings page
// offers against every shape the content region takes, so a style added to the
// list is asked the same question without anyone rewriting this.
func TestDividerCellsStayInTheStylesOwnGlyphs(t *testing.T) {
	for _, style := range config.BorderStyles {
		for _, dock := range []string{"top", "bottom", "hidden"} {
			for _, side := range sidebarSides() {
				t.Run(fmt.Sprintf("%s/%s-dock/%s", style, dock, sidebarName(side)), func(t *testing.T) {
					m := extentOSStyled(t, 4, dock, side, style)
					// The focused pane is outlined in the style's own focused weight, so a
					// divider that falls on its perimeter carries that style's heavy
					// glyphs. Both weights belong to this style; a glyph outside the
					// union is borrowed from a foreign one.
					own := styleGlyphs(config.Global.GetBorderForStyle()) +
						styleGlyphs(config.Global.GetFocusedBorderForStyle())
					g := frameCells(t, m)
					for _, c := range dividerCells(m) {
						if got := cellAt(g, c.X, c.Y); !strings.ContainsRune(own, got) {
							t.Errorf("the divider cell (%d,%d) is %q, which is not one of this style's own glyphs %q",
								c.X, c.Y, string(got), own)
						}
					}
				})
			}
		}
	}
}
