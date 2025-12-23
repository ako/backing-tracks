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

	case "stride":
		// Stride/ragtime: chords only on beats 2 & 4 (the "pah" in "oom-pah")
		// Pairs with stride bass style which plays bass on 1 & 3
		numBeats := int(duration / quarterNote)
		if numBeats == 0 {
			numBeats = 1
		}
		for i := 0; i < numBeats; i++ {
			tick := startTick + uint32(i)*quarterNote
			beat := (i % 4) + 1
			if beat == 2 || beat == 4 {
				// Chord stab on backbeats
				vel := uint8(75)
				if beat == 2 {
					vel = 80 // Slightly accent beat 2
				}
				// Quick strum for that percussive ragtime feel
				strumDelay := uint32(8)
				for j, note := range notes {
					noteTick := tick + uint32(j)*strumDelay
					events = append(events, midiEvent{noteTick, midi.NoteOn(0, note, vel)})
					events = append(events, midiEvent{tick + quarterNote - 50, midi.NoteOff(0, note)})
				}
			}
		}

	case "ragtime":
		// Ragtime with syncopated accents: similar to stride but with some anticipation
		numBeats := int(duration / quarterNote)
		if numBeats == 0 {
			numBeats = 1
		}
		anticipation := uint32(eighthNote / 2) // 16th note anticipation
		for i := 0; i < numBeats; i++ {
			tick := startTick + uint32(i)*quarterNote
			beat := (i % 4) + 1
			if beat == 2 || beat == 4 {
				// Main chord on backbeats
				vel := uint8(78)
				strumDelay := uint32(8)
				for j, note := range notes {
					noteTick := tick + uint32(j)*strumDelay
					events = append(events, midiEvent{noteTick, midi.NoteOn(0, note, vel)})
					events = append(events, midiEvent{tick + quarterNote - 50, midi.NoteOff(0, note)})
				}
			} else if beat == 1 {
				// Occasional syncopated anticipation before beat 2
				if i+1 < numBeats {
					syncopTick := tick + quarterNote - anticipation
					for _, note := range notes {
						events = append(events, midiEvent{syncopTick, midi.NoteOn(0, note, 65)})
						events = append(events, midiEvent{syncopTick + anticipation - 10, midi.NoteOff(0, note)})
					}
				}
			}
		}

	case "travis":
		// Travis picking: alternating bass with finger melody
		// Pattern: Bass-high-mid-high (thumb-index-middle-index)
		events = append(events, travisPicking(notes, startTick, duration, ticksPerBar)...)

	case "fingerpick":
		// Classic folk fingerpicking: Bass on 1,3 + arpeggiated treble
		// Pattern: Bass-mid-high-mid-Bass-mid-high-mid (per bar)
		events = append(events, folkFingerpick(notes, startTick, duration, ticksPerBar)...)

	case "arpeggio_up":
		// Ascending arpeggio on each beat
		events = append(events, arpeggioPattern(notes, startTick, duration, ticksPerBar, false)...)

	case "arpeggio_down":
		// Descending arpeggio on each beat
		events = append(events, arpeggioPattern(notes, startTick, duration, ticksPerBar, true)...)

	case "fingerpick_slow":
		// Slower, more sparse fingerpicking for ballads
		events = append(events, slowFingerpick(notes, startTick, duration, ticksPerBar)...)

	case "funk":
		// Classic funk: 16th note pattern, heavy on the ONE, syncopated chops
		events = append(events, funkRhythm(notes, startTick, duration, ticksPerBar, false)...)

	case "funk_muted", "funk_chop":
		// Choppy/muted funk - more percussive, shorter notes
		events = append(events, funkRhythm(notes, startTick, duration, ticksPerBar, true)...)

	case "sixteenth", "16th":
		// Straight 16th notes
		sixteenthNote := ticksPerBar / 16
		numSixteenths := int(duration / sixteenthNote)
		for i := 0; i < numSixteenths; i++ {
			tick := startTick + uint32(i)*sixteenthNote
			beat := (i / 4) % 4
			vel := uint8(60)
			if i%4 == 0 {
				vel = 75 // Accent on quarter note positions
			}
			if beat == 0 && i%4 == 0 {
				vel = 90 // Heavy accent on the ONE
			}
			for _, note := range notes {
				events = append(events, midiEvent{tick, midi.NoteOn(0, note, vel)})
				events = append(events, midiEvent{tick + sixteenthNote - 15, midi.NoteOff(0, note)})
			}
		}

	case "ska", "skank":
		// Ska off-beat "skank" - chords on the "and" of each beat
		// Classic Madness/Specials/Doe Maar style
		eighthNote := ticksPerBar / 8
		for i := 0; i < int(duration/eighthNote); i++ {
			// Only play on off-beats (1, 3, 5, 7 in 8th note grid)
			if i%2 == 1 {
				tick := startTick + uint32(i)*eighthNote
				vel := uint8(85)
				// Short, choppy chords
				noteLen := eighthNote * 2 / 3
				for _, note := range notes {
					events = append(events, midiEvent{tick, midi.NoteOn(0, note, vel)})
					events = append(events, midiEvent{tick + noteLen, midi.NoteOff(0, note)})
				}
			}
		}

	case "reggae", "one_drop":
		// Reggae rhythm - emphasis on beat 3, off-beat chops
		// Bob Marley/Peter Tosh style
		eighthNote := ticksPerBar / 8
		for i := 0; i < int(duration/eighthNote); i++ {
			tick := startTick + uint32(i)*eighthNote
			beat := i / 2
			isOffBeat := i%2 == 1

			if isOffBeat {
				// Off-beat chops (lighter)
				vel := uint8(65)
				noteLen := eighthNote / 2
				for _, note := range notes {
					events = append(events, midiEvent{tick, midi.NoteOn(0, note, vel)})
					events = append(events, midiEvent{tick + noteLen, midi.NoteOff(0, note)})
				}
			} else if beat == 2 {
				// Heavy accent on beat 3 (the "one drop" feel)
				vel := uint8(90)
				noteLen := eighthNote
				for _, note := range notes {
					events = append(events, midiEvent{tick, midi.NoteOn(0, note, vel)})
					events = append(events, midiEvent{tick + noteLen, midi.NoteOff(0, note)})
				}
			}
		}

	case "country", "train":
		// Country train beat / boom-chick pattern
		// Bass note on 1,3 - chord on 2,4
		quarterNote := ticksPerBar / 4
		for i := 0; i < int(duration/quarterNote); i++ {
			tick := startTick + uint32(i)*quarterNote
			if i%2 == 0 {
				// Beats 1, 3: bass note only (root)
				if len(notes) > 0 {
					vel := uint8(85)
					events = append(events, midiEvent{tick, midi.NoteOn(0, notes[0], vel)})
					events = append(events, midiEvent{tick + quarterNote - 20, midi.NoteOff(0, notes[0])})
				}
			} else {
				// Beats 2, 4: full chord (shorter, snappier)
				vel := uint8(75)
				noteLen := quarterNote * 2 / 3
				for _, note := range notes {
					events = append(events, midiEvent{tick, midi.NoteOn(0, note, vel)})
					events = append(events, midiEvent{tick + noteLen, midi.NoteOff(0, note)})
				}
			}
		}

	case "disco":
		// Disco - four on the floor with 16th note embellishments
		sixteenthNote := ticksPerBar / 16
		for i := 0; i < int(duration/sixteenthNote); i++ {
			tick := startTick + uint32(i)*sixteenthNote
			isQuarterBeat := i%4 == 0
			isOffBeat := i%2 == 1

			if isQuarterBeat {
				// Strong chord on quarter notes
				vel := uint8(85)
				noteLen := sixteenthNote * 3
				for _, note := range notes {
					events = append(events, midiEvent{tick, midi.NoteOn(0, note, vel)})
					events = append(events, midiEvent{tick + noteLen, midi.NoteOff(0, note)})
				}
			} else if isOffBeat && (i%4 == 1 || i%4 == 3) {
				// Light off-beat hits
				vel := uint8(55)
				noteLen := sixteenthNote / 2
				for _, note := range notes {
					events = append(events, midiEvent{tick, midi.NoteOn(0, note, vel)})
					events = append(events, midiEvent{tick + noteLen, midi.NoteOff(0, note)})
				}
			}
		}

	case "motown", "soul":
		// Motown/Soul - tight rhythm, emphasis on 2 and 4
		// Otis Redding / Stax style
		eighthNote := ticksPerBar / 8
		for i := 0; i < int(duration/eighthNote); i++ {
			tick := startTick + uint32(i)*eighthNote
			beat := i / 2
			isOnBeat := i%2 == 0

			vel := uint8(70)
			noteLen := eighthNote * 2 / 3

			if isOnBeat && (beat == 1 || beat == 3) {
				// Heavy backbeat on 2 and 4
				vel = 90
				noteLen = eighthNote
			} else if isOnBeat && beat == 0 {
				// Moderate on beat 1
				vel = 80
			}

			for _, note := range notes {
				events = append(events, midiEvent{tick, midi.NoteOn(0, note, vel)})
				events = append(events, midiEvent{tick + noteLen, midi.NoteOff(0, note)})
			}
		}

	case "flamenco", "rumba":
		// Flamenco rumba pattern - syncopated with strong accents
		sixteenthNote := ticksPerBar / 16
		// Classic flamenco pattern: accent pattern across 16ths
		// 1 . . 2 . . 3 . 4 . 5 . 6 . . .
		accentPattern := []int{0, 3, 6, 8, 10, 12} // positions to play
		accentVels := []uint8{95, 75, 75, 85, 75, 80}

		for bar := uint32(0); bar < duration/ticksPerBar; bar++ {
			for idx, pos := range accentPattern {
				tick := startTick + bar*ticksPerBar + uint32(pos)*sixteenthNote
				if tick >= startTick+duration {
					break
				}
				vel := accentVels[idx]
				noteLen := sixteenthNote
				if idx == 0 {
					noteLen = sixteenthNote * 2 // longer on the ONE
				}
				for _, note := range notes {
					events = append(events, midiEvent{tick, midi.NoteOn(0, note, vel)})
					events = append(events, midiEvent{tick + noteLen, midi.NoteOff(0, note)})
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

// travisPicking generates Travis-style fingerpicking
// Alternating bass with syncopated treble notes
// Pattern per beat: Bass-High-Mid-High (thumb plays bass, fingers play treble)
func travisPicking(notes ChordVoicing, startTick, duration, ticksPerBar uint32) []midiEvent {
	events := []midiEvent{}
	eighthNote := ticksPerBar / 8
	numBars := duration / ticksPerBar
	if numBars == 0 {
		numBars = 1
	}

	// Need at least 3 notes for proper Travis picking
	if len(notes) < 3 {
		return events
	}

	bass := notes[0]        // Root (thumb)
	fifth := notes[0] + 7   // Fifth for alternating bass
	mid := notes[1]         // Middle note (middle finger)
	high := notes[len(notes)-1] // Highest note (index finger)

	for bar := uint32(0); bar < numBars; bar++ {
		barStart := startTick + bar*ticksPerBar

		// 8 eighth notes per bar
		// Pattern: B1-H-M-H-B5-H-M-H (Bass1, High, Mid, High, Bass5, High, Mid, High)
		pattern := []struct {
			note uint8
			vel  uint8
		}{
			{bass, 80},  // 1: Bass on root
			{high, 60},  // &: High
			{mid, 55},   // 2: Mid
			{high, 60},  // &: High
			{fifth, 75}, // 3: Bass on fifth
			{high, 60},  // &: High
			{mid, 55},   // 4: Mid
			{high, 60},  // &: High
		}

		for i, p := range pattern {
			tick := barStart + uint32(i)*eighthNote
			noteDuration := eighthNote - 20
			events = append(events, midiEvent{tick, midi.NoteOn(0, p.note, p.vel)})
			events = append(events, midiEvent{tick + noteDuration, midi.NoteOff(0, p.note)})
		}
	}

	return events
}

// folkFingerpick generates classic folk fingerpicking pattern
// Similar to songs like "Dust in the Wind" or Leonard Cohen style
func folkFingerpick(notes ChordVoicing, startTick, duration, ticksPerBar uint32) []midiEvent {
	events := []midiEvent{}
	sixteenthNote := ticksPerBar / 16
	numBars := duration / ticksPerBar
	if numBars == 0 {
		numBars = 1
	}

	if len(notes) < 3 {
		return events
	}

	// Assign notes to fingers
	bass := notes[0]
	low := notes[0]
	if len(notes) > 1 {
		low = notes[1]
	}
	mid := notes[len(notes)/2]
	high := notes[len(notes)-1]

	for bar := uint32(0); bar < numBars; bar++ {
		barStart := startTick + bar*ticksPerBar

		// 16 sixteenth notes per bar
		// Classic pattern: B-L-M-H-M-L-B-L-M-H-M-L-B-L-M-H
		// (Bass, Low, Mid, High, Mid, Low, repeat with variations)
		pattern := []struct {
			note uint8
			vel  uint8
		}{
			{bass, 80}, // 1
			{mid, 55},  // e
			{high, 60}, // &
			{mid, 50},  // a
			{bass, 75}, // 2
			{mid, 55},  // e
			{high, 60}, // &
			{mid, 50},  // a
			{bass, 80}, // 3
			{mid, 55},  // e
			{high, 60}, // &
			{mid, 50},  // a
			{low, 70},  // 4
			{mid, 55},  // e
			{high, 60}, // &
			{mid, 50},  // a
		}

		for i, p := range pattern {
			tick := barStart + uint32(i)*sixteenthNote
			noteDuration := sixteenthNote*2 - 10 // Notes ring a bit
			events = append(events, midiEvent{tick, midi.NoteOn(0, p.note, p.vel)})
			events = append(events, midiEvent{tick + noteDuration, midi.NoteOff(0, p.note)})
		}
	}

	return events
}

// slowFingerpick generates a slower, more sparse fingerpicking for ballads
// Good for Leonard Cohen, Nick Drake style
func slowFingerpick(notes ChordVoicing, startTick, duration, ticksPerBar uint32) []midiEvent {
	events := []midiEvent{}
	eighthNote := ticksPerBar / 8
	numBars := duration / ticksPerBar
	if numBars == 0 {
		numBars = 1
	}

	if len(notes) < 2 {
		return events
	}

	bass := notes[0]
	mid := notes[len(notes)/2]
	high := notes[len(notes)-1]

	for bar := uint32(0); bar < numBars; bar++ {
		barStart := startTick + bar*ticksPerBar

		// Sparse pattern: B---M-H- B---H-M- (eighth notes, - = rest)
		// Positions: 0, 4, 5 for first half; 0, 4, 5 for second half with variation
		pattern := []struct {
			pos  int
			note uint8
			vel  uint8
		}{
			{0, bass, 80}, // Beat 1: Bass
			{2, mid, 55},  // Beat 2: Mid
			{3, high, 60}, // Beat 2&: High
			{4, bass, 75}, // Beat 3: Bass
			{6, high, 60}, // Beat 4: High
			{7, mid, 50},  // Beat 4&: Mid
		}

		for _, p := range pattern {
			tick := barStart + uint32(p.pos)*eighthNote
			noteDuration := eighthNote*2 - 10
			events = append(events, midiEvent{tick, midi.NoteOn(0, p.note, p.vel)})
			events = append(events, midiEvent{tick + noteDuration, midi.NoteOff(0, p.note)})
		}
	}

	return events
}

// funkRhythm generates classic funk rhythm guitar pattern
// Heavy on the ONE, syncopated 16th note scratches and chops
func funkRhythm(notes ChordVoicing, startTick, duration, ticksPerBar uint32, muted bool) []midiEvent {
	events := []midiEvent{}
	sixteenthNote := ticksPerBar / 16
	numBars := duration / ticksPerBar
	if numBars == 0 {
		numBars = 1
	}

	// Note duration - muted = very short/choppy, normal = slightly longer
	noteDur := sixteenthNote - 20
	if muted {
		noteDur = sixteenthNote / 2
	}

	for bar := uint32(0); bar < numBars; bar++ {
		barStart := startTick + bar*ticksPerBar

		// Classic funk pattern (16 sixteenth notes per bar):
		// Position:  1 e & a 2 e & a 3 e & a 4 e & a
		// Pattern:   X . x . . x X . x . x . . x . x
		// X = heavy hit, x = lighter hit, . = rest
		// This creates that syncopated, bouncy funk feel
		funkPattern := []struct {
			pos int
			vel uint8
			hit bool
		}{
			{0, 95, true},   // 1 - THE ONE (heavy!)
			{1, 0, false},   // e
			{2, 60, true},   // &
			{3, 0, false},   // a
			{4, 0, false},   // 2
			{5, 65, true},   // e (syncopation!)
			{6, 80, true},   // & (accent)
			{7, 0, false},   // a
			{8, 70, true},   // 3
			{9, 0, false},   // e
			{10, 60, true},  // &
			{11, 0, false},  // a
			{12, 0, false},  // 4
			{13, 65, true},  // e (syncopation!)
			{14, 0, false},  // a
			{15, 70, true},  // a (pickup to next bar)
		}

		for _, p := range funkPattern {
			if !p.hit {
				continue
			}
			tick := barStart + uint32(p.pos)*sixteenthNote
			vel := p.vel
			if muted {
				vel = vel - 10 // Muted is slightly softer
				if vel < 50 {
					vel = 50
				}
			}

			// Quick strum for that choppy funk sound
			strumDelay := uint32(5)
			for j, note := range notes {
				noteTick := tick + uint32(j)*strumDelay
				events = append(events, midiEvent{noteTick, midi.NoteOn(0, note, vel)})
				events = append(events, midiEvent{tick + noteDur, midi.NoteOff(0, note)})
			}
		}
	}

	return events
}

// arpeggioPattern generates ascending or descending arpeggios
func arpeggioPattern(notes ChordVoicing, startTick, duration, ticksPerBar uint32, descending bool) []midiEvent {
	events := []midiEvent{}
	quarterNote := ticksPerBar / 4
	numBeats := int(duration / quarterNote)
	if numBeats == 0 {
		numBeats = 1
	}

	noteCount := len(notes)
	if noteCount == 0 {
		return events
	}

	// Time per note within a beat
	noteSpacing := quarterNote / uint32(noteCount)

	for beat := 0; beat < numBeats; beat++ {
		beatStart := startTick + uint32(beat)*quarterNote
		vel := uint8(70)
		if beat%4 == 0 {
			vel = 80 // Accent beat 1
		}

		for i := 0; i < noteCount; i++ {
			noteIdx := i
			if descending {
				noteIdx = noteCount - 1 - i
			}

			tick := beatStart + uint32(i)*noteSpacing
			noteDuration := noteSpacing - 10
			noteVel := vel - uint8(i*3) // Softer for later notes
			if noteVel < 40 {
				noteVel = 40
			}

			events = append(events, midiEvent{tick, midi.NoteOn(0, notes[noteIdx], noteVel)})
			events = append(events, midiEvent{tick + noteDuration, midi.NoteOff(0, notes[noteIdx])})
		}
	}

	return events
}
