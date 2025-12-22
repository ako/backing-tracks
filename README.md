# Backing Tracks - v0.4

A terminal-based backing track player that uses the BTML (Backing Track Markup Language) DSL. Generate complete backing tracks with chords, bass, and drums!

## What Works

✅ Parse YAML-based BTML files
✅ Display track info and chord progressions in terminal
✅ Generate MIDI files from chord progressions
✅ **Bass line generation** (root, root_fifth, walking, swing_walking)
✅ **Drum patterns** (rock_beat, shuffle, jazz_swing)
✅ **Euclidean rhythms** for algorithmic drum patterns
✅ **Live visual display** - shows current chord and beat in real-time!
✅ Play backing tracks via FluidSynth (when installed)

## Installation

### 1. Build the Application

```bash
go build -o backing-tracks
```

### 2. Install FluidSynth and SoundFont

**Ubuntu/Debian:**
```bash
sudo apt install fluidsynth fluid-soundfont-gm
```

**Arch Linux:**
```bash
sudo pacman -S fluidsynth soundfont-fluid
```

**macOS:**
```bash
brew install fluid-synth
```

## Usage

```bash
./backing-tracks play examples/blues-a.btml
```

### Example Output

```
┌─ Slow Blues in A (Full Band) ──────────┐
│ Key: A | Tempo: 80 BPM | 4/4 | blues │
└──────────────────────────────────────┘

Chord Progression (24 bars, 2x):
  A7 | A7 | A7 | A7
  D7 | D7 | A7 | A7
  E7 | D7 | A7 | E7
  A7 | A7 | A7 | A7
  D7 | D7 | A7 | A7
  E7 | D7 | A7 | E7

♪ Bass: swing_walking (swing 60%)
♫ Drums: shuffle (70% intensity)

♪ Playing... (Press Ctrl+C to stop)

┌─────────────────────────────────────┐
│  Current Chord: A7                  │
└─────────────────────────────────────┘
Beat: ○ ○ ● ○  [Bar 1, Beat 3]
Progress: [=>                  ] 1/24
```

**Live Display Features:**
- Current chord displayed prominently
- Visual metronome (◉ highlights beat 1, ● shows current beat)
- Bar and beat numbers
- Progress bar through the entire progression

## BTML File Format

Create `.btml` files using simple YAML syntax:

```yaml
# blues-full.btml
track:
  title: "Slow Blues in A (Full Band)"
  key: A
  tempo: 80
  time_signature: 4/4
  style: blues

chord_progression:
  pattern: "A7 A7 A7 A7 D7 D7 A7 A7 E7 D7 A7 E7"
  bars_per_chord: 1
  repeat: 2

bass:
  style: swing_walking  # Options: root, root_fifth, walking, swing_walking
  swing: 0.6           # 0.5 = straight, 0.67 = triplet swing

drums:
  style: shuffle       # Options: rock_beat, shuffle, jazz_swing
  intensity: 0.7       # 0.0 to 1.0
```

## Supported Features

### Chord Progressions

**Basic Pattern:**
```yaml
chord_progression:
  pattern: "C G Am F"
  bars_per_chord: 2  # Each chord lasts 2 bars
  repeat: 4
```

**Inline Duration Notation:**
Mix different chord durations using `*` notation:
```yaml
chord_progression:
  # C for 2 bars, G for 1 bar, Am for 2 bars, F for 1 bar
  pattern: "C*2 G*1 Am*2 F*1"
  bars_per_chord: 1  # Default for chords without explicit duration
  repeat: 2
```

**Examples:**
- `Em*1` - Em chord for 1 bar
- `G*2` - G chord for 2 bars
- `Am` - Uses default `bars_per_chord` value
- `"C*2 G*1 Am F*2"` - Mixed durations in one progression

See `examples/mixed-durations.btml` for a complete example!

