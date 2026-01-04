package display

import (
	"fmt"
	"strings"

	"backing-tracks/parser"
)

// toggleEditMode enters or exits lyrics edit mode
func (m *TUIModel) toggleEditMode() {
	if m.editMode {
		// Already in edit mode - exit without saving
		m.discardEdits()
		return
	}

	// Build the editLyrics from track's lyrics data
	m.editLyrics = m.collectAllLyrics()

	// Copy form for editing
	m.editForm = make([]string, len(m.track.Form))
	copy(m.editForm, m.track.Form)

	// Pause playback when entering edit mode
	if m.player != nil && !m.player.IsPaused() {
		m.player.TogglePause()
	}

	m.editMode = true
	m.showLyrics = true    // Always enable lyrics display in edit mode for adding/editing
	m.editFocus = "lyrics" // Start with lyrics focus
	m.editBar = m.currentBar
	m.editBeat = 0
	m.editFormIndex = 0
	m.editSectionIndex = 0
	m.editChordIndex = -1 // -1 means "add new chord" mode
	m.editChordInput = ""
	m.editDirty = false
}

// collectAllLyrics gathers all beat lyrics from all sections or track-level lyrics
func (m *TUIModel) collectAllLyrics() []parser.BeatLyric {
	var result []parser.BeatLyric

	// Get beats per bar from time signature
	beatsPerBar := m.getBeatsPerBar()

	// First try section-level beat-mapped lyrics
	sectionInfos := m.track.GetSectionInfos()
	for _, info := range sectionInfos {
		for i := range m.track.Sections {
			if m.track.Sections[i].Name == info.Name && m.track.Sections[i].Lyrics != "" {
				lyrics := parser.ParseBeatLyrics(m.track.Sections[i].Lyrics, info.StartBar, beatsPerBar)
				// Only include entries that have actual lyrics text
				for _, bl := range lyrics {
					if bl.Lyrics != "" {
						result = append(result, bl)
					}
				}
				break
			}
		}
	}

	// If no section lyrics, try track-level per-bar lyrics
	if len(result) == 0 && len(m.track.Lyrics) > 0 {
		for bar, lyric := range m.track.Lyrics {
			if lyric != "" {
				result = append(result, parser.BeatLyric{
					Bar:    bar,
					Beat:   0, // Per-bar lyrics start at beat 0
					Lyrics: lyric,
				})
			}
		}
	}

	return result
}

// moveWordEarlier moves the word at selected bar/beat one beat earlier
func (m *TUIModel) moveWordEarlier() {
	idx := m.getEditLyricIndex(m.editBar, m.editBeat)
	if idx < 0 {
		return
	}

	beatsPerBar := m.getBeatsPerBar()
	bl := &m.editLyrics[idx]

	// Move one beat earlier
	bl.Beat--
	if bl.Beat < 0 {
		bl.Beat = beatsPerBar - 1
		bl.Bar--
		if bl.Bar < 0 {
			// Can't go before bar 0
			bl.Bar = 0
			bl.Beat = 0
		}
	}

	m.editDirty = true
}

// moveWordLater moves the word at selected bar/beat one beat later
func (m *TUIModel) moveWordLater() {
	idx := m.getEditLyricIndex(m.editBar, m.editBeat)
	if idx < 0 {
		return
	}

	beatsPerBar := m.getBeatsPerBar()
	bl := &m.editLyrics[idx]

	// Move one beat later
	bl.Beat++
	if bl.Beat >= beatsPerBar {
		bl.Beat = 0
		bl.Bar++
	}

	m.editDirty = true
}

// moveWordsEarlierFromHere moves all words from current bar onwards one beat earlier
func (m *TUIModel) moveWordsEarlierFromHere() {
	if len(m.editLyrics) == 0 {
		return
	}

	beatsPerBar := m.getBeatsPerBar()

	// Move all words from current bar onwards
	for i := range m.editLyrics {
		bl := &m.editLyrics[i]
		if bl.Bar < m.editBar || (bl.Bar == m.editBar && bl.Beat < m.editBeat) {
			continue // Skip words before current position
		}
		bl.Beat--
		if bl.Beat < 0 {
			bl.Beat = beatsPerBar - 1
			bl.Bar--
			if bl.Bar < 0 {
				bl.Bar = 0
				bl.Beat = 0
			}
		}
	}

	m.editDirty = true
}

