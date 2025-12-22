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
	MelodySimple   MelodyStyle = "simple"   // Mostly half/whole notes, chord tones
	MelodyModerate MelodyStyle = "moderate" // Quarter notes, some passing tones
	MelodyActive   MelodyStyle = "active"   // Eighth notes, more motion
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
	default:
		return MelodySimple
	}
}