### Chord Types
- **Major triads**: C, D, E, etc.
- **Minor triads**: Cm, Dm, Em, etc.
- **Dominant 7th**: C7, D7, E7, etc.
- **Major 7th**: Cmaj7, Dmaj7 (or C^7)
- **Minor 7th**: Cm7, Dm7, etc.
- **Power chords**: C5, D5, E5, etc.

### Bass Styles
- **root**: Simple root notes on downbeats
- **root_fifth**: Root on 1, fifth on 3
- **walking**: Walking bass pattern (root, 3rd, 5th, 7th)
- **swing_walking**: Swung walking bass for jazz/blues

### Drum Patterns
**Preset Styles:**
- **rock_beat**: Standard rock pattern (kick 1,3 | snare 2,4 | 8th note hihat)
- **shuffle**: Blues shuffle with triplet feel
- **jazz_swing**: Swinging jazz ride pattern with sparse kick/snare

**Euclidean Rhythms:**
Create algorithmic patterns using mathematical distribution:

```yaml
drums:
  kick:
    euclidean:
      hits: 5      # Number of hits
      steps: 8     # Total steps in pattern
      rotation: 0  # Offset rotation
  snare:
    euclidean:
      hits: 3
      steps: 8
      rotation: 2
  hihat:
    euclidean:
      hits: 7
      steps: 8
      rotation: 0
```

See `examples/rock-euclidean.btml` for a full example!

## Project Structure

```
backing-tracks/
├── main.go              # CLI entry point
├── parser/
│   └── parser.go        # BTML YAML parser (with bass/drums)
├── midi/
│   ├── generator.go     # MIDI file generation
│   ├── bass.go          # Bass pattern generator
│   └── drums.go         # Drum pattern generator (with Euclidean rhythms)
├── player/
│   └── fluidsynth.go    # FluidSynth integration
├── display/
│   └── terminal.go      # Terminal display formatting
├── examples/
│   ├── blues-a.btml         # Simple blues (chords only)
│   ├── blues-full.btml      # Full blues with bass & drums
│   ├── pop-progression.btml # Simple pop progression
│   ├── pop-full.btml        # Pop with bass & drums
│   ├── rock-euclidean.btml  # Rock with Euclidean drums
│   ├── jazz-swing.btml      # Jazz with walking bass
│   └── little-wing.btml     # Ballad (Hendrix style)
└── README.md                # This file
```

## How It Works

1. **Parse**: Read BTML file and parse YAML into Go structs (chords, bass, drums)
2. **Display**: Format and show track info in terminal
3. **Generate**:
   - Convert chord symbols to MIDI note events
   - Generate bass patterns (walking, root, etc.)
   - Generate drum patterns (preset or Euclidean rhythms)
4. **Play**: Execute FluidSynth to synthesize audio from MIDI

## Roadmap

**v0.1:** ✅ Basic chord progression playback
**v0.2:** ✅ Bass line generation (root, root_fifth, walking, swing_walking)
**v0.3:** ✅ Drum patterns (presets + Euclidean rhythms)
**v0.4:** ✅ Live visual display with current chord and beat tracking
**v0.5 (Next):** Scale display for soloing
**v0.6:** Mini-notation parser (Strudel-inspired)
**v0.7:** Interactive TUI with Bubbletea
**v0.8:** LLM integration for generating BTML from songs

## Dependencies

- `gopkg.in/yaml.v3` - YAML parsing
- `gitlab.com/gomidi/midi/v2` - MIDI file generation
- FluidSynth (external) - Audio synthesis

## Testing Without FluidSynth

The MIDI file is generated at `/tmp/backing-track.mid` regardless of whether FluidSynth is installed. You can:

1. Play it with any MIDI player
2. Import into a DAW (Ableton, Logic, FL Studio, etc.)
3. Use online MIDI players

## Example Tracks

The `examples/` directory contains several demo tracks:

