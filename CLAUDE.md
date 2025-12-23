# Claude Project Context: Backing Tracks

## Project Overview

This is a terminal-based backing track player written in Go that uses YAML-based DSL (Domain-Specific Language) called BTML (Backing Track Markup Language) to define complete backing tracks for guitar practice.

**Current Version:** v0.6

**Purpose:** Enable guitarists to create and play full-band backing tracks (chords, bass, drums) from simple YAML files, with real-time visual display showing current chord and beat.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│ BTML File (.yaml)                                       │
│ - Track metadata (tempo, key, time signature)          │
│ - Chord progression (pattern string)                   │
│ - Bass configuration (style, swing)                    │
│ - Drums configuration (presets or Euclidean rhythms)   │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│ Parser (parser/parser.go)                               │
│ - Parses YAML into Go structs                          │
│ - Validates and sets defaults                          │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│ Display (display/terminal.go)                           │
│ - Shows track info, chord grid, bass/drum info         │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│ MIDI Generation (midi/)                                 │
│ ├─ generator.go: Main MIDI file creation               │
│ ├─ bass.go: Bass pattern generation                    │
│ └─ drums.go: Drum pattern generation (w/ Euclidean)   │
│                                                         │
│ Output: 3-track MIDI file                              │
│ - Track 0: Tempo metadata                              │
│ - Track 1: Chords (channel 0, piano)                   │
│ - Track 2: Bass (channel 1, fingered bass)             │
│ - Track 3: Drums (channel 9, GM drum map)              │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│ Live Display (display/live.go)                          │
│ - Real-time chord and beat visualization               │
│ - Updates 10x/second via goroutine                     │
│ - Visual metronome, progress bar                       │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│ Player (player/fluidsynth.go)                           │
│ - Executes FluidSynth via exec.Command                 │
│ - Manages live display lifecycle                       │
│ - Synthesizes MIDI to audio using SoundFont            │
└─────────────────────────────────────────────────────────┘
```

## Key Components

### 1. Parser (`parser/parser.go`)

**Responsibility:** Parse BTML YAML files into Go structs

**Key Types:**
- `Track`: Root structure containing all track data
- `TrackInfo`: Metadata (title, key, tempo, time signature, style)
- `ChordProgression`: Pattern string, bars per chord, repeat count
- `Bass`: Style (root, root_fifth, walking, swing_walking), swing ratio
- `Drums`: Style presets OR explicit patterns
- `DrumPattern`: Can be Euclidean rhythm OR explicit beats
- `EuclideanRhythm`: Algorithmic rhythm (hits, steps, rotation)

**Important Methods:**
- `LoadTrack(filename)`: Main entry point, loads and validates BTML
- `GetChords()`: Expands pattern string into chord sequence with repeats
- `TotalBars()`: Calculates total bars in progression

### 2. MIDI Generation (`midi/`)

#### `generator.go`
**Responsibility:** Create multi-track MIDI file

**Key Functions:**
- `GenerateFromTrack(track)`: Main orchestrator
- `getChordVoicing(symbol)`: Converts chord symbols to MIDI notes
- `parseRoot(symbol)`: Extracts root note (C, D, E, etc.)
- `parseQuality(symbol)`: Extracts chord quality (7, maj7, m7, etc.)

**MIDI Structure:**
- Uses `gitlab.com/gomidi/midi/v2/smf` for MIDI file generation
- 480 ticks per quarter note
- 1920 ticks per bar (4/4 time)
- Channels: 0=chords, 1=bass, 9=drums (GM standard)

#### `bass.go`
**Responsibility:** Generate bass note patterns

**Styles:**
- `root`: Simple root notes on downbeats
- `root_fifth`: Root on beat 1, fifth on beat 3
- `walking`: Root, 3rd, 5th, 7th pattern
- `swing_walking`: Walking bass with swing feel (configurable ratio)

**Swing Implementation:**
- `swing=0.5`: Straight feel (50/50)
- `swing=0.6`: Slight swing (60/40)
- `swing=0.67`: Triplet swing (67/33)

**Key Functions:**
- `GenerateBassLine()`: Main generator
- `getThird()`: Returns major/minor 3rd based on chord quality
- `getSeventh()`: Returns appropriate 7th interval

#### `drums.go`
**Responsibility:** Generate drum patterns

**Preset Styles:**
- `rock_beat`: Kick 1,3 | Snare 2,4 | 8th note hihat
- `shuffle`: Blues shuffle with triplet feel
- `jazz_swing`: Ride cymbal pattern with sparse kick/snare

**Euclidean Rhythms:**
- Implements Bjorklund's algorithm
- Distributes N hits across M steps evenly
- Supports rotation offset
- Example: `euclidean(5, 8, 0)` = `x.x.x.x.` (5 hits in 8 steps)

**GM Drum Map:**
- Kick: 36 (Bass Drum 1)
- Snare: 38 (Acoustic Snare)
- Closed Hihat: 42
- Open Hihat: 46
- Ride: 51
- Crash: 49

**Key Functions:**
- `GenerateDrumPattern()`: Main generator
- `generateEuclideanRhythm()`: Bjorklund's algorithm implementation
- `generatePresetPattern()`: Preset drum patterns
- `rockBeat()`, `shuffleBeat()`, `jazzSwing()`: Specific patterns

### 3. Display (`display/`)

#### `terminal.go`
**Responsibility:** Static display of track info

**Shows:**
- Track header with metadata
- Chord grid (4 chords per line)
- Bass info (style, swing)
- Drums info (style/pattern, intensity)

#### `live.go`
**Responsibility:** Real-time playback visualization

**Features:**
- Current chord (large, prominent)
- Visual metronome (◉ = beat 1, ● = current beat, ○ = inactive)
- Bar and beat counter
- Progress bar through progression

**Implementation:**
- Runs in goroutine, updates 10x/second
- Uses ANSI escape codes to update in-place
- Calculates position based on elapsed time and tempo
- `timePerBeat = 1 second / (tempo/60)`

**Key Type:**
- `LiveDisplay`: Manages real-time display state

**Methods:**
- `NewLiveDisplay(track)`: Constructor
- `Start()`: Launches goroutine
- `Stop()`: Stops display
- `render()`: Updates display (called by ticker)

### 4. Player (`player/fluidsynth.go`)

**Responsibility:** Execute FluidSynth for audio playback

**Key Functions:**
- `PlayMIDIWithDisplay()`: New version with live visual display
- `PlayMIDI()`: Legacy version without display
- `findSoundFont()`: Locates .sf2 file on system

**FluidSynth Options:**
- `-ni`: No interactive mode
- `-q`: Quiet mode (suppress FluidSynth output)
- `-r 48000`: Sample rate 48kHz
- `-g 1.0`: Gain level

**External Dependency:**
- Requires FluidSynth installed (`apt install fluidsynth fluid-soundfont-gm`)
- Uses `/usr/share/sounds/sf2/FluidR3_GM.sf2` (or similar)

## File Structure

```
backing-tracks/
├── main.go                      # CLI entry point
├── go.mod, go.sum               # Go modules
├── parser/
│   └── parser.go                # YAML → Go structs
├── midi/
│   ├── generator.go             # MIDI file generation
│   ├── bass.go                  # Bass pattern generator
│   └── drums.go                 # Drum pattern generator
├── display/
│   ├── terminal.go              # Static track info display
│   └── live.go                  # Real-time playback display
├── player/
│   └── fluidsynth.go            # FluidSynth integration
├── examples/                    # Example BTML files
│   ├── blues-a.btml             # Simple blues
│   ├── blues-full.btml          # Blues with bass & drums
│   ├── pop-progression.btml     # Simple pop
│   ├── pop-full.btml            # Pop with full band
│   ├── rock-euclidean.btml      # Euclidean drum patterns
│   ├── jazz-swing.btml          # Jazz with walking bass
│   └── little-wing.btml         # Ballad (Hendrix style)
├── README.md                    # User documentation
├── DSL_PROPOSAL.md              # Full BTML specification
└── CLAUDE.md                    # This file
```

## Development Guidelines

### Adding New Features

**1. New Bass Style:**
- Edit `midi/bass.go`
- Add case to switch in `GenerateBassLine()`
- Document in README.md

**2. New Drum Preset:**
- Edit `midi/drums.go`
- Add case to `generatePresetPattern()`
- Create helper function (e.g., `funkBeat()`)
- Document in README.md

**3. New Chord Type:**
- Edit `midi/generator.go`
- Add case to `getChordVoicing()`
- Update `parseQuality()` if needed
- Document in README.md

### Common Tasks

**Build:**
```bash
go build -o backing-tracks
```

**Test Playback:**
```bash
./backing-tracks play examples/blues-full.btml
```

**Create New Example:**
1. Create `.btml` file in `examples/`
2. Add to README.md example table
3. Test with `./backing-tracks play`

**Debug MIDI Output:**
```bash
# MIDI file is at /tmp/backing-track.mid
# Can inspect with:
timidity /tmp/backing-track.mid  # If timidity installed
# Or import into DAW
```

## Technical Decisions

### Why YAML?
- LLM-friendly (easy for AI to generate)
- Human-readable and editable
- Native Go support (`gopkg.in/yaml.v3`)
- Supports comments

### Why FluidSynth?
- High-quality wavetable synthesis
- Low CPU usage
- Available on Linux/Mac/Windows
- Good SoundFont ecosystem

### Why MIDI Generation (not direct audio)?
- Portable and inspectable
- Can import into DAWs
- Easy to modify/debug
- Standard format

### Why Euclidean Rhythms?
- Mathematically interesting patterns
- Common in electronic music
- Easy to parameterize
- Creates non-obvious grooves

### Display Update Rate (10Hz)
- Balance between responsiveness and CPU usage
- Smooth enough for visual feedback
- Doesn't overload terminal

## Known Limitations

1. **4/4 Time Only**: Currently hardcoded (easy to extend)
2. **Single MIDI Tempo**: Can't handle tempo changes mid-song
3. **No Dynamics**: All notes same velocity within voice
4. **Terminal-Only**: No GUI (intentional, but could add)
5. **Linux-Focused**: Tested primarily on Linux (should work on Mac)

## Future Enhancements (Roadmap)

- **v0.5**: ✅ Scale display for soloing, chord charts, melody generation
- **v0.6**: ✅ 16th note rhythms, Bubbletea TUI
- **v0.7**: Mini-notation parser (Strudel-inspired)
- **v0.8**: LLM integration for generating BTML from songs

## Dependencies

```go
require (
    gopkg.in/yaml.v3 v3.0.1           // YAML parsing
    gitlab.com/gomidi/midi/v2 v2.3.18 // MIDI file generation
)
```

**External:**
- FluidSynth (audio synthesis)
- SoundFont file (.sf2)

## Testing Strategy

**Manual Testing:**
- Play each example file
- Verify visual display accuracy
- Listen to audio quality
- Check tempo accuracy

**Areas for Future Automated Testing:**
- Parser validation
- Chord symbol parsing
- Euclidean rhythm algorithm
- MIDI note generation

## Troubleshooting

**"fluidsynth not found"**
```bash
sudo apt install fluidsynth fluid-soundfont-gm
```

**"No SoundFont found"**
```bash
sudo apt install fluid-soundfont-gm
# Or download FluidR3_GM.sf2 manually
```

**Playback too slow/fast**
- Adjust `tempo` in BTML file
- Note: Slow blues is intentionally 60-80 BPM

**Display not updating**
- Check terminal supports ANSI escape codes
- Try different terminal emulator

## Code Style

- **Error handling**: Return errors, don't panic
- **Constants**: Use constants for magic numbers (e.g., `ticksPerBar`)
- **Comments**: Document complex algorithms (e.g., Euclidean rhythm)
- **Naming**: Descriptive names (`GenerateBassLine` not `GenBass`)

## Contributing

When adding features:
1. Update parser structs if needed
2. Implement generation logic
3. Add example BTML file
4. Update README.md
5. Document in code comments
6. Test with various tempos and styles

## Resources

- **MIDI Spec**: https://www.midi.org/specifications
- **GM Drum Map**: Standard General MIDI percussion
- **Euclidean Rhythms**: Bjorklund's algorithm paper
- **FluidSynth**: https://www.fluidsynth.org/
- **Strudel**: https://strudel.cc/ (inspiration for mini-notation)

## Version History

- **v0.6**: 16th note rhythms, Bubbletea TUI with three-column layout
- **v0.5**: Scale display, chord charts, melody generation, Strudel export
- **v0.4**: Live visual display with current chord and beat
- **v0.3**: Drum patterns with Euclidean rhythms
- **v0.2**: Bass line generation
- **v0.1**: Basic chord progression playback

---

**Last Updated:** 2025-12-23
**Maintained By:** Human + Claude
**License:** MIT
