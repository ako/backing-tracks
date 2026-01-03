# Backing Tracks - Architecture Document

## Overview

**backing-tracks** is a terminal-based MIDI backing track player written in Go. It enables guitarists to create and play full-band backing tracks (drums, bass, chords, melody) from YAML-based BTML files, with real-time visualization using Bubbletea TUI and audio synthesis via FluidSynth.

**Version:** v0.6 | **Language:** Go 1.24

---

## Component Architecture

```mermaid
graph TB
    subgraph CLI["CLI Layer"]
        main[main.go]
    end

    subgraph Parser["Parser Package"]
        parser[parser.go]
        lyrics[lyrics.go]
    end

    subgraph MIDI["MIDI Generation Package"]
        generator[generator.go]
        rhythm[rhythm.go]
        bass[bass.go]
        drums[drums.go]
        melody[melody.go]
        patterns[patterns.go]
        voicings[voicings.go]
        realtime_midi[realtime.go]
        tablature_midi[tablature.go]
    end

    subgraph Display["Display Package"]
        tui[tui.go]
        fretboard[fretboard.go]
        chords[chords.go]
        tablature[tablature.go]
        live[live.go]
    end

    subgraph Player["Player Package"]
        fluidsynth[fluidsynth.go]
        realtime_player[realtime.go]
    end

    subgraph Theory["Theory Package"]
        theory[theory.go]
    end

    subgraph External["External"]
        fs[(FluidSynth)]
        sf[(SoundFont .sf2)]
    end

    main --> parser
    main --> generator
    main --> tui
    main --> fluidsynth

    parser --> lyrics

    generator --> rhythm
    generator --> bass
    generator --> drums
    generator --> melody
    generator --> patterns
    generator --> voicings
    generator --> realtime_midi

    rhythm --> theory
    bass --> theory
    melody --> theory
    voicings --> theory

    tui --> fretboard
    tui --> chords
    tui --> tablature
    tui --> realtime_player

    fretboard --> theory
    chords --> theory

    fluidsynth --> realtime_player
    realtime_player --> realtime_midi
    realtime_player --> fs
    fs --> sf
```

---

## Package Structure

```
backing-tracks/
├── main.go                    # CLI entry point
├── parser/                    # BTML parsing
│   ├── parser.go              # Core data model & YAML parsing
│   └── lyrics.go              # Beat-mapped lyrics parsing
├── midi/                      # MIDI generation
│   ├── generator.go           # Orchestrator
│   ├── rhythm.go              # Chord rhythm patterns
│   ├── bass.go                # Bass line generation
│   ├── drums.go               # Drum patterns (incl. Euclidean)
│   ├── melody.go              # Melody generation
│   ├── patterns.go            # Fingerstyle pattern library
│   ├── voicings.go            # Guitar chord voicings
│   ├── realtime.go            # Real-time playback data
│   └── tablature.go           # Tablature generation
├── display/                   # User interface
│   ├── tui.go                 # Bubbletea TUI model
│   ├── fretboard.go           # Fretboard visualization
│   ├── chords.go              # Chord chart display
│   ├── tablature.go           # Tablature display
│   └── live.go                # Legacy ANSI display
├── player/                    # Audio playback
│   ├── fluidsynth.go          # FluidSynth integration
│   └── realtime.go            # Real-time player control
├── theory/                    # Music theory
│   └── theory.go              # Scales, tunings, intervals
└── strudel/                   # Export
    └── generator.go           # Strudel.cc export
```

---

## Data Model

