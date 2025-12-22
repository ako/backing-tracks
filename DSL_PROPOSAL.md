# Backing Track Markup Language (BTML)

## Overview

A YAML-based Domain-Specific Language (DSL) for defining guitar backing tracks. Designed to be both human-readable and LLM-friendly, enabling Claude and other AI models to generate backing tracks from YouTube videos or Spotify tracks.

## Design Goals

- **Simple**: Core structure is ~20 lines for basic tracks
- **LLM-friendly**: Clear structure, predictable format, easy to generate
- **Extensible**: Can add lyrics, sections, dynamics later
- **Terminal-ready**: Easy to parse and render as ASCII/Unicode in terminal applications
- **Multi-genre**: Support rock, blues, jazz, and other popular music styles
- **Expressive**: Inspired by Strudel's mini-notation for concise pattern representation
- **Flexible**: Support both compact notation and verbose structured formats

## DSL Structure

### Core Components

1. **track**: Metadata and global properties
2. **scale**: Soloing information and suggested scales
3. **chord_progression**: Core harmonic structure with sections and repeats
4. **bass**: Bass line specification (optional)
5. **drums**: Rhythm section patterns (optional)

## Example: Blues Backing Track

```yaml
---
track:
  title: "Slow Blues in A"
  key: A
  tempo: 80
  time_signature: 4/4
  style: blues
  duration: 48  # bars

scale:
  root: A
  type: minor_pentatonic
  notes: [A, C, D, E, G]
  positions: [5, 8, 12]  # fretboard positions for reference

chord_progression:
  bars_per_chord: 4
  sections:
    - name: verse
      repeat: 2
      chords:
        - { chord: A7, bars: 4 }
        - { chord: A7, bars: 4 }
        - { chord: D7, bars: 4 }
        - { chord: A7, bars: 4 }
        - { chord: E7, bars: 2 }
        - { chord: D7, bars: 2 }
        - { chord: A7, bars: 2 }
        - { chord: E7, bars: 2 }

bass:
  style: walking
  pattern: root_fifth_sixth_fifth  # or explicit notes
  swing: 0.6  # swing feel, 0.5 = straight, 0.67 = triplet swing

drums:
  style: shuffle
  groove: medium
  patterns:
    - name: main
      kick: [1, 3]  # beats in bar
      snare: [2, 4]
      hihat: eighth_note_shuffle
      intensity: 0.7
```

## Example: Jazz II-V-I

```yaml
---
track:
  title: "II-V-I in Bb"
  key: Bb
  tempo: 140
  time_signature: 4/4
  style: jazz

scale:
  root: Bb
  type: major
  notes: [Bb, C, D, Eb, F, G, A]
  chord_scales:
    - { chord: Cm7, scale: dorian, root: C }
    - { chord: F7, scale: mixolydian, root: F }
    - { chord: Bbmaj7, scale: ionian, root: Bb }

chord_progression:
  sections:
    - name: chorus
      repeat: 4
      chords:
        - { chord: Cm7, bars: 2 }
        - { chord: F7, bars: 2 }
        - { chord: Bbmaj7, bars: 4 }

bass:
  style: walking
  swing: 0.67

drums:
  style: jazz_swing
  ride_pattern: ding_ding_a_ding
  intensity: 0.5
```

## Example: Rock Progression

```yaml
---
track:
  title: "Power Chord Rock"
  key: E
  tempo: 120
  time_signature: 4/4
  style: rock

scale:
  root: E
  type: minor
  notes: [E, F#, G, A, B, C, D]
  positions: [0, 7, 12]

chord_progression:
  sections:
    - name: intro
      repeat: 1
      chords:
        - { chord: E5, bars: 2 }
        - { chord: G5, bars: 2 }
        - { chord: A5, bars: 4 }
    - name: verse
      repeat: 2
      chords:
        - { chord: E5, bars: 4 }
        - { chord: C5, bars: 2 }
        - { chord: D5, bars: 2 }
        - { chord: E5, bars: 4 }
        - { chord: G5, bars: 2 }
        - { chord: A5, bars: 2 }

bass:
  style: root_notes
  pattern: quarter_notes

drums:
  style: rock_beat
  groove: driving
  intensity: 0.9
```

## Data Model (Go Implementation)