| File | Description | Features |
|------|-------------|----------|
| `blues-a.btml` | Simple 12-bar blues | Chords only |
| `blues-full.btml` | Full band blues | Swing walking bass + shuffle drums |
| `pop-progression.btml` | I-V-vi-IV progression | Chords only |
| `pop-full.btml` | Pop with full band | Root bass + rock beat drums |
| `rock-euclidean.btml` | Rock jam | Root/fifth bass + Euclidean drum patterns |
| `jazz-swing.btml` | II-V-I in Bb | Walking bass + jazz swing drums |
| `little-wing.btml` | Ballad in Em (Hendrix style) | Root/fifth bass + gentle drums, 1 bar per chord |
| `mixed-durations.btml` | Mixed chord lengths | Demonstrates inline duration notation (C*2, G*1, etc.) |

Try them all:
```bash
./backing-tracks play examples/blues-full.btml
./backing-tracks play examples/rock-euclidean.btml
./backing-tracks play examples/jazz-swing.btml
./backing-tracks play examples/little-wing.btml
```

## Creating Your Own Backing Tracks

1. Create a new `.btml` file:

```yaml
track:
  title: "My Rock Track"
  key: C
  tempo: 120
  time_signature: 4/4
  style: rock

chord_progression:
  pattern: "C G Am F"
  bars_per_chord: 2
  repeat: 4

bass:
  style: root_fifth  # Try: root, root_fifth, walking, swing_walking

drums:
  style: rock_beat   # Try: rock_beat, shuffle, jazz_swing
  intensity: 0.8
```

2. Play it:

```bash
./backing-tracks play my-track.btml
```

3. Experiment with Euclidean rhythms:

```yaml
drums:
  kick:
    euclidean: { hits: 5, steps: 8, rotation: 0 }
  snare:
    euclidean: { hits: 3, steps: 8, rotation: 2 }
  hihat:
    euclidean: { hits: 7, steps: 8, rotation: 0 }
  intensity: 0.85
```

## Troubleshooting

### "fluidsynth not found"
Install FluidSynth: `sudo apt install fluidsynth fluid-soundfont-gm`

### "no SoundFont (.sf2) file found"
Install a SoundFont package: `sudo apt install fluid-soundfont-gm`

### No audio output
Check your system audio settings and ensure FluidSynth can access your audio device.

## License

[To be determined - suggest MIT]

## Contributing

Contributions welcome for:
- Additional chord voicings and extensions
- More bass patterns and styles
- More drum presets and patterns
- Scale display (v0.4)
- Mini-notation parser (v0.5)
- Interactive TUI (v0.6)
- UI improvements

See `DSL_PROPOSAL.md` for the full design specification.

## What's New

**v0.4 (Current)**
- ✅ **Live visual display during playback**
  - Current chord shown prominently
  - Visual metronome with beat indicators (◉ for beat 1, ● for other beats)
  - Bar and beat counter
  - Progress bar through the progression
  - Updates in real-time as you play!
- ✅ **Inline duration notation** for chord progressions
  - Mix different chord lengths: `"C*2 G*1 Am*2 F*1"`
  - Precise control over chord timing
  - See `examples/mixed-durations.btml`
- ✅ MIDI generation debugging output
- ✅ Cleaner FluidSynth output (quiet mode)

**v0.3**
- ✅ Bass line generation with 4 styles (root, root_fifth, walking, swing_walking)
- ✅ Drum pattern generation with 3 presets (rock_beat, shuffle, jazz_swing)
- ✅ Euclidean rhythm support for algorithmic drum patterns
- ✅ Swing feel for bass patterns
- ✅ Configurable drum intensity
- ✅ 7 example tracks demonstrating all features (blues, pop, rock, jazz, ballad)

**v0.2**
- ✅ Bass line generation (initial implementation)

**v0.1**
- ✅ Chord progression playback
- ✅ YAML-based DSL parsing
- ✅ FluidSynth integration