```mermaid
classDiagram
    class Track {
        +TrackInfo Info
        +ChordProgression Progression
        +[]Section Sections
        +[]string Form
        +*Rhythm Rhythm
        +*Bass Bass
        +*Drums Drums
        +*Melody Melody
        +*ScaleConfig Scale
    }

    class TrackInfo {
        +string Title
        +string Key
        +int Tempo
        +string TimeSignature
        +string Style
        +int Capo
        +string Tuning
    }

    class ChordProgression {
        +string Pattern
        +int BarsPerChord
        +int Repeat
        +GetChords() []Chord
    }

    class Chord {
        +string Symbol
        +float64 Bars
        +string Section
    }

    class Section {
        +string Name
        +ChordProgression ChordProgression
        +string Lyrics
    }

    class Rhythm {
        +string Style
        +string Pattern
        +float64 Swing
        +string Accent
        +string Instrument
    }

    class Bass {
        +string Style
        +float64 Swing
        +string Instrument
    }

    class Drums {
        +string Style
        +*DrumPattern Kick
        +*DrumPattern Snare
        +*DrumPattern Hihat
        +float64 Intensity
    }

    class DrumPattern {
        +*EuclideanRhythm Euclidean
        +[]int Beats
    }

    class EuclideanRhythm {
        +int Hits
        +int Steps
        +int Rotation
    }

    Track --> TrackInfo
    Track --> ChordProgression
    Track --> Section
    Track --> Rhythm
    Track --> Bass
    Track --> Drums
    ChordProgression --> Chord
    Section --> ChordProgression
    Drums --> DrumPattern
    DrumPattern --> EuclideanRhythm
```

---

## MIDI Generation Pipeline

```mermaid
flowchart TB
    subgraph Input
        btml[BTML File]
    end

    subgraph Parsing
        load[parser.LoadTrack]
        validate[Validate & Defaults]
    end

    subgraph Generation["MIDI Generation"]
        gen[GenerateFromTrack]

        subgraph Tracks
            t0[Track 0: Tempo]
            t1[Track 1: Chords]
            t2[Track 2: Bass]
            t3[Track 3: Drums]
            t4[Track 4: Melody]
        end

        subgraph Generators
            rhythm_gen[GenerateChordRhythm]
            bass_gen[GenerateBassLine]
            drum_gen[GenerateDrumPattern]
            melody_gen[GenerateMelody]
        end
    end

    subgraph Output
        midi[MIDI File]
        playback[PlaybackData]
    end

    btml --> load --> validate --> gen

    gen --> t0
    gen --> t1 --> rhythm_gen
    gen --> t2 --> bass_gen
    gen --> t3 --> drum_gen
    gen --> t4 --> melody_gen

    rhythm_gen --> midi
    bass_gen --> midi
    drum_gen --> midi
    melody_gen --> midi

    gen --> playback
```

### MIDI Track Configuration

| Track | Channel | Program | Content |
|-------|---------|---------|---------|
| 0 | - | - | Tempo metadata |
| 1 | 0 | Piano (0) | Chord voicings |
| 2 | 1 | Fingered Bass (33) | Bass notes |
| 3 | 9 | GM Drums | Drum hits |
| 4 | 2 | Steel Guitar (25) | Melody |

### Timing Constants

```go
TICKS_PER_QUARTER = 480    // Standard MIDI resolution
TICKS_PER_BAR     = 1920   // 4 quarters × 480 (4/4 time)
```

---

## Real-Time Playback Architecture

```mermaid
sequenceDiagram
    participant User
    participant TUI as TUI Model
    participant Player as RealtimePlayer
    participant PB as PlaybackLoop
    participant FS as FluidSynth

    User->>TUI: Start playback
    TUI->>Player: Start()
    Player->>FS: Launch subprocess
    Player->>PB: spawn goroutine

    rect rgb(240, 248, 255)
        note right of PB: Every 5ms
        PB->>PB: Calculate elapsed time
        PB->>PB: Find pending events
        PB->>FS: Send MIDI via stdin
        FS->>FS: Synthesize audio
    end

    rect rgb(255, 248, 240)
        note right of TUI: Every 50ms
        TUI->>Player: GetPlaybackState()
        Player-->>TUI: bar, beat, paused
        TUI->>TUI: Render display
    end

    User->>TUI: Press key
    TUI->>Player: TogglePause/Seek/Transpose
    Player->>PB: Update state
    Player->>FS: AllNotesOff if needed
```

### Player State Machine

