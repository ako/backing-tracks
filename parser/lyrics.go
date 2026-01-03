package parser

import (
	"regexp"
	"strings"
)

// BeatLyric represents lyrics at a specific bar/beat position
type BeatLyric struct {
	Bar    int    // Bar number (0-indexed)
	Beat   int    // Beat within bar (0-3 for 4/4)
	Chord  string // Chord symbol if chord changes here, empty if continuation
	Lyrics string // Lyrics text at this beat
}

// LyricsData holds all parsed lyrics mapped to bar/beat positions
type LyricsData struct {
	SectionName  string
	BeatLyrics   []BeatLyric // All lyrics with bar/beat positions
	StartBar     int
	EndBar       int
	BeatsPerBar  int
}

// ChordMark represents a chord position within a lyric line (legacy support)
type ChordMark struct {
	Position int    // Character position in the lyric line
	Chord    string // Chord symbol
}

// LyricLine represents a single line of lyrics with chord positions (legacy support)
type LyricLine struct {
	Text       string      // The lyric text
	ChordMarks []ChordMark // Chord positions within the line
}

// LyricsBlock represents parsed lyrics for a section (legacy support)
type LyricsBlock struct {
	SectionName string
	Lines       []LyricLine
	StartBar    int
	EndBar      int
}

// chordPattern matches common chord symbols
var chordPattern = regexp.MustCompile(`^[A-G][#b]?(m|maj|min|dim|aug|sus|add|M)?[0-9]*(\/[A-G][#b]?)?$`)

// BeatToken represents a token from a beat line
type BeatToken struct {
	StartPos int    // Character position where token starts
	EndPos   int    // Character position where token ends
	Token    string // The token itself (chord or "/")
	IsChord  bool   // True if this is a chord, false if "/"
}

// isBeatLine checks if a line is a beat/chord line (contains chords and "/" markers)
func isBeatLine(line string) bool {
	tokens := strings.Fields(line)
	if len(tokens) == 0 {
		return false
	}

	hasChord := false
	hasSlash := false
	validTokens := 0

	for _, token := range tokens {
		if token == "/" {
			hasSlash = true
			validTokens++
		} else if isChordSymbol(token) {
			hasChord = true
			validTokens++
		}
	}

	// Must have at least one chord and one slash, or multiple slashes
	// and most tokens must be valid beat markers
	return hasChord && (hasSlash || len(tokens) >= 2) && float64(validTokens)/float64(len(tokens)) >= 0.8
}

// parseBeatLine extracts beat tokens with their character positions
func parseBeatLine(line string) []BeatToken {
	var tokens []BeatToken

	i := 0
	for i < len(line) {
		// Skip whitespace
		if line[i] == ' ' || line[i] == '\t' {
			i++
			continue
		}

		// Find the end of this token
		start := i
		for i < len(line) && line[i] != ' ' && line[i] != '\t' {
			i++
		}

		tokenStr := line[start:i]
		isChord := isChordSymbol(tokenStr)
		isSlash := tokenStr == "/"

		if isChord || isSlash {
			tokens = append(tokens, BeatToken{
				StartPos: start,
				EndPos:   i,
				Token:    tokenStr,
				IsChord:  isChord,
			})
		}
	}

	return tokens
}

// extractLyricsForBeat extracts the lyrics text that falls under a beat token's position
func extractLyricsForBeat(lyricsLine string, currentToken BeatToken, nextToken *BeatToken) string {
	if lyricsLine == "" {
		return ""
	}

	startPos := currentToken.StartPos
	var endPos int

	if nextToken != nil {
		endPos = nextToken.StartPos
	} else {
		endPos = len(lyricsLine)
	}

	// Bounds checking
	if startPos >= len(lyricsLine) {
		return ""
	}
	if endPos > len(lyricsLine) {
		endPos = len(lyricsLine)
	}
	if startPos >= endPos {
		return ""
	}

	lyrics := lyricsLine[startPos:endPos]
	return strings.TrimSpace(lyrics)
}

