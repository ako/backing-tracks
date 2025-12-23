# Music Styles Guide

This guide explains the music theory behind each style supported by BTML, and how this software implements them through rhythm patterns, bass lines, and drum beats.

## Table of Contents

1. [Music Theory Fundamentals](#music-theory-fundamentals)
2. [Rock](#rock)
3. [Blues & Shuffle](#blues--shuffle)
4. [Jazz Swing](#jazz-swing)
5. [Folk & Fingerpicking](#folk--fingerpicking)
6. [Funk](#funk)
7. [Ska](#ska)
8. [Reggae](#reggae)
9. [Country](#country)
10. [Disco](#disco)
11. [Motown / Soul](#motown--soul)
12. [Flamenco](#flamenco)
13. [EDM & Trap](#edm--trap)
14. [Ragtime, Stride & Boogie-Woogie](#ragtime-stride--boogie-woogie)

---

## Music Theory Fundamentals

Before diving into specific styles, understanding some core concepts helps explain why each style sounds the way it does.

### Beat and Meter

Most Western music uses **4/4 time** (four beats per bar). The beats are numbered 1-2-3-4:

```
Beat:    1     2     3     4
         |     |     |     |
```

**Downbeats** (1 and 3) typically feel "strong" while **backbeats** (2 and 4) provide the groove. The emphasis on different beats is what distinguishes many styles.

### Subdivisions

Each beat can be subdivided:

- **Eighth notes**: Each beat divided in two (1 & 2 & 3 & 4 &)
- **Sixteenth notes**: Each beat divided in four (1 e & a 2 e & a 3 e & a 4 e & a)
- **Triplets**: Each beat divided in three (1-trip-let 2-trip-let...)

### Swing vs. Straight Feel

**Straight feel**: Subdivisions are evenly spaced (50/50)
```
Straight 8ths:  |-----|-----|-----|-----|-----|-----|-----|-----|
                1     &     2     &     3     &     4     &
```

**Swing feel**: The off-beat is delayed (typically 67/33 or "triplet swing")
```
Swing 8ths:     |--------|--|--------|--|--------|--|--------|--|
                1        &  2        &  3        &  4        &
```

This creates a "bouncy" or "lilting" groove characteristic of jazz, blues, and shuffle styles.

### Velocity and Dynamics

**Velocity** in MIDI determines how hard a note is struck (0-127). Accenting certain beats creates the rhythmic feel:

- Accenting beat 1 creates a "heavy" or "driving" feel
- Accenting beats 2 and 4 creates the "backbeat" feel (rock, soul)
- Accenting off-beats creates syncopation (funk, reggae)

---

## Rock

### Musical Characteristics

Rock music is built on the **backbeat** - a strong emphasis on beats 2 and 4. This creates the driving, energetic feel that defines the genre. Rock typically uses:

- **Straight eighth notes** (no swing)
- **Kick drum on 1 and 3** (the pulse)
- **Snare on 2 and 4** (the backbeat)
- **Eighth note hi-hats** (the ride)

### The Rock Beat Pattern

```
Beat:     1     &     2     &     3     &     4     &
Kick:     X                 X
Snare:                X                       X
Hi-hat:   x     x     x     x     x     x     x     x
```

### BTML Implementation

**Drums** (`rockBeat()`):
- Kick at ticks 0 and half-bar
- Snare at quarter-bar and three-quarter-bar
- Closed hi-hat on all eighth notes, softer on off-beats

**Bass styles that pair well**: `root_fifth`, `walking`

**Rhythm styles that pair well**: `eighth`, `strum_up_down`, `quarter`

### Example BTML

```yaml
rhythm:
  style: eighth

bass:
  style: root_fifth

drums:
  style: rock_beat
  intensity: 0.8
```

---

## Blues & Shuffle

### Musical Characteristics

Blues introduced the **shuffle feel** (also called "swing" or "triplet feel") where eighth notes are played with a long-short pattern based on triplets. This creates the characteristic "bouncy" blues groove.

The blues also uses **12/8 feel** - each beat feels like a triplet:

```
Straight 8ths:  1  &  2  &  3  &  4  &
Shuffle 8ths:   1  _a 2  _a 3  _a 4  _a  (where "_a" is delayed)
```

### Shuffle vs. Straight Comparison

```
Straight:  |--|--|--|--|--|--|--|--|
           1  &  2  &  3  &  4  &

Shuffle:   |-----|--|-----|--|-----|--|-----|--|
           1     a  2     a  3     a  4     a
```

### BTML Implementation

**Drums** (`shuffleBeat()`, `bluesShuffle()`):
- Divides the bar into 12 triplet-eighth positions
- Places hi-hats on positions 0, 2, 3, 5, 6, 8, 9, 11 (skipping middle triplets)
- `bluesShuffle` adds open hi-hat on up-beats and ghost notes on snare

**Bass** (`swing_walking`):
- Applies swing ratio to off-beats
- `swing: 0.67` gives classic triplet feel (67% first note, 33% second)

**Rhythm** (`shuffle_strum`):
- Strums follow the triplet pattern
- Accents on triplet downbeats (positions 0, 3, 6, 9)

### Example BTML

```yaml
rhythm:
  style: shuffle_strum
  swing: 0.67

bass:
  style: swing_walking
  swing: 0.67

drums:
  style: shuffle
  intensity: 0.6
```

---

## Jazz Swing

### Musical Characteristics

Jazz swing evolved from blues shuffle but is more sophisticated. Key characteristics:

- **Ride cymbal** instead of hi-hat (warmer, more complex sound)
- **Sparse kick drum** (just defines the "1")
- **Subtle snare on 2 and 4** (softer than rock)
- **Walking bass** that moves through chord tones

The classic jazz ride pattern:
```
Beat:     1        &     2        &     3        &     4        &
Ride:     DING     _a    DING     _a    DING     _a    DING     _a
          (strong)       (medium)       (strong)       (medium)
```

### BTML Implementation

**Drums** (`jazzSwing()`):
- Ride cymbal on swung eighth positions
- Very sparse kick (only beat 1)
- Soft snare on 2 and 4

**Bass** (`walking`, `swing_walking`):
- Root → 3rd → 5th → 7th pattern
- Creates melodic movement through the chord

**Rhythm**: Often uses `whole` or `half` for comping, letting the bass and drums carry the groove

### Example BTML

```yaml
rhythm:
  style: whole

bass:
  style: swing_walking
  swing: 0.67

drums:
  style: jazz_swing
  intensity: 0.5
```

---

## Folk & Fingerpicking

### Musical Characteristics

Folk music often uses **fingerpicking patterns** where the thumb plays bass notes while fingers play melody on higher strings. Common patterns include:

- **Travis picking**: Alternating bass with syncopated treble (named after Merle Travis)
- **Folk arpeggio**: Bass note followed by ascending chord tones
- **Boom-chick**: Bass on 1/3, chord stab on 2/4

### Travis Picking Pattern

```
Beat:    1     &     2     &     3     &     4     &
Thumb:   B                 B5
Index:         H                 H           H
Middle:              M                 M
         (Bass)      (High)(Mid)(High)(B5)  (H)  (M)  (H)
```

### BTML Implementation

**Rhythm styles**:
- `travis`: Alternating bass (root/fifth) with treble melody (high-mid-high pattern)
- `fingerpick`: Classic folk pattern (Bass-Low-Mid-High-Mid-Low per bar)
- `fingerpick_slow`: Sparse pattern for ballads (Nick Drake style)
- `folk`: Simple boom-chick (bass on 1/3, chord on 2/4)

### Example BTML

```yaml
rhythm:
  style: travis

bass:
  style: root  # Travis picking includes bass

drums:
  style: rock_beat
  intensity: 0.3  # Light brushes or no drums
```

---

## Funk

### Musical Characteristics

Funk is defined by **THE ONE** - an extremely heavy accent on beat 1 of each bar. Beyond that, funk features:

- **16th note subdivisions** (much busier than 8th note styles)
- **Syncopation** - accents on unexpected beats (especially "e" and "a")
- **Staccato/choppy** playing - short, percussive notes
- **Ghost notes** - very soft notes that add texture

### The Funk Pattern

```
Beat:     1  e  &  a  2  e  &  a  3  e  &  a  4  e  &  a
Pattern:  X  .  x  .  .  x  X  .  x  .  x  .  .  x  .  x
          95    60       65 80    70    60       65    70
          ↑                 ↑
          THE ONE!          accent
```

Notice how the accents fall on unexpected subdivisions (e of 2, & of 2, e of 4), creating that characteristic "off-kilter" funk feel.

### BTML Implementation

**Rhythm** (`funk`, `funk_muted`):
- 16th note pattern with heavy beat 1
- Syncopated accents on e's and a's
- `funk_muted` uses shorter notes for choppier feel

**Bass** (`funk`, `slap`, `funk_simple`):
- Syncopated pattern with octave jumps
- Ghost notes at lower velocities
- Heavy THE ONE accent

**Drums** (works with `rock_beat` or custom patterns):
- Typically uses sparse kick, active snare with ghost notes

### Example BTML

```yaml
rhythm:
  style: funk_muted

bass:
  style: funk

drums:
  style: rock_beat
  intensity: 0.8
```

---

## Ska

### Musical Characteristics

Ska originated in Jamaica in the late 1950s, characterized by the **off-beat guitar skank**:

- Guitar/keys play ONLY on the off-beats (the "&" of each beat)
- Walking bass with octave jumps
- Driving hi-hat pattern accenting off-beats
- Upbeat, energetic tempo (typically 140-180 BPM)

### The Ska Skank

```
Beat:     1     &     2     &     3     &     4     &
Guitar:         X           X           X           X
                ↑           ↑           ↑           ↑
                ONLY off-beats (the "skank")
```

This creates the instantly recognizable "cha-cha-cha-cha" ska rhythm.

### BTML Implementation

**Rhythm** (`ska`, `skank`):
- Plays chords only on positions 1, 3, 5, 7 of 8th note grid (off-beats)
- Short, choppy note duration
- Consistent velocity across all skanks

**Bass** (`ska`):
- Walking pattern: root → fifth → octave → fifth
- Octave jump on beat 3 is characteristic

**Drums** (`skaBeat()`):
- Standard kick on 1/3, snare on 2/4
- Hi-hat accents off-beats to support the skank

### Example BTML

```yaml
track:
  tempo: 160  # Ska is upbeat!

rhythm:
  style: ska

bass:
  style: ska

drums:
  style: ska
  intensity: 0.8
```

---

## Reggae

### Musical Characteristics

Reggae evolved from ska but at slower tempos with a "heavier" feel. The defining characteristic is the **one-drop** beat:

- **Kick AND snare hit together on beat 3** (the "one-drop")
- Silence on beat 1 (very unusual in Western music)
- Off-beat chord chops ("chank")
- Deep, spacious bass

### The One-Drop Pattern

```
Beat:     1     &     2     &     3     &     4     &
Kick:                             X
Snare:                            X           (soft)
Hi-hat:         x           x           x           x
Guitar:         x           x           X           x
                                        ↑
                                Heavy on beat 3
```

The emphasis on beat 3 creates the laid-back, hypnotic reggae groove.

### BTML Implementation

**Rhythm** (`reggae`, `one_drop`):
- Off-beat chops (like ska, but heavier accent on beat 3)
- Chord on beat 3 is accented (vel 90 vs 65 for others)

**Bass** (`reggae`, `one_drop`):
- Sparse: root on beat 1 (held long), fifth on beat 3
- Deep, sustained notes with space

**Drums** (`reggaeBeat()`):
- ONE-DROP: kick + snare together only on beat 3
- Off-beat hi-hats only
- Creates lots of space

### Example BTML

```yaml
track:
  tempo: 76  # Reggae is slower than ska

rhythm:
  style: reggae

bass:
  style: reggae

drums:
  style: reggae
  intensity: 0.7
```

---

## Country

### Musical Characteristics

Country uses the **boom-chick** or **train beat** pattern:

- "Boom" = bass note on beats 1 and 3
- "Chick" = chord on beats 2 and 4
- Alternating root-fifth bass movement
- Steady, driving hi-hat

This creates the "two-step" or "shuffle" feel associated with country music.

### The Train Beat

```
Beat:     1     &     2     &     3     &     4     &
Bass:     B                 B5
Guitar:               chord             chord
Hi-hat:   x     x     x     x     x     x     x     x
```

### BTML Implementation

**Rhythm** (`country`, `train`):
- Bass note (root) on beats 1 and 3
- Full chord on beats 2 and 4
- Creates the characteristic boom-chick

**Bass** (`country`, `train`):
- Alternating: root → fifth → root → fifth
- Steady quarter notes
- Classic Nashville bass style

**Drums** (`countryBeat()`):
- Standard kick 1/3, snare 2/4
- Steady 8th note hi-hat (the "train")

### Example BTML

```yaml
track:
  tempo: 120

rhythm:
  style: country

bass:
  style: country

drums:
  style: country
  intensity: 0.7
```

---

## Disco

### Musical Characteristics

Disco introduced **four-on-the-floor** - kick drum on every beat:

- Kick on beats 1, 2, 3, AND 4 (hence "four on the floor")
- Snare on 2 and 4 (backbeat)
- 16th note hi-hats with open hi-hat accents
- Driving, danceable groove

### The Disco Beat

```
Beat:     1  e  &  a  2  e  &  a  3  e  &  a  4  e  &  a
Kick:     X           X           X           X
Snare:                X                       X
Hi-hat:   x  x  O  x  x  x  O  x  x  x  O  x  x  x  O  x
                ↑                 ↑
                Open hi-hat on & of each beat
```

The open hi-hat on the "&" creates the characteristic "shimmering" disco feel.

### BTML Implementation

**Drums** (`discoBeat()`):
- Four-on-the-floor kick
- Open hi-hat on position 2 of each 4-position group (the "&")
- 16th note closed hi-hat otherwise

**Bass** (`disco`):
- Driving octave pattern on 8th notes
- Root on downbeats, octave up on off-beats
- Creates the "pumping" disco bass

**Rhythm** (`disco`):
- Strong chord on quarter notes
- Light off-beat hits
- Supports the four-on-the-floor feel

### Example BTML

```yaml
track:
  tempo: 120

rhythm:
  style: disco

bass:
  style: disco

drums:
  style: disco
  intensity: 0.85
```

---

## Motown / Soul

### Musical Characteristics

The "Motown Sound" was created at Hitsville U.S.A. in Detroit, characterized by:

- **Heavy backbeat** on 2 and 4 (even heavier than rock)
- **Melodic, syncopated bass** (James Jamerson style)
- **Tambourine** feel on 8th notes
- Gospel-influenced chord progressions

The Motown rhythm section (The Funk Brothers) emphasized:
- Tight, locked grooves
- Call-and-response between instruments
- Rich harmonic bass lines that move through chord tones

### BTML Implementation

**Rhythm** (`motown`, `soul`):
- 8th note pattern
- Heavy accents on beats 2 and 4 (vel 90)
- Moderate accent on beat 1 (vel 80)

**Bass** (`motown`, `soul`):
- Melodic, syncopated pattern moving through chord tones
- Pattern: root → fifth → root → third → fifth → root → third → fifth
- James Jamerson-inspired voice leading

**Drums** (`motownBeat()`):
- Extra-heavy snare on 2 and 4
- Kick pickup before beat 3
- Tambourine-style 8th note hi-hat

### Example BTML

```yaml
track:
  tempo: 112

rhythm:
  style: motown

bass:
  style: motown

drums:
  style: motown
  intensity: 0.75
```

---

## Flamenco

### Musical Characteristics

Flamenco music from Spain uses unique rhythmic patterns called **compás**. The rumba flamenca pattern is the most accessible:

- **Syncopated accents** that don't align with standard 4/4
- **Strong accent on beat 1** (the "golpe")
- **Rasgueado** strumming technique
- Complex interplay between guitar and percussion (palmas/cajón)

### The Rumba Pattern (Simplified)

```
Position: 1  2  3  4  5  6  7  8  9  10 11 12 13 14 15 16
Accent:   X        x        x     x     x     x
Velocity: 95       75       75    85    75    80
```

The accents at positions 0, 3, 6, 8, 10, 12 create a 3-3-2-2-2-4 grouping that gives flamenco its distinctive feel.

### BTML Implementation

**Rhythm** (`flamenco`, `rumba`):
- Accent pattern: positions 0, 3, 6, 8, 10, 12 (16th note grid)
- Heaviest accent on beat 1 (vel 95)
- Longer note on beat 1

**Drums** (`flamencoBeat()`):
- Simulates cajón: low tones (kick) and slaps (snare)
- Kick on positions 0, 6, 10
- Slaps on positions 3, 8, 12
- Ghost notes (finger rolls) on 2, 5, 9, 14

### Example BTML

```yaml
track:
  tempo: 100
  key: Am  # Flamenco often uses minor/Phrygian

rhythm:
  style: flamenco

drums:
  style: flamenco
  intensity: 0.8
```

---

## EDM & Trap

### Musical Characteristics

**EDM (Electronic Dance Music)** builds on disco's four-on-the-floor but with:
- **Heavier kick drum** (sidechain compression feel)
- **16th note hi-hats** with variations
- **Build-ups and drops**
- **808 sub bass**

**Trap** adds:
- **Rolling hi-hats** (32nd notes, triplet rolls)
- **Sparse, syncopated kick**
- **Very heavy 808 bass**
- **Heavy snare on 2 and 4**

### EDM Pattern

```
Beat:     1  e  &  a  2  e  &  a  3  e  &  a  4  e  &  a
Kick:     X           X           X           X
Snare:                X                       X
Hi-hat:   x  x  O  x  x  x  O  x  x  x  O  x  x  x  O  x
```

### Trap Pattern

```
Beat:     1  e  &  a  2  e  &  a  3  e  &  a  4  e  &  a
Kick:     X                 x                 X
Snare:                X                       X
Hi-hat:   x  x  x  x  x  x  x  x  x  x [rolling 32nds here]
```

### BTML Implementation

**Drums** (`fourOnFloor()`):
- Four-on-the-floor kick
- 16th note hi-hats with open hi-hat accents

**Drums** (`trapBeat()`):
- Syncopated kick (1, &-of-2, 4)
- Rolling hi-hats with 32nd note fills

**Bass** (`808`, `sub`, `808_octave`, `edm`):
- Very low sustained notes (MIDI note 28)
- Syncopated pattern: hit on 1, &-of-2, 4
- `808_octave` adds octave jumps for movement

### Example BTML (EDM)

```yaml
rhythm:
  style: sixteenth

bass:
  style: edm

drums:
  style: edm
  intensity: 0.9
```

### Example BTML (Trap)

```yaml
rhythm:
  style: sixteenth

bass:
  style: 808

drums:
  style: trap
  intensity: 0.85
```

---

## Ragtime, Stride & Boogie-Woogie

### Musical Characteristics

These early American styles share the **oom-pah** feel but differ in complexity:

**Ragtime** (1890s-1910s):
- Left hand: bass on 1/3, chord on 2/4
- Syncopated, "ragged" right hand melodies
- Moderate tempo

**Stride** (1920s):
- Evolution of ragtime with wider bass jumps
- More virtuosic left hand
- Faster tempos

**Boogie-Woogie** (1930s):
- Driving 8th note bass pattern
- 12-bar blues form
- Bass pattern: 1-1-5-6-b7-6-5-5

### The Oom-Pah Pattern (Ragtime/Stride)

```
Beat:     1     2     3     4
Left:     B   chord   B   chord
          (oom) (pah) (oom) (pah)
```

### The Boogie Pattern

```
Beat:     1     &     2     &     3     &     4     &
Bass:     1     1     5     6     b7    6     5     5
```

This creates the driving, rolling boogie-woogie feel.

### BTML Implementation

**Rhythm** (`stride`, `ragtime`):
- `stride`: Chord stabs only on beats 2 and 4
- `ragtime`: Similar but with syncopated anticipations

**Bass** (`stride`, `boogie`):
- `stride`: Low bass on beats 1/3 (the "oom")
- `boogie`: 8th note pattern through 1-1-5-6-b7-6-5-5

**Drums** (use light or no drums, or jazz_swing for variety)

### Example BTML (Stride)

```yaml
rhythm:
  style: stride

bass:
  style: stride

drums:
  style: jazz_swing
  intensity: 0.4
```

### Example BTML (Boogie-Woogie)

```yaml
track:
  tempo: 140

rhythm:
  style: stride  # or custom pattern

bass:
  style: boogie

drums:
  style: shuffle
  intensity: 0.7
```

---

## Quick Reference Table

| Style | Rhythm | Bass | Drums | Tempo | Instruments |
|-------|--------|------|-------|-------|-------------|
| Rock | `eighth` | `root_fifth` | `rock_beat` | 90-140 | `steel_guitar`, `overdrive` |
| Blues | `shuffle_strum` | `swing_walking` | `shuffle` | 60-120 | `steel_guitar`, `fretless_bass` |
| Jazz | `whole`, `half` | `swing_walking` | `jazz_swing` | 100-180 | `jazz_guitar`, `contrabass` |
| Folk | `travis`, `fingerpick` | `root` | Light/none | 80-140 | `nylon_guitar`, `steel_guitar` |
| Funk | `funk`, `funk_muted` | `slap` | `rock_beat` | 90-120 | `clean_guitar`, `slap_bass` |
| Classical | `arpeggio_up` | `root` | none | 60-100 | `nylon_guitar`, `strings` |
| Country | `country`, `train` | `root_fifth` | `country` | 100-140 | `steel_guitar`, `picked_bass` |
| Disco | `disco` | `disco` | `disco` | 110-130 | `clean_guitar`, `fingered_bass` |
| EDM | `sixteenth` | `808_octave` | `four_on_floor` | 120-150 | `synth_lead`, `synth_bass` |
| Ragtime | `stride` | `stride` | Light | 80-120 | `honky_tonk`, `piano` |
| Boogie | `stride` | `boogie` | `shuffle` | 120-160 | `honky_tonk`, `piano` |

---

## Tips for Creating Authentic Styles

1. **Match bass and drum styles** - They should complement each other
2. **Use appropriate tempos** - Reggae at 160 BPM won't feel right
3. **Consider swing** - Blues and jazz need it; funk and disco don't
4. **Adjust intensity** - Ballads need lower intensity; dance music needs higher
5. **Choose fitting instruments** - Nylon guitar for classical, slap bass for funk
6. **Listen to examples** - The best way to understand a style is to listen to it

## Further Reading

- **The Drummer's Bible** by Mick Berry & Jason Gianni - Comprehensive drum pattern reference
- **Standing in the Shadows of Motown** - James Jamerson's bass lines
- **Funk Drumming** by Jim Payne - Detailed funk analysis
- **The Latin Real Book** - Authentic Latin/Flamenco patterns
