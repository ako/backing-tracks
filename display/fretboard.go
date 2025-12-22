package display

import (
	"fmt"
	"strings"

	"backing-tracks/theory"
)

// FretboardDisplay manages guitar neck visualization
type FretboardDisplay struct {
	scale        *theory.Scale
	numFrets     int
	positions    [][]bool // [string][fret] = in scale
	roots        [][]bool // [string][fret] = is root
	highlighted  []int    // Currently playing MIDI notes
	compactMode  bool     // Use compact display for narrow terminals
}

// NewFretboardDisplay creates a new fretboard display
func NewFretboardDisplay(scale *theory.Scale, numFrets int) *FretboardDisplay {
	fd := &FretboardDisplay{
		scale:       scale,
		numFrets:    numFrets,
		highlighted: []int{},
		compactMode: false,
	}
	fd.updatePositions()
	return fd
}

// updatePositions recalculates scale positions on fretboard
func (fd *FretboardDisplay) updatePositions() {
	if fd.scale != nil {
		fd.positions, fd.roots = fd.scale.GetFretboardPositions(fd.numFrets)
	}
}

// SetScale updates the displayed scale
func (fd *FretboardDisplay) SetScale(scale *theory.Scale) {
	fd.scale = scale
	fd.updatePositions()
}

// SetCompactMode enables/disables compact display
func (fd *FretboardDisplay) SetCompactMode(compact bool) {
	fd.compactMode = compact
}

// HighlightNote marks a note as currently playing
func (fd *FretboardDisplay) HighlightNote(midiNote int) {
	fd.highlighted = append(fd.highlighted, midiNote)
}

// ClearHighlights removes all highlighted notes
func (fd *FretboardDisplay) ClearHighlights() {
	fd.highlighted = []int{}
}

// isHighlighted checks if a fret position is currently highlighted
func (fd *FretboardDisplay) isHighlighted(stringIdx, fret int) bool {
	midiNote := theory.GuitarTuning[stringIdx] + fret
	for _, h := range fd.highlighted {
		if h == midiNote {
			return true
		}
	}
	return false
}

// Render returns the fretboard as a slice of strings (one per line)
func (fd *FretboardDisplay) Render() []string {
	if fd.scale == nil {
		return []string{"No scale set"}
	}

	if fd.compactMode {
		return fd.renderCompact()
	}
	return fd.renderFull()
}

// renderFull renders the full fretboard (frets 0-15)
func (fd *FretboardDisplay) renderFull() []string {
	lines := []string{}

	// Scale name header
	lines = append(lines, fmt.Sprintf(" %s", fd.scale.Name))
	lines = append(lines, "")

	// Fret numbers header
	fretHeader := "   "
	for fret := 0; fret <= fd.numFrets; fret++ {
		fretHeader += fmt.Sprintf("%2d ", fret)
	}
	lines = append(lines, fretHeader)

	// Top nut/border
	nutLine := "   ╔"
	for fret := 0; fret <= fd.numFrets; fret++ {
		if fret == fd.numFrets {
			nutLine += "══╗"
		} else {
			nutLine += "══╤"
		}
	}
	lines = append(lines, nutLine)

	// Guitar strings (high to low for display: e, B, G, D, A, E)
	stringOrder := []int{5, 4, 3, 2, 1, 0} // Reverse for display (high e at top)

	for i, stringIdx := range stringOrder {
		stringName := theory.GuitarStringNames[stringIdx]
		line := fmt.Sprintf(" %s ║", stringName)

		for fret := 0; fret <= fd.numFrets; fret++ {
			symbol := fd.getFretSymbol(stringIdx, fret)
			line += symbol
			if fret < fd.numFrets {
				line += "│"
			} else {
				line += "║"
			}
		}
		lines = append(lines, line)

		// Add separator between strings (except after last)
		if i < len(stringOrder)-1 {
			sepLine := "   ╟"
			for fret := 0; fret <= fd.numFrets; fret++ {
				if fret == fd.numFrets {
					sepLine += "──╢"
				} else {
					sepLine += "──┼"
				}
			}
			lines = append(lines, sepLine)
		}
	}

	// Bottom border
	bottomLine := "   ╚"
	for fret := 0; fret <= fd.numFrets; fret++ {
		if fret == fd.numFrets {
			bottomLine += "══╝"
		} else {
			bottomLine += "══╧"
		}
	}
	lines = append(lines, bottomLine)

	// Fret markers
	markerLine := "      "
	for fret := 0; fret <= fd.numFrets; fret++ {
		if fret == 3 || fret == 5 || fret == 7 || fret == 9 || fret == 15 {
			markerLine += " ● "
		} else if fret == 12 {
			markerLine += "●● "
		} else {
			markerLine += "   "
		}
	}
	lines = append(lines, markerLine)

	// Legend
	lines = append(lines, "")
	lines = append(lines, " ◆ Root  ● Scale  ○ Playing")

	return lines
}

