package strudel

import (
	"fmt"
	"strings"

	"backing-tracks/parser"
)

// GenerateStrudel converts a BTML track to Strudel code
func GenerateStrudel(track *parser.Track) string {
	var sb strings.Builder

	// Header comment
	sb.WriteString(fmt.Sprintf("// %s\n", track.Info.Title))
	sb.WriteString(fmt.Sprintf("// Key: %s | Tempo: %d BPM | Style: %s\n", track.Info.Key, track.Info.Tempo, track.Info.Style))
	sb.WriteString("// Generated from BTML\n\n")

	// Build layers
	layers := []string{}

	// Chord progression
	chordPattern := generateChordPattern(track)
	if chordPattern != "" {
		layers = append(layers, chordPattern)
	}

	// Bass line
	if track.Bass != nil {
		bassPattern := generateBassPattern(track)
		if bassPattern != "" {
			layers = append(layers, bassPattern)
		}
	}

	// Drums
	if track.Drums != nil {
		drumPatterns := generateDrumPatterns(track)
		layers = append(layers, drumPatterns...)
	}

	// Combine all layers with stack()
	if len(layers) == 1 {
		sb.WriteString(layers[0])
	} else if len(layers) > 1 {
		sb.WriteString("stack(\n")
		for i, layer := range layers {
			sb.WriteString("  " + layer)
			if i < len(layers)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		sb.WriteString(")")
	}

	// Add tempo
	sb.WriteString(fmt.Sprintf("\n  .cpm(%d/4)", track.Info.Tempo))

	return sb.String()
}

// generateChordPattern creates Strudel note patterns for chords
func generateChordPattern(track *parser.Track) string {
	chords := track.Progression.GetChords()
	if len(chords) == 0 {
		return ""
	}

	// Convert chords to Strudel notation
	// Use angle brackets for sequence and @ for duration
	var patterns []string

	for _, chord := range chords {
		notes := chordToNotes(chord.Symbol)
		duration := chord.Bars

		// Format: [c3,e3,g3] for chord, with @duration for bars
		noteStr := fmt.Sprintf("[%s]", strings.Join(notes, ","))
		if duration != 1.0 {
			noteStr = fmt.Sprintf("%s@%g", noteStr, duration)
		}
		patterns = append(patterns, noteStr)
	}

	// Determine rhythm pattern
	rhythm := "1"
	if track.Rhythm != nil {
		rhythm = rhythmToStrudel(track.Rhythm)
	}

	pattern := strings.Join(patterns, " ")

	// Apply rhythm subdivision if not just whole notes
	if rhythm != "1" {
		return fmt.Sprintf("note(\"%s\").s(\"piano\").struct(\"%s\")", pattern, rhythm)
	}

	return fmt.Sprintf("note(\"%s\").s(\"piano\")", pattern)
}

// chordToNotes converts a chord symbol to Strudel note names
func chordToNotes(symbol string) []string {
	root, octave := parseRoot(symbol)
	quality := parseQuality(symbol)

	// Base MIDI note (octave 3)
	rootNote := noteToMidi(root)
	baseOctave := 3

	// Get intervals based on quality
	intervals := getIntervals(quality)

	// Convert to note names
	notes := make([]string, len(intervals))
	for i, interval := range intervals {
		midi := rootNote + interval
		noteName := midiToNote(midi%12, baseOctave+octave+(midi/12))
		notes[i] = noteName
	}

	return notes
}

// getIntervals returns intervals for a chord quality
func getIntervals(quality string) []int {
	switch quality {
	case "m", "min":
		return []int{0, 3, 7} // Minor
	case "7":
		return []int{0, 4, 7, 10} // Dominant 7
	case "maj7":
		return []int{0, 4, 7, 11} // Major 7
	case "m7", "min7":
		return []int{0, 3, 7, 10} // Minor 7
	case "dim":
		return []int{0, 3, 6} // Diminished
	case "aug":
		return []int{0, 4, 8} // Augmented
	case "sus4":
		return []int{0, 5, 7} // Suspended 4
	case "sus2":
		return []int{0, 2, 7} // Suspended 2
	case "5":
		return []int{0, 7} // Power chord
	default:
		return []int{0, 4, 7} // Major
	}
}

// parseRoot extracts the root note from a chord symbol
func parseRoot(symbol string) (string, int) {
	if len(symbol) == 0 {
		return "C", 0
	}

	root := string(symbol[0])
	rest := symbol[1:]
	octaveOffset := 0

	// Check for accidentals
	if len(rest) > 0 {
		if rest[0] == '#' || rest[0] == 'b' {
			root += string(rest[0])
		}
	}

	return root, octaveOffset
}

// parseQuality extracts the chord quality from symbol
func parseQuality(symbol string) string {
	// Remove root note (and accidental)
	quality := symbol
	if len(quality) > 0 {
		quality = quality[1:] // Remove first char
	}
	if len(quality) > 0 && (quality[0] == '#' || quality[0] == 'b') {
		quality = quality[1:] // Remove accidental
	}

	return quality
}

// noteToMidi converts note name to MIDI offset (0-11)
func noteToMidi(note string) int {
	notes := map[string]int{
		"C": 0, "C#": 1, "Db": 1,
		"D": 2, "D#": 3, "Eb": 3,
		"E": 4, "Fb": 4, "E#": 5,
		"F": 5, "F#": 6, "Gb": 6,
		"G": 7, "G#": 8, "Ab": 8,
		"A": 9, "A#": 10, "Bb": 10,
		"B": 11, "Cb": 11, "B#": 0,
	}
	if val, ok := notes[note]; ok {
		return val
	}
	return 0
}

// midiToNote converts MIDI offset + octave to note name
func midiToNote(offset int, octave int) string {
	noteNames := []string{"c", "cs", "d", "ds", "e", "f", "fs", "g", "gs", "a", "as", "b"}
	return fmt.Sprintf("%s%d", noteNames[offset%12], octave)
}

// rhythmToStrudel converts BTML rhythm to Strudel struct pattern
func rhythmToStrudel(rhythm *parser.Rhythm) string {
	if rhythm.Pattern != "" {
		return customPatternToStrudel(rhythm.Pattern)
	}

	switch rhythm.Style {
	case "whole":
		return "1"
	case "half":
		return "1 1"
	case "quarter":
		return "1 1 1 1"
	case "eighth":
		return "1 1 1 1 1 1 1 1"
	case "strum_down":
		return "1 ~ ~ 1 ~ ~ 1 ~"
	case "strum_up_down":
		return "1 ~ 1 ~ 1 ~ 1 ~"
	case "folk":
		return "1 ~ 1 ~ 1 ~ 1 ~"
	case "shuffle_strum":
		return "1 ~ ~ 1 ~ 1"
	case "travis", "fingerpick":
		return "1 ~ 1 ~ 1 ~ 1 ~"
	case "fingerpick_slow":
		return "1 ~ ~ ~ 1 ~ ~ ~"
	case "arpeggio_up", "arpeggio_down":
		return "1 1 1 1"
	default:
		return "1"
	}
}

// customPatternToStrudel converts D/U/x/. pattern to Strudel
func customPatternToStrudel(pattern string) string {
	var parts []string
	for _, c := range pattern {
		switch c {
		case 'D', 'd', 'U', 'u':
			parts = append(parts, "1")
		case 'x':
			parts = append(parts, "1")
		case '.', '-':
			parts = append(parts, "~")
		}
	}
	return strings.Join(parts, " ")
}

// generateBassPattern creates Strudel pattern for bass
func generateBassPattern(track *parser.Track) string {
	chords := track.Progression.GetChords()
	if len(chords) == 0 {
		return ""
	}

	var patterns []string

	for _, chord := range chords {
		root, _ := parseRoot(chord.Symbol)
		quality := parseQuality(chord.Symbol)
		rootMidi := noteToMidi(root)

		// Create bass pattern based on style
		var bassNotes []string
		octave := 2 // Bass octave

		switch track.Bass.Style {
		case "root":
			bassNotes = []string{midiToNote(rootMidi, octave)}
		case "root_fifth":
			fifth := (rootMidi + 7) % 12
			bassNotes = []string{
				midiToNote(rootMidi, octave),
				midiToNote(fifth, octave),
			}
		case "walking", "swing_walking":
			third := rootMidi + 4
			if strings.HasPrefix(quality, "m") {
				third = rootMidi + 3
			}
			fifth := rootMidi + 7
			seventh := rootMidi + 10
			bassNotes = []string{
				midiToNote(rootMidi%12, octave),
				midiToNote(third%12, octave),
				midiToNote(fifth%12, octave),
				midiToNote(seventh%12, octave),
			}
		default:
			bassNotes = []string{midiToNote(rootMidi, octave)}
		}

		// Format with duration
		noteStr := strings.Join(bassNotes, " ")
		if chord.Bars != 1.0 {
			noteStr = fmt.Sprintf("<%s>@%g", noteStr, chord.Bars)
		} else if len(bassNotes) > 1 {
			noteStr = fmt.Sprintf("<%s>", noteStr)
		}
		patterns = append(patterns, noteStr)
	}

	return fmt.Sprintf("note(\"%s\").s(\"bass\")", strings.Join(patterns, " "))
}

// generateDrumPatterns creates Strudel patterns for drums
func generateDrumPatterns(track *parser.Track) []string {
	drums := track.Drums
	var patterns []string

	// Handle preset styles
	if drums.Style != "" {
		switch drums.Style {
		case "rock_beat":
			patterns = append(patterns, "s(\"bd ~ ~ ~ bd ~ ~ ~\")") // Kick on 1, 3
			patterns = append(patterns, "s(\"~ ~ sd ~ ~ ~ sd ~\")") // Snare on 2, 4
			patterns = append(patterns, "s(\"hh hh hh hh hh hh hh hh\")") // 8th note hats
		case "shuffle", "blues_shuffle":
			patterns = append(patterns, "s(\"bd ~ ~ bd ~ ~ bd ~ ~ bd ~ ~\").slow(1.5)") // Shuffle kick
			patterns = append(patterns, "s(\"~ ~ ~ sd ~ ~ ~ ~ ~ sd ~ ~\").slow(1.5)") // Shuffle snare
			patterns = append(patterns, "s(\"hh ~ hh hh ~ hh hh ~ hh hh ~ hh\").slow(1.5)") // Shuffle hats
		case "jazz_swing":
			patterns = append(patterns, "s(\"~ ~ bd ~ ~ ~ ~ ~ bd ~ ~ ~\").slow(1.5)") // Sparse kick
			patterns = append(patterns, "s(\"~ ~ ~ ~ ~ sd ~ ~ ~ ~ ~ ~\").slow(1.5)") // Sparse snare
			patterns = append(patterns, "s(\"ride ~ ride ride ~ ride ride ~ ride ride ~ ride\").slow(1.5)") // Ride pattern
		default:
			// Minimal default
			patterns = append(patterns, "s(\"bd ~ ~ ~ bd ~ ~ ~\")")
			patterns = append(patterns, "s(\"~ ~ sd ~ ~ ~ sd ~\")")
		}
		return patterns
	}

	// Handle custom patterns
	if drums.Kick != nil {
		kickPattern := drumPatternToStrudel(drums.Kick, "bd")
		if kickPattern != "" {
			patterns = append(patterns, kickPattern)
		}
	}

	if drums.Snare != nil {
		snarePattern := drumPatternToStrudel(drums.Snare, "sd")
		if snarePattern != "" {
			patterns = append(patterns, snarePattern)
		}
	}

	if drums.Hihat != nil {
		hihatPattern := drumPatternToStrudel(drums.Hihat, "hh")
		if hihatPattern != "" {
			patterns = append(patterns, hihatPattern)
		}
	}

	if drums.Ride != nil {
		ridePattern := drumPatternToStrudel(drums.Ride, "ride")
		if ridePattern != "" {
			patterns = append(patterns, ridePattern)
		}
	}

	return patterns
}

// drumPatternToStrudel converts a BTML drum pattern to Strudel
func drumPatternToStrudel(pattern *parser.DrumPattern, sound string) string {
	// Handle Euclidean rhythm
	if pattern.Euclidean != nil {
		return fmt.Sprintf("s(\"%s\").euclid(%d,%d,%d)",
			sound, pattern.Euclidean.Hits, pattern.Euclidean.Steps, pattern.Euclidean.Rotation)
	}

	// Handle explicit beats
	if len(pattern.Beats) > 0 {
		// Convert beat positions to pattern (assuming 8 subdivisions per bar)
		steps := make([]string, 8)
		for i := range steps {
			steps[i] = "~"
		}
		for _, beat := range pattern.Beats {
			// Convert 1-4 beats to 0-7 indices (2 subdivisions per beat)
			if beat >= 1 && beat <= 4 {
				idx := (beat - 1) * 2
				steps[idx] = sound
			}
		}
		return fmt.Sprintf("s(\"%s\")", strings.Join(steps, " "))
	}

	return ""
}
