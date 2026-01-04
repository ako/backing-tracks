package display

import (
	"fmt"

	"backing-tracks/parser"
	"backing-tracks/theory"

	tea "github.com/charmbracelet/bubbletea"
)

// handleEditModeKeys handles keyboard input during edit mode
func (m *TUIModel) handleEditModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	beatsPerBar := m.getBeatsPerBar()
	key := msg.String()

	// Global edit mode keys (work regardless of focus)
	switch key {
	case "esc":
		// Discard changes and exit edit mode
		m.discardEdits()
		return m, nil
	case "ctrl+s":
		// Save changes and exit edit mode
		m.saveEdits()
		return m, nil
	case "ctrl+f":
		// Cycle focus: lyrics -> form -> sections -> lyrics
		m.cycleEditFocus()
		return m, nil
	case "ctrl+n":
		// Create a new section
		m.createNewSection()
		return m, nil
	}

	// Handle keys based on current focus
	switch m.editFocus {
	case "lyrics":
		return m.handleLyricsEditKeys(key, beatsPerBar)
	case "form":
		return m.handleFormEditKeys(key)
	case "sections":
		return m.handleSectionsEditKeys(key, msg)
	case "track":
		return m.handleTrackEditKeys(key, msg)
	}

	// Ignore other keys in edit mode
	return m, nil
}

// handleLyricsEditKeys handles keyboard input when focus is on lyrics
func (m *TUIModel) handleLyricsEditKeys(key string, beatsPerBar int) (tea.Model, tea.Cmd) {
	switch key {
	case "left":
		// Move to previous beat (wrap to previous bar)
		m.editBeat--
		if m.editBeat < 0 {
			if m.editBar > 0 {
				m.editBar--
				m.editBeat = beatsPerBar - 1
			} else {
				m.editBeat = 0
			}
		}
		return m, nil
	case "right":
		// Move to next beat (wrap to next bar)
		m.editBeat++
		if m.editBeat >= beatsPerBar {
			if m.editBar < len(m.bars)-1 {
				m.editBar++
				m.editBeat = 0
			} else {
				m.editBeat = beatsPerBar - 1
			}
		}
		return m, nil
	case "up":
		// Move to previous line (2 bars up)
		if m.editBar >= 2 {
			m.editBar -= 2
		} else {
			m.editBar = 0
		}
		return m, nil
	case "down":
		// Move to next line (2 bars down)
		if m.editBar < len(m.bars)-2 {
			m.editBar += 2
		} else {
			m.editBar = len(m.bars) - 1
		}
		return m, nil
	case "shift+left":
		// Move word at current position one beat earlier
		m.moveWordEarlier()
		return m, nil
	case "shift+right":
		// Move word at current position one beat later
		m.moveWordLater()
		return m, nil
	case "shift+tab":
		// Move all words from current bar one beat earlier
		m.moveWordsEarlierFromHere()
		return m, nil
	case "tab":
		// Move all words from current bar one beat later
		m.moveWordsLaterFromHere()
		return m, nil
	case "backspace":
		// Delete last character of word at selected bar/beat
		if idx := m.getEditLyricIndex(m.editBar, m.editBeat); idx >= 0 {
			word := m.editLyrics[idx].Lyrics
			if len(word) > 0 {
				m.editLyrics[idx].Lyrics = word[:len(word)-1]
				m.editDirty = true
			}
		}
		return m, nil
	case "delete":
		// Delete entire word at selected bar/beat
		if idx := m.getEditLyricIndex(m.editBar, m.editBeat); idx >= 0 {
			m.editLyrics[idx].Lyrics = ""
			m.editDirty = true
		}
		return m, nil
	case "enter":
		// Save and exit
		m.saveEdits()
		return m, nil
	default:
		// Handle printable characters - append to word at selected bar/beat
		if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
			idx := m.getEditLyricIndex(m.editBar, m.editBeat)
			if idx >= 0 {
				// Append to existing word
				m.editLyrics[idx].Lyrics += key
				m.editDirty = true
			} else {
				// Create new word at this position
				m.editLyrics = append(m.editLyrics, parser.BeatLyric{
					Bar:    m.editBar,
					Beat:   m.editBeat,
					Lyrics: key,
				})
				m.editDirty = true
			}
			return m, nil
		}
	}
	return m, nil
}

