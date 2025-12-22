package display

import (
	"fmt"
	"strings"

	"backing-tracks/parser"
)

// ShowTrack displays the track information in the terminal
func ShowTrack(track *parser.Track) {
	// Header box
	title := track.Info.Title
	info := fmt.Sprintf("Key: %s | Tempo: %d BPM | %s | %s",
		track.Info.Key,
		track.Info.Tempo,
		track.Info.TimeSignature,
		track.Info.Style,
	)

	maxLen := len(title)
	if len(info) > maxLen {
		maxLen = len(info)
	}

	// Print header
	fmt.Printf("┌─ %s %s┐\n", title, strings.Repeat("─", maxLen-len(title)+1))
	fmt.Printf("│ %s%s │\n", info, strings.Repeat(" ", maxLen-len(info)))
	fmt.Printf("└%s┘\n\n", strings.Repeat("─", maxLen+2))

	// Chord progression
	chords := track.Progression.GetChords()
	totalBars := track.Progression.TotalBars()

	repeatInfo := ""
	if track.Progression.Repeat > 1 {
		repeatInfo = fmt.Sprintf(", %dx", track.Progression.Repeat)
	}

	fmt.Printf("Chord Progression (%d bars%s):\n", totalBars, repeatInfo)

	// Display chords in a grid (4 per line for 12-bar blues)
	chordsPerLine := 4
	for i := 0; i < len(chords); i += chordsPerLine {
		end := i + chordsPerLine
		if end > len(chords) {
			end = len(chords)
		}

		line := make([]string, 0, chordsPerLine)
		for j := i; j < end; j++ {
			line = append(line, chords[j].Symbol)
		}

		fmt.Printf("  %s\n", strings.Join(line, " | "))
	}

	// Rhythm info
	if track.Rhythm != nil {
		var rhythmInfo string
		if track.Rhythm.Pattern != "" {
			rhythmInfo = fmt.Sprintf("♬ Rhythm: pattern \"%s\"", track.Rhythm.Pattern)
		} else {
			rhythmInfo = fmt.Sprintf("♬ Rhythm: %s", track.Rhythm.Style)
		}
		if track.Rhythm.Swing > 0 && track.Rhythm.Swing != 0.5 {
			swingPercent := int(track.Rhythm.Swing * 100)
			rhythmInfo += fmt.Sprintf(" (swing %d%%)", swingPercent)
		}
		if track.Rhythm.Accent != "" {
			rhythmInfo += fmt.Sprintf(" [accent: %s]", track.Rhythm.Accent)
		}
		fmt.Println(rhythmInfo)
	}

	// Bass info
	if track.Bass != nil {
		bassInfo := fmt.Sprintf("♪ Bass: %s", track.Bass.Style)
		if track.Bass.Swing > 0 && track.Bass.Swing != 0.5 {
			swingPercent := int(track.Bass.Swing * 100)
			bassInfo += fmt.Sprintf(" (swing %d%%)", swingPercent)
		}
		fmt.Println(bassInfo)
	}

	// Drums info
	if track.Drums != nil {
		drumsInfo := "♫ Drums: "
		if track.Drums.Style != "" {
			drumsInfo += track.Drums.Style
		} else {
			// Show custom pattern info
			parts := []string{}
			if track.Drums.Kick != nil {
				parts = append(parts, "kick")
			}
			if track.Drums.Snare != nil {
				parts = append(parts, "snare")
			}
			if track.Drums.Hihat != nil {
				parts = append(parts, "hihat")
			}
			if track.Drums.Ride != nil {
				parts = append(parts, "ride")
			}
			drumsInfo += strings.Join(parts, "+")
		}

		if track.Drums.Intensity > 0 {
			intensityPercent := int(track.Drums.Intensity * 100)
			drumsInfo += fmt.Sprintf(" (%d%% intensity)", intensityPercent)
		}

		fmt.Println(drumsInfo)
	}

	if track.Bass != nil || track.Drums != nil {
		fmt.Println()
	}
}