// ParseBeatLyrics parses the new Beatles-style format into beat-mapped lyrics
func ParseBeatLyrics(raw string, startBar int, beatsPerBar int) []BeatLyric {
	if raw == "" {
		return nil
	}

	if beatsPerBar <= 0 {
		beatsPerBar = 4 // Default to 4/4
	}

	lines := strings.Split(raw, "\n")
	var result []BeatLyric

	currentBeat := 0 // Total beat counter from start of section

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		if isBeatLine(line) {
			beatTokens := parseBeatLine(line)
			if len(beatTokens) == 0 {
				continue
			}

			// Look for lyrics line below
			lyricsLine := ""
			if i+1 < len(lines) {
				nextLine := lines[i+1]
				if strings.TrimSpace(nextLine) != "" && !isBeatLine(nextLine) {
					lyricsLine = nextLine
					i++ // Skip the lyrics line
				}
			}

			// Process each beat token
			for j, token := range beatTokens {
				bar := startBar + (currentBeat / beatsPerBar)
				beat := currentBeat % beatsPerBar

				// Get lyrics for this beat
				var nextToken *BeatToken
				if j+1 < len(beatTokens) {
					nextToken = &beatTokens[j+1]
				}
				lyrics := extractLyricsForBeat(lyricsLine, token, nextToken)

				// Determine chord (empty string for "/" continuation)
				chord := ""
				if token.IsChord {
					chord = token.Token
				}

				result = append(result, BeatLyric{
					Bar:    bar,
					Beat:   beat,
					Chord:  chord,
					Lyrics: lyrics,
				})

				currentBeat++
			}
		}
	}

	return result
}

// GetLyricsAtPosition returns lyrics for a specific bar and beat
func (ld *LyricsData) GetLyricsAtPosition(bar, beat int) *BeatLyric {
	for i := range ld.BeatLyrics {
		if ld.BeatLyrics[i].Bar == bar && ld.BeatLyrics[i].Beat == beat {
			return &ld.BeatLyrics[i]
		}
	}
	return nil
}

// GetLyricsInRange returns all lyrics between two bar positions
func (ld *LyricsData) GetLyricsInRange(startBar, endBar int) []BeatLyric {
	var result []BeatLyric
	for _, bl := range ld.BeatLyrics {
		if bl.Bar >= startBar && bl.Bar < endBar {
			result = append(result, bl)
		}
	}
	return result
}

// BuildLyricsData creates LyricsData from track sections
func BuildLyricsData(sections []Section, sectionInfos []SectionInfo, beatsPerBar int) []LyricsData {
	var result []LyricsData

	// Create a map of section name to section for quick lookup
	sectionMap := make(map[string]*Section)
	for i := range sections {
		sectionMap[sections[i].Name] = &sections[i]
	}

	// Process each section info (which has bar positions)
	for _, info := range sectionInfos {
		section, ok := sectionMap[info.Name]
		if !ok || section.Lyrics == "" {
			continue
		}

		beatLyrics := ParseBeatLyrics(section.Lyrics, info.StartBar, beatsPerBar)
		if len(beatLyrics) == 0 {
			continue
		}

		result = append(result, LyricsData{
			SectionName: info.Name,
			BeatLyrics:  beatLyrics,
			StartBar:    info.StartBar,
			EndBar:      info.EndBar,
			BeatsPerBar: beatsPerBar,
		})
	}

	return result
}

// GetBeatLyricsAt finds lyrics at a specific bar/beat across all sections
func GetBeatLyricsAt(lyricsData []LyricsData, bar, beat int) *BeatLyric {
	for i := range lyricsData {
		if bl := lyricsData[i].GetLyricsAtPosition(bar, beat); bl != nil {
			return bl
		}
	}
	return nil
}