```mermaid
stateDiagram-v2
    [*] --> Stopped
    Stopped --> Playing: Start()
    Playing --> Paused: TogglePause()
    Paused --> Playing: TogglePause()
    Playing --> Playing: Seek/Transpose/Mute
    Paused --> Paused: Seek/Transpose/Mute
    Playing --> Stopped: Stop()
    Paused --> Stopped: Stop()
    Stopped --> [*]
```

---

## TUI Display Architecture

```mermaid
graph TB
    subgraph TUI["TUI Model (tea.Model)"]
        state[Playback State]
        handlers[Key Handlers]
        renderer[View Renderer]
    end

    subgraph Components["Display Components"]
        fretboard[Fretboard Display]
        chordchart[Chord Chart]
        tablature[Tablature Display]
        progress[Progress Bar]
        lyrics[Lyrics Display]
    end

    subgraph Layout["3-Column Layout"]
        left[Left: Fretboard + Scale]
        center[Center: Chords + Progress]
        right[Right: Tablature + Patterns]
    end

    state --> renderer
    handlers --> state

    renderer --> fretboard --> left
    renderer --> chordchart --> center
    renderer --> tablature --> right
    renderer --> progress --> center
    renderer --> lyrics --> center
```

### Keyboard Controls

| Key | Action |
|-----|--------|
| Space | Pause/Resume |
| ←/→ | Seek by 1 bar |
| ↑/↓ | Transpose semitone |
| Shift+↑/↓ | Adjust tempo ±5 BPM |
| [/] | Adjust capo (with audio transpose) |
| {/} | Visual capo only (no audio change) |
| </> | Cycle through tunings |
| 1-5 | Mute track (1=drums, 2=bass, 3=chords, 4=melody, 5=fingerstyle) |
| Shift+1-9 | Loop N bars from current position |
| Shift+0 | Loop current section |
| ;/' | Cycle fingerstyle pattern |
| l | Toggle lyrics display |
| m | Toggle metronome/beat counter |
| s | Toggle strum pattern display |
| t | Toggle inline tablature |
| c | Toggle chord names display |
| e | Enter lyrics edit mode |
| q | Quit |

### Lyrics Edit Mode (when in edit mode)

| Key | Action |
|-----|--------|
| ← | Select previous word |
| → | Select next word |
| Shift+← | Move selected word one beat earlier |
| Shift+→ | Move selected word one beat later |
| Enter | Save changes and exit edit mode |
| Esc | Discard changes and exit edit mode |

---

## Key Algorithms

### 1. Euclidean Rhythm (Bjorklund's Algorithm)

Distributes N hits evenly across M steps. Used for drum patterns.

```mermaid
flowchart LR
    subgraph Input
        hits[Hits: 5]
        steps[Steps: 8]
        rotation[Rotation: 0]
    end

    subgraph Algorithm
        dist[Distribute evenly]
        rotate[Apply rotation]
    end

    subgraph Output
        pattern["x.x.x.x."]
    end

    hits --> dist
    steps --> dist
    dist --> rotate
    rotation --> rotate
    rotate --> pattern
```

**Example patterns:**
- `(3, 8, 0)` → `x..x..x.` (Tresillo)
- `(5, 8, 0)` → `x.x.x.x.` (Cinquillo)
- `(4, 12, 0)` → `x..x..x..x..` (Triplet feel)

### 2. Swing Implementation

Adjusts timing of off-beat notes for swing feel.

```
Swing ratio: 0.5 = straight, 0.67 = triplet swing

Straight (0.5):   |----|----|----|----|
                  1    2    3    4

Swing (0.67):     |------|----|------|----|
                  1      2    3      4
```

### 3. Chord Voicing Generation

```mermaid
flowchart TB
    symbol[Chord Symbol: Am7]

    subgraph Parse
        root[Extract Root: A]
        quality[Extract Quality: m7]
    end

    subgraph Build
        base[Root MIDI: 57]
        intervals[Add Intervals]
    end

    subgraph Output
        voicing["[57, 60, 64, 67]"]
    end

    symbol --> root --> base
    symbol --> quality --> intervals
    base --> intervals --> voicing
```