// handleFormEditKeys handles keyboard input when focus is on form
func (m *TUIModel) handleFormEditKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		// Move selection up
		if m.editFormIndex > 0 {
			m.editFormIndex--
		}
		return m, nil
	case "down", "j":
		// Move selection down
		if m.editFormIndex < len(m.editForm)-1 {
			m.editFormIndex++
		}
		return m, nil
	case "K", "shift+up":
		// Move selected entry up in the form
		if m.editFormIndex > 0 && len(m.editForm) > 1 {
			m.editForm[m.editFormIndex], m.editForm[m.editFormIndex-1] = m.editForm[m.editFormIndex-1], m.editForm[m.editFormIndex]
			m.editFormIndex--
			m.editDirty = true
		}
		return m, nil
	case "J", "shift+down":
		// Move selected entry down in the form
		if m.editFormIndex < len(m.editForm)-1 && len(m.editForm) > 1 {
			m.editForm[m.editFormIndex], m.editForm[m.editFormIndex+1] = m.editForm[m.editFormIndex+1], m.editForm[m.editFormIndex]
			m.editFormIndex++
			m.editDirty = true
		}
		return m, nil
	case "+", "=":
		// Add section to form - cycle through available sections
		m.cycleFormSectionPicker()
		return m, nil
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Add specific section by number (1 = first section, etc.)
		sectionIdx := int(key[0] - '1') // Convert '1'-'9' to 0-8
		if sectionIdx < len(m.track.Sections) {
			m.addSectionToForm(m.track.Sections[sectionIdx].Name)
		}
		return m, nil
	case "-", "d", "backspace", "delete":
		// Remove selected entry from form
		if len(m.editForm) > 0 && m.editFormIndex < len(m.editForm) {
			m.editForm = append(m.editForm[:m.editFormIndex], m.editForm[m.editFormIndex+1:]...)
			if m.editFormIndex >= len(m.editForm) && m.editFormIndex > 0 {
				m.editFormIndex--
			}
			m.editDirty = true
		}
		return m, nil
	case "enter":
		// Save and exit
		m.saveEdits()
		return m, nil
	default:
		// Number keys 1-9 to quickly add a section by index
		if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
			sectionIdx := int(key[0] - '1')
			if sectionIdx < len(m.track.Sections) {
				sectionName := m.track.Sections[sectionIdx].Name
				// Insert after current position
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
			return m, nil
		}
	}
	return m, nil
}

