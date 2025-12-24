package theory

import (
	"strings"
)

// ScaleType defines different scale types
type ScaleType string

const (
	ScalePentatonicMinor ScaleType = "pentatonic_minor"
	ScalePentatonicMajor ScaleType = "pentatonic_major"
	ScaleBlues           ScaleType = "blues"
	ScaleNaturalMinor    ScaleType = "natural_minor"
	ScaleNaturalMajor    ScaleType = "natural_major"
	ScaleDorian          ScaleType = "dorian"
	ScaleMixolydian      ScaleType = "mixolydian"
	ScaleHarmonicMinor   ScaleType = "harmonic_minor"
)

// ScaleIntervals maps scale types to their interval patterns (semitones from root)
var ScaleIntervals = map[ScaleType][]int{
	ScalePentatonicMinor: {0, 3, 5, 7, 10},           // R, b3, 4, 5, b7
	ScalePentatonicMajor: {0, 2, 4, 7, 9},            // R, 2, 3, 5, 6
	ScaleBlues:           {0, 3, 5, 6, 7, 10},        // R, b3, 4, b5, 5, b7
	ScaleNaturalMinor:    {0, 2, 3, 5, 7, 8, 10},     // R, 2, b3, 4, 5, b6, b7
	ScaleNaturalMajor:    {0, 2, 4, 5, 7, 9, 11},     // R, 2, 3, 4, 5, 6, 7
	ScaleDorian:          {0, 2, 3, 5, 7, 9, 10},     // R, 2, b3, 4, 5, 6, b7
	ScaleMixolydian:      {0, 2, 4, 5, 7, 9, 10},     // R, 2, 3, 4, 5, 6, b7
	ScaleHarmonicMinor:   {0, 2, 3, 5, 7, 8, 11},     // R, 2, b3, 4, 5, b6, 7
}

// ScaleNames maps scale types to display names
var ScaleNames = map[ScaleType]string{
	ScalePentatonicMinor: "Minor Pentatonic",
	ScalePentatonicMajor: "Major Pentatonic",
	ScaleBlues:           "Blues",
	ScaleNaturalMinor:    "Natural Minor",
	ScaleNaturalMajor:    "Major",
	ScaleDorian:          "Dorian",
	ScaleMixolydian:      "Mixolydian",
	ScaleHarmonicMinor:   "Harmonic Minor",
}

// NoteNames for display (sharps)
var NoteNames = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

// NoteNamesFlat for display (flats)
var NoteNamesFlat = []string{"C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"}

// GuitarTuning is standard tuning MIDI note numbers (low to high: E2, A2, D3, G3, B3, E4)
var GuitarTuning = []int{40, 45, 50, 55, 59, 64}

// GuitarStringNames for display
var GuitarStringNames = []string{"E", "A", "D", "G", "B", "e"}

// Tuning represents a guitar tuning configuration
type Tuning struct {
	Notes []int    // MIDI note numbers for each string (low to high)
	Names []string // String names for display
}

