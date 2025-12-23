# BTML - Backing Track Markup Language

## User & LLM Manual

BTML is a YAML-based notation for defining complete backing tracks with chords, rhythm patterns, bass lines, drums, and melody. This manual serves as both user documentation and LLM instructions for generating backing tracks.

---

## Quick Start

```yaml
track:
  title: "My Song"
  key: G
  tempo: 120
  time_signature: 4/4
  style: rock

chord_progression:
  pattern: "G D Em C"
  bars_per_chord: 1
  repeat: 4

rhythm:
  style: eighth

bass:
  style: root_fifth

drums:
  style: rock_beat
  intensity: 0.7

melody:
  enabled: true
  style: moderate
  density: 0.5

scale:
  type: pentatonic_minor
```

---

## Structure Overview

A BTML file has these sections:

| Section | Required | Description |
|---------|----------|-------------|
| `track` | Yes | Metadata (title, key, tempo, style) |
| `chord_progression` | Yes | The chord sequence |
| `rhythm` | No | How chords are played (strum/pick pattern) |
| `bass` | No | Bass line style |
| `drums` | No | Drum pattern |
| `melody` | No | Auto-generated melody line |
| `scale` | No | Scale override for display/melody |

---

## Track Metadata

```yaml
track:
  title: "Song Name"        # Display name
  key: C                    # Musical key (C, G, Am, F#, Bb, etc.)
  tempo: 120                # BPM (beats per minute)
  time_signature: 4/4       # Currently only 4/4 supported
  style: rock               # Genre hint (rock, blues, jazz, folk, pop, ballad, funk, edm)
```

### Common Tempos by Genre
| Genre | Typical BPM |
|-------|-------------|
| Slow ballad | 50-70 |
| Blues | 60-90 |
| Folk | 80-110 |
| Funk | 90-110 |
| Pop | 100-130 |
| Rock | 110-140 |
| EDM | 120-140 |
| Fast rock | 140-180 |

---

## Chord Progression

```yaml
chord_progression:
  pattern: "C G Am F"       # Space-separated chord symbols
  bars_per_chord: 1         # Default duration per chord
  repeat: 2                 # How many times to repeat
```

### Chord Symbols

| Symbol | Type | Example |
|--------|------|---------|
| `C`, `G`, `D` | Major triad | C major |
| `Am`, `Em`, `Dm` | Minor triad | A minor |
| `A7`, `E7`, `D7` | Dominant 7th | A dominant 7 |
| `Cmaj7`, `Gmaj7` | Major 7th | C major 7 |
| `Am7`, `Em7` | Minor 7th | A minor 7 |
| `E9`, `A9` | Dominant 9th | E dominant 9 |
| `E5`, `A5`, `G5` | Power chord | E power chord |
| `Asus4`, `Dsus4` | Suspended 4th | A sus 4 |
| `Asus2`, `Dsus2` | Suspended 2nd | A sus 2 |
| `E7sus4` | Dominant 7 sus 4 | E7 suspended |
| `Bb`, `F#`, `Eb` | Accidentals | B flat major |

### Slash Chords (Bass Note)

Specify a different bass note using slash notation:

```yaml
pattern: "Am Am/G Am/F Am/E"    # Descending bass line
pattern: "C/E F G C"            # C with E in bass
```

The note after `/` becomes the bass note while the chord voicing stays the same.

### Inline Duration Notation

Override bar duration for individual chords:

```yaml
pattern: "Em G Am Em Bm*0.5 Bb*0.5 Am*0.5 C*0.5"
```

| Notation | Meaning |
|----------|---------|
| `C` | Uses default `bars_per_chord` |
| `C*2` | 2 bars |
| `C*1` | 1 bar |
| `C*0.5` | Half bar (2 beats) |

### Common Progressions

```yaml
# 12-Bar Blues
pattern: "A7 A7 A7 A7 D7 D7 A7 A7 E7 D7 A7 E7"

# Pop (I-V-vi-IV)
pattern: "C G Am F"

# Jazz II-V-I
pattern: "Dm7 G7 Cmaj7 Cmaj7"

# Folk
pattern: "G C G D"

# Minor Ballad
pattern: "Am G F E"

# Funk Vamp
pattern: "E9 E9 E9 E9 A9 E9 E9 E9"

# Descending Bass (Stairway/Babe I'm Gonna Leave You)
pattern: "Am Am/G Am/F Am/E"
```

---

## Rhythm Section

The rhythm section defines HOW chords are played.

### Preset Styles

```yaml
rhythm:
  style: quarter            # Use a preset style
  swing: 0.55               # Optional swing feel (0.5 = straight)
  accent: "1,3"             # Optional beat accents
```