// handleSectionsEditKeys handles keyboard input when focus is on sections
func (m *TUIModel) handleSectionsEditKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle rename mode first
	if m.editSectionRename {
		switch key {
		case "enter":
			// Confirm rename
			if m.editSectionName != "" && m.editSectionIndex >= 0 && m.editSectionIndex < len(m.track.Sections) {
				oldName := m.track.Sections[m.editSectionIndex].Name
				newName := m.editSectionName
				m.track.Sections[m.editSectionIndex].Name = newName
				// Update form references to this section
				for i, formEntry := range m.editForm {
					if formEntry == oldName {
						m.editForm[i] = newName
					}
				}
				m.editDirty = true
			}
			m.editSectionRename = false
			m.editSectionName = ""
			return m, nil
		case "esc":
			// Cancel rename
			m.editSectionRename = false
			m.editSectionName = ""
			return m, nil
		case "backspace":
			// Remove last character
			if len(m.editSectionName) > 0 {
				m.editSectionName = m.editSectionName[:len(m.editSectionName)-1]
			}
			return m, nil
		default:
			// Add typed characters to name
			if len(msg.Runes) > 0 {
				for _, r := range msg.Runes {
					if r >= 32 && r <= 126 {
						m.editSectionName += string(r)
					}
				}
			}
			return m, nil
		}
	}

	// Get chord count for current section
	chordCount := 0
	if m.editSectionIndex >= 0 && m.editSectionIndex < len(m.track.Sections) {
		chordCount = len(m.track.Sections[m.editSectionIndex].Progression.GetChords())
	}

	// Check for rename trigger (r key when not typing a chord)
	if key == "r" && m.editChordInput == "" && m.editChordIndex == -1 {
		if m.editSectionIndex >= 0 && m.editSectionIndex < len(m.track.Sections) {
			m.editSectionRename = true
			m.editSectionName = "" // Start with empty name
			return m, nil
		}
	}

	// First, check if this has printable runes (handles all typing including *, #, ., 0-9, etc.)
	// Check runes regardless of key type since some terminals report differently
	if len(msg.Runes) > 0 {
		// Don't capture space here - it's handled specially below
		if !(len(msg.Runes) == 1 && msg.Runes[0] == ' ') {
			for _, r := range msg.Runes {
				if r >= 32 && r <= 126 {
					m.editChordInput += string(r)
				}
			}
			return m, nil
		}
	}

	switch key {
	case "up":
		// Move selection up to previous section
		if m.editSectionIndex > 0 {
			m.editSectionIndex--
			m.editChordIndex = -1 // Reset to "add new" mode
			m.editChordInput = ""
		}
		return m, nil
	case "down":
		// Move selection down to next section
		if m.editSectionIndex < len(m.track.Sections)-1 {
			m.editSectionIndex++
			m.editChordIndex = -1 // Reset to "add new" mode
			m.editChordInput = ""
		}
		return m, nil
	case "left":
		// Move to previous chord (or wrap to "add new" mode)
		m.editChordInput = "" // Clear any pending input
		if m.editChordIndex == -1 {
			// Wrap from "add new" to last chord
			if chordCount > 0 {
				m.editChordIndex = chordCount - 1
			}
		} else if m.editChordIndex > 0 {
			m.editChordIndex--
		}
		return m, nil
	case "right":
		// Move to next chord (or to "add new" mode)
		m.editChordInput = "" // Clear any pending input
		if m.editChordIndex == -1 {
			// From "add new" mode, move to first chord
			if chordCount > 0 {
				m.editChordIndex = 0
			}
		} else if m.editChordIndex < chordCount-1 {
			m.editChordIndex++
		} else {
			m.editChordIndex = -1 // Move to "add new" mode
		}
		return m, nil
	case "backspace":
		if len(m.editChordInput) > 0 {
			// Remove last character from chord input
			m.editChordInput = m.editChordInput[:len(m.editChordInput)-1]
		} else if m.editChordIndex >= 0 {
			// Delete selected chord
			m.removeChordFromSection(m.editChordIndex)
			// Adjust selection after deletion
			if m.editChordIndex >= chordCount-1 {
				m.editChordIndex = chordCount - 2
			}
			if m.editChordIndex < 0 {
				m.editChordIndex = -1
			}
		}
		return m, nil
	case "delete":
		// Delete selected chord or last chord
		if m.editChordIndex >= 0 {
			m.removeChordFromSection(m.editChordIndex)
			if m.editChordIndex >= chordCount-1 {
				m.editChordIndex = chordCount - 2
			}
			if m.editChordIndex < 0 {
				m.editChordIndex = -1
			}
		} else if chordCount > 0 {
			m.removeLastChordFromSection()
		}
		return m, nil
	case " ":
		// Space commits the chord
		if m.editChordInput != "" {
			if m.editChordIndex >= 0 {
				// Replace selected chord
				m.replaceChordInSection(m.editChordIndex, m.editChordInput)
				m.editChordIndex++ // Move to next chord
				newChordCount := len(m.track.Sections[m.editSectionIndex].Progression.GetChords())
				if m.editChordIndex >= newChordCount {
					m.editChordIndex = -1 // Move to "add new" mode
				}
			} else {
				// Add new chord
				m.addChordToSection(m.editChordInput)
			}
			m.editChordInput = ""
		}
		return m, nil
	case "enter":
		// Commit any pending chord, then save and exit
		if m.editChordInput != "" {
			if m.editChordIndex >= 0 {
				m.replaceChordInSection(m.editChordIndex, m.editChordInput)
			} else {
				m.addChordToSection(m.editChordInput)
			}
			m.editChordInput = ""
		}
		m.saveEdits()
		return m, nil
	case ".", "*", "#", "/", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Fallback for chord notation characters if msg.Runes was empty
		m.editChordInput += key
		return m, nil
	}
	return m, nil
}

// Track field constants
const (
	trackFieldTitle = iota
	trackFieldKey
	trackFieldTempo
	trackFieldTime
	trackFieldStyle
	trackFieldCapo
	trackFieldTuning
	trackFieldRhythm
	trackFieldBass
	trackFieldDrums
	trackFieldCount // Total number of fields
)

// Available options for cycling through with left/right keys
var trackStyleOptions = []string{"blues", "jazz", "rock", "pop", "folk", "country", "ballad", "fingerstyle", "funk", "reggae", "latin"}
var bassStyleOptions = []string{"root", "root_fifth", "walking", "swing_walking", "stride", "boogie", "funk", "slap", "ska", "reggae", "country", "disco", "motown", "808"}
var drumsStyleOptions = []string{"rock_beat", "shuffle", "blues_shuffle", "jazz_swing", "four_on_floor", "trap", "ska", "reggae", "country", "disco", "motown", "flamenco"}

// handleTrackEditKeys handles keyboard input when focus is on track properties
func (m *TUIModel) handleTrackEditKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key {
	case "up":
		// Move to previous field
		if m.editTrackField > 0 {
			m.commitTrackField() // Commit current field before moving
			m.editTrackField--
			m.editTrackInput = ""
		}
		return m, nil
	case "down":
		// Move to next field
		if m.editTrackField < trackFieldCount-1 {
			m.commitTrackField() // Commit current field before moving
			m.editTrackField++
			m.editTrackInput = ""
		}
		return m, nil
	case "left", "right":
		// Cycle through options for fields that have predefined choices
		return m.cycleTrackFieldOption(key == "right"), nil
	case "backspace":
		// Remove last character from input
		if len(m.editTrackInput) > 0 {
			m.editTrackInput = m.editTrackInput[:len(m.editTrackInput)-1]
		}
		return m, nil
	case "enter":
		// Commit current field and save
		m.commitTrackField()
		m.saveEdits()
		return m, nil
	case "tab":
		// Commit and move to next field
		m.commitTrackField()
		m.editTrackField = (m.editTrackField + 1) % trackFieldCount
		m.editTrackInput = ""
		return m, nil
	default:
		// Handle printable characters
		if len(msg.Runes) > 0 {
			for _, r := range msg.Runes {
				if r >= 32 && r <= 126 {
					m.editTrackInput += string(r)
				}
			}
			return m, nil
		}
	}
	return m, nil
}