// Tunings maps tuning names to their configurations
var Tunings = map[string]Tuning{
	// Standard and drop tunings
	"standard": {[]int{40, 45, 50, 55, 59, 64}, []string{"E", "A", "D", "G", "B", "e"}},
	"drop_d":   {[]int{38, 45, 50, 55, 59, 64}, []string{"D", "A", "D", "G", "B", "e"}},
	"drop_c":   {[]int{36, 43, 48, 53, 57, 62}, []string{"C", "G", "C", "F", "A", "d"}},
	"d_standard": {[]int{38, 43, 48, 53, 57, 62}, []string{"D", "G", "C", "F", "A", "d"}},  // All strings down 1 whole step
	"eb_standard": {[]int{39, 44, 49, 54, 58, 63}, []string{"Eb", "Ab", "Db", "Gb", "Bb", "eb"}}, // All strings down 1/2 step

	// Open tunings
	"open_e": {[]int{40, 47, 52, 56, 59, 64}, []string{"E", "B", "E", "G#", "B", "e"}},  // Open E major
	"open_d": {[]int{38, 45, 50, 54, 57, 62}, []string{"D", "A", "D", "F#", "A", "d"}},  // Open D major
	"open_g": {[]int{38, 43, 50, 55, 59, 62}, []string{"D", "G", "D", "G", "B", "d"}},   // Open G major (Keith Richards)
	"open_a": {[]int{40, 45, 52, 57, 61, 64}, []string{"E", "A", "E", "A", "C#", "e"}},  // Open A major

	// Modal/Celtic tunings
	"dadgad": {[]int{38, 45, 50, 55, 57, 62}, []string{"D", "A", "D", "G", "A", "d"}},   // DADGAD (Celtic)
	"dadgbd": {[]int{38, 45, 50, 55, 59, 62}, []string{"D", "A", "D", "G", "B", "d"}},   // Double drop D

	// Other tunings
	"open_c": {[]int{36, 43, 48, 55, 60, 64}, []string{"C", "G", "C", "G", "C", "e"}},   // Open C
	"nashville": {[]int{52, 57, 62, 67, 71, 76}, []string{"e", "a", "d", "g", "b", "e"}}, // Nashville (high strung)
}

// TuningNames is an ordered list of tuning names for cycling through
var TuningNames = []string{
	"standard",
	"drop_d",
	"drop_c",
	"d_standard",
	"eb_standard",
	"open_e",
	"open_d",
	"open_g",
	"open_a",
	"dadgad",
	"dadgbd",
	"open_c",
	"nashville",
}

// GetTuning returns a tuning by name, defaulting to standard if not found
func GetTuning(name string) Tuning {
	if name == "" {
		return Tunings["standard"]
	}
	if tuning, ok := Tunings[name]; ok {
		return tuning
	}
	return Tunings["standard"]
}

// GetTuningIndex returns the index of a tuning name in TuningNames, or 0 if not found
func GetTuningIndex(name string) int {
	if name == "" {
		return 0
	}
	for i, n := range TuningNames {
		if n == name {
			return i
		}
	}
	return 0
}

// Scale represents a musical scale with intervals from root
type Scale struct {
	Name      string    // e.g., "A Minor Pentatonic"
	Type      ScaleType // The scale type
	Root      int       // MIDI note offset (0-11, where C=0)
	RootName  string    // Display name of root (e.g., "A", "Bb")
	Intervals []int     // Semitones from root
}

// NewScale creates a new scale with the given root and type
func NewScale(root int, scaleType ScaleType) *Scale {
	root = root % 12 // Normalize to 0-11
	intervals, ok := ScaleIntervals[scaleType]
	if !ok {
		intervals = ScaleIntervals[ScalePentatonicMinor] // Default
		scaleType = ScalePentatonicMinor
	}

	scaleName := ScaleNames[scaleType]
	rootName := NoteNames[root]

	return &Scale{
		Name:      rootName + " " + scaleName,
		Type:      scaleType,
		Root:      root,
		RootName:  rootName,
		Intervals: intervals,
	}
}

// ParseKey parses a key string (e.g., "Am", "Bb", "F#m") and returns root (0-11) and isMinor
func ParseKey(keyStr string) (root int, isMinor bool) {
	keyStr = strings.TrimSpace(keyStr)
	if keyStr == "" {
		return 0, false // Default to C major
	}

	// Check for minor indicator
	isMinor = strings.HasSuffix(strings.ToLower(keyStr), "m") &&
		!strings.HasSuffix(strings.ToLower(keyStr), "maj")

	// Remove minor suffix for parsing
	rootStr := keyStr
	if isMinor {
		rootStr = keyStr[:len(keyStr)-1]
	}

	// Parse root note
	root = NoteToMidi(rootStr)
	return root, isMinor
}

