package midi

import (
	"backing-tracks/parser"
)

// DrumNote represents a single drum hit
type DrumNote struct {
	Note     uint8  // MIDI drum note (GM drum map)
	Tick     uint32 // When to play
	Velocity uint8  // Hit velocity
}

// GM Drum Map (General MIDI standard percussion)
const (
	KickDrum      = 36 // Bass Drum 1
	SnareDrum     = 38 // Acoustic Snare
	ClosedHihat   = 42 // Closed Hi-Hat
	OpenHihat     = 46 // Open Hi-Hat
	RideCymbal    = 51 // Ride Cymbal 1
	CrashCymbal   = 49 // Crash Cymbal 1
)

// GenerateDrumPattern creates drum notes for the entire track
func GenerateDrumPattern(totalBars int, drums *parser.Drums, ticksPerBar uint32) []DrumNote {
	if drums == nil {
		return nil
	}

	notes := []DrumNote{}

	// Get intensity (default 0.7)
	intensity := 0.7
	if drums.Intensity > 0 {
		intensity = drums.Intensity
	}
	baseVelocity := uint8(float64(100) * intensity)

	// Use style presets if no explicit patterns
	if drums.Style != "" && drums.Kick == nil && drums.Snare == nil && drums.Hihat == nil {
		return generatePresetPattern(drums.Style, totalBars, ticksPerBar, baseVelocity)
	}

	// Generate from explicit patterns
	for bar := 0; bar < totalBars; bar++ {
		barStartTick := uint32(bar) * ticksPerBar

		// Kick drum
		if drums.Kick != nil {
			notes = append(notes, generateDrumVoice(drums.Kick, KickDrum, barStartTick, ticksPerBar, baseVelocity+10)...)
		}

		// Snare drum
		if drums.Snare != nil {
			notes = append(notes, generateDrumVoice(drums.Snare, SnareDrum, barStartTick, ticksPerBar, baseVelocity)...)
		}

		// Hi-hat
		if drums.Hihat != nil {
			notes = append(notes, generateDrumVoice(drums.Hihat, ClosedHihat, barStartTick, ticksPerBar, baseVelocity-20)...)
		}

		// Ride cymbal
		if drums.Ride != nil {
			notes = append(notes, generateDrumVoice(drums.Ride, RideCymbal, barStartTick, ticksPerBar, baseVelocity-15)...)
		}
	}

	return notes
}

// generateDrumVoice creates notes for a single drum voice
func generateDrumVoice(pattern *parser.DrumPattern, note uint8, startTick, ticksPerBar uint32, velocity uint8) []DrumNote {
	notes := []DrumNote{}

	// Euclidean rhythm
	if pattern.Euclidean != nil {
		rhythm := generateEuclideanRhythm(pattern.Euclidean.Hits, pattern.Euclidean.Steps, pattern.Euclidean.Rotation)
		ticksPerStep := ticksPerBar / uint32(pattern.Euclidean.Steps)

		for i, hit := range rhythm {
			if hit {
				notes = append(notes, DrumNote{
					Note:     note,
					Tick:     startTick + uint32(i)*ticksPerStep,
					Velocity: velocity,
				})
			}
		}
	}

	// Explicit beats
	if pattern.Beats != nil && len(pattern.Beats) > 0 {
		quarterNote := ticksPerBar / 4
		for _, beat := range pattern.Beats {
			notes = append(notes, DrumNote{
				Note:     note,
				Tick:     startTick + uint32(beat-1)*quarterNote,
				Velocity: velocity,
			})
		}
	}

	return notes
}

// generateEuclideanRhythm implements Bjorklund's algorithm for Euclidean rhythms
func generateEuclideanRhythm(hits, steps, rotation int) []bool {
	if hits >= steps {
		// All hits
		result := make([]bool, steps)
		for i := range result {
			result[i] = true
		}
		return result
	}

	if hits == 0 {
		// No hits
		return make([]bool, steps)
	}

	// Bjorklund's algorithm
	pattern := make([][]bool, steps)

	// Initialize with hits and rests
	for i := 0; i < hits; i++ {
		pattern[i] = []bool{true}
	}
	for i := hits; i < steps; i++ {
		pattern[i] = []bool{false}
	}

	// Distribute evenly
	count := steps
	for {
		smaller := min(hits, count-hits)
		if smaller <= 1 {
			break
		}

		// Concatenate pairs
		for i := 0; i < smaller; i++ {
			pattern[i] = append(pattern[i], pattern[count-smaller+i]...)
		}

		count -= smaller
		if hits > count-hits {
			hits = count - hits
		}
	}

	// Flatten to single array
	result := []bool{}
	for i := 0; i < count; i++ {
		result = append(result, pattern[i]...)
	}

	// Apply rotation
	if rotation != 0 {
		rotation = rotation % len(result)
		if rotation < 0 {
			rotation = len(result) + rotation
		}
		result = append(result[rotation:], result[:rotation]...)
	}

	return result
}