| Style | Description | Best For |
|-------|-------------|----------|
| `whole` | Sustained chords | Slow ballads |
| `half` | Two strums per bar | Ballads |
| `quarter` | Four strums per bar | Folk, pop |
| `eighth` | Eight strums per bar | Rock, pop |
| `sixteenth` | Straight 16th note strumming | Funk, disco |
| `funk_16th` | Funky 16ths with ghost notes | Funk, R&B |
| `strum_down` | Arpeggiated downstrums | Acoustic rock |
| `strum_up_down` | Alternating up/down | Pop, folk |
| `folk` | Bass on 1,3 + chord on 2,4 | Country, folk |
| `shuffle_strum` | Triplet shuffle | Blues, swing |
| `stride` | Stride piano pattern | Jazz, ragtime |
| `ragtime` | Classic ragtime | Ragtime |
| `travis` | Travis picking (alternating bass) | Fingerstyle, country |
| `fingerpick` | 16th note fingerpicking | Folk, classical |
| `fingerpick_slow` | Sparse picking | Ballads, Leonard Cohen |
| `arpeggio_up` | Ascending arpeggio | Ambient, new wave |
| `arpeggio_down` | Descending arpeggio | Ballads, post-punk |
| `funk` | Syncopated 16th notes (heavy on the one) | Funk, R&B |
| `funk_muted` | Heavily muted/choppy funk | Funk rock |

### Custom Strum Patterns

Define exact strum patterns using notation:

```yaml
rhythm:
  pattern: "D.DU.UDU"       # Custom pattern
  swing: 0.6                # Optional swing
```

#### Pattern Notation

| Character | Meaning | Velocity |
|-----------|---------|----------|
| `D` | Down strum (loud) | 85 |
| `d` | Down strum (soft) | 65 |
| `U` | Up strum (loud) | 75 |
| `u` | Up strum (soft) | 55 |
| `x` | Muted/ghost strum | 50 |
| `.` | Rest (silence) | - |
| `-` | Hold/tie | - |

#### Pattern Length = Subdivision

| Length | Subdivision |
|--------|-------------|
| 4 chars | Quarter notes |
| 8 chars | Eighth notes |
| 16 chars | Sixteenth notes |

#### Common Patterns

```yaml
# Pop/Rock (8th notes)
pattern: "D.DU.UDU"

# Folk (8th notes)
pattern: "D.dUD.dU"

# Reggae off-beat
pattern: ".xD.xDxU"

# Blues shuffle
pattern: ".D.D.D.U"

# Heavy rock with mutes
pattern: "DxDxDxDx"

# Ballad arpeggios (16th notes)
pattern: "D.u.d.u.D.u.d.u."
```

### Swing Feel

```yaml
rhythm:
  style: shuffle_strum
  swing: 0.67               # Triplet swing (67/33)
```

| Value | Feel |
|-------|------|
| 0.5 | Straight (no swing) |
| 0.55 | Slight swing |
| 0.6 | Moderate swing |
| 0.67 | Triplet swing (blues/jazz) |

---

## Bass Section

```yaml
bass:
  style: walking            # Bass style
  swing: 0.6                # Optional swing
```

### Bass Styles

| Style | Description | Best For |
|-------|-------------|----------|
| `root` | Root notes only | Pop, rock, ballads |
| `root_fifth` | Root on 1, fifth on 3 | Folk, country, rock |
| `walking` | Root-3rd-5th-7th pattern | Jazz, blues |
| `swing_walking` | Walking with swing feel | Blues, jazz |
| `stride` | Stride piano bass (octave jumps) | Jazz, ragtime |
| `boogie` | Boogie-woogie pattern | Blues rock, boogie |
| `808` / `sub` | Heavy sustained sub bass | EDM, trap, hip-hop |
| `808_octave` / `edm` | Sub bass with octave jumps | EDM, house |
| `funk` / `slap` | Syncopated slap bass | Funk, R&B |
| `funk_simple` | Simpler funk bass | Funk soul |

---

## Drums Section

### Preset Styles

```yaml
drums:
  style: rock_beat
  intensity: 0.8            # 0.0 to 1.0
```

| Style | Description |
|-------|-------------|
| `rock_beat` | Kick 1,3 / Snare 2,4 / 8th hi-hat |
| `shuffle` | Blues shuffle with triplet feel |
| `blues_shuffle` | Driving blues with ghost notes & open hi-hat |
| `jazz_swing` | Ride pattern with sparse kick/snare |
| `four_on_floor` / `edm` | Four kicks per bar with 16th hi-hats |
| `trap` | Trap-style with rolling hi-hats and 808 kick |
| `funk` | Tight funk groove |
| `kick_only` | Minimal kick drum only |

