package midi

import (
	"backing-tracks/parser"
	"strings"

	"gitlab.com/gomidi/midi/v2"
)

// ChordEvent represents a chord hit with timing
type ChordEvent struct {
	Notes    []uint8 // MIDI notes in the chord
	Tick     uint32  // When to play
	Duration uint32  // How long to hold
	Velocity uint8   // Volume
}

// GenerateChordRhythm creates chord events based on rhythm style
func GenerateChordRhythm(chords []parser.Chord, rhythm *parser.Rhythm, ticksPerBar uint32) []midiEvent {
	events := []midiEvent{}
	currentTick := uint32(0)

	// Default to whole notes if no rhythm specified
	style := "whole"
	swing := 0.5
	pattern := ""
	if rhythm != nil {
		if rhythm.Style != "" {
			style = rhythm.Style
		}
		if rhythm.Swing > 0 {
			swing = rhythm.Swing
		}
		if rhythm.Pattern != "" {
			pattern = rhythm.Pattern
			style = "pattern" // Override style when pattern is specified
		}
	}

	// Parse accent beats
	accentBeats := map[int]bool{1: true} // Default accent on beat 1
	if rhythm != nil && rhythm.Accent != "" {
		accentBeats = parseAccentBeats(rhythm.Accent)
	}

	for _, chord := range chords {
		notes := getChordVoicing(chord.Symbol)
		duration := uint32(chord.Bars * float64(ticksPerBar))

		var chordEvents []midiEvent
		if style == "pattern" {
			chordEvents = generateCustomPattern(pattern, notes, currentTick, duration, ticksPerBar, swing)
		} else {
			chordEvents = generateRhythmPattern(style, notes, currentTick, duration, ticksPerBar, swing, accentBeats)
		}
		events = append(events, chordEvents...)

		currentTick += duration
	}

	return events
}

