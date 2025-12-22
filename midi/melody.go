package midi

import (
	"math/rand"
	"strings"
	"time"

	"backing-tracks/parser"
	"backing-tracks/theory"
)

// MelodyNote represents a generated melody note
type MelodyNote struct {
	Note     uint8  // MIDI note number
	Tick     uint32 // When to play
	Duration uint32 // How long to hold
	Velocity uint8  // Note velocity
}

// MelodyStyle defines the complexity/density of generated melody
type MelodyStyle string

const (
	MelodySimple       MelodyStyle = "simple"        // Mostly half/whole notes, chord tones
	MelodyModerate     MelodyStyle = "moderate"      // Quarter notes, some passing tones
	MelodyActive       MelodyStyle = "active"        // Eighth notes, more motion
	MelodyBluesHead    MelodyStyle = "blues_head"    // Classic AAB 12-bar blues vocal pattern
	MelodyCallResponse MelodyStyle = "call_response" // Same as blues_head
)

// MelodyConfig holds melody generation settings
type MelodyConfig struct {
	Style         MelodyStyle
	Octave        int     // Base octave (default 4)
	Density       float64 // 0.0-1.0, how many notes to play
	UseChordTones bool    // Prioritize chord tones on strong beats
}

// DefaultMelodyConfig returns sensible defaults
func DefaultMelodyConfig() *MelodyConfig {
	return &MelodyConfig{
		Style:         MelodySimple,
		Octave:        4,
		Density:       0.5,
		UseChordTones: true,
	}
}

// GenerateMelody creates a melody line for the track
func GenerateMelody(chords []parser.Chord, key string, style string, config *MelodyConfig, ticksPerBar uint32) []MelodyNote {
	if config == nil {
		config = DefaultMelodyConfig()
	}

	// Seed random for variation
	rand.Seed(time.Now().UnixNano())

	// Use special generator for blues head / call-response style
	if config.Style == MelodyBluesHead || config.Style == MelodyCallResponse {
		return generateBluesHead(chords, key, style, config, ticksPerBar)
	}

	notes := []MelodyNote{}
	currentTick := uint32(0)

	// Start in comfortable guitar range (MIDI 52-72 = E3-C5)
	baseNote := 52 + (config.Octave-3)*12
	currentNote := baseNote + 7 // Start on 5th degree
	direction := 1              // 1 = ascending, -1 = descending

	// Determine note durations based on style
	var noteDuration uint32
	var noteSpacing uint32

	switch config.Style {
	case MelodySimple:
		noteDuration = ticksPerBar / 2  // Half notes
		noteSpacing = ticksPerBar / 2
	case MelodyModerate:
		noteDuration = ticksPerBar / 4  // Quarter notes
		noteSpacing = ticksPerBar / 4
	case MelodyActive:
		noteDuration = ticksPerBar / 8  // Eighth notes
		noteSpacing = ticksPerBar / 8
	default:
		noteDuration = ticksPerBar / 4
		noteSpacing = ticksPerBar / 4
	}

	for _, chord := range chords {
		chordDuration := uint32(chord.Bars * float64(ticksPerBar))
		chordEndTick := currentTick + chordDuration

		// Get scale for this chord
		scale := theory.GetScaleForStyle(key, style, chord.Symbol)
		scaleNotes := scale.GetScaleNotes(baseNote-12, baseNote+24) // 3 octave range

		// Get chord tones for emphasis
		chordTones := theory.GetChordTones(chord.Symbol)

		// Generate notes for this chord
		for tick := currentTick; tick < chordEndTick; tick += noteSpacing {
			// Random skip based on density
			if rand.Float64() > config.Density {
				continue
			}

			// Determine if this is a strong beat
			beatInBar := (tick % ticksPerBar) / (ticksPerBar / 4)
			isStrongBeat := beatInBar == 0 || beatInBar == 2

			// Choose next note
			if isStrongBeat && config.UseChordTones && len(chordTones) > 0 {
				// Strong beat: prefer chord tone
				currentNote = chooseChordTone(chordTones, currentNote, scaleNotes, baseNote)
			} else {
				// Weak beat or passing tone: stepwise motion in scale
				currentNote = chooseScaleNote(scaleNotes, currentNote, direction)
			}

			// Keep in playable range
			if currentNote > baseNote+19 { // Getting too high
				direction = -1
				currentNote = chooseScaleNote(scaleNotes, currentNote, direction)
			} else if currentNote < baseNote-5 { // Getting too low
				direction = 1
				currentNote = chooseScaleNote(scaleNotes, currentNote, direction)
			}

			// Occasionally change direction for more musical phrases
			if rand.Float64() < 0.15 {
				direction = -direction
			}

			// Occasionally make a larger leap (3rd or 4th)
			if rand.Float64() < 0.1 {
				leapAmount := 2 + rand.Intn(2) // 2 or 3 scale degrees
				for i := 0; i < leapAmount; i++ {
					currentNote = chooseScaleNote(scaleNotes, currentNote, direction)
				}
			}

			// Add the note
			velocity := uint8(65 + rand.Intn(20)) // Slight velocity variation
			if isStrongBeat {
				velocity += 10 // Accent strong beats
			}

			// Slight duration variation for more natural feel
			dur := noteDuration - uint32(rand.Intn(int(noteDuration/8)+1))
			if dur < noteDuration/2 {
				dur = noteDuration / 2
			}

			notes = append(notes, MelodyNote{
				Note:     uint8(currentNote),
				Tick:     tick,
				Duration: dur,
				Velocity: velocity,
			})
		}

		currentTick = chordEndTick
	}

	return notes
}