### Custom Drum Patterns

```yaml
drums:
  kick:
    beats: [1, 3]           # Explicit beat positions
  snare:
    beats: [2, 4]
  hihat:
    euclidean:              # Euclidean rhythm
      hits: 8
      steps: 8
      rotation: 0
  ride:
    euclidean:
      hits: 3
      steps: 8
      rotation: 2
  intensity: 0.7
```

### Euclidean Rhythms

Distributes N hits evenly across M steps:

```yaml
euclidean:
  hits: 5                   # Number of hits
  steps: 8                  # Total steps per bar
  rotation: 0               # Offset rotation
```

| Pattern | Result |
|---------|--------|
| (3, 8, 0) | `x..x..x.` - Tresillo |
| (5, 8, 0) | `x.x.x.xx` - Cinquillo |
| (7, 8, 0) | `x.xxxxxx` - Almost all |
| (4, 12, 0) | `x..x..x..x..` - Triplet feel |

---

## Melody Section

Auto-generate a melody line that follows the chord progression:

```yaml
melody:
  enabled: true             # Turn on melody generation
  style: moderate           # Complexity level
  density: 0.5              # 0.0-1.0, how sparse/dense
  octave: 4                 # Base octave (default 4)
```

### Melody Styles

| Style | Description | Best For |
|-------|-------------|----------|
| `simple` | Half/whole notes, chord tones | Ballads, learning |
| `moderate` | Quarter notes, passing tones | Pop, rock |
| `active` | Eighth notes, more motion | Jazz, funk |
| `blues_head` | Classic AAB 12-bar blues vocal | Blues |
| `call_response` | Same as blues_head | Blues |

---

## Scale Section

Override the auto-detected scale for fretboard display and melody generation:

```yaml
scale:
  type: pentatonic_minor    # Scale type
```

### Scale Types

| Type | Notes | Best For |
|------|-------|----------|
| `pentatonic_minor` | 1-b3-4-5-b7 | Blues, rock solos |
| `pentatonic_major` | 1-2-3-5-6 | Country, pop |
| `blues` | 1-b3-4-#4-5-b7 | Blues |
| `natural_minor` | 1-2-b3-4-5-b6-b7 | Minor keys |
| `natural_major` | 1-2-3-4-5-6-7 | Major keys |
| `dorian` | 1-2-b3-4-5-6-b7 | Jazz, funk |
| `mixolydian` | 1-2-3-4-5-6-b7 | Dominant chords, rock |

---

## Complete Examples

### Blues in A

```yaml
track:
  title: "Slow Blues in A"
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
  swing: 0.67

bass:
  style: walking
  swing: 0.67

drums:
  style: blues_shuffle
  intensity: 0.7

melody:
  enabled: true
  style: blues_head
  density: 0.6

scale:
  type: blues
```

### 70s Funk

```yaml
track:
  title: "70s Funk in E"
  key: E
  tempo: 100
  time_signature: 4/4
  style: funk

chord_progression:
  pattern: "E9 E9 E9 E9 E9 E9 A9 E9"
  bars_per_chord: 1
  repeat: 4

rhythm:
  style: funk
  accent: "1"

bass:
  style: slap

drums:
  style: funk
  intensity: 0.85

melody:
  enabled: true
  style: active
  density: 0.6

scale:
  type: mixolydian
```

### Folk Fingerpicking

```yaml
track:
  title: "Folk Ballad"
  key: C
  tempo: 72
  time_signature: 4/4
  style: folk

chord_progression:
  pattern: "C Am F G"
  bars_per_chord: 2
  repeat: 4

rhythm:
  style: fingerpick_slow

bass:
  style: root

drums:
  kick:
    beats: [1]
  intensity: 0.2
```

### EDM/808

```yaml
track:
  title: "EDM Drop"
  key: Am
  tempo: 128
  time_signature: 4/4
  style: edm

chord_progression:
  pattern: "Am F C G"
  bars_per_chord: 2
  repeat: 4

rhythm:
  style: sixteenth

bass:
  style: 808

drums:
  style: four_on_floor
  intensity: 0.9

melody:
  enabled: true
  style: active
  density: 0.7
```

### Descending Bass Ballad

```yaml
track:
  title: "Folk Rock Ballad"
  key: Am
  tempo: 134
  time_signature: 4/4
  style: folk

chord_progression:
  pattern: "Am Am/G Am/F Am/E Am Am/G D E"
  bars_per_chord: 1
  repeat: 4

rhythm:
  style: arpeggio_down

bass:
  style: root

drums:
  style: rock_beat
  intensity: 0.6
```