// generateRhythmPattern creates the actual rhythm pattern for a chord
func generateRhythmPattern(style string, notes ChordVoicing, startTick, duration, ticksPerBar uint32, swing float64, accentBeats map[int]bool) []midiEvent {
	events := []midiEvent{}
	quarterNote := ticksPerBar / 4
	eighthNote := ticksPerBar / 8
	tripletEighth := ticksPerBar / 12

	switch style {
	case "whole":
		// Whole note - one strum per chord, held for full duration
		for _, note := range notes {
			events = append(events, midiEvent{startTick, midi.NoteOn(0, note, 80)})
			events = append(events, midiEvent{startTick + duration - 10, midi.NoteOff(0, note)})
		}

	case "half":
		// Half notes - two strums per bar
		numHalves := int(duration / (ticksPerBar / 2))
		if numHalves == 0 {
			numHalves = 1
		}
		halfDuration := duration / uint32(numHalves)
		for i := 0; i < numHalves; i++ {
			tick := startTick + uint32(i)*halfDuration
			vel := uint8(75)
			if i == 0 {
				vel = 85 // Accent first
			}
			for _, note := range notes {
				events = append(events, midiEvent{tick, midi.NoteOn(0, note, vel)})
				events = append(events, midiEvent{tick + halfDuration - 10, midi.NoteOff(0, note)})
			}
		}

	case "quarter":
		// Quarter notes - four strums per bar
		numQuarters := int(duration / quarterNote)
		if numQuarters == 0 {
			numQuarters = 1
		}
		for i := 0; i < numQuarters; i++ {
			tick := startTick + uint32(i)*quarterNote
			beat := (i % 4) + 1
			vel := uint8(70)
			if accentBeats[beat] {
				vel = 85
			}
			for _, note := range notes {
				events = append(events, midiEvent{tick, midi.NoteOn(0, note, vel)})
				events = append(events, midiEvent{tick + quarterNote - 10, midi.NoteOff(0, note)})
			}
		}

	case "eighth":
		// Eighth notes - eight strums per bar
		numEighths := int(duration / eighthNote)
		if numEighths == 0 {
			numEighths = 1
		}
		for i := 0; i < numEighths; i++ {
			tick := startTick + uint32(i)*eighthNote
			beat := (i/2)%4 + 1
			vel := uint8(65)
			if i%2 == 0 {
				vel = 75 // Downbeats louder
			}
			if accentBeats[beat] && i%2 == 0 {
				vel = 85
			}
			for _, note := range notes {
				events = append(events, midiEvent{tick, midi.NoteOn(0, note, vel)})
				events = append(events, midiEvent{tick + eighthNote - 10, midi.NoteOff(0, note)})
			}
		}

	case "strum_down":
		// Arpeggiated downstrum on each beat
		numQuarters := int(duration / quarterNote)
		if numQuarters == 0 {
			numQuarters = 1
		}
		for i := 0; i < numQuarters; i++ {
			tick := startTick + uint32(i)*quarterNote
			beat := (i % 4) + 1
			vel := uint8(70)
			if accentBeats[beat] {
				vel = 85
			}
			// Strum from low to high with slight delay
			strumDelay := uint32(15) // 15 ticks between each note
			for j, note := range notes {
				noteTick := tick + uint32(j)*strumDelay
				events = append(events, midiEvent{noteTick, midi.NoteOn(0, note, vel)})
				events = append(events, midiEvent{tick + quarterNote - 10, midi.NoteOff(0, note)})
			}
		}

	case "strum_up_down":
		// Alternating up/down strums on eighth notes
		numEighths := int(duration / eighthNote)
		if numEighths == 0 {
			numEighths = 1
		}
		for i := 0; i < numEighths; i++ {
			tick := startTick + uint32(i)*eighthNote
			vel := uint8(70)
			if i%2 == 0 {
				vel = 80 // Downstrums louder
			}
			strumDelay := uint32(12)
			noteOrder := notes
			if i%2 == 1 {
				// Upstrum - reverse note order
				noteOrder = reverseNotes(notes)
			}
			for j, note := range noteOrder {
				noteTick := tick + uint32(j)*strumDelay
				events = append(events, midiEvent{noteTick, midi.NoteOn(0, note, vel)})
				events = append(events, midiEvent{tick + eighthNote - 10, midi.NoteOff(0, note)})
			}
		}

	case "folk":
		// Folk/country strum: Bass note on 1,3 + chord on 2,4
		numBeats := int(duration / quarterNote)
		if numBeats == 0 {
			numBeats = 1
		}
		for i := 0; i < numBeats; i++ {
			tick := startTick + uint32(i)*quarterNote
			beat := (i % 4) + 1
			if beat == 1 || beat == 3 {
				// Bass note only (root)
				if len(notes) > 0 {
					events = append(events, midiEvent{tick, midi.NoteOn(0, notes[0], 85)})
					events = append(events, midiEvent{tick + quarterNote - 10, midi.NoteOff(0, notes[0])})
				}
			} else {
				// Full chord strum (higher notes)
				for j := 1; j < len(notes); j++ {
					events = append(events, midiEvent{tick, midi.NoteOn(0, notes[j], 70)})
					events = append(events, midiEvent{tick + quarterNote - 10, midi.NoteOff(0, notes[j])})
				}
			}
		}

	case "shuffle_strum":
		// Shuffle rhythm strumming (triplet feel)
		numBars := duration / ticksPerBar
		if numBars == 0 {
			numBars = 1
		}
		for bar := uint32(0); bar < numBars; bar++ {
			barStart := startTick + bar*ticksPerBar
			// Shuffle pattern: hit on triplet positions 0, 2, 3, 5, 6, 8, 9, 11
			shufflePattern := []int{0, 2, 3, 5, 6, 8, 9, 11}
			for _, pos := range shufflePattern {
				tick := barStart + uint32(pos)*tripletEighth
				// Apply swing
				if pos%3 == 2 {
					// This is an upbeat - could apply swing offset
					swingOffset := uint32(float64(tripletEighth) * (swing - 0.5) * 2)
					tick += swingOffset
				}
				vel := uint8(70)
				if pos%3 == 0 {
					vel = 80 // Accent downbeats
				}
				strumDelay := uint32(10)
				for j, note := range notes {
					noteTick := tick + uint32(j)*strumDelay
					events = append(events, midiEvent{noteTick, midi.NoteOn(0, note, vel)})
					events = append(events, midiEvent{tick + tripletEighth*2, midi.NoteOff(0, note)})
				}
			}
		}

	default:
		// Default to whole notes
		for _, note := range notes {
			events = append(events, midiEvent{startTick, midi.NoteOn(0, note, 80)})
			events = append(events, midiEvent{startTick + duration - 10, midi.NoteOff(0, note)})
		}
	}

	return events
}