// NoteToMidi converts a note name to MIDI offset (0-11)
func NoteToMidi(note string) int {
	note = strings.TrimSpace(note)
	if note == "" {
		return 0
	}

	// Map of note names to MIDI offsets
	noteMap := map[string]int{
		"C": 0, "C#": 1, "Db": 1,
		"D": 2, "D#": 3, "Eb": 3,
		"E": 4, "Fb": 4, "E#": 5,
		"F": 5, "F#": 6, "Gb": 6,
		"G": 7, "G#": 8, "Ab": 8,
		"A": 9, "A#": 10, "Bb": 10,
		"B": 11, "Cb": 11, "B#": 0,
	}

	// Try exact match first
	if midi, ok := noteMap[note]; ok {
		return midi
	}

	// Try first character + optional accidental
	if len(note) >= 1 {
		base := strings.ToUpper(string(note[0]))
		if len(note) >= 2 {
			accidental := string(note[1])
			if accidental == "#" || accidental == "b" {
				if midi, ok := noteMap[base+accidental]; ok {
					return midi
				}
			}
		}
		if midi, ok := noteMap[base]; ok {
			return midi
		}
	}

	return 0 // Default to C
}

// MidiToNote converts a MIDI offset (0-11) to note name
func MidiToNote(midi int) string {
	return NoteNames[midi%12]
}

// GetScaleForStyle returns the appropriate scale based on track style and key
func GetScaleForStyle(key string, style string, currentChord string) *Scale {
	root, isMinor := ParseKey(key)
	style = strings.ToLower(style)

	// Check for specific style keywords
	switch {
	case strings.Contains(style, "blues"):
		// Blues: Always use blues scale
		return NewScale(root, ScaleBlues)

	case strings.Contains(style, "jazz"):
		// Jazz: Use modes based on chord or key
		if currentChord != "" {
			return getJazzScaleForChord(currentChord, root, isMinor)
		}
		if isMinor {
			return NewScale(root, ScaleDorian)
		}
		return NewScale(root, ScaleMixolydian)

	case strings.Contains(style, "rock"):
		// Rock: Pentatonic
		if isMinor {
			return NewScale(root, ScalePentatonicMinor)
		}
		// For major rock, often use minor pentatonic anyway (blues-rock feel)
		return NewScale(root, ScalePentatonicMinor)

	case strings.Contains(style, "pop"):
		// Pop: Natural scales
		if isMinor {
			return NewScale(root, ScaleNaturalMinor)
		}
		return NewScale(root, ScaleNaturalMajor)

	case strings.Contains(style, "folk"):
		// Folk: Natural scales or pentatonic
		if isMinor {
			return NewScale(root, ScaleNaturalMinor)
		}
		return NewScale(root, ScalePentatonicMajor)

	case strings.Contains(style, "funk") || strings.Contains(style, "soul"):
		// Funk/Soul: Minor pentatonic or Dorian
		return NewScale(root, ScaleDorian)

	case strings.Contains(style, "country"):
		// Country: Major pentatonic
		return NewScale(root, ScalePentatonicMajor)

	default:
		// Default: Pentatonic (works over everything)
		if isMinor {
			return NewScale(root, ScalePentatonicMinor)
		}
		return NewScale(root, ScalePentatonicMinor) // Minor penta works over major too
	}
}

// getJazzScaleForChord returns appropriate jazz scale for a chord
func getJazzScaleForChord(chordSymbol string, keyRoot int, keyIsMinor bool) *Scale {
	chordSymbol = strings.TrimSpace(chordSymbol)
	if chordSymbol == "" {
		if keyIsMinor {
			return NewScale(keyRoot, ScaleDorian)
		}
		return NewScale(keyRoot, ScaleMixolydian)
	}

	// Parse chord root
	chordRoot := parseChordRoot(chordSymbol)
	quality := strings.ToLower(chordSymbol)

	// Determine scale based on chord quality
	switch {
	case strings.Contains(quality, "maj7") || strings.Contains(quality, "maj9"):
		// Major 7th: Lydian or Major
		return NewScale(chordRoot, ScaleNaturalMajor)

	case strings.Contains(quality, "m7") || strings.Contains(quality, "min7"):
		// Minor 7th: Dorian
		return NewScale(chordRoot, ScaleDorian)

	case strings.Contains(quality, "7") && !strings.Contains(quality, "maj"):
		// Dominant 7th: Mixolydian
		return NewScale(chordRoot, ScaleMixolydian)

	case strings.Contains(quality, "dim") || strings.Contains(quality, "o"):
		// Diminished: Use half-whole diminished or harmonic minor
		return NewScale(chordRoot, ScaleHarmonicMinor)

	case strings.Contains(quality, "m") || strings.Contains(quality, "min"):
		// Minor: Dorian
		return NewScale(chordRoot, ScaleDorian)

	default:
		// Major chord: Mixolydian (works well for improv)
		return NewScale(chordRoot, ScaleMixolydian)
	}
}