**Interval mappings:**
| Quality | Intervals (semitones) |
|---------|----------------------|
| Major | 0, 4, 7 |
| Minor (m) | 0, 3, 7 |
| Dominant 7 | 0, 4, 7, 10 |
| Major 7 | 0, 4, 7, 11 |
| Minor 7 | 0, 3, 7, 10 |
| Power (5) | 0, 7, 12 |

### 4. Bass Line Generation

```mermaid
flowchart LR
    input[Bass Style]

    subgraph Styles
        root[root: Root only]
        rf[root_fifth: R + 5th]
        walk[walking: R-3-5-7]
        swing_walk[swing_walking: Walking + swing]
    end

    input --> Styles
```

### 5. Beat-Mapped Lyrics Parsing

Parses Beatles-style chord/beat notation:

```
Input:
  Am   /    /    /    G    /    /    /
  Words placed un-der beats where sung

Output:
  BeatLyric{Bar:0, Beat:0, Chord:"Am", Lyrics:"Words"}
  BeatLyric{Bar:0, Beat:1, Chord:"",   Lyrics:"placed"}
  BeatLyric{Bar:0, Beat:2, Chord:"",   Lyrics:"un-der"}
  BeatLyric{Bar:0, Beat:3, Chord:"",   Lyrics:"beats"}
  BeatLyric{Bar:1, Beat:0, Chord:"G",  Lyrics:"where"}
  ...
```

---

## Data Flow: Complete Execution

```mermaid
flowchart TB
    subgraph CLI
        cmd[backing-tracks play file.btml]
    end

    subgraph Parse
        load[LoadTrack]
        expand[Expand Sections]
    end

    subgraph Generate
        midi_gen[GenerateFromTrack]
        playback_gen[GeneratePlaybackData]
    end

    subgraph Display
        show[ShowTrack]
        tui_init[NewTUIModel]
    end

    subgraph Playback
        player_init[NewRealtimePlayer]
        fs_start[Start FluidSynth]
        loop[PlaybackLoop]
    end

    subgraph Runtime
        tui_run[TUI Event Loop]
        audio[Audio Output]
    end

    cmd --> load --> expand
    expand --> midi_gen
    expand --> playback_gen
    expand --> show

    playback_gen --> player_init
    player_init --> fs_start
    fs_start --> loop

    show --> tui_init
    tui_init --> tui_run

    loop --> audio
    tui_run <--> loop
```

---

## Dependencies

### Go Modules

```go
require (
    github.com/charmbracelet/bubbletea v1.3.10   // TUI framework
    github.com/charmbracelet/lipgloss v1.1.0     // Terminal styling
    gitlab.com/gomidi/midi/v2 v2.3.18            // MIDI generation
    golang.org/x/term v0.38.0                    // Terminal utilities
    gopkg.in/yaml.v3 v3.0.1                      // YAML parsing
)
```

### External

- **FluidSynth** - Software synthesizer for MIDI playback
- **SoundFont (.sf2)** - Instrument samples (e.g., FluidR3_GM.sf2)

---

## Technical Decisions

| Decision | Rationale |
|----------|-----------|
| YAML for BTML | LLM-friendly, human-readable, supports comments |
| MIDI intermediate format | Portable, inspectable, DAW-compatible |
| FluidSynth | High-quality synthesis, low CPU, cross-platform |
| Bubbletea TUI | Composable components, keyboard-driven |
| Real-time player | Enables seeking, transposing, muting during playback |
| Euclidean rhythms | Mathematically interesting, easy to parameterize |
| 480 ticks/quarter | Standard MIDI resolution for precise timing |

---

## Limitations

1. **4/4 time only** - Other time signatures not yet supported
2. **Single tempo** - No tempo changes mid-song
3. **Terminal-only** - No GUI (intentional design choice)
4. **Linux-focused** - Primary testing on Linux (should work on macOS)

---

## Code Statistics

- **Total Go lines:** ~12,000
- **Packages:** 8
- **Source files:** 22
- **Types defined:** 50+
- **Key interfaces:** tea.Model, PlayerController
