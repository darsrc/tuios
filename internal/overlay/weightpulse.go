package overlay

import "time"

// weightPulseFrames is the weight-animated boundary's four frames, one per
// 50 ms: light, medium, full, medium. DAR §action control animates a control's
// weight on entry rather than its colour.
var weightPulseFrames = [4]rune{'▏', '▍', '█', '▍'}

// WeightPulse is the boundary's glyph at now, measured from start: the four
// frames above at 50 ms each, 200 ms total, off monotonic time. Past 200 ms —
// or before start — it returns 0 and the caller draws the settled cap.
//
// It lives in overlay rather than the motion package because a dialog's frame
// is drawn here and overlay may not import the motion package (it would cycle
// through terminal back to overlay).
func WeightPulse(start, now time.Time) rune {
	f := int(now.Sub(start) / (50 * time.Millisecond))
	if f < 0 || f >= len(weightPulseFrames) {
		return 0
	}
	return weightPulseFrames[f]
}