---

## LLM Generation Guidelines

When generating BTML files, follow these guidelines:

### 1. Match Style to Genre

| Genre | Rhythm | Bass | Drums | Tempo |
|-------|--------|------|-------|-------|
| Blues | shuffle_strum, swing 0.67 | walking/swing_walking | blues_shuffle/shuffle | 60-90 |
| Jazz | custom pattern, swing 0.67 | walking/stride | jazz_swing | 120-180 |
| Folk | fingerpick/travis/folk | root/root_fifth | minimal or none | 80-110 |
| Pop | eighth/strum_up_down | root | rock_beat | 100-130 |
| Rock | eighth or DxDx patterns | root_fifth | rock_beat | 110-150 |
| Ballad | fingerpick_slow/arpeggio | root | minimal | 50-80 |
| Funk | funk/funk_muted | funk/slap | funk | 90-110 |
| EDM | sixteenth | 808/edm | four_on_floor/trap | 120-140 |

### 2. Use Appropriate Chord Types

| Genre | Typical Chords |
|-------|----------------|
| Blues | Dominant 7ths (A7, D7, E7) |
| Jazz | 7th chords (Cm7, F7, Bbmaj7) |
| Folk | Major/minor triads |
| Rock | Power chords (E5, A5) or triads |
| Pop | Major/minor triads, some 7ths |
| Funk | 9th chords (E9, A9), 7ths |
| EDM | Minor triads, simple progressions |

### 3. Set Intensity Appropriately

| Feel | Intensity |
|------|-----------|
| Gentle/sparse | 0.2 - 0.4 |
| Moderate | 0.5 - 0.7 |
| Driving | 0.7 - 0.85 |
| Aggressive | 0.85 - 1.0 |

### 4. Common Song Structures

**Verse-Chorus (repeat 4x):**
```yaml
repeat: 4
# 4 chords x 4 repeats = 16 bars
```

**12-Bar Blues (repeat 2x):**
```yaml
pattern: "A7 A7 A7 A7 D7 D7 A7 A7 E7 D7 A7 E7"
repeat: 2
# 12 bars x 2 = 24 bars
```

**32-Bar Form:**
```yaml
pattern: "..." # 8 chords
bars_per_chord: 2
repeat: 2
# 8 x 2 x 2 = 32 bars
```

### 5. Creating Specific Feels

**Emotional Ballad:**
```yaml
tempo: 60
rhythm:
  style: fingerpick_slow
bass:
  style: root
drums:
  intensity: 0.25
```

**Driving Rock:**
```yaml
tempo: 130
rhythm:
  pattern: "DxDxDxDx"
bass:
  style: root_fifth
drums:
  intensity: 0.9
```

**Funk Groove:**
```yaml
tempo: 100
rhythm:
  style: funk
  accent: "1"
bass:
  style: slap
drums:
  style: funk
  intensity: 0.85
```

**80s New Wave:**
```yaml
tempo: 126
rhythm:
  style: arpeggio_up
bass:
  style: root_fifth
drums:
  style: rock_beat
  intensity: 0.7
```

---

## Command Reference

```bash
# Play a backing track
./backing-tracks play examples/blues-a.btml

# Play with custom SoundFont
./backing-tracks play --soundfont ~/soundfonts/SGM.sf2 examples/blues-a.btml

# Export to MIDI file
./backing-tracks export examples/blues-a.btml

# Export with custom output path
./backing-tracks export examples/blues-a.btml my-track.mid

# Export to Strudel code
./backing-tracks strudel examples/blues-a.btml

# List available SoundFonts
./backing-tracks soundfonts
```

### Environment Variables

```bash
# Set default SoundFont
export SOUNDFONT=~/soundfonts/FluidR3_GM.sf2
```

### SoundFont Tips

- Place `.sf2` files in `./soundfonts/` directory for auto-detection
- Set `SOUNDFONT` env var for a permanent default
- Use `--soundfont` flag to override for specific playback

---

## Version

BTML v0.6 - Backing Tracks Player

Features:
- Chord progressions with fractional bar notation and slash chords
- Custom strum patterns (D/U/x/. notation)
- 16th note rhythm styles (sixteenth, funk_16th, funk_muted)
- Fingerpicking and funk rhythm styles
- Multiple bass styles (root, walking, funk, 808)
- Drum presets and Euclidean rhythms
- Auto-generated melody with multiple styles
- Scale display with fretboard visualization
- Bubbletea TUI with three-column layout
- Custom SoundFont support
- MIDI and Strudel export
