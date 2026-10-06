package ui

import "time"

// FilamentFrames is the Living Filament: DAR §38 modules 1–20 played in the
// valid chain 1 → 4 → 15 → 20 → 3 → 11 → 16 → 17 → 2 → 9 → 14, concatenated
// so each module's start state equals the previous module's end state. The
// frames play exactly as authored — no variants, no substitutions, no
// capability fallback — and the chain ends on U+1F372, so it loops back to
// module 1 seamlessly.
var FilamentFrames = [207]rune{
	'🭲', '🭲', '🭲', '🭵', '▕', '▕', '🭵', '🭴',
	'🭳', '🭲', '🭱', '🭰', '▏', '🭰', '🭱', '🭲',
	'🭳', '🭲', '🭱', '🭲', '🭳', '🭳', '🭲', '🭲',
	'┃', '▮', '▮', '┃', '🭲', '🭲', '🭳', '┃',
	'▮', '▮', '┃', '🭳', '🭲', '🭲', '🭲', '╰',
	'∪', '∪', 'ᑌ', 'ᑌ', '∪', '∪', '╯', '🭲',
	'🭳', '🭳', '🭴', '🭴', '🭴', '🭴', '🭴', ')',
	')', '❩', 'ᑐ', 'ᑐ', '○', '◉', '●', '◉',
	'○', 'ᑕ', 'ᑕ', '❨', '(', '(', '🭳', '🭲',
	'🭲', '🭲', '(', '(', '❨', 'ᑕ', 'ᑕ', '○',
	'○', '◦', '∘', '∙', '∙', '∘', '◦', '○',
	'○', 'ᑐ', 'ᑐ', '❩', ')', ')', '🭲', '🭳',
	'🭳', '🭳', '╵', '╴', '╴', '─', '🭹', '🭹',
	'🭺', '▁', '🭿', '▕', '🭵', '🭴', '🭴', '🭴',
	'🭴', '🭴', '╭', '⌢', '∩', '⌒', '⌒', '∩',
	'⌢', '╮', '🭳', '🭳', '🭲', '🭲', '🭲', '🭲',
	'🭲', '🭱', '🭰', '▏', '▏', '🭰', '🭱', '🭲',
	'🭳', '🭴', '🭵', '▕', '▕', '▕', '🭵', '🭴',
	'🭳', '🭲', '🭱', '🭰', '▏', '▏', '▏', '🭰',
	'🭱', '🭲', '🭳', '🭳', '🭳', '🭴', '🭵', '▕',
	'▕', '🭵', '🭴', '🭳', '🭳', '🭲', '🭲', '🭱',
	'🭰', '▏', '🭰', '🭱', '🭲', '🭲', '🭲', '╵',
	'╵', '˙', '·', '∙', '∙', '•', '∙', '·',
	'˙', '╷', '╷', '🭲', '🭱', '🭱', '🭱', '🭱',
	'🭱', '(', '(', '❨', 'ᑕ', '⊂', '○', '⊃',
	'ᑐ', '❩', ')', ')', '🭲', '🭲', '🭲',
}

// Filament is the Living Filament's clock: one start instant from which every
// frame is a pure function of the elapsed time, so a dropped frame lands on
// the right glyph rather than the wrong one.
type Filament struct {
	started time.Time
}

// NewFilament returns a filament started now.
func NewFilament() Filament { return Filament{started: time.Now()} }

// IsZero reports whether the filament has not been started: its zero value,
// which is the resting state between working agents.
func (f Filament) IsZero() bool { return f.started.IsZero() }

// Frame is the filament's glyph at now: 20 FPS off monotonic time, wrapped
// across the whole table so the loop is seamless. A now before the start
// instant (the first tick carries its scheduled time, a hair before the
// time.Now() the filament started with) is frame 0, not a negative index.
func (f Filament) Frame(now time.Time) rune {
	elapsed := now.Sub(f.started)
	if elapsed < 0 {
		elapsed = 0
	}
	return FilamentFrames[int(elapsed/50*time.Millisecond)%len(FilamentFrames)]
}