// generateCustomPattern creates chord events from a custom pattern string
// Pattern notation:
//   D = down strum (loud, low to high)
//   U = up strum (softer, high to low)
//   d = soft down strum
//   u = soft up strum
//   x = muted/ghost strum (very short, percussive)
//   . = rest (silence)
//   - = tie/hold previous
// Pattern length determines subdivision (8 chars = 8th notes, 16 chars = 16th notes)
func generateCustomPattern(pattern string, notes ChordVoicing, startTick, duration, ticksPerBar uint32, swing float64) []midiEvent {
	events := []midiEvent{}

	if len(pattern) == 0 {
		return events
	}

	// Calculate how many times to repeat the pattern for this chord's duration
	patternLen := len(pattern)

	// Determine subdivision based on pattern length
	// Common patterns: 4 = quarter, 8 = eighth, 16 = sixteenth
	var ticksPerStep uint32
	numBars := duration / ticksPerBar
	if numBars == 0 {
		numBars = 1
	}

	// Pattern applies per bar, so total steps = patternLen * numBars
	ticksPerStep = ticksPerBar / uint32(patternLen)

	strumDelay := uint32(12) // Delay between notes in arpeggio

	for bar := uint32(0); bar < numBars; bar++ {
		barStart := startTick + bar*ticksPerBar

		for i, char := range pattern {
			stepTick := barStart + uint32(i)*ticksPerStep

			// Apply swing to off-beats (odd positions in pairs)
			if swing > 0.5 && i%2 == 1 {
				// Swing delays the off-beat
				swingAmount := uint32(float64(ticksPerStep) * (swing - 0.5) * 2)
				stepTick += swingAmount
			}

			switch char {
			case 'D': // Loud down strum
				events = append(events, strumChord(notes, stepTick, ticksPerStep, 85, strumDelay, false)...)

			case 'd': // Soft down strum
				events = append(events, strumChord(notes, stepTick, ticksPerStep, 65, strumDelay, false)...)

			case 'U': // Loud up strum
				events = append(events, strumChord(notes, stepTick, ticksPerStep, 75, strumDelay, true)...)

			case 'u': // Soft up strum
				events = append(events, strumChord(notes, stepTick, ticksPerStep, 55, strumDelay, true)...)

			case 'x', 'X': // Muted/ghost strum (short, percussive)
				events = append(events, strumChord(notes, stepTick, ticksPerStep/4, 50, strumDelay/2, false)...)

			case '.', '-', ' ':
				// Rest or hold - no new notes

			default:
				// Unknown character - treat as rest
			}
		}
	}

	return events
}

// strumChord creates strum events for a chord
func strumChord(notes ChordVoicing, startTick, duration uint32, velocity uint8, strumDelay uint32, upStrum bool) []midiEvent {
	events := []midiEvent{}

	noteOrder := notes
	if upStrum {
		noteOrder = reverseNotes(notes)
	}

	for i, note := range noteOrder {
		noteTick := startTick + uint32(i)*strumDelay
		// Slight velocity variation for more natural feel
		vel := velocity
		if i > 0 {
			vel -= uint8(i * 2) // Slightly softer for later notes
		}
		if vel < 30 {
			vel = 30
		}

		events = append(events, midiEvent{noteTick, midi.NoteOn(0, note, vel)})
		events = append(events, midiEvent{startTick + duration - 10, midi.NoteOff(0, note)})
	}

	return events
}

// parseAccentBeats parses accent string like "1,3" into a map
func parseAccentBeats(accent string) map[int]bool {
	result := map[int]bool{}
	parts := strings.Split(accent, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		switch p {
		case "1":
			result[1] = true
		case "2":
			result[2] = true
		case "3":
			result[3] = true
		case "4":
			result[4] = true
		}
	}
	return result
}

// reverseNotes returns a reversed copy of the notes slice
func reverseNotes(notes ChordVoicing) ChordVoicing {
	reversed := make(ChordVoicing, len(notes))
	for i, note := range notes {
		reversed[len(notes)-1-i] = note
	}
	return reversed
}
