# Backing Tracks - v0.5

A terminal-based backing track player that uses the BTML (Backing Track Markup Language) DSL. Generate complete backing tracks with chords, bass, drums, and auto-generated melodies with real-time scale visualization!

## What Works

✅ Parse YAML-based BTML files
✅ Display track info and chord progressions in terminal
✅ Generate MIDI files from chord progressions
✅ **Bass line generation** (root, root_fifth, walking, swing_walking, stride, boogie)
✅ **Drum patterns** (rock_beat, shuffle, jazz_swing, kick_only)
✅ **Rhythm styles** (strumming, fingerpicking, travis, arpeggio, stride, ragtime)
✅ **Euclidean rhythms** for algorithmic drum patterns
✅ **Live visual display** - shows current chord, beat, scale fretboard, and chord charts!
✅ **Auto-generated melody** - improvisation based on scale and style
✅ **Guitar fretboard display** - shows scale positions for soloing
✅ **Chord chart display** - shows finger positions for current chord
✅ **Strudel export** - export to Strudel live coding format
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
# Play a backing track
./backing-tracks play examples/blues-full.btml

# Export to MIDI file
./backing-tracks export examples/blues-full.btml output.mid

# Export to Strudel (live coding)
./backing-tracks strudel examples/blues-full.btml output.strudel.js
```

### Live Display

During playback, you'll see:
- Current chord displayed prominently
- Visual metronome with beat indicators
- Strum pattern visualization
- **Guitar fretboard** showing the scale for improvisation
- **Chord diagrams** with finger positions
- Progress bar through the progression

```
  Slow Blues in A (Full Band)                       A | 80 BPM  │  Scale: A Blues
  ══════════════════════════════════════════════════════════════════

                A7                               A7                     │   A Blues
                                                                        │    0 1 2 3 4 5 6 7 8 9 101112
   .   ↓   .   ↓   .   ↓   .   ↑    .   ↓   .   ↓   .   ↓   .   ↑      │  e ● · · ● · ◆ · · ● · ● ● ●
   ●       ○       ○       ○        1       2       3       4          │  B · ● · ● ● ● · · ● · ◆ · ·
  ──────────────────────────────────────────────────────────────────    │  G ● · ◆ · · ● · ● ● ● · · ●
                A7                               A7                     │  D ● ● ● · · ● · ◆ · · ● · ●
                                                                        │  A ◆ · · ● · ● ● ● · · ● · ◆
                                                                        │  E ● · · ● · ◆ · · ● · ● ● ●
                                                                        │
                                                                        │   A7 [x02020]
                                                                        │   E  A  D  G  B  e

  ▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  20% (bar 5/24)
```

## BTML File Format

Create `.btml` files using simple YAML syntax:

```yaml
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

rhythm:
  style: shuffle_strum
  swing: 0.6

bass:
  style: swing_walking
  swing: 0.6

drums:
  style: shuffle
  intensity: 0.7

melody:
  enabled: true
  style: simple      # simple, moderate, active
  density: 0.5       # 0.0 to 1.0
```

## Supported Features

### Chord Progressions

**Basic Pattern:**
```yaml
chord_progression:
  pattern: "C G Am F"
  bars_per_chord: 2
  repeat: 4
```

**Inline Duration Notation:**
```yaml
chord_progression:
  pattern: "C*2 G*1 Am*2 F*1"  # C for 2 bars, G for 1 bar, etc.
