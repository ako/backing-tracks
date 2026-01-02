# Fingerstyle Tablature Display Proposal

**Version:** 1.0
**Date:** 2025-12-26
**Status:** Proposal

## Overview

Add real-time guitar tablature display that shows a complete fingerstyle arrangement combining bass, chords, and melody on a single nylon guitar. The tab scrolls at the bottom of the TUI during playback, helping guitarists learn to play the complete song solo.

## Goals

1. **Learn fingerstyle technique**: Show how bass, chord, and melody interweave
2. **Real-time display**: Tab advances in sync with playback
3. **Playable arrangements**: Generate musically sensible fingerstyle patterns
4. **Visual clarity**: Clear 6-line tab format readable in terminal

## Example Display

```
┌─ Die With a Smile ────────────────────────────────────────────────────────┐
│ Key: A | Tempo: 105 BPM | 6/8 | ballad | [verse]                          │
├───────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│              Amaj7                              D                         │
│         ◉    ○    ○    ○    ○    ○         Bar 3/48                      │
│         1    2    3    4    5    6                                        │
│                                                                           │
│  ▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  │
├───────────────────────────────────────────────────────────────────────────┤
│  Amaj7                          │ D                                       │
│e ├──0─────0───0─────0───────────┼───2─────2───2─────2───────────┤        │
│B ├──2───────────2───────2───────┼───3───────────3───────3───────┤        │
│G ├──1─────1───────1───────1─────┼───2─────2───────2───────2─────┤        │
│D ├──2───────────────────────────┼───0───────────────────────────┤        │
│A ├──0───────────────────────────┼─────────────────────────0─────┤        │
│E ├──────────────────────────────┼─────────────────────────────── │        │
│    1 . 2 . 3 . 4 . 5 . 6 .       1 . 2 . 3 . 4 . 5 . 6 .                 │
│    ▲                             ▲                                        │
│    └─ NOW                        └─ NEXT                                  │
└───────────────────────────────────────────────────────────────────────────┘
```

## Fingerstyle Pattern Generation

### Pattern Types

```yaml
fingerstyle:
  pattern: travis        # Pattern name
  bass_strings: [6, 5, 4] # Which strings for bass notes
  melody_strings: [1, 2, 3] # Which strings for melody
```

#### 1. Travis Picking (4/4)
Alternating bass with melody/chord fills.

```
Beat:    1   &   2   &   3   &   4   &
Bass:    B       B       B       B
Melody:      M       M       M       M
```

```
e ├──────0───────0───────0───────0───┤
B ├──────────────────────────────────┤
G ├──────────1───────1───────1───────┤
D ├──────────────────────────────────┤
A ├──────────────2───────────────────┤  (alternating bass)
E ├──0───────────────0───────────────┤
     1   &   2   &   3   &   4   &
```

#### 2. Arpeggio Pattern (4/4)
Rolling chord tones from bass to treble.

```
e ├──────────────0───────────────0───┤
B ├──────────1───────────────1───────┤
G ├──────0───────────────0───────────┤
D ├──────────────────────────────────┤
A ├──0───────────────0───────────────┤
E ├──────────────────────────────────┤
     1   &   2   &   3   &   4   &
```

#### 3. Folk/Ballad Pattern (6/8)
Bass on 1, arpeggiated chord on 2-3-4-5-6.

```
e ├──────────────0───0───────────────┤
B ├──────────1───────────1───────────┤
G ├──────0───────────────────0───────┤
D ├──────────────────────────────────┤
A ├──0───────────────────────────────┤
E ├──────────────────────────────────┤
     1   2   3   4   5   6
```

#### 4. Classical Arpeggio (PIMA)
Thumb-Index-Middle-Ring pattern.

```
e ├──────────────────a───────────────┤  a = ring finger
B ├──────────────m───────────────────┤  m = middle finger
G ├──────────i───────────────────────┤  i = index finger
D ├──────────────────────────────────┤
A ├──────p───────────────────────────┤  p = thumb
E ├──────────────────────────────────┤
```

#### 5. Bossa Nova Pattern
Syncopated Brazilian pattern.