// renderCompact renders a compact fretboard (narrower)
func (fd *FretboardDisplay) renderCompact() []string {
	lines := []string{}
	maxFret := 12
	if fd.numFrets < 12 {
		maxFret = fd.numFrets
	}

	// Scale name
	lines = append(lines, fmt.Sprintf(" %s", fd.scale.Name))

	// Fret numbers (compact)
	fretHeader := "  "
	for fret := 0; fret <= maxFret; fret++ {
		if fret < 10 {
			fretHeader += fmt.Sprintf("%d ", fret)
		} else {
			fretHeader += fmt.Sprintf("%d", fret)
		}
	}
	lines = append(lines, fretHeader)

	// Strings (high to low)
	stringOrder := []int{5, 4, 3, 2, 1, 0}
	for _, stringIdx := range stringOrder {
		stringName := theory.GuitarStringNames[stringIdx]
		line := fmt.Sprintf("%s ", stringName)

		for fret := 0; fret <= maxFret; fret++ {
			symbol := fd.getCompactSymbol(stringIdx, fret)
			line += symbol
		}
		lines = append(lines, line)
	}

	// Fret markers
	markerLine := "  "
	for fret := 0; fret <= maxFret; fret++ {
		if fret == 3 || fret == 5 || fret == 7 || fret == 9 {
			markerLine += "· "
		} else if fret == 12 {
			markerLine += ": "
		} else {
			markerLine += "  "
		}
	}
	lines = append(lines, markerLine)

	return lines
}

// getFretSymbol returns the display symbol for a fret position
func (fd *FretboardDisplay) getFretSymbol(stringIdx, fret int) string {
	if fd.isHighlighted(stringIdx, fret) {
		return "\033[33m○\033[0m" // Yellow circle for playing
	}
	if fd.roots[stringIdx][fret] {
		return "\033[31m◆\033[0m" // Red diamond for root
	}
	if fd.positions[stringIdx][fret] {
		return "\033[32m●\033[0m" // Green dot for scale note
	}
	return "─" // Empty fret
}

// getCompactSymbol returns the compact display symbol for a fret position
func (fd *FretboardDisplay) getCompactSymbol(stringIdx, fret int) string {
	if fd.isHighlighted(stringIdx, fret) {
		return "\033[33m○\033[0m " // Yellow circle for playing
	}
	if fd.roots[stringIdx][fret] {
		return "\033[31m◆\033[0m " // Red diamond for root
	}
	if fd.positions[stringIdx][fret] {
		return "\033[32m●\033[0m " // Green dot for scale note
	}
	return "· " // Empty fret
}

// GetWidth returns the approximate width of the rendered fretboard
func (fd *FretboardDisplay) GetWidth() int {
	if fd.compactMode {
		return 30 // Compact mode width
	}
	return (fd.numFrets + 1) * 3 + 6 // Full mode width
}

// RenderSimple returns a simplified one-line scale indicator
func (fd *FretboardDisplay) RenderSimple() string {
	if fd.scale == nil {
		return ""
	}

	// Show scale name and notes
	notes := []string{}
	for _, interval := range fd.scale.Intervals {
		noteIdx := (fd.scale.Root + interval) % 12
		notes = append(notes, theory.NoteNames[noteIdx])
	}

	return fmt.Sprintf("%s: %s", fd.scale.Name, strings.Join(notes, "-"))
}