// chooseChordTone selects a chord tone near the current note
func chooseChordTone(chordTones []int, currentNote int, scaleNotes []int, baseNote int) int {
	if len(chordTones) == 0 {
		return currentNote
	}

	// Find chord tones in the range near current note
	candidates := []int{}
	for _, ct := range chordTones {
		// Check multiple octaves
		for oct := -1; oct <= 2; oct++ {
			candidate := ct + (baseNote/12)*12 + oct*12
			if candidate >= baseNote-5 && candidate <= baseNote+19 {
				candidates = append(candidates, candidate)
			}
		}
	}

	if len(candidates) == 0 {
		return currentNote
	}

	// Find the closest candidate to current note
	closest := candidates[0]
	closestDist := abs(candidates[0] - currentNote)

	for _, c := range candidates {
		dist := abs(c - currentNote)
		// Prefer notes within a 4th (5 semitones) but allow some variety
		if dist < closestDist || (dist <= 5 && rand.Float64() < 0.3) {
			closest = c
			closestDist = dist
		}
	}

	return closest
}

// chooseScaleNote finds the next scale note in the given direction
func chooseScaleNote(scaleNotes []int, currentNote int, direction int) int {
	if len(scaleNotes) == 0 {
		return currentNote + direction
	}

	// Find current position in scale notes
	currentIdx := -1
	for i, n := range scaleNotes {
		if n == currentNote {
			currentIdx = i
			break
		}
		// If not exact, find closest
		if n > currentNote {
			if direction > 0 {
				return n // Next note up
			}
			if i > 0 {
				return scaleNotes[i-1] // Previous note down
			}
			return n
		}
	}

	// Move in direction
	if currentIdx >= 0 {
		newIdx := currentIdx + direction
		if newIdx >= 0 && newIdx < len(scaleNotes) {
			return scaleNotes[newIdx]
		}
	}

	// Fallback: find closest scale note
	closest := scaleNotes[0]
	for _, n := range scaleNotes {
		if abs(n-currentNote) < abs(closest-currentNote) {
			closest = n
		}
	}

	// Move one step from closest
	for i, n := range scaleNotes {
		if n == closest {
			newIdx := i + direction
			if newIdx >= 0 && newIdx < len(scaleNotes) {
				return scaleNotes[newIdx]
			}
		}
	}

	return currentNote + direction*2 // Fallback: move by whole step
}

// abs returns absolute value of int
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// MelodyStyleFromString converts a string to MelodyStyle
func MelodyStyleFromString(s string) MelodyStyle {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "simple":
		return MelodySimple
	case "moderate", "medium":
		return MelodyModerate
	case "active", "busy":
		return MelodyActive
	case "blues_head", "blueshead", "blues-head":
		return MelodyBluesHead
	case "call_response", "callresponse", "call-response", "aab":
		return MelodyCallResponse
	default:
		return MelodySimple
	}
}

