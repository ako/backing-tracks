# Lyrics Support Proposal

**Version:** 1.0
**Date:** 2024-12-25
**Status:** Proposal

## Overview

Add support for displaying lyrics synchronized with chord changes during playback. This feature allows BTML files to include lyrics in a chord-sheet format that gets displayed in the TUI.

## New BTML Syntax

### Lyrics Field in Sections

```yaml
sections:
  - name: verse
    chord_progression:
      pattern: "Dm G Cmaj7 Em Am"
      bars_per_chord: 1
    lyrics: |
      Dm                         G
      Don't cry, snowman, not in front of me
      Cmaj7                              Em          Am
      Who'll catch your tears if you can't catch me, darling?
```

### Format Rules

1. **Chord lines**: Lines containing only chord symbols (with whitespace for positioning)
2. **Lyric lines**: Lines containing the actual lyrics text
3. **Chord positioning**: Chord symbols are positioned above the syllable where the chord change occurs
4. **Multi-line verses**: Empty lines separate verses/phrases within a section

### Detection Heuristics

A line is considered a **chord line** if:
- It matches the pattern of known chord symbols separated by whitespace
- Contains primarily chord-like tokens (e.g., `C`, `Am`, `Dm7`, `G#m`, `Cmaj7`, `F/A`)
- Has no lowercase words longer than 2 characters (to distinguish from lyrics)

## Data Structures

### Parser Changes

```go
// Section gains a Lyrics field
type Section struct {
    Name        string           `yaml:"name"`
    Progression ChordProgression `yaml:"chord_progression"`
    Lyrics      string           `yaml:"lyrics,omitempty"`
}

// Parsed lyrics structure for display
type LyricLine struct {
    Text        string       // The lyric text
    ChordMarks  []ChordMark  // Chord positions within the line
}

type ChordMark struct {
    Position int    // Character position in the lyric line
    Chord    string // Chord symbol
}

// LyricsBlock represents parsed lyrics for a section
type LyricsBlock struct {
    SectionName string
    Lines       []LyricLine
    StartBar    int
    EndBar      int
}
```

### Playback Data Changes

```go
type PlaybackData struct {
    // ... existing fields ...
    Lyrics []LyricsBlock // Lyrics for each section
}

// Get lyrics for current position
func (p *PlaybackData) GetLyricsAtBar(bar int) *LyricLine
```

## Display Integration

### TUI Layout with Lyrics

```
┌─ Snowman ─────────────────────────────────────────────────┐
│ Key: C | Tempo: 70 BPM | 4/4 | pop_ballad | [verse]       │
└───────────────────────────────────────────────────────────┘

                         Dm
           Don't cry, snowman, not in front of me

          ◉    ○    ○    ○           Bar 3/42
          1    2    3    4

  ▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░░░░░░░░

┌─ Chord Grid ──────────────────────────────────────────────┐
│  [Dm] │  G  │ Cmaj7 │  Em  │  Am  │  Dm  │  G  │ Cmaj7   │
└───────────────────────────────────────────────────────────┘
```

### Lyrics Display Features

1. **Current line highlight**: Show the lyric line for the current bar
2. **Chord emphasis**: Highlight the chord symbol above the current position
3. **Lookahead**: Optionally show next line in dimmed text
4. **Scroll**: For long verses, scroll lyrics as playback progresses

### Display Modes

| Mode | Description |
|------|-------------|
| `lyrics_off` | Don't show lyrics (current behavior) |
| `lyrics_inline` | Show current lyric line above metronome |
| `lyrics_panel` | Dedicated lyrics panel (for larger terminals) |

Toggle with keyboard shortcut: `L` or `Shift+L`

## Implementation Plan

### Phase 1: Parser Support

1. Add `Lyrics` field to `Section` struct
2. Create `ParseLyrics()` function to parse chord-over-lyrics format
3. Store parsed lyrics in `PlaybackData`
4. Unit tests for lyrics parsing

**Files to modify:**
- `parser/parser.go` - Add Lyrics field
- `parser/lyrics.go` - New file for lyrics parsing
- `midi/realtime.go` - Include lyrics in PlaybackData

### Phase 2: Basic Display

1. Add lyrics display area to TUI
2. Show current lyric line based on bar position
3. Sync lyrics with section boundaries
4. Handle sections without lyrics gracefully

**Files to modify:**
- `display/tui.go` - Add lyrics rendering
- `player/realtime.go` - Expose lyrics access methods

### Phase 3: Enhanced Features

1. Chord highlighting in lyrics
2. Keyboard toggle for lyrics display
3. Lookahead/lookbehind lines
4. Transpose chord symbols in lyrics display

## Lyrics Parsing Algorithm