// parseChordRoot extracts the root note from a chord symbol
func parseChordRoot(chordSymbol string) int {
	if len(chordSymbol) == 0 {
		return 0
	}

	// Get the root (first 1-2 chars)
	rootStr := string(chordSymbol[0])
	if len(chordSymbol) > 1 {
		second := chordSymbol[1]
		if second == '#' || second == 'b' {
			rootStr += string(second)
		}
	}

	return NoteToMidi(rootStr)
}

// ContainsNote checks if a MIDI note is in the scale
func (s *Scale) ContainsNote(midiNote int) bool {
	noteInOctave := midiNote % 12
	relativeToRoot := (noteInOctave - s.Root + 12) % 12

	for _, interval := range s.Intervals {
		if interval == relativeToRoot {
			return true
		}
	}
	return false
}

// IsRoot checks if a MIDI note is the root of the scale
func (s *Scale) IsRoot(midiNote int) bool {
	return midiNote%12 == s.Root
}

// GetFretboardPositions returns a 2D array [string][fret] indicating scale notes
// Returns: positions[stringIndex][fretIndex] = true if note is in scale
// Also returns: roots[stringIndex][fretIndex] = true if note is root
func (s *Scale) GetFretboardPositions(numFrets int) (positions [][]bool, roots [][]bool) {
	return s.GetFretboardPositionsWithTuning(numFrets, Tunings["standard"])
}

// GetFretboardPositionsWithTuning returns fretboard positions using a specific tuning
func (s *Scale) GetFretboardPositionsWithTuning(numFrets int, tuning Tuning) (positions [][]bool, roots [][]bool) {
	numStrings := len(tuning.Notes)
	positions = make([][]bool, numStrings)
	roots = make([][]bool, numStrings)

	for stringIdx := 0; stringIdx < numStrings; stringIdx++ {
		positions[stringIdx] = make([]bool, numFrets+1)
		roots[stringIdx] = make([]bool, numFrets+1)
		openNote := tuning.Notes[stringIdx]

		for fret := 0; fret <= numFrets; fret++ {
			midiNote := openNote + fret
			positions[stringIdx][fret] = s.ContainsNote(midiNote)
			roots[stringIdx][fret] = s.IsRoot(midiNote)
		}
	}

	return positions, roots
}

// GetScaleNotes returns all MIDI notes in the scale within a range
func (s *Scale) GetScaleNotes(lowNote, highNote int) []int {
	notes := []int{}
	for midi := lowNote; midi <= highNote; midi++ {
		if s.ContainsNote(midi) {
			notes = append(notes, midi)
		}
	}
	return notes
}

// GetChordTones returns the chord tones (R, 3, 5, 7) for a chord symbol
func GetChordTones(chordSymbol string) []int {
	root := parseChordRoot(chordSymbol)
	quality := strings.ToLower(chordSymbol)

	// Base triad intervals
	var intervals []int

	switch {
	case strings.Contains(quality, "dim"):
		intervals = []int{0, 3, 6} // R, b3, b5
	case strings.Contains(quality, "aug"):
		intervals = []int{0, 4, 8} // R, 3, #5
	case strings.Contains(quality, "m") || strings.Contains(quality, "min"):
		intervals = []int{0, 3, 7} // R, b3, 5
	default:
		intervals = []int{0, 4, 7} // R, 3, 5 (major)
	}

	// Add 7th if present
	if strings.Contains(quality, "maj7") {
		intervals = append(intervals, 11) // Major 7th
	} else if strings.Contains(quality, "7") {
		intervals = append(intervals, 10) // Minor 7th (dominant)
	}

	// Convert to absolute MIDI offsets
	tones := make([]int, len(intervals))
	for i, interval := range intervals {
		tones[i] = (root + interval) % 12
	}

	return tones
}