```

### Chord Types
- **Major triads**: C, D, E, F, G, A, B
- **Minor triads**: Cm, Dm, Em, Am, etc.
- **Dominant 7th**: C7, D7, E7, A7, etc.
- **Major 7th**: Cmaj7, Dmaj7, Fmaj7, etc.
- **Minor 7th**: Cm7, Dm7, Em7, Am7, etc.
- **Suspended**: Csus4, Dsus2, etc.
- **Power chords**: C5, D5, E5, etc.

### Rhythm Styles

| Style | Description | Best For |
|-------|-------------|----------|
| `whole` | One strum per bar | Slow ballads |
| `half` | Two strums per bar | Ballads |
| `quarter` | Four strums per bar | Pop, rock |
| `eighth` | Eight strums per bar | Rock, punk |
| `strum_down` | Arpeggiated downstrum | Folk |
| `strum_up_down` | Alternating strums | Pop, rock |
| `folk` | Bass note + chord pattern | Folk, country |
| `shuffle_strum` | Triplet shuffle | Blues |
| `travis` | Travis picking pattern | Country, folk |
| `fingerpick` | Folk fingerpicking | Singer-songwriter |
| `fingerpick_slow` | Sparse fingerpicking | Ballads |
| `arpeggio_up` | Ascending arpeggio | Classical, ambient |
| `arpeggio_down` | Descending arpeggio | Classical, ballads |
| `stride` | Chords on 2 & 4 | Ragtime, stride piano |
| `ragtime` | Stride with syncopation | Ragtime |

### Bass Styles

| Style | Description | Best For |
|-------|-------------|----------|
| `root` | Root notes on downbeats | Simple accompaniment |
| `root_fifth` | Root on 1, fifth on 3 | Folk, country, rock |
| `walking` | Root, 3rd, 5th, 7th pattern | Jazz |
| `swing_walking` | Swung walking bass | Jazz, blues |
| `stride` | Low bass on 1 & 3 | Ragtime, stride piano |
| `boogie` | Driving eighth note pattern | Boogie-woogie, rock & roll |

### Drum Patterns

**Preset Styles:**
| Style | Description |
|-------|-------------|
| `rock_beat` | Kick 1,3 / Snare 2,4 / 8th hihat |
| `shuffle` | Blues shuffle with triplet feel |
| `jazz_swing` | Swinging ride with sparse kick/snare |
| `kick_only` | Just kick drum (for stripped-down tracks) |

**Euclidean Rhythms:**
```yaml
drums:
  kick:
    euclidean: { hits: 5, steps: 8, rotation: 0 }
  snare:
    euclidean: { hits: 3, steps: 8, rotation: 2 }
  hihat:
    euclidean: { hits: 7, steps: 8, rotation: 0 }
```

### Melody Generation

Auto-generate an improvisation track:

```yaml
melody:
  enabled: true
  style: simple      # simple (half notes), moderate (quarters), active (eighths)
  density: 0.5       # 0.0 to 1.0 - how many notes to play
```

The melody uses scale-appropriate notes based on the track style:
- **Blues** → Blues scale
- **Jazz** → Dorian/Mixolydian modes
- **Rock** → Pentatonic minor
- **Folk/Pop** → Natural major/minor

### Scale Override

Force a specific scale instead of auto-detection:

```yaml
scale:
  type: blues  # pentatonic_minor, pentatonic_major, blues, dorian, mixolydian, natural_minor, natural_major