// generatePresetPattern creates preset drum patterns
func generatePresetPattern(style string, totalBars int, ticksPerBar uint32, velocity uint8) []DrumNote {
	notes := []DrumNote{}

	for bar := 0; bar < totalBars; bar++ {
		barStartTick := uint32(bar) * ticksPerBar

		switch style {
		case "rock_beat":
			// Standard rock: Kick 1,3 | Snare 2,4 | Hihat 8ths
			notes = append(notes, rockBeat(barStartTick, ticksPerBar, velocity)...)

		case "shuffle":
			// Blues shuffle
			notes = append(notes, shuffleBeat(barStartTick, ticksPerBar, velocity)...)

		case "blues_shuffle":
			// Driving blues shuffle with open hihat accents
			notes = append(notes, bluesShuffle(barStartTick, ticksPerBar, velocity)...)

		case "jazz_swing":
			// Jazz swing ride pattern
			notes = append(notes, jazzSwing(barStartTick, ticksPerBar, velocity)...)

		case "four_on_floor", "edm":
			// EDM/House: kick every beat, snare on 2&4, 16th hihats
			notes = append(notes, fourOnFloor(barStartTick, ticksPerBar, velocity)...)

		case "trap":
			// Trap: rolling hihats, sparse kick, heavy snare
			notes = append(notes, trapBeat(barStartTick, ticksPerBar, velocity)...)

		case "ska":
			// Ska: driving offbeat feel
			notes = append(notes, skaBeat(barStartTick, ticksPerBar, velocity)...)

		case "reggae", "one_drop":
			// Reggae one-drop: kick and snare on beat 3
			notes = append(notes, reggaeBeat(barStartTick, ticksPerBar, velocity)...)

		case "country", "train":
			// Country train beat
			notes = append(notes, countryBeat(barStartTick, ticksPerBar, velocity)...)

		case "disco":
			// Classic disco beat
			notes = append(notes, discoBeat(barStartTick, ticksPerBar, velocity)...)

		case "motown", "soul":
			// Motown/Soul beat
			notes = append(notes, motownBeat(barStartTick, ticksPerBar, velocity)...)

		case "flamenco", "rumba":
			// Flamenco rumba (cajon style)
			notes = append(notes, flamencoBeat(barStartTick, ticksPerBar, velocity)...)

		default:
			// Simple 4/4 beat
			notes = append(notes, rockBeat(barStartTick, ticksPerBar, velocity)...)
		}
	}

	return notes
}

// rockBeat generates a standard rock beat
func rockBeat(startTick, ticksPerBar uint32, velocity uint8) []DrumNote {
	notes := []DrumNote{}
	eighthNote := ticksPerBar / 8

	// Kick: beats 1 and 3
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick, Velocity: velocity + 10})
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick + ticksPerBar/2, Velocity: velocity + 10})

	// Snare: beats 2 and 4
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + ticksPerBar/4, Velocity: velocity})
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + 3*ticksPerBar/4, Velocity: velocity})

	// Hi-hat: eighth notes
	for i := 0; i < 8; i++ {
		vel := velocity - 20
		if i%2 == 1 {
			vel -= 10 // Softer on offbeats
		}
		notes = append(notes, DrumNote{Note: ClosedHihat, Tick: startTick + uint32(i)*eighthNote, Velocity: uint8(vel)})
	}

	return notes
}

// shuffleBeat generates a shuffle/blues beat
func shuffleBeat(startTick, ticksPerBar uint32, velocity uint8) []DrumNote {
	notes := []DrumNote{}

	// Kick: beats 1 and 3
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick, Velocity: velocity + 10})
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick + ticksPerBar/2, Velocity: velocity + 10})

	// Snare: beats 2 and 4
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + ticksPerBar/4, Velocity: velocity})
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + 3*ticksPerBar/4, Velocity: velocity})

	// Shuffle hi-hat (triplet feel)
	// Divide bar into 12 (triplet eighths)
	tripletEighth := ticksPerBar / 12
	shufflePattern := []int{0, 2, 3, 5, 6, 8, 9, 11} // Swung eighths

	for _, pos := range shufflePattern {
		vel := velocity - 20
		if pos%3 == 2 {
			vel -= 10 // Softer on triplet upbeats
		}
		notes = append(notes, DrumNote{Note: ClosedHihat, Tick: startTick + uint32(pos)*tripletEighth, Velocity: uint8(vel)})
	}

	return notes
}