// generateBluesHead creates the classic AAB 12-bar blues vocal melody pattern
// Structure per 12 bars:
//   Bars 1-2: Call phrase (A)
//   Bars 3-4: Response/rest
//   Bars 5-6: Repeat call (A)
//   Bars 7-8: Response/rest
//   Bars 9-10: Resolution phrase (B)
//   Bars 11-12: Turnaround/rest
func generateBluesHead(chords []parser.Chord, key string, style string, config *MelodyConfig, ticksPerBar uint32) []MelodyNote {
	notes := []MelodyNote{}

	// Calculate total bars
	totalBars := 0
	for _, chord := range chords {
		totalBars += int(chord.Bars)
	}

	// Base note in comfortable vocal/guitar range
	baseNote := 52 + (config.Octave-3)*12 // E3 for octave 3

	// Get the blues scale for the key
	scale := theory.GetScaleForStyle(key, style, "")
	scaleNotes := scale.GetScaleNotes(baseNote-5, baseNote+12)

	// Process in 12-bar chunks
	currentTick := uint32(0)
	barIndex := 0

	for _, chord := range chords {
		chordBars := int(chord.Bars)
		chordTones := theory.GetChordTones(chord.Symbol)

		for b := 0; b < chordBars; b++ {
			// Where are we in the 12-bar structure?
			positionIn12 := barIndex % 12

			barStartTick := currentTick + uint32(b)*ticksPerBar

			switch positionIn12 {
			case 0, 1: // Bars 1-2: First call phrase (A)
				phraseNotes := generateCallPhrase(barStartTick, ticksPerBar, scaleNotes, chordTones, baseNote, positionIn12, config.Density)
				notes = append(notes, phraseNotes...)

			case 2, 3: // Bars 3-4: Response (sparse or rest)
				if rand.Float64() < 0.3 { // Sometimes add a response lick
					responseNotes := generateResponsePhrase(barStartTick, ticksPerBar, scaleNotes, baseNote)
					notes = append(notes, responseNotes...)
				}

			case 4, 5: // Bars 5-6: Repeat call phrase (A) - similar to first
				phraseNotes := generateCallPhrase(barStartTick, ticksPerBar, scaleNotes, chordTones, baseNote, positionIn12-4, config.Density)
				notes = append(notes, phraseNotes...)

			case 6, 7: // Bars 7-8: Response (sparse or rest)
				if rand.Float64() < 0.3 {
					responseNotes := generateResponsePhrase(barStartTick, ticksPerBar, scaleNotes, baseNote)
					notes = append(notes, responseNotes...)
				}

			case 8, 9: // Bars 9-10: Resolution phrase (B) - different melody
				resolveNotes := generateResolutionPhrase(barStartTick, ticksPerBar, scaleNotes, chordTones, baseNote, positionIn12-8, config.Density)
				notes = append(notes, resolveNotes...)

			case 10, 11: // Bars 11-12: Turnaround (sparse or characteristic lick)
				if positionIn12 == 10 && rand.Float64() < 0.5 {
					turnaroundNotes := generateTurnaroundPhrase(barStartTick, ticksPerBar, scaleNotes, baseNote)
					notes = append(notes, turnaroundNotes...)
				}
			}

			barIndex++
		}

		currentTick += uint32(chordBars) * ticksPerBar
	}

	return notes
}

// generateCallPhrase creates the "call" melody (sung line A)
// Typical blues vocal phrasing: starts on/near root, moves through scale, ends on chord tone
func generateCallPhrase(startTick, ticksPerBar uint32, scaleNotes []int, chordTones []int, baseNote int, barInPhrase int, density float64) []MelodyNote {
	notes := []MelodyNote{}

	if barInPhrase == 0 {
		// First bar of phrase: main melodic content
		// Start on the 5th or root, move stepwise

		// Pickup/anacrusis feel - start slightly before beat 2
		tick := startTick + ticksPerBar/8

		// Starting note: 5th degree (common blues start)
		startNote := findClosestScaleNote(scaleNotes, baseNote+7)

		// First note - longer, emphasized
		notes = append(notes, MelodyNote{
			Note:     uint8(startNote),
			Tick:     tick,
			Duration: ticksPerBar / 4,
			Velocity: 85,
		})

		// Second note - step down or repeat
		tick += ticksPerBar / 4
		secondNote := chooseScaleNote(scaleNotes, startNote, -1)
		notes = append(notes, MelodyNote{
			Note:     uint8(secondNote),
			Tick:     tick,
			Duration: ticksPerBar / 4,
			Velocity: 75,
		})

		// Third note - continue descending or jump
		if rand.Float64() < density {
			tick += ticksPerBar / 4
			thirdNote := chooseScaleNote(scaleNotes, secondNote, -1)
			notes = append(notes, MelodyNote{
				Note:     uint8(thirdNote),
				Tick:     tick,
				Duration: ticksPerBar / 4,
				Velocity: 70,
			})
		}

	} else {
		// Second bar of phrase: resolution/tail
		// Usually shorter, ends on a held note

		tick := startTick + ticksPerBar/8

		// Find a chord tone to land on
		targetNote := baseNote
		if len(chordTones) > 0 {
			targetNote = findClosestNote(chordTones, baseNote, baseNote-5, baseNote+12)
		}

		// Approach note
		approachNote := findClosestScaleNote(scaleNotes, targetNote+2)
		notes = append(notes, MelodyNote{
			Note:     uint8(approachNote),
			Tick:     tick,
			Duration: ticksPerBar / 8,
			Velocity: 70,
		})

		// Landing note - held longer
		tick += ticksPerBar / 6
		notes = append(notes, MelodyNote{
			Note:     uint8(targetNote),
			Tick:     tick,
			Duration: ticksPerBar / 2,
			Velocity: 80,
		})
	}

	return notes
}