```go
// Core structs for parsing
type Track struct {
    Info       TrackInfo          `yaml:"track"`
    Scale      Scale              `yaml:"scale"`
    Progression ChordProgression  `yaml:"chord_progression"`
    Bass       *Bass              `yaml:"bass,omitempty"`
    Drums      *Drums             `yaml:"drums,omitempty"`
}

type TrackInfo struct {
    Title         string `yaml:"title"`
    Key           string `yaml:"key"`
    Tempo         int    `yaml:"tempo"`
    TimeSignature string `yaml:"time_signature"`
    Style         string `yaml:"style"`
    Duration      int    `yaml:"duration,omitempty"`
}

type Scale struct {
    Root        string        `yaml:"root"`
    Type        string        `yaml:"type"`
    Notes       []string      `yaml:"notes"`
    Positions   []int         `yaml:"positions,omitempty"`
    ChordScales []ChordScale  `yaml:"chord_scales,omitempty"`
}

type ChordScale struct {
    Chord string `yaml:"chord"`
    Scale string `yaml:"scale"`
    Root  string `yaml:"root"`
}

type ChordProgression struct {
    BarsPerChord int       `yaml:"bars_per_chord,omitempty"`
    Sections     []Section `yaml:"sections"`
}

type Section struct {
    Name   string  `yaml:"name"`
    Repeat int     `yaml:"repeat"`
    Chords []Chord `yaml:"chords"`
}

type Chord struct {
    Chord string `yaml:"chord"`
    Bars  int    `yaml:"bars"`
}

type Bass struct {
    Style   string  `yaml:"style"`
    Pattern string  `yaml:"pattern,omitempty"`
    Swing   float64 `yaml:"swing,omitempty"`
}

type Drums struct {
    Style       string    `yaml:"style"`
    Groove      string    `yaml:"groove,omitempty"`
    RidePattern string    `yaml:"ride_pattern,omitempty"`
    Patterns    []Pattern `yaml:"patterns,omitempty"`
    Intensity   float64   `yaml:"intensity,omitempty"`
}

type Pattern struct {
    Name      string   `yaml:"name"`
    Kick      []int    `yaml:"kick,omitempty"`
    Snare     []int    `yaml:"snare,omitempty"`
    Hihat     string   `yaml:"hihat,omitempty"`
    Intensity float64  `yaml:"intensity,omitempty"`
}
```

## Terminal Output Concept

Example rendering for terminal display:

```
┌─ Slow Blues in A ────────────────────────┐
│ Key: A | Tempo: 80 BPM | 4/4 | Blues     │
└──────────────────────────────────────────┘

Scale: A Minor Pentatonic
  A  C  D  E  G
  [Positions: 5, 8, 12]

Progression (12-bar blues):
┌────┬────┬────┬────┐
│ A7 │ A7 │ A7 │ A7 │
├────┼────┼────┼────┤
│ D7 │ D7 │ A7 │ A7 │
├────┼────┼────┼────┤
│ E7 │ D7 │ A7 │ E7 │
└────┴────┴────┴────┘

♪ Bass: Walking (swing 60%)
♫ Drums: Shuffle (medium)
```

## Flexibility Levels

The DSL supports multiple complexity levels:

### Simple (Pattern Names)
```yaml
bass:
  style: walking
drums:
  style: shuffle
```

### Detailed (Explicit Patterns)
```yaml
drums:
  patterns:
    - name: main
      kick: [1, 3]
      snare: [2, 4]
      hihat: eighth_note_shuffle
```

### Hybrid (Mix Both)
```yaml
bass:
  style: walking
  swing: 0.6
  pattern: root_fifth_sixth_fifth
```

## LLM Generation Guidelines

### Sample Prompt for Claude

```
Analyze this YouTube video/Spotify track of [SONG_NAME] and generate a BTML
(Backing Track Markup Language) document with:

1. Chord progression with proper sections and repeats
2. Suggested scale(s) for soloing
3. Bass pattern style
4. Drum style and groove
5. Tempo and time signature

Output in YAML format following the BTML specification.
```

### What LLMs Should Extract

- **From audio analysis**: Key, tempo, time signature, style/genre
- **From chord detection**: Progression structure, sections, repeats
- **From rhythm analysis**: Drum patterns, bass style, swing feel
- **From music theory**: Appropriate scales for soloing over the progression

## File Format Conventions

- **Extension**: `.btml` or `.yaml`
- **Encoding**: UTF-8
- **Comments**: Use YAML comments (`#`) for annotations
- **Naming**: `song-name-key.btml` (e.g., `slow-blues-a.btml`)

## Strudel-Inspired Features

