package midi

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"backing-tracks/parser"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

// midiEvent represents a MIDI event with absolute timing
type midiEvent struct {
	tick    uint32
	message midi.Message
}

// ChordVoicing represents MIDI note numbers for a chord
type ChordVoicing []uint8

// GenerateFromTrack creates a MIDI file from a track
func GenerateFromTrack(track *parser.Track) (string, error) {
	// Create temporary MIDI file
	tmpFile := "/tmp/backing-track.mid"

	// Create SMF (Standard MIDI File)
	s := smf.New()
	s.TimeFormat = smf.MetricTicks(480) // 480 ticks per quarter note

	// Track 0: Tempo and metadata
	var track0 smf.Track
	track0.Add(0, smf.MetaTempo(float64(track.Info.Tempo)))
	track0.Close(0)
	s.Add(track0)

	// Track 1: Chord progression
	var track1 smf.Track

	// Set program (0 = Acoustic Grand Piano)
	track1.Add(0, midi.ProgramChange(0, 0))

	chords := track.Progression.GetChords()

	// Calculate ticks per bar (4/4 time signature assumed)
	// 480 ticks per quarter note * 4 quarter notes = 1920 ticks per bar
	ticksPerBar := uint32(1920)

	// Generate chord events using rhythm pattern
	chordEvents := GenerateChordRhythm(chords, track.Rhythm, ticksPerBar)

	// Calculate total duration for later use
	currentTick := uint32(0)
	for _, chord := range chords {
		currentTick += uint32(chord.Bars * float64(ticksPerBar))
	}

	// Sort events by absolute tick
	sort.Slice(chordEvents, func(i, j int) bool {
		return chordEvents[i].tick < chordEvents[j].tick
	})

	// Add events with DELTA times (Track.Add expects delta, not absolute!)
	prevTick := uint32(0)
	for _, evt := range chordEvents {
		delta := evt.tick - prevTick
		track1.Add(delta, evt.message)
		prevTick = evt.tick
	}

	track1.Close(0)
	s.Add(track1)

	// Track 2: Bass (channel 1)
	bassCount := 0
	if track.Bass != nil {
		var track2 smf.Track
		// Set program (33 = Fingered Bass)
		track2.Add(0, midi.ProgramChange(1, 33))

		bassNotes := GenerateBassLine(chords, track.Bass, ticksPerBar)
		bassCount = len(bassNotes)
		// Debug: print first few bass notes
		if len(bassNotes) > 0 {
			fmt.Printf("[MIDI] Sample bass notes:\n")
			for i := 0; i < min(4, len(bassNotes)); i++ {
				fmt.Printf("  Note %d: MIDI#%d at tick %d (bar %.1f)\n",
					i+1, bassNotes[i].Note, bassNotes[i].Tick, float64(bassNotes[i].Tick)/float64(ticksPerBar))
			}
		}

		// Collect bass events with absolute ticks
		var bassEvents []midiEvent
		for _, note := range bassNotes {
			bassEvents = append(bassEvents, midiEvent{note.Tick, midi.NoteOn(1, note.Note, note.Velocity)})
			bassEvents = append(bassEvents, midiEvent{note.Tick + note.Duration, midi.NoteOff(1, note.Note)})
		}
		sort.Slice(bassEvents, func(i, j int) bool {
			return bassEvents[i].tick < bassEvents[j].tick
		})

		// Add with delta times
		prevTick := uint32(0)
		for _, evt := range bassEvents {
			delta := evt.tick - prevTick
			track2.Add(delta, evt.message)
			prevTick = evt.tick
		}

		track2.Close(0)
		s.Add(track2)
	}

	// Track 3: Drums (channel 9 - standard MIDI drum channel)
	drumCount := 0
	if track.Drums != nil {
		var track3 smf.Track

		totalBars := track.Progression.TotalBars()
		drumNotes := GenerateDrumPattern(totalBars, track.Drums, ticksPerBar)
		drumCount = len(drumNotes)

		// Collect drum events with absolute ticks
		var drumEvents []midiEvent
		for _, note := range drumNotes {
			drumEvents = append(drumEvents, midiEvent{note.Tick, midi.NoteOn(9, note.Note, note.Velocity)})
			drumEvents = append(drumEvents, midiEvent{note.Tick + 10, midi.NoteOff(9, note.Note)})
		}
		sort.Slice(drumEvents, func(i, j int) bool {
			return drumEvents[i].tick < drumEvents[j].tick
		})

		// Add with delta times
		prevTick := uint32(0)
		for _, evt := range drumEvents {
			delta := evt.tick - prevTick
			track3.Add(delta, evt.message)
			prevTick = evt.tick
		}

		track3.Close(0)
		s.Add(track3)
	}

	// Debug output
	chordEventCount := len(chordEvents) / 2 // Divide by 2 since each note has on+off
	fmt.Printf("\n[MIDI] Generated %d chord events, %d bass notes, %d drum hits\n", chordEventCount, bassCount, drumCount)
	fmt.Printf("[MIDI] Tracks: %d (tempo + chords + bass + drums)\n", len(s.Tracks))
	fmt.Printf("[MIDI] Channels: Chords=0 (Piano), Bass=1 (Fingered Bass), Drums=9 (GM Drums)\n")
	fmt.Printf("[MIDI] Total duration: %d ticks (%d bars)\n", currentTick, currentTick/ticksPerBar)

	// Write to file
	f, err := os.Create(tmpFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := s.WriteTo(f); err != nil {
		return "", err
	}

	return tmpFile, nil
}

// getChordVoicing returns MIDI note numbers for a chord symbol
func getChordVoicing(symbol string) ChordVoicing {
	// Parse chord symbol
	root := parseRoot(symbol)
	quality := parseQuality(symbol)

	// Base octave (middle C = 60, we'll use octave 3 and 4)
	rootNote := root + 48 // Octave 3

	switch quality {
	case "7": // Dominant 7th (e.g., A7)
		return ChordVoicing{
			rootNote,           // Root
			rootNote + 4,       // Major 3rd
			rootNote + 7,       // Perfect 5th
			rootNote + 10,      // Minor 7th
		}
	case "maj7", "^7": // Major 7th
		return ChordVoicing{
			rootNote,
			rootNote + 4,
			rootNote + 7,
			rootNote + 11,
		}
	case "m7": // Minor 7th
		return ChordVoicing{
			rootNote,
			rootNote + 3,  // Minor 3rd
			rootNote + 7,
			rootNote + 10,
		}
	case "m": // Minor triad
		return ChordVoicing{
			rootNote,
			rootNote + 3,
			rootNote + 7,
		}
	case "5": // Power chord
		return ChordVoicing{
			rootNote,
			rootNote + 7,
			rootNote + 12, // Octave
		}
	default: // Major triad
		return ChordVoicing{
			rootNote,
			rootNote + 4,
			rootNote + 7,
		}
	}
}

// parseRoot extracts the root note from a chord symbol
func parseRoot(symbol string) uint8 {
	// Get first character(s) for root note
	root := strings.ToUpper(string(symbol[0]))

	// Check for sharp or flat
	if len(symbol) > 1 {
		if symbol[1] == '#' || symbol[1] == 'b' {
			root += string(symbol[1])
		}
	}

	// Map to MIDI note number (C=0, C#=1, D=2, etc.)
	noteMap := map[string]uint8{
		"C":  0,
		"C#": 1, "DB": 1,
		"D":  2,
		"D#": 3, "EB": 3,
		"E":  4,
		"F":  5,
		"F#": 6, "GB": 6,
		"G":  7,
		"G#": 8, "AB": 8,
		"A":  9,
		"A#": 10, "BB": 10,
		"B":  11,
	}

	if note, ok := noteMap[root]; ok {
		return note
	}

	return 0 // Default to C
}

// parseQuality extracts the chord quality/type
func parseQuality(symbol string) string {
	// Remove root note
	quality := symbol
	if len(symbol) > 0 {
		quality = symbol[1:]
	}
	if len(quality) > 0 && (quality[0] == '#' || quality[0] == 'b') {
		quality = quality[1:]
	}

	// Common quality patterns
	if quality == "" {
		return "major"
	}
	if strings.HasPrefix(quality, "maj7") || quality == "^7" {
		return "maj7"
	}
	if strings.HasPrefix(quality, "m7") {
		return "m7"
	}
	if quality == "7" {
		return "7"
	}
	if quality == "m" {
		return "m"
	}
	if quality == "5" {
		return "5"
	}

	return "major"
}