// generateResponsePhrase creates sparse instrumental response
func generateResponsePhrase(startTick, ticksPerBar uint32, scaleNotes []int, baseNote int) []MelodyNote {
	notes := []MelodyNote{}

	// Simple 2-3 note response, often descending
	tick := startTick + ticksPerBar/4

	note1 := findClosestScaleNote(scaleNotes, baseNote+5)
	notes = append(notes, MelodyNote{
		Note:     uint8(note1),
		Tick:     tick,
		Duration: ticksPerBar / 6,
		Velocity: 65,
	})

	if rand.Float64() < 0.6 {
		tick += ticksPerBar / 4
		note2 := chooseScaleNote(scaleNotes, note1, -1)
		notes = append(notes, MelodyNote{
			Note:     uint8(note2),
			Tick:     tick,
			Duration: ticksPerBar / 4,
			Velocity: 60,
		})
	}

	return notes
}

// generateResolutionPhrase creates the "B" line (resolution/answer)
// Different melodic contour than the A phrase
func generateResolutionPhrase(startTick, ticksPerBar uint32, scaleNotes []int, chordTones []int, baseNote int, barInPhrase int, density float64) []MelodyNote {
	notes := []MelodyNote{}

	if barInPhrase == 0 {
		// First bar: Start higher than the A phrase, descend
		tick := startTick + ticksPerBar/8

		// Start on a higher note (octave or 7th)
		startNote := findClosestScaleNote(scaleNotes, baseNote+10)

		notes = append(notes, MelodyNote{
			Note:     uint8(startNote),
			Tick:     tick,
			Duration: ticksPerBar / 4,
			Velocity: 85,
		})

		// Descend more dramatically
		tick += ticksPerBar / 4
		note2 := chooseScaleNote(scaleNotes, startNote, -1)
		note2 = chooseScaleNote(scaleNotes, note2, -1) // Two steps down
		notes = append(notes, MelodyNote{
			Note:     uint8(note2),
			Tick:     tick,
			Duration: ticksPerBar / 4,
			Velocity: 80,
		})

		if rand.Float64() < density {
			tick += ticksPerBar / 4
			note3 := chooseScaleNote(scaleNotes, note2, -1)
			notes = append(notes, MelodyNote{
				Note:     uint8(note3),
				Tick:     tick,
				Duration: ticksPerBar / 4,
				Velocity: 75,
			})
		}

	} else {
		// Second bar: Strong resolution to root
		tick := startTick + ticksPerBar/8

		// Approach from above
		approachNote := findClosestScaleNote(scaleNotes, baseNote+3)
		notes = append(notes, MelodyNote{
			Note:     uint8(approachNote),
			Tick:     tick,
			Duration: ticksPerBar / 6,
			Velocity: 75,
		})

		// Land on root with emphasis
		tick += ticksPerBar / 5
		notes = append(notes, MelodyNote{
			Note:     uint8(baseNote),
			Tick:     tick,
			Duration: ticksPerBar * 2 / 3,
			Velocity: 90,
		})
	}

	return notes
}

// generateTurnaroundPhrase creates the classic turnaround lick
// Descending chromatic or scale-based line
func generateTurnaroundPhrase(startTick, ticksPerBar uint32, scaleNotes []int, baseNote int) []MelodyNote {
	notes := []MelodyNote{}

	// Classic turnaround: descending line
	// Start on 5th or 6th, descend to root
	tick := startTick + ticksPerBar/4
	eighthNote := ticksPerBar / 8

	// Descending pattern: 5-4-b3-2-1 or similar
	startNote := findClosestScaleNote(scaleNotes, baseNote+7) // 5th

	for i := 0; i < 4; i++ {
		notes = append(notes, MelodyNote{
			Note:     uint8(startNote),
			Tick:     tick,
			Duration: eighthNote - 10,
			Velocity: uint8(75 - i*5),
		})
		tick += eighthNote
		startNote = chooseScaleNote(scaleNotes, startNote, -1)
		if startNote < baseNote-2 {
			break
		}
	}

	return notes
}

// findClosestScaleNote finds the scale note closest to target
func findClosestScaleNote(scaleNotes []int, target int) int {
	if len(scaleNotes) == 0 {
		return target
	}

	closest := scaleNotes[0]
	for _, n := range scaleNotes {
		if abs(n-target) < abs(closest-target) {
			closest = n
		}
	}
	return closest
}

// findClosestNote finds the note from candidates closest to target within range
func findClosestNote(candidates []int, target int, minNote, maxNote int) int {
	closest := target
	closestDist := 999

	for _, c := range candidates {
		// Check in multiple octaves
		for oct := -1; oct <= 2; oct++ {
			note := c + 12*oct + (target/12)*12 - 12
			if note >= minNote && note <= maxNote {
				dist := abs(note - target)
				if dist < closestDist {
					closest = note
					closestDist = dist
				}
			}
		}
	}
	return closest
}