// moveWordsLaterFromHere moves all words from current bar onwards one beat later
func (m *TUIModel) moveWordsLaterFromHere() {
	if len(m.editLyrics) == 0 {
		return
	}

	beatsPerBar := m.getBeatsPerBar()

	// Move all words from current bar onwards
	for i := range m.editLyrics {
		bl := &m.editLyrics[i]
		if bl.Bar < m.editBar || (bl.Bar == m.editBar && bl.Beat < m.editBeat) {
			continue // Skip words before current position
		}
		bl.Beat++
		if bl.Beat >= beatsPerBar {
			bl.Beat = 0
			bl.Bar++
		}
	}

	m.editDirty = true
}

// saveEdits saves the edited lyrics back to the BTML file
func (m *TUIModel) saveEdits() {
	if !m.editDirty {
		m.editMode = false
		return
	}

	// Get beats per bar
	beatsPerBar := m.getBeatsPerBar()

	// Check if we have section-level lyrics
	hasSectionLyrics := false
	for _, section := range m.track.Sections {
		if section.Lyrics != "" {
			hasSectionLyrics = true
			break
		}
	}

	if hasSectionLyrics || len(m.track.Sections) > 0 {
		// Save to section-level beat-mapped lyrics
		sectionInfos := m.track.GetSectionInfos()

		for i := range m.track.Sections {
			section := &m.track.Sections[i]

			// Find the section info
			var sectionStart, sectionEnd int
			for _, info := range sectionInfos {
				if info.Name == section.Name {
					sectionStart = info.StartBar
					sectionEnd = info.EndBar
					break
				}
			}

			// Collect edited lyrics for this section
			var sectionLyrics []parser.BeatLyric
			for _, bl := range m.editLyrics {
				if bl.Bar >= sectionStart && bl.Bar < sectionEnd {
					// Adjust bar to be relative to section start for serialization
					adjusted := bl
					adjusted.Bar -= sectionStart
					sectionLyrics = append(sectionLyrics, adjusted)
				}
			}

			// Serialize back to beat notation format (even for sections that didn't have lyrics before)
			if len(sectionLyrics) > 0 {
				section.Lyrics = parser.SerializeBeatLyrics(sectionLyrics, beatsPerBar)
			} else if section.Lyrics != "" {
				// Clear lyrics if all were removed
				section.Lyrics = ""
			}
		}
	} else {
		// Save to track-level per-bar lyrics
		// Find the max bar number to size the array
		maxBar := 0
		for _, bl := range m.editLyrics {
			if bl.Bar > maxBar {
				maxBar = bl.Bar
			}
		}

		// Create new lyrics array
		newLyrics := make([]string, maxBar+1)
		for _, bl := range m.editLyrics {
			// For per-bar format, we put the lyric at its bar position
			// If multiple words land on same bar, concatenate them
			if newLyrics[bl.Bar] != "" {
				newLyrics[bl.Bar] += " " + bl.Lyrics
			} else {
				newLyrics[bl.Bar] = bl.Lyrics
			}
		}
		m.track.Lyrics = newLyrics
	}

	// Save edited form back to track
	m.track.Form = m.editForm

	// Re-expand sections after form change (updates Progression)
	if len(m.track.Sections) > 0 && len(m.track.Form) > 0 {
		m.track.Progression.Pattern = "" // Clear to force re-expansion
		// Force re-expansion by calling LoadTrack logic
		m.expandTrackSections()
	}

	// Save the track to file
	if m.editFilename != "" {
		if err := parser.SaveTrack(m.track, m.editFilename); err != nil {
			// TODO: Show error to user
		}
	}

	// Update the player's lyrics so the display reflects the changes
	if m.player != nil {
		m.player.UpdateLyrics(m.editLyrics)
	}

	m.editMode = false
	m.editDirty = false
}

// discardEdits exits edit mode without saving
func (m *TUIModel) discardEdits() {
	m.editMode = false
	m.editLyrics = nil
	m.editDirty = false
}