// Legacy functions for backward compatibility

// ParseLyrics parses chord-over-lyrics format into structured LyricLines (legacy)
func ParseLyrics(raw string) []LyricLine {
	if raw == "" {
		return nil
	}

	lines := strings.Split(raw, "\n")
	var result []LyricLine

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Check for new beat-based format first
		if isBeatLine(line) {
			// Convert beat format to legacy format for compatibility
			beatTokens := parseBeatLine(line)
			lyricsLine := ""
			if i+1 < len(lines) {
				nextLine := lines[i+1]
				if strings.TrimSpace(nextLine) != "" && !isBeatLine(nextLine) {
					lyricsLine = nextLine
					i++
				}
			}

			// Create chord marks from beat tokens
			var chordMarks []ChordMark
			for _, token := range beatTokens {
				if token.IsChord {
					chordMarks = append(chordMarks, ChordMark{
						Position: token.StartPos,
						Chord:    token.Token,
					})
				}
			}

			if len(chordMarks) > 0 || lyricsLine != "" {
				result = append(result, LyricLine{
					Text:       lyricsLine,
					ChordMarks: chordMarks,
				})
			}
		} else if isChordLine(line) {
			// Old format: chord line followed by lyrics
			chordPositions := extractChordPositions(line)

			lyricText := ""
			if i+1 < len(lines) {
				nextLine := lines[i+1]
				if strings.TrimSpace(nextLine) != "" && !isChordLine(nextLine) {
					lyricText = nextLine
					i++
				}
			}

			if len(chordPositions) > 0 || lyricText != "" {
				result = append(result, LyricLine{
					Text:       lyricText,
					ChordMarks: chordPositions,
				})
			}
		} else {
			// Lyric line without chords above it
			result = append(result, LyricLine{
				Text:       line,
				ChordMarks: nil,
			})
		}
	}

	return result
}

// isChordLine determines if a line contains only chord symbols (legacy format)
func isChordLine(line string) bool {
	tokens := strings.Fields(line)
	if len(tokens) == 0 {
		return false
	}

	// If it looks like a beat line, it's not an old-style chord line
	if isBeatLine(line) {
		return false
	}

	chordCount := 0
	for _, token := range tokens {
		if isChordSymbol(token) {
			chordCount++
		}
	}

	return float64(chordCount)/float64(len(tokens)) >= 0.5
}

// isChordSymbol checks if a string looks like a chord symbol
func isChordSymbol(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	if len(s) == 0 || !strings.ContainsAny(string(s[0]), "ABCDEFG") {
		return false
	}

	return chordPattern.MatchString(s)
}

// extractChordPositions finds chord symbols and their character positions in a line
func extractChordPositions(line string) []ChordMark {
	var marks []ChordMark

	i := 0
	for i < len(line) {
		if line[i] == ' ' || line[i] == '\t' {
			i++
			continue
		}

		start := i
		for i < len(line) && line[i] != ' ' && line[i] != '\t' {
			i++
		}

		token := line[start:i]
		if isChordSymbol(token) {
			marks = append(marks, ChordMark{
				Position: start,
				Chord:    token,
			})
		}
	}

	return marks
}

// GetLineAtBar returns the lyric line for a given bar position within this block (legacy)
func (lb *LyricsBlock) GetLineAtBar(bar int) *LyricLine {
	if bar < lb.StartBar || bar >= lb.EndBar {
		return nil
	}

	if len(lb.Lines) == 0 {
		return nil
	}

	sectionBars := lb.EndBar - lb.StartBar
	if sectionBars == 0 {
		return &lb.Lines[0]
	}

	relativeBar := bar - lb.StartBar
	lineIndex := (relativeBar * len(lb.Lines)) / sectionBars
	if lineIndex >= len(lb.Lines) {
		lineIndex = len(lb.Lines) - 1
	}

	return &lb.Lines[lineIndex]
}

