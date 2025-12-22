package midi

import (
	"backing-tracks/parser"
)

// BassNote represents a single bass note with timing
type BassNote struct {
	Note     uint8   // MIDI note number
	Tick     uint32  // When to play (in ticks)
	Duration uint32  // How long to hold (in ticks)
	Velocity uint8   // Note velocity (volume)
}

// GenerateBassLine creates bass notes from a chord progression
func GenerateBassLine(chords []parser.Chord, bass *parser.Bass, ticksPerBar uint32) []BassNote {
	if bass == nil {
		return nil
	}

	notes := []BassNote{}
	currentTick := uint32(0)

	// Determine swing ratio (0.5 = straight, 0.67 = triplet swing)
	swing := 0.5
	if bass.Swing > 0 {
		swing = bass.Swing
	}

	for _, chord := range chords {
		root := parseRoot(chord.Symbol)
		// Support fractional bars by multiplying float first
		barDuration := uint32(float64(ticksPerBar) * chord.Bars)

		switch bass.Style {
		case "root":
			// Just root notes on downbeats
			notes = append(notes, BassNote{
				Note:     root + 36, // Bass octave (E1 = 28, A1 = 33)
				Tick:     currentTick,
				Duration: barDuration - 10,
				Velocity: 90,
			})

		case "root_fifth":
			// Root on 1, fifth on 3
			notes = append(notes, BassNote{
				Note:     root + 36,
				Tick:     currentTick,
				Duration: ticksPerBar/2 - 10,
				Velocity: 90,
			})
			notes = append(notes, BassNote{
				Note:     root + 36 + 7, // Fifth
				Tick:     currentTick + ticksPerBar/2,
				Duration: ticksPerBar/2 - 10,
				Velocity: 85,
			})

		case "walking":
			// Walking bass: root, 3rd, 5th, 6th (or 7th)
			quarterNote := ticksPerBar / 4
			third := getThird(chord.Symbol)
			seventh := getSeventh(chord.Symbol)

			pattern := []uint8{
				root + 36,        // Root
				root + 36 + third, // 3rd
				root + 36 + 7,    // 5th
				root + 36 + seventh, // 7th or 6th
			}

			for i, note := range pattern {
				tick := currentTick + uint32(i)*quarterNote
				notes = append(notes, BassNote{
					Note:     note,
					Tick:     tick,
					Duration: quarterNote - 10,
					Velocity: 85,
				})
			}

		case "swing_walking":
			// Swung walking bass (for jazz/blues)
			quarterNote := ticksPerBar / 4
			third := getThird(chord.Symbol)
			seventh := getSeventh(chord.Symbol)

			pattern := []uint8{
				root + 36,
				root + 36 + third,
				root + 36 + 7,
				root + 36 + seventh,
			}

			for i, note := range pattern {
				// Apply swing feel to each beat pair
				var tick uint32
				beatPair := i / 2  // Which pair (0 or 1)
				isOffbeat := i % 2 == 1

				pairStart := currentTick + uint32(beatPair*2)*quarterNote

				if !isOffbeat {
					// On beats - normal timing
					tick = pairStart
				} else {
					// Off beats - delayed based on swing ratio
					// swing=0.5 means 50/50 (straight), swing=0.67 means 67/33 (triplet)
					tick = pairStart + uint32(float64(quarterNote*2)*swing)
				}

				notes = append(notes, BassNote{
					Note:     note,
					Tick:     tick,
					Duration: quarterNote - 10,
					Velocity: 85,
				})
			}

		default:
			// Default to simple root notes
			notes = append(notes, BassNote{
				Note:     root + 36,
				Tick:     currentTick,
				Duration: barDuration - 10,
				Velocity: 90,
			})
		}

		currentTick += barDuration
	}

	return notes
}

// getThird returns the third interval (major or minor)
func getThird(chordSymbol string) uint8 {
	quality := parseQuality(chordSymbol)
	if quality == "m" || quality == "m7" {
		return 3 // Minor third
	}
	return 4 // Major third
}

// getSeventh returns the seventh interval
func getSeventh(chordSymbol string) uint8 {
	quality := parseQuality(chordSymbol)
	switch quality {
	case "7":
		return 10 // Minor 7th (dominant)
	case "maj7", "^7":
		return 11 // Major 7th
	case "m7":
		return 10 // Minor 7th
	default:
		return 9 // Major 6th (as a substitute)
	}
}