```
e ├──0───────0───────0───────────────┤
B ├──────1───────1───────1───────────┤
G ├──────────────────────────0───────┤
D ├──2───────────────2───────────────┤
A ├──────────────────────────────────┤
E ├──────────────────────────────────┤
     1   &   2   &   3   &   4   &
```

### Chord Voicing Selection

The system selects appropriate chord voicings based on:

1. **Bass note requirement**: Root in bass (or specified bass note for slash chords)
2. **Melody note**: Highest note should support melody line
3. **Playability**: Prefer open chords, common shapes
4. **Voice leading**: Minimize finger movement between chords

```go
type ChordVoicing struct {
    Name      string   // "Amaj7", "D", etc.
    Frets     [6]int   // -1 = muted, 0 = open, 1+ = fret
    Fingers   [6]int   // Which finger (1-4) or 0 for open/muted
    BassNote  int      // MIDI note of lowest sounding string
    MelodyNote int     // MIDI note of highest sounding string
}

// Example voicings
var Amaj7_open = ChordVoicing{
    Name:   "Amaj7",
    Frets:  [6]int{0, 0, 1, 2, 2, 0},  // E A D G B e
    Fingers: [6]int{0, 0, 1, 3, 2, 0},
    BassNote: 40,  // E2
    MelodyNote: 64, // E4
}
```

## Data Structures

### Parser Additions

```go
// New section in BTML
type FingerstyleConfig struct {
    Enabled     bool   `yaml:"enabled"`
    Pattern     string `yaml:"pattern"`      // travis, arpeggio, folk, classical, bossa
    BassStrings []int  `yaml:"bass_strings"` // Default: [6, 5, 4]
    Complexity  string `yaml:"complexity"`   // simple, moderate, advanced
}

// In Track struct
type Track struct {
    // ... existing fields ...
    Fingerstyle FingerstyleConfig `yaml:"fingerstyle,omitempty"`
}
```

### Tab Generation

```go
// TabNote represents a single note in tablature
type TabNote struct {
    String   int     // 1-6 (1=high e, 6=low E)
    Fret     int     // 0-24
    Beat     float64 // Position within bar (1.0, 1.5, 2.0, etc.)
    Duration float64 // In beats
    Finger   string  // p, i, m, a (right hand) or 1-4 (left hand)
}

// TabBar represents one bar of tablature
type TabBar struct {
    ChordName string
    Notes     []TabNote
    BarNumber int
}

// TabDisplay manages the scrolling tab view
type TabDisplay struct {
    CurrentBar int
    Bars       []TabBar
    ViewWidth  int      // How many bars visible
    Tempo      float64
}
```

### Tab Generation Algorithm

```go
func GenerateFingerstyleTab(track *Track) []TabBar {
    var bars []TabBar

    for _, section := range track.Sections {
        chords := section.GetChords()

        for i, chord := range chords {
            voicing := selectVoicing(chord, section.Fingerstyle)
            pattern := getPattern(section.Fingerstyle.Pattern, track.TimeSignature)

            bar := TabBar{
                ChordName: chord,
                BarNumber: i + 1,
            }

            // Apply pattern to voicing
            for _, patternNote := range pattern {
                note := TabNote{
                    String:   patternNote.String,
                    Fret:     voicing.Frets[6-patternNote.String],
                    Beat:     patternNote.Beat,
                    Duration: patternNote.Duration,
                }
                bar.Notes = append(bar.Notes, note)
            }

            // Add melody notes if melody is enabled
            if track.Melody.Enabled {
                melodyNotes := getMelodyForBar(track, i)
                bar.Notes = mergeMelodyIntoTab(bar.Notes, melodyNotes, voicing)
            }

            bars = append(bars, bar)
        }
    }

    return bars
}
```

## Display Implementation

### Tab Rendering