// ChordVoicing represents a chord fingering on guitar
type ChordVoicing struct {
	Frets    [6]int // -1 = muted, 0 = open, 1+ = fret number
	BaseFret int    // Starting fret for display
}

// GenerateChordVoicing creates a chord voicing for any tuning
func GenerateChordVoicing(chordSymbol string, tuning Tuning) ChordVoicing {
	chordTones := GetChordTones(chordSymbol)
	if len(chordTones) == 0 {
		return ChordVoicing{Frets: [6]int{-1, -1, -1, -1, -1, -1}}
	}

	root := chordTones[0]
	numStrings := len(tuning.Notes)
	if numStrings > 6 {
		numStrings = 6
	}

	// Find all possible fret positions for each string
	// stringFrets[string][chordToneIndex] = fret position (-1 if not available in range)
	type fretOption struct {
		fret      int
		toneIndex int // 0=root, 1=3rd, 2=5th, 3=7th
		isRoot    bool
	}

	stringOptions := make([][]fretOption, numStrings)
	for str := 0; str < numStrings; str++ {
		openNote := tuning.Notes[str] % 12
		for fret := 0; fret <= 12; fret++ {
			noteAtFret := (openNote + fret) % 12
			for toneIdx, tone := range chordTones {
				if noteAtFret == tone {
					stringOptions[str] = append(stringOptions[str], fretOption{
						fret:      fret,
						toneIndex: toneIdx,
						isRoot:    tone == root,
					})
				}
			}
		}
	}

	// Try to find the best voicing
	// Strategy: Find root on bass string, then fill in other notes within 4 frets
	bestVoicing := ChordVoicing{Frets: [6]int{-1, -1, -1, -1, -1, -1}}
	bestScore := -1

	// Try each possible root position on lower strings (0, 1, 2)
	for bassStr := 0; bassStr < 3 && bassStr < numStrings; bassStr++ {
		for _, rootOpt := range stringOptions[bassStr] {
			if !rootOpt.isRoot {
				continue
			}

			voicing := [6]int{-1, -1, -1, -1, -1, -1}
			voicing[bassStr] = rootOpt.fret
			baseFret := rootOpt.fret
			if baseFret == 0 {
				baseFret = 1
			}

			// For each higher string, find a chord tone within 4 frets of base
			usedTones := make(map[int]bool)
			usedTones[0] = true // Root is used

			for str := bassStr + 1; str < numStrings; str++ {
				bestOpt := fretOption{fret: -1}
				bestOptScore := -1

				for _, opt := range stringOptions[str] {
					// Check if within playable range
					fretDiff := opt.fret - baseFret
					if opt.fret > 0 && (fretDiff < -1 || fretDiff > 4) {
						continue
					}

					// Score this option
					score := 0
					if opt.fret == 0 {
						score += 3 // Prefer open strings
					}
					if !usedTones[opt.toneIndex] {
						score += 2 // Prefer adding new chord tones
					}
					if opt.toneIndex == 0 {
						score += 1 // Roots are good
					}

					if score > bestOptScore {
						bestOptScore = score
						bestOpt = opt
					}
				}

				if bestOpt.fret >= 0 {
					voicing[str] = bestOpt.fret
					usedTones[bestOpt.toneIndex] = true
				}
			}

			// Mute strings below the bass note
			for str := 0; str < bassStr; str++ {
				voicing[str] = -1
			}

			// Score this voicing
			score := 0
			tonesUsed := len(usedTones)
			stringsUsed := 0
			openStrings := 0
			for _, f := range voicing {
				if f >= 0 {
					stringsUsed++
				}
				if f == 0 {
					openStrings++
				}
			}

			score = tonesUsed*10 + stringsUsed*5 + openStrings*3
			if baseFret <= 3 {
				score += 5 // Prefer lower positions
			}

			if score > bestScore {
				bestScore = score
				bestVoicing = ChordVoicing{Frets: voicing, BaseFret: baseFret}
			}
		}
	}

	// If no root-based voicing found, try any voicing
	if bestScore < 0 {
		// Fallback: just find any playable combination
		for baseFret := 0; baseFret <= 5; baseFret++ {
			voicing := [6]int{-1, -1, -1, -1, -1, -1}
			found := false
			for str := 0; str < numStrings; str++ {
				for _, opt := range stringOptions[str] {
					if opt.fret >= baseFret && opt.fret <= baseFret+4 {
						voicing[str] = opt.fret
						found = true
						break
					}
				}
			}
			if found {
				bestVoicing = ChordVoicing{Frets: voicing, BaseFret: baseFret}
				break
			}
		}
	}

	return bestVoicing
}