// jazzSwing generates a jazz swing ride pattern
func jazzSwing(startTick, ticksPerBar uint32, velocity uint8) []DrumNote {
	notes := []DrumNote{}

	// Sparse kick (just 1 and sometimes 3)
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick, Velocity: velocity})

	// Snare: beats 2 and 4 (backbeat)
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + ticksPerBar/4, Velocity: velocity - 10})
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + 3*ticksPerBar/4, Velocity: velocity - 10})

	// Ride cymbal: swung pattern (ding ding-a ding)
	tripletEighth := ticksPerBar / 12
	ridePattern := []int{0, 2, 3, 5, 6, 8, 9, 11}

	for _, pos := range ridePattern {
		vel := velocity - 15
		if pos%3 == 2 {
			vel -= 10
		}
		notes = append(notes, DrumNote{Note: RideCymbal, Tick: startTick + uint32(pos)*tripletEighth, Velocity: uint8(vel)})
	}

	return notes
}

// bluesShuffle generates a driving blues shuffle pattern
// Classic 12/8 feel with accented open hihats and ghost notes
func bluesShuffle(startTick, ticksPerBar uint32, velocity uint8) []DrumNote {
	notes := []DrumNote{}
	tripletEighth := ticksPerBar / 12

	// Kick: 1 and 3, with pickup on the "a" of 2 and 4
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick, Velocity: velocity + 10})
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick + 5*tripletEighth, Velocity: velocity - 5})  // "a" of 2
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick + 6*tripletEighth, Velocity: velocity + 10}) // beat 3
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick + 11*tripletEighth, Velocity: velocity - 5}) // "a" of 4

	// Snare: 2 and 4 with ghost notes
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + 3*tripletEighth, Velocity: velocity})      // beat 2
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + 9*tripletEighth, Velocity: velocity})      // beat 4
	// Ghost notes (very soft)
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + 2*tripletEighth, Velocity: velocity - 35}) // before 2
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + 8*tripletEighth, Velocity: velocity - 35}) // before 4

	// Hi-hat: shuffled pattern with open hihat accents on upbeats
	// Pattern: closed-closed-OPEN, closed-closed-OPEN, closed-closed-OPEN, closed-closed-OPEN
	for beat := 0; beat < 4; beat++ {
		beatStart := uint32(beat * 3)
		// Downbeat - closed, accented
		notes = append(notes, DrumNote{
			Note:     ClosedHihat,
			Tick:     startTick + beatStart*tripletEighth,
			Velocity: velocity - 10,
		})
		// Skip beat - closed, softer
		notes = append(notes, DrumNote{
			Note:     ClosedHihat,
			Tick:     startTick + (beatStart+1)*tripletEighth,
			Velocity: velocity - 25,
		})
		// Upbeat - OPEN hihat for shuffle feel
		notes = append(notes, DrumNote{
			Note:     OpenHihat,
			Tick:     startTick + (beatStart+2)*tripletEighth,
			Velocity: velocity - 15,
		})
	}

	return notes
}

// fourOnFloor generates an EDM/House beat with four-on-the-floor kick
func fourOnFloor(startTick, ticksPerBar uint32, velocity uint8) []DrumNote {
	notes := []DrumNote{}
	quarterNote := ticksPerBar / 4
	sixteenthNote := ticksPerBar / 16

	// Kick: every beat (four on the floor)
	for beat := 0; beat < 4; beat++ {
		notes = append(notes, DrumNote{
			Note:     KickDrum,
			Tick:     startTick + uint32(beat)*quarterNote,
			Velocity: velocity + 15,
		})
	}

	// Clap/Snare: beats 2 and 4
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + quarterNote, Velocity: velocity + 5})
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + 3*quarterNote, Velocity: velocity + 5})

	// Hi-hat: 16th notes with accents on offbeats
	for i := 0; i < 16; i++ {
		vel := velocity - 25
		if i%4 == 2 { // Accent on "and" of each beat
			vel = velocity - 10
		}
		if i%2 == 1 { // Offbeat 16ths slightly softer
			vel -= 5
		}
		// Open hihat on certain offbeats for energy
		hihatNote := uint8(ClosedHihat)
		if i == 6 || i == 14 { // Open hihat before beats 2 and 4
			hihatNote = OpenHihat
			vel = velocity - 15
		}
		notes = append(notes, DrumNote{
			Note:     hihatNote,
			Tick:     startTick + uint32(i)*sixteenthNote,
			Velocity: uint8(vel),
		})
	}

	return notes
}