```go
func (td *TabDisplay) Render(currentBeat float64) string {
    var sb strings.Builder

    // Calculate which bars to show (current + next)
    currentBar := td.getCurrentBar(currentBeat)

    // Render chord names row
    sb.WriteString(fmt.Sprintf("  %-16s │ %-16s\n",
        td.Bars[currentBar].ChordName,
        td.Bars[currentBar+1].ChordName))

    // Render 6 string lines
    strings := []string{"e", "B", "G", "D", "A", "E"}
    for i, stringName := range strings {
        sb.WriteString(fmt.Sprintf("%s ├", stringName))
        sb.WriteString(td.renderStringLine(currentBar, 6-i))
        sb.WriteString("┼")
        sb.WriteString(td.renderStringLine(currentBar+1, 6-i))
        sb.WriteString("┤\n")
    }

    // Render beat markers
    sb.WriteString(td.renderBeatMarkers(currentBeat))

    // Render "NOW" indicator
    sb.WriteString(td.renderPlayhead(currentBeat))

    return sb.String()
}

func (td *TabDisplay) renderStringLine(barIndex, stringNum int) string {
    bar := td.Bars[barIndex]
    line := make([]rune, td.BarWidth)
    for i := range line {
        line[i] = '─'
    }

    for _, note := range bar.Notes {
        if note.String == stringNum {
            pos := td.beatToPosition(note.Beat)
            fretStr := strconv.Itoa(note.Fret)
            for j, c := range fretStr {
                if pos+j < len(line) {
                    line[pos+j] = c
                }
            }
        }
    }

    return string(line)
}
```

### Scrolling Behavior

The display shows 2 bars at a time:
1. **Current bar**: Being played now, playhead indicator shows exact position
2. **Next bar**: Coming up, gives player time to prepare

```
│ Amaj7 (current)        │ D (next)               │
│◄── playing this ──────►│◄── prepare for this ──►│
```

When the playhead reaches the end of the current bar, the view shifts:
- Current bar becomes the left bar
- Next bar loads into right position

### Playhead Animation

```go
func (td *TabDisplay) renderPlayhead(currentBeat float64) string {
    barBeat := math.Mod(currentBeat, float64(td.BeatsPerBar))
    position := int((barBeat / float64(td.BeatsPerBar)) * float64(td.BarWidth))

    line := strings.Repeat(" ", td.BarWidth*2 + 3)
    runes := []rune(line)
    runes[position+2] = '▲'

    return string(runes) + "\n    └─ NOW\n"
}
```

## BTML Configuration

### Simple Usage

```yaml
fingerstyle:
  enabled: true
  pattern: travis
```

### Advanced Usage

```yaml
fingerstyle:
  enabled: true
  pattern: classical
  complexity: advanced
  bass_strings: [6, 5]
  custom_pattern: |
    1: p
    1.5: i
    2: m
    2.5: a
    3: p
    3.5: i
    4: m
    4.5: a
```

### Per-Section Override

```yaml
sections:
  - name: verse
    chord_progression:
      pattern: "Am G F E"
    fingerstyle:
      pattern: arpeggio

  - name: chorus
    chord_progression:
      pattern: "C G Am F"
    fingerstyle:
      pattern: travis
      complexity: moderate
```

## Implementation Plan

### Phase 1: Basic Tab Display

1. Create `display/tablature.go` with tab rendering logic
2. Add static tab display to TUI (bottom panel)
3. Support basic chord voicings (open chords)
4. Implement single pattern (Travis picking)

**Files:**
- `display/tablature.go` (new)
- `display/tui.go` (add tab panel)
- `midi/voicings.go` (new - chord voicing database)

### Phase 2: Pattern Library

1. Implement pattern types (travis, arpeggio, folk, classical, bossa)
2. Add pattern selection based on time signature and style
3. Create pattern generation from templates

**Files:**
- `midi/patterns.go` (new)
- `parser/parser.go` (add fingerstyle config)

### Phase 3: Real-time Sync

1. Sync tab scroll with playback position
2. Implement playhead animation
3. Add bar-to-bar transitions
4. Highlight current notes being played

**Files:**
- `display/tablature.go` (enhance)
- `player/realtime.go` (expose timing)

### Phase 4: Melody Integration

1. Merge melody notes into fingerstyle tab
2. Voice leading optimization
3. Automatic position selection (which fret to play melody)

**Files:**
- `midi/tablature.go` (new - tab generation from melody)
- `midi/voicings.go` (enhance)

### Phase 5: Advanced Features

1. Chord diagrams alongside tab
2. Finger position hints (PIMA, 1-4)
3. Export to PDF/image
4. Custom pattern editor

## Chord Voicing Database

### Open Chord Voicings

