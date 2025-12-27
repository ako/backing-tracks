---
name: btml-generation
description: Generate BTML backing track files for guitar practice
---

# BTML Generation Skill

Generate BTML (Backing Track Markup Language) files for guitar practice backing tracks.

## What is BTML?

BTML is a YAML-based notation for defining complete backing tracks with chords, rhythm patterns, bass lines, drums, and melody. Files use the `.btml` extension and are played with the `backing-tracks` CLI.

## File Structure

```yaml
track:
  title: "Song Name"
  key: C                    # Musical key (C, G, Am, F#, Bb, etc.)
  tempo: 120                # BPM
  time_signature: 4/4
  style: rock               # Genre hint
  tuning: standard          # Guitar tuning (optional)
  capo: 0                   # Capo position (optional)

chord_progression:
  pattern: "C G Am F"       # Space-separated chord symbols
  bars_per_chord: 1
  repeat: 4

rhythm:
  style: eighth             # Chord rhythm style

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

## Chord Symbols

| Symbol | Type |
|--------|------|
| `C`, `G`, `D` | Major triad |
| `Am`, `Em`, `Dm` | Minor triad |
| `A7`, `E7`, `D7` | Dominant 7th |
| `Cmaj7`, `Gmaj7` | Major 7th |
| `Am7`, `Em7` | Minor 7th |
| `E9`, `A9` | Dominant 9th |
| `E5`, `A5` | Power chord |
| `Asus4`, `Dsus2` | Suspended |
| `Bb`, `F#`, `Eb` | Accidentals |
| `Am/G`, `C/E` | Slash chords (bass note) |
| `C*2`, `Am*0.5` | Duration override (bars) |

## Genre Guidelines

### Blues
```yaml
track:
  key: A
  tempo: 80
  style: blues
chord_progression:
  pattern: "A7 A7 A7 A7 D7 D7 A7 A7 E7 D7 A7 E7"  # 12-bar blues
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
  style: blues_head
scale:
  type: blues
```

### Jazz
```yaml
track:
  key: C
  tempo: 140
  style: jazz
chord_progression:
  pattern: "Dm7 G7 Cmaj7 Cmaj7"  # II-V-I
rhythm:
  style: stride
  swing: 0.67
bass:
  style: walking
drums:
  style: jazz_swing
  intensity: 0.6
```

### Folk/Fingerpicking
```yaml
track:
  key: C
  tempo: 72
  style: folk
chord_progression:
  pattern: "C Am F G"
  bars_per_chord: 2
rhythm:
  style: fingerpick_slow
bass:
  style: root
drums:
  intensity: 0.2
```

### Rock
```yaml
track:
  key: E
  tempo: 130
  style: rock
chord_progression:
  pattern: "E5 G5 A5 D5"
rhythm:
  style: eighth
bass:
  style: root_fifth
drums:
  style: rock_beat
  intensity: 0.85
```

### Funk
```yaml
track:
  key: E
  tempo: 100
  style: funk
chord_progression:
  pattern: "E9 E9 E9 A9"
rhythm:
  style: funk
  accent: "1"
bass:
  style: slap
drums:
  style: funk
  intensity: 0.85
scale:
  type: mixolydian
```

### EDM/Electronic
```yaml
track:
  key: Am
  tempo: 128
  style: edm
chord_progression:
  pattern: "Am F C G"
  bars_per_chord: 2
rhythm:
  style: sixteenth
bass:
  style: 808
drums:
  style: four_on_floor
  intensity: 0.9
```

### Pop Ballad
```yaml
track:
  key: G
  tempo: 60
  style: ballad
chord_progression:
  pattern: "G D Em C"
  bars_per_chord: 2
rhythm:
  style: arpeggio_down
bass:
  style: root
drums:
  intensity: 0.3
```

## Rhythm Styles

| Style | Description | Best For |
|-------|-------------|----------|
| `whole` | Sustained chords | Slow ballads |
| `half` | Two strums per bar | Ballads |
| `quarter` | Four strums per bar | Folk, pop |
| `eighth` | Eight strums per bar | Rock, pop |
| `sixteenth` | 16th note strumming | Funk, disco |
| `funk` | Syncopated 16th notes | Funk, R&B |
| `shuffle_strum` | Triplet shuffle | Blues, swing |
| `fingerpick` | 16th note fingerpicking | Folk, classical |
| `fingerpick_slow` | Sparse picking | Ballads |
| `travis` | Travis picking | Country, folk |
| `arpeggio_up` | Ascending arpeggio | New wave |
| `arpeggio_down` | Descending arpeggio | Ballads |
| `stride` | Stride piano pattern | Jazz, ragtime |

Custom patterns: `pattern: "D.DU.UDU"` (D=down, U=up, x=mute, .=rest)

## Bass Styles

| Style | Description |
|-------|-------------|
| `root` | Root notes only |
| `root_fifth` | Root on 1, fifth on 3 |
| `walking` | Root-3rd-5th-7th |
| `swing_walking` | Walking with swing |
| `stride` | Stride piano bass |
| `boogie` | Boogie-woogie |
| `808` / `sub` | Heavy sub bass |
| `funk` / `slap` | Syncopated slap bass |

## Drum Styles

| Style | Description |
|-------|-------------|
| `rock_beat` | Kick 1,3 / Snare 2,4 / 8th hi-hat |
| `shuffle` | Blues shuffle |
| `blues_shuffle` | Driving blues |
| `jazz_swing` | Ride pattern |
| `four_on_floor` | Four kicks per bar (EDM) |
| `trap` | Trap-style with rolling hi-hats |
| `funk` | Tight funk groove |

Intensity: 0.0-1.0 (0.3=gentle, 0.7=moderate, 0.9=aggressive)

## Sections and Form

For complex songs with verses/choruses:

```yaml
sections:
  - name: verse
    chord_progression:
      pattern: "C G Am F"
    lyrics: |
      C              G
      Lyrics go here with chords above
      Am             F
      Positioned where they should be played

  - name: chorus
    chord_progression:
      pattern: "F G C Am"

form:
  - verse
  - verse
  - chorus
  - verse
  - chorus
```

## Guitar Tunings

| Tuning | Notes | Use Case |
|--------|-------|----------|
| `standard` | E A D G B e | Default |
| `drop_d` | D A D G B e | Heavy riffs |
| `open_e` | E B E G# B e | Slide blues |
| `open_g` | D G D G B d | Rolling Stones |
| `dadgad` | D A D G A d | Celtic |

## Common Progressions

```yaml
# Pop (I-V-vi-IV)
pattern: "C G Am F"

# 12-Bar Blues
pattern: "A7 A7 A7 A7 D7 D7 A7 A7 E7 D7 A7 E7"

# Jazz II-V-I
pattern: "Dm7 G7 Cmaj7 Cmaj7"

# Minor Ballad
pattern: "Am G F E"

# Descending Bass
pattern: "Am Am/G Am/F Am/E"

# Funk Vamp
pattern: "E9 E9 E9 A9"
```

## Best Practices

1. **Match tempo to genre**: Blues 60-90, Folk 80-110, Pop 100-130, Rock 110-150, EDM 120-140
2. **Use appropriate chords**: Blues uses dominant 7ths, Funk uses 9ths, Folk uses simple triads
3. **Set swing for blues/jazz**: Use `swing: 0.67` for triplet feel
4. **Balance intensity**: Start with 0.6-0.7, adjust based on energy level
5. **Use sections for structure**: Prefer sections+form over long repeating patterns
6. **Include scale for soloing**: Match scale to key (pentatonic_minor for blues/rock)

## Output Location

Save BTML files to `examples/` directory with `.btml` extension.
