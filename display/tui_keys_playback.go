package display

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// handlePlaybackKeys handles keyboard input during playback mode
func (m *TUIModel) handlePlaybackKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "esc":
		m.quitting = true
		return m, tea.Quit
	case "e":
		// Toggle edit mode (if track has editable content)
		if m.trackCanEdit() {
			m.toggleEditMode()
		}
	case " ":
		// Toggle pause
		if m.player != nil {
			m.player.TogglePause()
		} else {
			if m.paused {
				m.pausedTotal += time.Since(m.pausedAt)
				m.paused = false
			} else {
				m.pausedAt = time.Now()
				m.paused = true
			}
		}
	case "left":
		// Jump to previous bar
		if m.player != nil {
			m.player.SeekRelative(-1)
		} else {
			timePerBar := m.timePerBeat * 4
			if m.currentBar > 0 {
				m.seekOffset -= timePerBar
			}
		}
	case "right":
		// Jump to next bar
		if m.player != nil {
			m.player.SeekRelative(1)
		} else {
			timePerBar := m.timePerBeat * 4
			if m.currentBar < len(m.bars)-1 {
				m.seekOffset += timePerBar
			}
		}
	case "up":
		// Transpose up one semitone
		if m.player != nil {
			m.player.Transpose(1)
			m.transposeOffset = m.player.GetTranspose()
		} else {
			m.transposeOffset++
		}
		m.updateTransposedScale()
	case "down":
		// Transpose down one semitone
		if m.player != nil {
			m.player.Transpose(-1)
			m.transposeOffset = m.player.GetTranspose()
		} else {
			m.transposeOffset--
		}
		m.updateTransposedScale()
	case "1":
		// Toggle drums
		if m.player != nil {
			m.player.ToggleTrackMute(0)
		}
	case "2":
		// Toggle bass
		if m.player != nil {
			m.player.ToggleTrackMute(1)
		}
	case "3":
		// Toggle chords
		if m.player != nil {
			m.player.ToggleTrackMute(2)
		}
	case "4":
		// Toggle melody
		if m.player != nil {
			m.player.ToggleTrackMute(3)
		}
	case "5":
		// Toggle fingerstyle
		if m.player != nil {
			m.player.ToggleTrackMute(4)
		}
	case "[":
		// Move capo down (with audio transpose)
		if m.capoPosition > 0 {
			m.capoPosition--
			if m.player != nil {
				m.player.SetCapo(m.capoPosition)
			}
			m.updateTablatureConfig()
		}
	case "]":
		// Move capo up (with audio transpose)
		if m.capoPosition < 12 {
			m.capoPosition++
			if m.player != nil {
				m.player.SetCapo(m.capoPosition)
			}
			m.updateTablatureConfig()
		}
	case "{":
		// Move capo down (visual only, no audio transpose)
		if m.capoPosition > 0 {
			m.capoPosition--
			m.updateTablatureConfig()
		}
	case "}":
		// Move capo up (visual only, no audio transpose)
		if m.capoPosition < 12 {
			m.capoPosition++
			m.updateTablatureConfig()
		}
	case ",", "<":
		// Previous tuning
		m.cycleTuning(-1)
	case ".", ">":
		// Next tuning
		m.cycleTuning(1)
	case "!":
		// Loop 1 bar (Shift+1)
		if m.player != nil {
			m.player.ToggleLoop(1)
		}
	case "@":
		// Loop 2 bars (Shift+2)
		if m.player != nil {
			m.player.ToggleLoop(2)
		}
	case "#":
		// Loop 3 bars (Shift+3)
		if m.player != nil {
			m.player.ToggleLoop(3)
		}
	case "$":
		// Loop 4 bars (Shift+4)
		if m.player != nil {
			m.player.ToggleLoop(4)
		}
	case "%":
		// Loop 5 bars (Shift+5)
		if m.player != nil {
			m.player.ToggleLoop(5)
		}
	case "^":
		// Loop 6 bars (Shift+6)
		if m.player != nil {
			m.player.ToggleLoop(6)
		}
	case "&":
		// Loop 7 bars (Shift+7)
		if m.player != nil {
			m.player.ToggleLoop(7)
		}
	case "*":
		// Loop 8 bars (Shift+8)
		if m.player != nil {
			m.player.ToggleLoop(8)
		}
	case "(":
		// Loop 9 bars (Shift+9)
		if m.player != nil {
			m.player.ToggleLoop(9)
		}
	case "shift+up":
		// Increase tempo by 5 BPM
		if m.player != nil {
			m.player.AdjustTempo(5)
		}
	case "shift+down":
		// Decrease tempo by 5 BPM
		if m.player != nil {
			m.player.AdjustTempo(-5)
		}
	case ")":
		// Loop current section (Shift+0)
		if m.player != nil {
			m.player.LoopCurrentSection()
		}
	case "l":
		// Toggle lyrics display
		if m.player != nil && m.player.HasLyrics() {
			m.showLyrics = !m.showLyrics
		}
	case "m":
		// Toggle metronome/beat counter
		m.showMetronome = !m.showMetronome
	case "s":
		// Toggle strum pattern
		m.showStrumPattern = !m.showStrumPattern
	case "c":
		// Toggle chord names display
		m.showChordNames = !m.showChordNames
	case "t":
		// Toggle inline tablature display in left column
		m.showTablature = !m.showTablature
	case ";":
		// Previous pattern type
		if m.tablature != nil {
			newPattern := m.tablature.PrevPattern()
			m.tablature.RegenerateTablature(m.track)
			// Also update the audio player
			if m.player != nil {
				m.player.SetFingerstylePattern(newPattern)
			}
		}
	case "'":
		// Next pattern type
		if m.tablature != nil {
			newPattern := m.tablature.NextPattern()
			m.tablature.RegenerateTablature(m.track)
			// Also update the audio player
			if m.player != nil {
				m.player.SetFingerstylePattern(newPattern)
			}
		}
	}

	return m, nil
}