// cycleTrackFieldOption cycles through available options for the current field
func (m *TUIModel) cycleTrackFieldOption(forward bool) *TUIModel {
	var options []string
	var currentValue string

	switch m.editTrackField {
	case trackFieldStyle:
		options = trackStyleOptions
		currentValue = m.track.Info.Style
	case trackFieldTuning:
		options = theory.TuningNames
		currentValue = m.track.Info.Tuning
		if currentValue == "" {
			currentValue = "standard"
		}
	case trackFieldRhythm:
		// Rhythm uses same style options as track style
		options = trackStyleOptions
		if m.track.Rhythm != nil {
			currentValue = m.track.Rhythm.Style
		}
	case trackFieldBass:
		options = bassStyleOptions
		if m.track.Bass != nil {
			currentValue = m.track.Bass.Style
		}
	case trackFieldDrums:
		options = drumsStyleOptions
		if m.track.Drums != nil {
			currentValue = m.track.Drums.Style
		}
	default:
		// No cycling for other fields
		return m
	}

	if len(options) == 0 {
		return m
	}

	// Find current index
	currentIdx := 0
	for i, opt := range options {
		if opt == currentValue {
			currentIdx = i
			break
		}
	}

	// Calculate new index
	var newIdx int
	if forward {
		newIdx = (currentIdx + 1) % len(options)
	} else {
		newIdx = (currentIdx - 1 + len(options)) % len(options)
	}

	// Apply the new value
	newValue := options[newIdx]
	switch m.editTrackField {
	case trackFieldStyle:
		m.track.Info.Style = newValue
		m.editDirty = true
	case trackFieldTuning:
		m.track.Info.Tuning = newValue
		m.editDirty = true
	case trackFieldRhythm:
		if m.track.Rhythm == nil {
			m.track.Rhythm = &parser.Rhythm{}
		}
		m.track.Rhythm.Style = newValue
		m.editDirty = true
	case trackFieldBass:
		if m.track.Bass == nil {
			m.track.Bass = &parser.Bass{}
		}
		m.track.Bass.Style = newValue
		m.editDirty = true
	case trackFieldDrums:
		if m.track.Drums == nil {
			m.track.Drums = &parser.Drums{}
		}
		m.track.Drums.Style = newValue
		m.editDirty = true
	}

	return m
}

// commitTrackField commits the current track field input to the track
func (m *TUIModel) commitTrackField() {
	if m.editTrackInput == "" {
		return
	}

	switch m.editTrackField {
	case trackFieldTitle:
		m.track.Info.Title = m.editTrackInput
		m.editDirty = true
	case trackFieldKey:
		m.track.Info.Key = m.editTrackInput
		m.editDirty = true
	case trackFieldTempo:
		var tempo int
		if _, err := fmt.Sscanf(m.editTrackInput, "%d", &tempo); err == nil && tempo > 0 {
			m.track.Info.Tempo = tempo
			m.editDirty = true
		}
	case trackFieldTime:
		m.track.Info.TimeSignature = m.editTrackInput
		m.editDirty = true
	case trackFieldStyle:
		m.track.Info.Style = m.editTrackInput
		m.editDirty = true
	case trackFieldCapo:
		var capo int
		if _, err := fmt.Sscanf(m.editTrackInput, "%d", &capo); err == nil && capo >= 0 && capo <= 12 {
			m.track.Info.Capo = capo
			m.editDirty = true
		}
	case trackFieldTuning:
		m.track.Info.Tuning = m.editTrackInput
		m.editDirty = true
	case trackFieldRhythm:
		if m.track.Rhythm == nil {
			m.track.Rhythm = &parser.Rhythm{}
		}
		m.track.Rhythm.Style = m.editTrackInput
		m.editDirty = true
	case trackFieldBass:
		if m.track.Bass == nil {
			m.track.Bass = &parser.Bass{}
		}
		m.track.Bass.Style = m.editTrackInput
		m.editDirty = true
	case trackFieldDrums:
		if m.track.Drums == nil {
			m.track.Drums = &parser.Drums{}
		}
		m.track.Drums.Style = m.editTrackInput
		m.editDirty = true
	}

	m.editTrackInput = ""
}