```

## Example Tracks

The `examples/` directory contains many demo tracks:

### Blues Styles
| File | Description |
|------|-------------|
| `blues-full.btml` | Full band blues with swing walking bass |
| `blues-delta.btml` | Raw, sparse Delta blues (72 BPM) |
| `blues-chicago.btml` | Electric Chicago shuffle (116 BPM) |
| `blues-texas.btml` | Clean, jazzy Texas blues (126 BPM) |
| `blues-jump.btml` | Uptempo jump blues (168 BPM) |
| `blues-slow.btml` | Soulful slow blues (58 BPM) |
| `blues-west-coast.btml` | Jazzy West Coast blues (96 BPM) |
| `blues-boogie.btml` | John Lee Hooker style one-chord boogie |

### Fingerpicking
| File | Description |
|------|-------------|
| `fingerpick-travis.btml` | Travis picking in G (country/folk) |
| `fingerpick-folk.btml` | Classic folk fingerpicking |
| `fingerpick-ballad.btml` | Slow ballad fingerpicking |
| `fingerpick-arpeggio.btml` | Classical ascending arpeggios |
| `fingerpick-spanish.btml` | Spanish romance style |
| `fingerpick-songwriter.btml` | Singer-songwriter style |

### Ragtime & Boogie
| File | Description |
|------|-------------|
| `ragtime.btml` | Classic stride piano ragtime |
| `boogie-woogie.btml` | Driving boogie-woogie piano |
| `rock-n-roll-piano.btml` | 50s rock & roll (Jerry Lee Lewis style) |

### Other Styles
| File | Description |
|------|-------------|
| `pop-full.btml` | Pop with bass & drums |
| `rock-euclidean.btml` | Rock with Euclidean drum patterns |
| `jazz-swing.btml` | Jazz II-V-I with walking bass |
| `little-wing.btml` | Ballad in Em |

Try them:
```bash
./backing-tracks play examples/blues-chicago.btml
./backing-tracks play examples/fingerpick-travis.btml
./backing-tracks play examples/ragtime.btml
./backing-tracks play examples/rock-n-roll-piano.btml
```

## Project Structure

```
backing-tracks/
├── main.go              # CLI entry point
├── parser/
│   └── parser.go        # BTML YAML parser
├── midi/
│   ├── generator.go     # MIDI file generation
│   ├── bass.go          # Bass pattern generator
│   ├── drums.go         # Drum pattern generator
│   ├── rhythm.go        # Chord rhythm patterns
│   └── melody.go        # Melody generation
├── player/
│   └── fluidsynth.go    # FluidSynth integration
├── display/
│   ├── terminal.go      # Terminal display formatting
│   ├── live.go          # Real-time playback display
│   ├── fretboard.go     # Guitar fretboard visualization
│   └── chords.go        # Chord diagram display
├── theory/
│   └── theory.go        # Music theory (scales, keys)
├── strudel/
│   └── generator.go     # Strudel export
├── examples/            # Example BTML files
└── README.md
```

## Roadmap

- **v0.1:** ✅ Basic chord progression playback
- **v0.2:** ✅ Bass line generation
- **v0.3:** ✅ Drum patterns (presets + Euclidean rhythms)
- **v0.4:** ✅ Live visual display with chord and beat tracking
- **v0.5:** ✅ Scale display, chord charts, melody generation, Strudel export
- **v0.6:** Mini-notation parser (Strudel-inspired)
- **v0.7:** Interactive TUI with Bubbletea
- **v0.8:** LLM integration for generating BTML from songs

## Dependencies

- `gopkg.in/yaml.v3` - YAML parsing
- `gitlab.com/gomidi/midi/v2` - MIDI file generation
- FluidSynth (external) - Audio synthesis

## Troubleshooting

### "fluidsynth not found"
Install FluidSynth: `sudo apt install fluidsynth fluid-soundfont-gm`

### "no SoundFont (.sf2) file found"
Install a SoundFont package: `sudo apt install fluid-soundfont-gm`

### No audio output
Check your system audio settings and ensure FluidSynth can access your audio device.

## License

MIT

## Contributing

Contributions welcome! See `BTML_MANUAL.md` for the full BTML specification and `CLAUDE.md` for development guidelines.

## What's New

**v0.5 (Current)**
- ✅ **Guitar fretboard display** showing scale positions for improvisation
- ✅ **Chord chart display** with finger positions (open + barre voicings)
- ✅ **Auto-generated melody track** based on scale and chord tones
- ✅ **Smart scale detection** based on track style (blues, jazz, rock, etc.)
- ✅ **Strudel export** for live coding
- ✅ **New rhythm styles**: stride, ragtime, fingerpicking variants
- ✅ **New bass styles**: stride, boogie
- ✅ **Many new examples**: blues styles, fingerpicking, ragtime, boogie-woogie

**v0.4**
- ✅ Live visual display during playback
- ✅ Inline duration notation for chords
- ✅ Karaoke-style scrolling display with lyrics support

**v0.3**
- ✅ Bass line generation (4 styles)
- ✅ Drum patterns (3 presets + Euclidean)
- ✅ Swing feel for bass

**v0.2**
- ✅ Bass line generation (initial)

**v0.1**
- ✅ Chord progression playback
- ✅ FluidSynth integration