// getEditLyricAt returns the edited lyrics at a specific bar/beat, or nil
func (m *TUIModel) getEditLyricAt(bar, beat int) *parser.BeatLyric {
	for i := range m.editLyrics {
		if m.editLyrics[i].Bar == bar && m.editLyrics[i].Beat == beat {
			return &m.editLyrics[i]
		}
	}
	return nil
}

// getEditLyricIndex returns the index of lyrics at a specific bar/beat, or -1 if not found
func (m *TUIModel) getEditLyricIndex(bar, beat int) int {
	for i := range m.editLyrics {
		if m.editLyrics[i].Bar == bar && m.editLyrics[i].Beat == beat {
			return i
		}
	}
	return -1
}

// isEditSelected returns true if the given bar/beat is currently selected
func (m *TUIModel) isEditSelected(bar, beat int) bool {
	if !m.editMode {
		return false
	}
	return m.editBar == bar && m.editBeat == beat
}

// isEditBarSelected returns true if the given bar is currently selected
func (m *TUIModel) isEditBarSelected(bar int) bool {
	if !m.editMode {
		return false
	}
	return m.editBar == bar
}

// cycleEditFocus cycles through the focus areas: lyrics -> form -> sections -> track -> lyrics
func (m *TUIModel) cycleEditFocus() {
	switch m.editFocus {
	case "lyrics":
		m.editFocus = "form"
	case "form":
		m.editFocus = "sections"
	case "sections":
		m.editFocus = "track"
		m.editTrackField = 0
		m.editTrackInput = ""
	case "track":
		m.editFocus = "lyrics"
	default:
		m.editFocus = "lyrics"
	}
}

// cycleFormSectionPicker cycles through available sections when pressing +
// Each press of + changes the last added section to the next available section
func (m *TUIModel) cycleFormSectionPicker() {
	if len(m.track.Sections) == 0 {
		return
	}

	// If we just added a section, cycle it to the next one
	// Otherwise, add a new section
	if len(m.editForm) > 0 && m.editFormIndex < len(m.editForm) {
		currentSection := m.editForm[m.editFormIndex]
		// Find current section index
		currentIdx := -1
		for i, s := range m.track.Sections {
			if s.Name == currentSection {
				currentIdx = i
				break
			}
		}
		// Cycle to next section
		nextIdx := (currentIdx + 1) % len(m.track.Sections)
		m.editForm[m.editFormIndex] = m.track.Sections[nextIdx].Name
		m.editDirty = true
	} else {
		// No form entries, add the first section
		m.addSectionToForm(m.track.Sections[0].Name)
	}
}

// createNewSection creates a new empty section with a unique name
func (m *TUIModel) createNewSection() {
	// Generate a unique section name
	baseName := "new"
	name := baseName
	counter := 1
	for {
		exists := false
		for _, s := range m.track.Sections {
			if s.Name == name {
				exists = true
				break
			}
		}
		if !exists {
			break
		}
		counter++
		name = fmt.Sprintf("%s%d", baseName, counter)
	}

	// Create new section with empty chord progression
	newSection := parser.Section{
		Name: name,
		Progression: parser.ChordProgression{
			Pattern:      "",
			BarsPerChord: 1,
		},
	}

	// Add to track sections
	m.track.Sections = append(m.track.Sections, newSection)

	// Switch to sections focus and select the new section
	m.editFocus = "sections"
	m.editSectionIndex = len(m.track.Sections) - 1
	m.editChordIndex = -1
	m.editChordInput = ""
	m.editDirty = true
}

// addSectionToForm adds a section by name to the form after the current position
func (m *TUIModel) addSectionToForm(sectionName string) {
	insertIdx := m.editFormIndex + 1
	if len(m.editForm) == 0 {
		insertIdx = 0
	}
	newForm := make([]string, 0, len(m.editForm)+1)
	newForm = append(newForm, m.editForm[:insertIdx]...)
	newForm = append(newForm, sectionName)
	newForm = append(newForm, m.editForm[insertIdx:]...)
	m.editForm = newForm
	m.editFormIndex = insertIdx
	m.editDirty = true
}