// GenerateMultipleVoicings creates several voicing options for a chord
func GenerateMultipleVoicings(chordSymbol string, tuning Tuning, maxVoicings int) []ChordVoicing {
	chordTones := GetChordTones(chordSymbol)
	if len(chordTones) == 0 {
		return nil
	}

	root := chordTones[0]
	numStrings := len(tuning.Notes)
	if numStrings > 6 {
		numStrings = 6
	}

	// Find all fret positions for each string
	type fretOption struct {
		fret      int
		toneIndex int
		isRoot    bool
	}

	stringOptions := make([][]fretOption, numStrings)
	for str := 0; str < numStrings; str++ {
		openNote := tuning.Notes[str] % 12
		for fret := 0; fret <= 14; fret++ {
			noteAtFret := (openNote + fret) % 12
			for toneIdx, tone := range chordTones {
				if noteAtFret == tone {
					stringOptions[str] = append(stringOptions[str], fretOption{
						fret:      fret,
						toneIndex: toneIdx,
						isRoot:    tone == root,
					})
				}
			}
		}
	}

	var voicings []ChordVoicing
	seen := make(map[string]bool)

	// Generate voicings starting from different bass positions
	for bassStr := 0; bassStr < 3 && bassStr < numStrings; bassStr++ {
		for _, rootOpt := range stringOptions[bassStr] {
			if !rootOpt.isRoot {
				continue
			}
			if len(voicings) >= maxVoicings {
				break
			}

			voicing := [6]int{-1, -1, -1, -1, -1, -1}
			voicing[bassStr] = rootOpt.fret
			baseFret := rootOpt.fret
			if baseFret == 0 {
				baseFret = 1
			}

			// Fill higher strings
			for str := bassStr + 1; str < numStrings; str++ {
				for _, opt := range stringOptions[str] {
					fretDiff := opt.fret - baseFret
					if opt.fret == 0 || (fretDiff >= -1 && fretDiff <= 4) {
						voicing[str] = opt.fret
						break
					}
				}
			}

			// Mute lower strings
			for str := 0; str < bassStr; str++ {
				voicing[str] = -1
			}

			// Create key for deduplication
			key := ""
			for _, f := range voicing {
				key += string(rune('0' + f + 2)) // Offset to avoid negative
			}

			if !seen[key] {
				seen[key] = true
				voicings = append(voicings, ChordVoicing{Frets: voicing, BaseFret: baseFret})
			}
		}
	}

	return voicings
}

// ScaleTypeFromString converts a string to ScaleType
func ScaleTypeFromString(s string) ScaleType {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "pentatonic_minor", "minor_pentatonic", "pentatonic minor":
		return ScalePentatonicMinor
	case "pentatonic_major", "major_pentatonic", "pentatonic major":
		return ScalePentatonicMajor
	case "blues":
		return ScaleBlues
	case "natural_minor", "minor", "aeolian":
		return ScaleNaturalMinor
	case "natural_major", "major", "ionian":
		return ScaleNaturalMajor
	case "dorian":
		return ScaleDorian
	case "mixolydian":
		return ScaleMixolydian
	case "harmonic_minor":
		return ScaleHarmonicMinor
	default:
		return ScalePentatonicMinor
	}
}
