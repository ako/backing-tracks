# BTML - Backing Track Markup Language

## User & LLM Manual

BTML is a YAML-based notation for defining complete backing tracks with chords, rhythm patterns, bass lines, and drums. This manual serves as both user documentation and LLM instructions for generating backing tracks.

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

---

## Track Metadata

```yaml
track:
  title: "Song Name"        # Display name
  key: C                    # Musical key (C, G, Am, F#, Bb, etc.)
  tempo: 120                # BPM (beats per minute)
  time_signature: 4/4       # Currently only 4/4 supported
  style: rock               # Genre hint (rock, blues, jazz, folk, pop, ballad)
```

### Common Tempos by Genre
| Genre | Typical BPM |
|-------|-------------|
| Slow ballad | 50-70 |
| Blues | 60-90 |
| Folk | 80-110 |
| Pop | 100-130 |
| Rock | 110-140 |
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
| `E5`, `A5`, `G5` | Power chord | E power chord |
| `Bb`, `F#`, `Eb` | Accidentals | B flat major |

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
| `strum_down` | Arpeggiated downstrums | Acoustic rock |
| `strum_up_down` | Alternating up/down | Pop, folk |
| `folk` | Bass on 1,3 + chord on 2,4 | Country, folk |
| `shuffle_strum` | Triplet shuffle | Blues, swing |
| `travis` | Travis picking (alternating bass) | Fingerstyle, country |
| `fingerpick` | 16th note fingerpicking | Folk, classical |
| `fingerpick_slow` | Sparse picking | Ballads, Leonard Cohen |
| `arpeggio_up` | Ascending arpeggio | Ambient, classical |
| `arpeggio_down` | Descending arpeggio | Ballads |

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
| `root` | Root notes only | Pop, rock, simple |
| `root_fifth` | Root on 1, fifth on 3 | Folk, country, rock |
| `walking` | Root-3rd-5th-7th pattern | Jazz, blues |
| `swing_walking` | Walking with swing feel | Blues, jazz |

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

### Rock with Custom Strum

```yaml
track:
  title: "Power Rock"
  key: E
  tempo: 130
  time_signature: 4/4
  style: rock

chord_progression:
  pattern: "E5 G5 A5 E5"
  bars_per_chord: 2
  repeat: 4

rhythm:
  pattern: "DxDxDxDx"

bass:
  style: root_fifth

drums:
  style: rock_beat
  intensity: 0.9
```

### Jazz with Walking Bass

```yaml
track:
  title: "Jazz Standards"
  key: Bb
  tempo: 140
  time_signature: 4/4
  style: jazz

chord_progression:
  pattern: "Cm7 F7 Bbmaj7 Bbmaj7"
  bars_per_chord: 2
  repeat: 4

rhythm:
  pattern: "D..Ud..U"
  swing: 0.67

bass:
  style: walking
  swing: 0.67

drums:
  style: jazz_swing
  intensity: 0.6
```

---

## LLM Generation Guidelines

When generating BTML files, follow these guidelines:

### 1. Match Style to Genre

| Genre | Rhythm | Bass | Drums | Tempo |
|-------|--------|------|-------|-------|
| Blues | shuffle_strum, swing 0.67 | walking/swing_walking | blues_shuffle/shuffle | 60-90 |
| Jazz | custom pattern, swing 0.67 | walking | jazz_swing | 120-180 |
| Folk | fingerpick/travis/folk | root/root_fifth | minimal or none | 80-110 |
| Pop | eighth/strum_up_down | root | rock_beat | 100-130 |
| Rock | custom DxDx patterns | root_fifth | rock_beat | 110-150 |
| Ballad | fingerpick_slow/arpeggio | root | minimal | 50-80 |

### 2. Use Appropriate Chord Types

| Genre | Typical Chords |
|-------|----------------|
| Blues | Dominant 7ths (A7, D7, E7) |
| Jazz | 7th chords (Cm7, F7, Bbmaj7) |
| Folk | Major/minor triads |
| Rock | Power chords (E5, A5) or triads |
| Pop | Major/minor triads |

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
# 4 chords × 4 repeats = 16 bars
```

**12-Bar Blues (repeat 2x):**
```yaml
pattern: "A7 A7 A7 A7 D7 D7 A7 A7 E7 D7 A7 E7"
repeat: 2
# 12 bars × 2 = 24 bars
```

**32-Bar Form:**
```yaml
pattern: "..." # 8 chords
bars_per_chord: 2
repeat: 2
# 8 × 2 × 2 = 32 bars
```

### 5. Creating Specific Feels

**Emotional Ballad:**
```yaml
tempo: 60
rhythm:
  style: fingerpick_slow
drums:
  intensity: 0.25
```

**Driving Rock:**
```yaml
tempo: 130
rhythm:
  pattern: "DxDxDxDx"
drums:
  intensity: 0.9
```

**Laid-back Soul:**
```yaml
tempo: 85
rhythm:
  pattern: "D.d.D.dU"
  swing: 0.55
```

---

## Command Reference

```bash
# Play a backing track
./backing-tracks play examples/blues-a.btml

# Export to MIDI file
./backing-tracks export examples/blues-a.btml

# Export with custom output path
./backing-tracks export examples/blues-a.btml my-track.mid
```

---

## Version

BTML v0.4 - Backing Tracks Player

Features:
- Chord progressions with fractional bar notation
- Custom strum patterns (D/U/x/. notation)
- Fingerpicking styles (travis, fingerpick, fingerpick_slow)
- Multiple bass styles (root, walking, swing)
- Drum presets and Euclidean rhythms
- MIDI export for DAW integration
