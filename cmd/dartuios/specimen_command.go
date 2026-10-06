package main

import (
	"fmt"

	"github.com/darsrc/tuios/internal/app"
	"github.com/spf13/cobra"
)

func newSpecimenCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "specimen",
		Short: "Render the DAR standardized-elements ledger as a 100×30 frame",
		Long: `Render the DAR ledger of standardized elements (DAR §45) as a 100×30 frame
through the ordinary renderer and print it to stdout.

Every element the dartuios chrome draws — the anchored panel, the line
weights, the row rail, the agent status marks, the action chip and its
pulse, the Living Filament — appears in its chrome position, so the frame
is a reference the docs and the review of a glyph-set change both point
at. ` + "`dartuios shot`" + ` saves what it draws as PNG or SVG.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			frame := app.Specimen()
			if frame == "" {
				return fmt.Errorf("the frame would not compose")
			}
			_, err := fmt.Fprint(cmd.OutOrStdout(), frame+"\n")
			return err
		},
	}
}