// trapBeat generates a trap-style beat with rolling hihats
func trapBeat(startTick, ticksPerBar uint32, velocity uint8) []DrumNote {
	notes := []DrumNote{}
	quarterNote := ticksPerBar / 4
	sixteenthNote := ticksPerBar / 16
	thirtySecondNote := ticksPerBar / 32

	// Kick: sparse, syncopated (1, and of 2, 4)
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick, Velocity: velocity + 20})
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick + quarterNote + quarterNote/2, Velocity: velocity + 15})
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick + 3*quarterNote, Velocity: velocity + 15})

	// Snare: heavy on 2 and 4
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + quarterNote, Velocity: velocity + 10})
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + 3*quarterNote, Velocity: velocity + 10})

	// Rolling hi-hats: mix of 16ths and 32nds with triplet rolls
	// Basic 16th pattern with rolls
	hihatPattern := []struct {
		offset   uint32
		velocity int
	}{
		{0, 0}, {sixteenthNote, -10}, {2 * sixteenthNote, 0}, {3 * sixteenthNote, -10},
		{4 * sixteenthNote, 0}, {5 * sixteenthNote, -10}, {6 * sixteenthNote, 0}, {7 * sixteenthNote, -10},
		{8 * sixteenthNote, 0}, {9 * sixteenthNote, -10}, {10 * sixteenthNote, 0}, {11 * sixteenthNote, -10},
		// Roll before beat 4
		{12 * sixteenthNote, 0},
		{12*sixteenthNote + thirtySecondNote, -15},
		{13 * sixteenthNote, -5},
		{13*sixteenthNote + thirtySecondNote, -15},
		{14 * sixteenthNote, 0},
		{14*sixteenthNote + thirtySecondNote, -15},
		{15 * sixteenthNote, -5},
		{15*sixteenthNote + thirtySecondNote, -20},
	}

	for _, h := range hihatPattern {
		notes = append(notes, DrumNote{
			Note:     ClosedHihat,
			Tick:     startTick + h.offset,
			Velocity: uint8(int(velocity) - 20 + h.velocity),
		})
	}

	return notes
}

// skaBeat generates a ska beat - driving offbeat feel
func skaBeat(startTick, ticksPerBar uint32, velocity uint8) []DrumNote {
	notes := []DrumNote{}
	eighthNote := ticksPerBar / 8

	// Kick: beats 1 and 3
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick, Velocity: velocity + 10})
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick + ticksPerBar/2, Velocity: velocity + 10})

	// Snare: beats 2 and 4 (backbeat)
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + ticksPerBar/4, Velocity: velocity + 5})
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + 3*ticksPerBar/4, Velocity: velocity + 5})

	// Hi-hat: driving 8th notes with accents on offbeats
	for i := 0; i < 8; i++ {
		vel := velocity - 15
		if i%2 == 1 {
			vel = velocity - 5 // Accent offbeats (the skank)
		}
		notes = append(notes, DrumNote{
			Note:     ClosedHihat,
			Tick:     startTick + uint32(i)*eighthNote,
			Velocity: uint8(vel),
		})
	}

	return notes
}

// reggaeBeat generates a reggae one-drop beat
func reggaeBeat(startTick, ticksPerBar uint32, velocity uint8) []DrumNote {
	notes := []DrumNote{}
	quarterNote := ticksPerBar / 4
	eighthNote := ticksPerBar / 8

	// One-drop: kick AND snare together on beat 3 only
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick + 2*quarterNote, Velocity: velocity + 15})
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + 2*quarterNote, Velocity: velocity + 10})

	// Cross-stick on beat 4 (optional accent)
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + 3*quarterNote, Velocity: velocity - 20})

	// Hi-hat: sparse, offbeat focused
	// Play on the "and" of each beat
	for i := 0; i < 8; i++ {
		if i%2 == 1 { // Only offbeats
			notes = append(notes, DrumNote{
				Note:     ClosedHihat,
				Tick:     startTick + uint32(i)*eighthNote,
				Velocity: velocity - 10,
			})
		}
	}

	return notes
}