// addChordToSection adds a chord to the current section's progression
func (m *TUIModel) addChordToSection(chord string) {
	if m.editSectionIndex < 0 || m.editSectionIndex >= len(m.track.Sections) {
		return
	}
	section := &m.track.Sections[m.editSectionIndex]
	currentPattern := string(section.Progression.Pattern)
	if currentPattern == "" {
		section.Progression.Pattern = parser.StringOrList(chord)
	} else {
		section.Progression.Pattern = parser.StringOrList(currentPattern + " " + chord)
	}
	m.editDirty = true
}

// removeLastChordFromSection removes the last chord from the current section's progression
func (m *TUIModel) removeLastChordFromSection() {
	if m.editSectionIndex < 0 || m.editSectionIndex >= len(m.track.Sections) {
		return
	}
	section := &m.track.Sections[m.editSectionIndex]
	currentPattern := string(section.Progression.Pattern)
	parts := strings.Fields(currentPattern)
	if len(parts) > 0 {
		parts = parts[:len(parts)-1]
		section.Progression.Pattern = parser.StringOrList(strings.Join(parts, " "))
		m.editDirty = true
	}
}

// removeChordFromSection removes a chord at a specific index from the current section's progression
func (m *TUIModel) removeChordFromSection(index int) {
	if m.editSectionIndex < 0 || m.editSectionIndex >= len(m.track.Sections) {
		return
	}
	section := &m.track.Sections[m.editSectionIndex]
	currentPattern := string(section.Progression.Pattern)
	parts := strings.Fields(currentPattern)
	if index < 0 || index >= len(parts) {
		return
	}
	// Remove the chord at index
	newParts := make([]string, 0, len(parts)-1)
	newParts = append(newParts, parts[:index]...)
	newParts = append(newParts, parts[index+1:]...)
	section.Progression.Pattern = parser.StringOrList(strings.Join(newParts, " "))
	m.editDirty = true
}

// replaceChordInSection replaces a chord at a specific index in the current section's progression
func (m *TUIModel) replaceChordInSection(index int, newChord string) {
	if m.editSectionIndex < 0 || m.editSectionIndex >= len(m.track.Sections) {
		return
	}
	section := &m.track.Sections[m.editSectionIndex]
	currentPattern := string(section.Progression.Pattern)
	parts := strings.Fields(currentPattern)
	if index < 0 || index >= len(parts) {
		return
	}
	parts[index] = newChord
	section.Progression.Pattern = parser.StringOrList(strings.Join(parts, " "))
	m.editDirty = true
}

// expandTrackSections rebuilds the Progression from Sections and Form
func (m *TUIModel) expandTrackSections() {
	// Build a map of section name -> section
	sectionMap := make(map[string]*parser.Section)
	for i := range m.track.Sections {
		section := &m.track.Sections[i]
		sectionMap[section.Name] = section
		// Set defaults for each section
		if section.Progression.BarsPerChord == 0 {
			section.Progression.BarsPerChord = 1
		}
	}

	// Expand form into a single pattern
	var allChords []string
	for _, sectionName := range m.track.Form {
		section, ok := sectionMap[sectionName]
		if !ok {
			continue // Skip unknown sections
		}
		// Add section marker so GetChords() can track section boundaries
		allChords = append(allChords, "["+sectionName+"]")
		// Get chords from this section (without repeat applied)
		chords := section.Progression.GetChords()
		for _, chord := range chords {
			// Reconstruct the chord notation with duration
			if chord.Bars == 1.0 {
				allChords = append(allChords, chord.Symbol)
			} else {
				allChords = append(allChords, fmt.Sprintf("%s*%g", chord.Symbol, chord.Bars))
			}
		}
	}

	// Set the expanded progression
	m.track.Progression.Pattern = parser.StringOrList(strings.Join(allChords, " "))
	m.track.Progression.BarsPerChord = 1 // Already handled in the pattern
	m.track.Progression.Repeat = 1       // Form already specifies the structure

	// Rebuild bars from new progression
	m.bars = processChordsIntoBars(m.track)
	m.chords = m.track.Progression.GetChords()
}
