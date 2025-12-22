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
		root := parseBassNote(chord.Symbol) // Use bass note for slash chords (Am/G → G)
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

		case "stride":
			// Stride bass for ragtime/stride piano: low bass on 1 & 3
			// The "oom" in "oom-pah" - pairs with stride rhythm style for chords on 2 & 4
			quarterNote := ticksPerBar / 4
			fifth := root + 7

			// Beat 1: Root (low octave)
			notes = append(notes, BassNote{
				Note:     root + 28, // Low bass octave
				Tick:     currentTick,
				Duration: quarterNote - 20,
				Velocity: 95,
			})
			// Beat 3: Fifth (low octave)
			notes = append(notes, BassNote{
				Note:     fifth + 28, // Low bass octave
				Tick:     currentTick + quarterNote*2,
				Duration: quarterNote - 20,
				Velocity: 90,
			})

		case "boogie":
			// Boogie-woogie bass: driving eighth note pattern
			eighthNote := ticksPerBar / 8

			// Classic boogie pattern: 1-1-5-6-b7-6-5-5
			boogiePattern := []uint8{
				root + 36,      // 1
				root + 36,      // 1
				root + 36 + 7,  // 5
				root + 36 + 9,  // 6
				root + 36 + 10, // b7
				root + 36 + 9,  // 6
				root + 36 + 7,  // 5
				root + 36 + 7,  // 5
			}

			for i, note := range boogiePattern {
				notes = append(notes, BassNote{
					Note:     note,
					Tick:     currentTick + uint32(i)*eighthNote,
					Duration: eighthNote - 15,
					Velocity: uint8(85 + (i%2)*5), // Slight accent on downbeats
				})
			}

		case "808", "sub":
			// 808 sub bass: heavy sustained notes with syncopation
			// Low, long notes that sustain through the bar
			quarterNote := ticksPerBar / 4
			eighthNote := ticksPerBar / 8

			// Pattern: hit on 1, and-of-2, 4 (common EDM pattern)
			notes = append(notes, BassNote{
				Note:     root + 28, // Very low octave for sub bass
				Tick:     currentTick,
				Duration: quarterNote + eighthNote, // Sustain through beat 2
				Velocity: 110,                      // Heavy!
			})
			notes = append(notes, BassNote{
				Note:     root + 28,
				Tick:     currentTick + quarterNote + eighthNote, // And of 2
				Duration: quarterNote,
				Velocity: 100,
			})
			notes = append(notes, BassNote{
				Note:     root + 28,
				Tick:     currentTick + 3*quarterNote, // Beat 4
				Duration: quarterNote - 20,
				Velocity: 105,
			})

		case "808_octave", "edm":
			// EDM bass with octave jumps
			quarterNote := ticksPerBar / 4
			eighthNote := ticksPerBar / 8

			// Pattern with octave movement
			notes = append(notes, BassNote{
				Note:     root + 28, // Low
				Tick:     currentTick,
				Duration: eighthNote,
				Velocity: 110,
			})
			notes = append(notes, BassNote{
				Note:     root + 40, // High octave
				Tick:     currentTick + eighthNote,
				Duration: eighthNote - 10,
				Velocity: 95,
			})
			notes = append(notes, BassNote{
				Note:     root + 28,
				Tick:     currentTick + 2*eighthNote,
				Duration: quarterNote,
				Velocity: 105,
			})
			notes = append(notes, BassNote{
				Note:     root + 28,
				Tick:     currentTick + 2*quarterNote,
				Duration: eighthNote,
				Velocity: 110,
			})
			notes = append(notes, BassNote{
				Note:     root + 40,
				Tick:     currentTick + 2*quarterNote + eighthNote,
				Duration: eighthNote - 10,
				Velocity: 90,
			})
			notes = append(notes, BassNote{
				Note:     root + 28,
				Tick:     currentTick + 3*quarterNote,
				Duration: quarterNote - 20,
				Velocity: 100,
			})

		case "funk", "slap":
			// Funk/slap bass: syncopated 16th note pattern with octaves
			// Heavy on the ONE, ghost notes, and syncopation
			sixteenthNote := ticksPerBar / 16

			// Classic funk bass pattern - emphasizes the one, adds octaves and ghost notes
			// Pattern: ROOT . oct . . r . R . oct . r . . oct .
			funkBassPattern := []struct {
				pos      int
				interval int // 0=root, 12=octave, 7=fifth
				vel      uint8
			}{
				{0, 0, 100},  // 1 - THE ONE
				{2, 12, 70},  // & - octave (softer)
				{5, 0, 60},   // 2e - ghost note
				{6, 0, 90},   // 2& - accent
				{8, 0, 85},   // 3
				{10, 12, 70}, // 3& - octave
				{12, 0, 65},  // 4 - softer
				{14, 12, 75}, // 4& - octave pickup
			}

			for _, p := range funkBassPattern {
				tick := currentTick + uint32(p.pos)*sixteenthNote
				notes = append(notes, BassNote{
					Note:     root + 36 + uint8(p.interval),
					Tick:     tick,
					Duration: sixteenthNote - 15,
					Velocity: p.vel,
				})
			}

		case "funk_simple":
			// Simpler funk bass - root and fifth with syncopation
			sixteenthNote := ticksPerBar / 16
			fifth := root + 7

			// Simpler pattern: Root on 1, syncopated hits
			simpleFunk := []struct {
				pos  int
				note uint8
				vel  uint8
			}{
				{0, root + 36, 95},   // 1
				{6, fifth + 36, 80},  // 2&
				{10, root + 36, 75},  // 3&
				{12, fifth + 36, 70}, // 4
				{15, root + 36, 80},  // pickup
			}

			for _, p := range simpleFunk {
				tick := currentTick + uint32(p.pos)*sixteenthNote
				notes = append(notes, BassNote{
					Note:     p.note,
					Tick:     tick,
					Duration: sixteenthNote*2 - 15,
					Velocity: p.vel,
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