// BuildLyricsBlocks creates LyricsBlocks from track sections (legacy)
func BuildLyricsBlocks(sections []Section, sectionInfos []SectionInfo) []LyricsBlock {
	var blocks []LyricsBlock

	sectionMap := make(map[string]*Section)
	for i := range sections {
		sectionMap[sections[i].Name] = &sections[i]
	}

	for _, info := range sectionInfos {
		section, ok := sectionMap[info.Name]
		if !ok || section.Lyrics == "" {
			continue
		}

		lines := ParseLyrics(section.Lyrics)
		if len(lines) == 0 {
			continue
		}

		blocks = append(blocks, LyricsBlock{
			SectionName: info.Name,
			Lines:       lines,
			StartBar:    info.StartBar,
			EndBar:      info.EndBar,
		})
	}

	return blocks
}

// GetLyricsAtBar finds the lyric line at a specific bar position (legacy)
func GetLyricsAtBar(blocks []LyricsBlock, bar int) *LyricLine {
	for i := range blocks {
		if line := blocks[i].GetLineAtBar(bar); line != nil {
			return line
		}
	}
	return nil
}

// SerializeBeatLyrics converts a slice of BeatLyric back to beat notation format
// This is used when saving edited lyrics back to the BTML file
func SerializeBeatLyrics(lyrics []BeatLyric, beatsPerBar int) string {
	if len(lyrics) == 0 {
		return ""
	}
	if beatsPerBar <= 0 {
		beatsPerBar = 4
	}

	// Find the bar range
	minBar := lyrics[0].Bar
	maxBar := lyrics[0].Bar
	for _, bl := range lyrics {
		if bl.Bar < minBar {
			minBar = bl.Bar
		}
		if bl.Bar > maxBar {
			maxBar = bl.Bar
		}
	}

	// Create a map for quick lookup: bar*beatsPerBar + beat -> BeatLyric
	lyricsMap := make(map[int]*BeatLyric)
	for i := range lyrics {
		key := lyrics[i].Bar*beatsPerBar + lyrics[i].Beat
		lyricsMap[key] = &lyrics[i]
	}

	// We also need to track chords that were in the original
	// For simplicity, we'll use "/" for continuation beats
	// The chord that was playing at each beat position

	var result strings.Builder
	barsPerLine := 2 // Output 2 bars per line (common format)
	beatWidth := 5   // Width for each beat column

	for startBar := minBar; startBar <= maxBar; startBar += barsPerLine {
		endBar := startBar + barsPerLine
		if endBar > maxBar+1 {
			endBar = maxBar + 1
		}

		// Build chord line and lyrics line
		var chordLine strings.Builder
		var lyricsLine strings.Builder

		for bar := startBar; bar < endBar; bar++ {
			for beat := 0; beat < beatsPerBar; beat++ {
				key := bar*beatsPerBar + beat
				bl := lyricsMap[key]

				// Chord/slash for this beat
				chordToken := "/"
				lyricText := ""

				if bl != nil {
					if bl.Chord != "" {
						chordToken = bl.Chord
					}
					lyricText = bl.Lyrics
				}

				// Pad to beat width
				chordPadded := chordToken
				for len(chordPadded) < beatWidth {
					chordPadded += " "
				}
				chordLine.WriteString(chordPadded)

				lyricPadded := lyricText
				for len(lyricPadded) < beatWidth {
					lyricPadded += " "
				}
				lyricsLine.WriteString(lyricPadded)
			}
		}

		// Add the lines (trim trailing whitespace)
		chordStr := strings.TrimRight(chordLine.String(), " ")
		lyricsStr := strings.TrimRight(lyricsLine.String(), " ")

		result.WriteString(chordStr)
		result.WriteString("\n")
		if lyricsStr != "" {
			result.WriteString(lyricsStr)
			result.WriteString("\n")
		}
	}

	return strings.TrimRight(result.String(), "\n")
}