```go
var OpenVoicings = map[string]ChordVoicing{
    "C":     {Frets: [6]int{-1, 3, 2, 0, 1, 0}},
    "Am":    {Frets: [6]int{-1, 0, 2, 2, 1, 0}},
    "Am7":   {Frets: [6]int{-1, 0, 2, 0, 1, 0}},
    "G":     {Frets: [6]int{3, 2, 0, 0, 0, 3}},
    "D":     {Frets: [6]int{-1, -1, 0, 2, 3, 2}},
    "Dm":    {Frets: [6]int{-1, -1, 0, 2, 3, 1}},
    "E":     {Frets: [6]int{0, 2, 2, 1, 0, 0}},
    "Em":    {Frets: [6]int{0, 2, 2, 0, 0, 0}},
    "Em7":   {Frets: [6]int{0, 2, 0, 0, 0, 0}},
    "A":     {Frets: [6]int{-1, 0, 2, 2, 2, 0}},
    "Amaj7": {Frets: [6]int{-1, 0, 2, 1, 2, 0}},
    "F":     {Frets: [6]int{1, 3, 3, 2, 1, 1}},  // Barre
    "Fmaj7": {Frets: [6]int{-1, -1, 3, 2, 1, 0}}, // Easy version
    // ... more voicings
}
```

### Voicing Selection Logic

```go
func selectVoicing(chord string, config FingerstyleConfig) ChordVoicing {
    // 1. Try open voicing first (easier to play)
    if v, ok := OpenVoicings[chord]; ok {
        return v
    }

    // 2. Try barre chord based on root
    root := parseRoot(chord)
    quality := parseQuality(chord)

    // 3. Generate moveable shape
    return generateBarreVoicing(root, quality)
}
```

## Keyboard Controls

| Key | Action |
|-----|--------|
| `t` | Toggle tablature display on/off |
| `Shift+T` | Cycle tab display modes (2-bar, 4-bar, full) |
| `;` | Previous pattern type |
| `'` | Next pattern type |
| `p` | Cycle pattern complexity (simple → moderate → advanced) |

**Note:** `[` and `]` are already used for capo control with audio transpose.

## Display Modes

### Mode 1: Compact (2 bars)
Default view showing current and next bar.

### Mode 2: Extended (4 bars)
Shows current bar + 3 upcoming bars for more preparation time.

### Mode 3: Full Section
Shows entire current section's tab (scrollable).

## Edge Cases

1. **Chord not in voicing database**: Generate from chord tones + root
2. **Melody note conflicts with chord**: Adjust voicing to include melody note
3. **Impossible stretches**: Suggest capo or alternate voicing
4. **Very fast tempo**: Simplify pattern automatically
5. **Non-standard tuning**: Allow tuning specification in BTML

## Example: Complete Song Tab

For "Die With a Smile" with fingerstyle tab enabled:

```yaml
fingerstyle:
  enabled: true
  pattern: folk      # 6/8 arpeggio pattern
  complexity: moderate
```

Would generate:

```
  Amaj7                          │ D
e ├──0─────────0───0─────────────┼───2─────────2───2─────────────┤
B ├──────2───────────────2───────┼───────3───────────────3───────┤
G ├──────────1───────1───────────┼───────────2───────2───────────┤
D ├──────────────────────────────┼───0───────────────────────────┤
A ├──0───────────────────────────┼───────────────────────────────┤
E ├──────────────────────────────┼───────────────────────────────┤
     1   2   3   4   5   6         1   2   3   4   5   6
     ▲
     └─ NOW
```

## Benefits

1. **Self-contained practice**: No separate tab needed
2. **Learn fingerstyle**: See bass/chord/melody interaction
3. **Real-time feedback**: Tab syncs with audio
4. **Gradual complexity**: Start simple, increase pattern complexity
5. **Style variety**: Different patterns for different genres

## Future Enhancements

1. **Tab export**: Generate static tab files (ASCII, PDF, Guitar Pro)
2. **Practice mode**: Slow down tempo, loop sections
3. **Difficulty levels**: Auto-simplify for beginners
4. **Custom voicings**: User-defined chord shapes
5. **Capo support**: Transpose display for capo position
6. **Left-hand fingering**: Show which fingers to use

---

**Status**: Proposal
**Complexity**: High (significant new feature)
**Dependencies**: Existing chord/melody generation