```go
func ParseLyrics(raw string) []LyricLine {
    lines := strings.Split(raw, "\n")
    var result []LyricLine

    for i := 0; i < len(lines); i++ {
        line := lines[i]

        if isChordLine(line) {
            // This is a chord line, next line should be lyrics
            chordPositions := extractChordPositions(line)

            if i+1 < len(lines) && !isChordLine(lines[i+1]) {
                lyricText := lines[i+1]
                result = append(result, LyricLine{
                    Text:       lyricText,
                    ChordMarks: chordPositions,
                })
                i++ // Skip the lyric line we just processed
            }
        } else if strings.TrimSpace(line) != "" {
            // Lyric line without chords above it
            result = append(result, LyricLine{
                Text:       line,
                ChordMarks: nil,
            })
        }
    }

    return result
}

func isChordLine(line string) bool {
    tokens := strings.Fields(line)
    if len(tokens) == 0 {
        return false
    }

    chordCount := 0
    for _, token := range tokens {
        if isChordSymbol(token) {
            chordCount++
        }
    }

    // Line is a chord line if >50% of tokens are chords
    return float64(chordCount)/float64(len(tokens)) > 0.5
}

func isChordSymbol(s string) bool {
    // Match patterns like: C, Am, Dm7, G#m, Cmaj7, F/A, Bb, etc.
    pattern := `^[A-G][#b]?(m|maj|min|dim|aug|sus|add)?[0-9]*(\/[A-G][#b]?)?$`
    matched, _ := regexp.MatchString(pattern, s)
    return matched
}

func extractChordPositions(line string) []ChordMark {
    var marks []ChordMark

    // Find each chord and its character position
    pos := 0
    for _, segment := range splitKeepingPosition(line) {
        if isChordSymbol(strings.TrimSpace(segment.text)) {
            marks = append(marks, ChordMark{
                Position: segment.start,
                Chord:    strings.TrimSpace(segment.text),
            })
        }
    }

    return marks
}
```

## Mapping Lyrics to Bars

Since lyrics are defined per-section, we need to map lyric lines to bar positions:

```go
func (lb *LyricsBlock) GetLineAtBar(bar int) *LyricLine {
    if bar < lb.StartBar || bar >= lb.EndBar {
        return nil
    }

    // Calculate relative position within section
    sectionBars := lb.EndBar - lb.StartBar
    relativeBar := bar - lb.StartBar

    // Simple mapping: distribute lines evenly across bars
    // More sophisticated: use chord positions for alignment
    if len(lb.Lines) == 0 {
        return nil
    }

    lineIndex := (relativeBar * len(lb.Lines)) / sectionBars
    if lineIndex >= len(lb.Lines) {
        lineIndex = len(lb.Lines) - 1
    }

    return &lb.Lines[lineIndex]
}
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `l` | Toggle lyrics display on/off |
| `Shift+L` | Cycle lyrics display modes |

## Edge Cases

1. **No lyrics in section**: Display nothing, graceful fallback
2. **Lyrics longer than section**: Scroll or compress display
3. **Chords in lyrics don't match progression**: Display as-is (lyrics are documentation)
4. **Unicode/special characters**: Ensure proper handling in terminal
5. **Very long lines**: Wrap or truncate with ellipsis
6. **Transpose active**: Transpose chord symbols in lyrics display too

## Example: Full Song with Lyrics

```yaml
track:
  title: "Snowman"
  key: C
  tempo: 70
  time_signature: 4/4
  style: pop_ballad

sections:
  - name: verse
    chord_progression:
      pattern: "Dm G Cmaj7 Em Am"
      bars_per_chord: 1
    lyrics: |
      Dm                         G
      Don't cry, snowman, not in front of me
      Cmaj7                              Em          Am
      Who'll catch your tears if you can't catch me, darling?

  - name: chorus
    chord_progression:
      pattern: "E Am G C E Am Dm G"
      bars_per_chord: 1
    lyrics: |
      E                   Am
      I want you to know that I'm never leaving
             G                                    C
      'Cause I'm Mrs. Snow, 'til death we'll be freezing
      E                   Am
      You are my home, my home for all seasons
         Dm        G
      So come on, let's go

form:
  - verse
  - chorus
  - verse
  - chorus
```

## Benefits

1. **Practice aid**: See lyrics while playing along
2. **Learning songs**: Understand song structure with lyrics context
3. **Complete lead sheets**: BTML becomes a full lead sheet format
4. **LLM generation**: Claude can generate complete songs with lyrics from sheet music

## Compatibility

- **Backwards compatible**: `lyrics` field is optional
- **Parser ignores unknown fields**: Existing files continue to work
- **Display fallback**: No lyrics = current behavior

## Future Enhancements

1. **Karaoke mode**: Highlight words as they're sung
2. **Lyrics search**: Find songs by lyric content
3. **Export**: Generate printable lead sheets
4. **Melody sync**: Tie lyrics to melody notes (if melody is present)