[Strudel](https://strudel.cc) is a browser-based live coding environment that uses TidalCycles' pattern language in JavaScript. We can adopt several powerful concepts:

### 1. Mini-Notation for Patterns

Compact string-based pattern notation for rhythms and chord sequences:

```yaml
chord_progression:
  # Compact notation: chords in quotes with spacing for rhythm
  pattern: "A7 A7 D7 A7 | E7 D7 A7 E7"
  bars_per_section: 4
```

**Advantages:**
- More concise for simple progressions
- Easier for LLMs to generate from audio analysis
- Natural representation of musical rhythm

### 2. Advanced Chord Notation

Support jazz/extended chord symbols like Strudel:

```yaml
scale:
  chord_scales:
    - { chord: "C^7", scale: "ionian" }      # Major 7th
    - { chord: "A7b13", scale: "altered" }   # Dominant 7 flat 13
    - { chord: "Dm7", scale: "dorian" }
    - { chord: "G7#9", scale: "diminished" }
```

**Chord Symbol Reference:**
- `^7` or `maj7` = Major 7th
- `7` = Dominant 7th
- `m7` = Minor 7th
- `ø7` or `m7b5` = Half-diminished
- `#9`, `b9`, `#11`, `b13` = Alterations

### 3. Euclidean Rhythms for Drums

Algorithmically generate drum patterns using Euclidean rhythm notation `(beats, steps, rotation)`:

```yaml
drums:
  patterns:
    - name: main
      kick:
        euclidean: [3, 8, 0]      # 3 hits across 8 steps: x..x..x.
      snare:
        euclidean: [4, 16, 2]     # 4 hits across 16 steps, rotated by 2
      hihat:
        euclidean: [7, 8, 0]      # Almost every step: xxxxxxx.
      intensity: 0.8
```

**Use Cases:**
- Generate interesting, non-standard drum patterns
- Create polyrhythmic textures
- Quickly experiment with different rhythmic densities

### 4. Pattern Sequencing Syntax

Use angle brackets for alternating/sequencing patterns (like Strudel's `<>` notation):

```yaml
chord_progression:
  # Alternate between two progressions
  pattern: "<A7 D7 A7 E7 | A7 D7 E7 A7>"

bass:
  # Alternate bass patterns
  pattern: "<root fifth | root third fifth octave>"
  swing: 0.6
```

### 5. Nested Rhythmic Subdivisions

Support bracket notation `[]` for rhythmic subdivisions:

```yaml
drums:
  # Four kicks, but third beat is subdivided into two hits
  kick_pattern: "x x [x x] x"
  # Translates to: kick on 1, 2, 3-and, 4

bass:
  # Quarter notes, but bar 3-4 uses eighth notes
  rhythm: "q q [e e e e]"  # q=quarter, e=eighth
  notes: "A A [A C D E]"
```

### 6. Rest Notation

Use `~` for explicit rests:

```yaml
chord_progression:
  pattern: "A7 ~ D7 A7 | E7 D7 ~ E7"
  # Rests create space in the progression

bass:
  pattern: "root ~ fifth ~ | root third ~ fifth"
  # Syncopated bass with rests
```

### 7. Polyrhythm Support

Layer multiple patterns with different time divisions:

```yaml
drums:
  polyrhythms:
    - instrument: kick
      pattern: "x ~ ~ x ~ ~ ~"  # Grouping in 3s
      subdivision: 3
    - instrument: snare
      pattern: "~ x ~ ~"         # Grouping in 4s
      subdivision: 4
  # Creates 3-against-4 polyrhythm
```

### Comparison: Verbose vs Mini-Notation

**Verbose structured format:**
```yaml
chord_progression:
  sections:
    - name: verse
      repeat: 2
      chords:
        - { chord: A7, bars: 4 }
        - { chord: A7, bars: 4 }
        - { chord: D7, bars: 4 }
        - { chord: A7, bars: 4 }
```

**Mini-notation (Strudel-inspired):**
```yaml
chord_progression:
  pattern: "A7*4 A7*4 D7*4 A7*4"  # *4 means 4 bars
  repeat: 2
```

**Even more compact:**
```yaml
chord_progression:
  pattern: "[A7*4]!2 D7*4 A7*4"  # !2 means repeat section twice
```

Both are valid - choose based on complexity and preference!

### Hybrid Approach: Best of Both Worlds

Combine structured YAML with mini-notation for optimal flexibility:

```yaml
---
track:
  title: "Strudel-Style Jazz Blues"
  key: Bb
  tempo: 120
  time_signature: 4/4

scale:
  root: Bb
  type: mixolydian
  notes: [Bb, C, D, Eb, F, G, Ab]

chord_progression:
  # Mini-notation for quick entry
  pattern: "Bb7 Eb7 Bb7 Bb7 | Eb7 Eb7 Bb7 Bb7 | F7 Eb7 Bb7 F7"
  bars_per_chord: 1

  # OR structured for complex voicings
  sections:
    - name: head
      chords:
        - { chord: "Bb7#9", bars: 1, voicing: "rootless_a" }
        - { chord: "Eb7b13", bars: 1, voicing: "rootless_b" }

bass:
  # Compact pattern notation
  pattern: "<root fifth | root third fifth sixth>"
  swing: 0.67

drums:
  style: jazz_swing
  patterns:
    - name: main
      ride:
        pattern: "ding ding-a ding"  # Swing ride pattern
      kick:
        euclidean: [3, 8, 1]         # Sparse, syncopated
      snare:
        euclidean: [2, 8, 3]         # Backbeat with variation
```

### Implementation Considerations

**Parsing Strategy:**
1. Support both notations (mini-notation strings and structured objects)
2. Parse mini-notation into internal event sequences
3. Allow mixing: use mini-notation where convenient, structured where precise

**Go Parser Components:**
```go
type ChordProgression struct {
    // Option 1: Mini-notation string
    Pattern string `yaml:"pattern,omitempty"`

    // Option 2: Structured sections
    Sections []Section `yaml:"sections,omitempty"`

    BarsPerChord int `yaml:"bars_per_chord,omitempty"`
}

type DrumPattern struct {
    // Simple string pattern
    Pattern string `yaml:"pattern,omitempty"`

    // OR Euclidean rhythm spec
    Euclidean *EuclideanRhythm `yaml:"euclidean,omitempty"`

    // OR explicit beat positions
    Beats []int `yaml:"beats,omitempty"`
}

type EuclideanRhythm struct {
    Hits     int `yaml:"hits"`      // Number of hits
    Steps    int `yaml:"steps"`     // Total steps
    Rotation int `yaml:"rotation"`  // Offset rotation
}
```

**Benefits:**
- **Conciseness**: Simple patterns in 1 line vs 10+ lines
- **Power**: Complex polyrhythms and algorithmic patterns
- **LLM-Friendly**: Both formats are easy for AI to generate
- **Human-Friendly**: Musicians can choose their preferred notation style

## Future Extensions

Potential additions to the DSL:

1. **Lyrics and Markers**
   ```yaml
   markers:
     - { bar: 1, text: "Intro" }
     - { bar: 13, text: "Solo section starts" }
   ```

2. **Dynamics and Articulation**
   ```yaml
   dynamics:
     - { section: verse, level: 0.6 }
     - { section: chorus, level: 0.9 }
   ```

3. **Multiple Instruments**
   ```yaml
   instruments:
     - type: rhythm_guitar
       pattern: strumming
       voicing: barre_chords
   ```

4. **Click Track / Metronome**
   ```yaml
   metronome:
     enabled: true
     accent_downbeat: true
     count_in: 2  # bars
   ```

## Why YAML?

1. **LLM Excellence**: Language models are trained on extensive YAML and generate it reliably
2. **Human-Readable**: Musicians can easily read and edit files
3. **Native Go Support**: Standard library via `gopkg.in/yaml.v3`
4. **Comments**: Support for annotations and explanations
5. **Hierarchical**: Natural fit for musical structure (track → sections → chords)
6. **Whitespace-Significant**: Encourages clean, organized structure

## Implementation Libraries

### Go Dependencies

```go
import (
    "gopkg.in/yaml.v3"  // YAML parsing
    "github.com/charmbracelet/lipgloss"  // Terminal styling
    "github.com/charmbracelet/bubbletea"  // TUI framework
)
```

### Parsing Example

```go
func LoadTrack(filename string) (*Track, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var track Track
    if err := yaml.Unmarshal(data, &track); err != nil {
        return nil, err
    }

    return &track, nil
}
```

## Advantages Summary

1. **Simple**: Core structure is concise (~20 lines minimum)
2. **Extensible**: Easy to add new features without breaking existing files
3. **LLM-friendly**: Clear structure, predictable format
4. **Terminal-ready**: Straightforward to parse and render
5. **Multi-genre**: Supports blues, rock, jazz, and more
6. **Theory-aware**: Built-in support for scales, modes, and chord-scale relationships
7. **Practice-oriented**: Focuses on what guitarists need for practice sessions
8. **Expressive**: Strudel-inspired mini-notation for concise patterns
9. **Algorithmic**: Euclidean rhythms for generating interesting drum patterns
10. **Dual-mode**: Choose between compact notation or verbose structured format

## License

[To be determined - suggest MIT or Apache 2.0 for open source]

## Version

Current Specification Version: 0.1.0 (Draft)