// countryBeat generates a country train beat
func countryBeat(startTick, ticksPerBar uint32, velocity uint8) []DrumNote {
	notes := []DrumNote{}
	quarterNote := ticksPerBar / 4
	eighthNote := ticksPerBar / 8

	// Kick: beats 1 and 3
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick, Velocity: velocity + 10})
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick + 2*quarterNote, Velocity: velocity + 10})

	// Snare: beats 2 and 4 (strong backbeat)
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + quarterNote, Velocity: velocity + 5})
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + 3*quarterNote, Velocity: velocity + 5})

	// Hi-hat: steady 8th notes (train rhythm)
	for i := 0; i < 8; i++ {
		vel := velocity - 15
		if i%2 == 0 {
			vel = velocity - 10 // Slightly accent downbeats
		}
		notes = append(notes, DrumNote{
			Note:     ClosedHihat,
			Tick:     startTick + uint32(i)*eighthNote,
			Velocity: uint8(vel),
		})
	}

	return notes
}

// discoBeat generates a classic disco beat
func discoBeat(startTick, ticksPerBar uint32, velocity uint8) []DrumNote {
	notes := []DrumNote{}
	quarterNote := ticksPerBar / 4
	sixteenthNote := ticksPerBar / 16

	// Four-on-the-floor kick
	for beat := 0; beat < 4; beat++ {
		notes = append(notes, DrumNote{
			Note:     KickDrum,
			Tick:     startTick + uint32(beat)*quarterNote,
			Velocity: velocity + 10,
		})
	}

	// Snare: beats 2 and 4
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + quarterNote, Velocity: velocity + 5})
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + 3*quarterNote, Velocity: velocity + 5})

	// Hi-hat: 16th notes with open hihat on offbeats
	for i := 0; i < 16; i++ {
		hihatNote := uint8(ClosedHihat)
		vel := velocity - 20

		if i%4 == 2 { // "and" of each beat - open hihat
			hihatNote = OpenHihat
			vel = velocity - 10
		} else if i%4 == 0 { // On the beat
			vel = velocity - 15
		}

		notes = append(notes, DrumNote{
			Note:     hihatNote,
			Tick:     startTick + uint32(i)*sixteenthNote,
			Velocity: uint8(vel),
		})
	}

	return notes
}

// motownBeat generates a Motown/Soul beat
func motownBeat(startTick, ticksPerBar uint32, velocity uint8) []DrumNote {
	notes := []DrumNote{}
	quarterNote := ticksPerBar / 4
	eighthNote := ticksPerBar / 8

	// Kick: beats 1 and 3 with occasional pickup
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick, Velocity: velocity + 10})
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick + 2*quarterNote, Velocity: velocity + 10})
	// Pickup before beat 3
	notes = append(notes, DrumNote{Note: KickDrum, Tick: startTick + quarterNote + 3*eighthNote/2, Velocity: velocity - 5})

	// Snare: heavy on 2 and 4 (the Motown backbeat)
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + quarterNote, Velocity: velocity + 15})
	notes = append(notes, DrumNote{Note: SnareDrum, Tick: startTick + 3*quarterNote, Velocity: velocity + 15})

	// Tambourine feel on 8th notes
	for i := 0; i < 8; i++ {
		vel := velocity - 20
		if i%2 == 0 {
			vel = velocity - 15
		}
		notes = append(notes, DrumNote{
			Note:     ClosedHihat,
			Tick:     startTick + uint32(i)*eighthNote,
			Velocity: uint8(vel),
		})
	}

	return notes
}

// flamencoBeat generates a flamenco rumba beat (cajon style)
func flamencoBeat(startTick, ticksPerBar uint32, velocity uint8) []DrumNote {
	notes := []DrumNote{}
	sixteenthNote := ticksPerBar / 16

	// Flamenco rumba pattern - syncopated kicks and snares
	// Classic pattern: 1 . . 2 . . 3 . 4 . 5 . 6 . . .

	// Low tones (kick) on 1 and syncopated positions
	kickPositions := []int{0, 6, 10}
	for _, pos := range kickPositions {
		vel := velocity + 10
		if pos == 0 {
			vel = velocity + 15 // Accent the ONE
		}
		notes = append(notes, DrumNote{
			Note:     KickDrum,
			Tick:     startTick + uint32(pos)*sixteenthNote,
			Velocity: uint8(vel),
		})
	}

	// High tones (snare/slap) on offbeats
	slapPositions := []int{3, 8, 12}
	for _, pos := range slapPositions {
		notes = append(notes, DrumNote{
			Note:     SnareDrum,
			Tick:     startTick + uint32(pos)*sixteenthNote,
			Velocity: velocity,
		})
	}

	// Finger rolls/ghost notes
	ghostPositions := []int{2, 5, 9, 14}
	for _, pos := range ghostPositions {
		notes = append(notes, DrumNote{
			Note:     SnareDrum,
			Tick:     startTick + uint32(pos)*sixteenthNote,
			Velocity: velocity - 30,
		})
	}

	return notes
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
