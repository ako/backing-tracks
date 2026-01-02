# Spark Go Amp Integration Research

## Overview

This document explores integration possibilities between the backing-tracks project and Positive Grid Spark Go guitar amplifier.

**Goal**: Combine BTML backing tracks with Spark Go's amp modeling for a more realistic practice experience.

## Spark Go Capabilities

The Spark Go is a portable smart guitar amp with:
- Bluetooth Audio streaming (acts as Bluetooth speaker)
- Bluetooth Low Energy (BLE) for app control
- USB Audio interface functionality
- Amp/effect modeling via the Spark app

### Dual Bluetooth Connections

The Spark Go exposes two Bluetooth devices:
1. **"Spark GO Audio"** - Standard Bluetooth audio streaming
2. **"Spark GO BLE"** - App control protocol (presets, effects, parameters)

## Integration Options

### Option 1: Audio Routing (Simple)

Route backing track audio to the Spark Go while playing guitar through it.

#### Bluetooth Audio
1. Pair computer/phone with "Spark GO Audio"
2. Set system audio output to Spark GO
3. Run backing tracks - audio plays through the amp
4. Play guitar through the amp simultaneously

#### USB Audio Interface
1. Connect Spark Go via USB
2. Configure FluidSynth to output to Spark Go USB audio device:
   ```bash
   fluidsynth -a alsa -o audio.alsa.device=hw:SparkGO \
              /usr/share/sounds/sf2/FluidR3_GM.sf2 \
              /tmp/backing-track.mid
   ```
3. Guitar and backing track both come through Spark Go

**Pros**: No code changes needed, works today
**Cons**: No synchronized preset/effect changes

### Option 2: BLE Protocol Control (Advanced)

The Spark 40 Bluetooth protocol has been reverse-engineered by the community.

#### Key Resources

- **[paulhamsh/Spark](https://github.com/paulhamsh/Spark)** - Main protocol documentation and ESP32 code
- **[paulhamsh/SparkMIDI](https://github.com/paulhamsh/SparkMIDI)** - MIDI to Spark bridge
- **[happyhappysundays/SparkBox](https://github.com/happyhappysundays/SparkBox)** - BLE foot pedal project

#### Protocol Capabilities

The BLE protocol can:
- Load/switch presets (hardware presets 1-4 or custom)
- Toggle individual effects (drive, mod, delay, reverb, noise gate)
- Change effect models (swap amp types, pedal types)
- Adjust parameters (gain, tone, level, etc.)

The protocol does **NOT** handle audio - only configuration messages.

#### Hardware Required

- ESP32 microcontroller board (~$10)
  - M5 Stack Core2
  - M5 Stick C
  - Heltec WiFi Kit
  - Generic ESP32 DevKit

#### Spark Go Compatibility

**Status: Unconfirmed**

The Spark Go uses "Spark GO BLE" similar to how Spark 40 uses "Spark 40 BLE". The protocol may be compatible, but this needs testing. Key questions:

1. Does Spark Go use the same BLE service UUIDs?
   - Spark 40 uses: FFCA (from amp), FFC9 (to amp)
2. Are the message formats identical?
3. Are the effect/amp model names the same?

#### Potential Integration

If compatible, we could extend BTML to include preset changes:

```yaml
track:
  title: "Blues in A"
  tempo: 80

progression:
  pattern: "| A7 | D7 | A7 | A7 | D7 | D7 | A7 | A7 | E7 | D7 | A7 | E7 |"

# New: Spark amp integration
spark:
  preset: "Clean Blues"  # Initial preset
  sections:
    - bar: 1
      preset: "Clean Blues"
    - bar: 9
      preset: "Crunchy Lead"  # Switch for turnaround
```

### Option 3: MIDI Control Bridge

Use MIDI Control Change messages to trigger Spark preset changes via ESP32.

#### Architecture

```
┌─────────────────────┐
│ BTML File           │
│ (with CC messages)  │
└─────────┬───────────┘
          ↓
┌─────────────────────┐
│ MIDI Generator      │
│ (adds CC events)    │
└─────────┬───────────┘
          ↓
┌─────────────────────┐     ┌─────────────────────┐
│ MIDI Output         │────→│ ESP32 + SparkMIDI   │
│ (USB or Bluetooth)  │     │                     │
└─────────────────────┘     └─────────┬───────────┘
                                      ↓ BLE
                            ┌─────────────────────┐
                            │ Spark Go Amp        │
                            └─────────────────────┘
```

#### MIDI CC Mapping Example

```
CC#0  = Preset select (0-3 for hardware presets)
CC#1  = Drive toggle (0=off, 127=on)
CC#2  = Mod toggle
CC#3  = Delay toggle
CC#4  = Reverb toggle
```

#### Implementation Steps

1. Extend `midi/generator.go` to support CC events
2. Add `spark` section to BTML parser
3. Generate CC messages at specified bar positions
4. ESP32 running SparkMIDI translates CC → BLE commands

## Recommended Implementation Path

### Phase 1: Audio Routing (No code changes)
- Document USB audio setup for Spark Go
- Test Bluetooth audio streaming with backing tracks
- Verify latency is acceptable

### Phase 2: Protocol Testing
- Acquire ESP32 dev board
- Flash paulhamsh/Spark firmware
- Test BLE connection with Spark Go
- Document compatibility findings

### Phase 3: MIDI CC Integration (If Phase 2 succeeds)
- Add MIDI CC support to generator
- Extend BTML schema for spark control
- Build preset-per-section feature

### Phase 4: Real-time Control
- Integrate ESP32 control into backing-tracks binary
- Serial or network communication with ESP32
- Or: Direct BLE from Go (using tinygo-bluetooth)

## Alternative: Direct Go BLE

Instead of ESP32, we could implement BLE control directly in Go:

```go
import "tinygo.org/x/bluetooth"

// Connect to Spark Go and send preset change
func changeSparkPreset(preset int) error {
    adapter := bluetooth.DefaultAdapter
    // Scan for "Spark GO BLE"
    // Connect and send protocol message
}
```

**Pros**: Single binary, no extra hardware
**Cons**: More complex, bluetooth library maturity

## References

### Community Projects
- [paulhamsh/Spark](https://github.com/paulhamsh/Spark) - Protocol docs & ESP32 code
- [paulhamsh/SparkMIDI](https://github.com/paulhamsh/SparkMIDI) - MIDI bridge
- [paulhamsh/Spark-Control-X](https://github.com/paulhamsh/Spark-Control-X) - Control X protocol
- [happyhappysundays/SparkBox](https://github.com/happyhappysundays/SparkBox) - BLE pedal
- [jamesguitar3/sparkpal](https://github.com/jamesguitar3/sparkpal) - Node.js control
- [soundshed/spark-guide](https://github.com/soundshed/spark-guide) - Getting started guide

### Official Documentation
- [Spark Go Product Page](https://www.positivegrid.com/products/spark-go)
- [Spark Go Bluetooth Help](https://help.positivegrid.com/hc/en-us/articles/13744428707597-Connect-the-mobile-device-to-Spark-GO-via-Bluetooth)
- [Spark Go Firmware Notes](https://help.positivegrid.com/hc/en-us/articles/23639946402061-Spark-GO-Firmware-Release-Notes)

### Technical Resources
- Spark Protocol Description v3.2.pdf (in paulhamsh/Spark repo)
- BLE UUIDs: FFCA (from amp), FFC9 (to amp)
- GM Drum Map (for backing track drums)

## Open Questions

1. Is the Spark Go BLE protocol identical to Spark 40?
2. What is the BLE latency for preset changes?
3. Can we achieve gapless preset switching?
4. Does Spark Go support all the same effect models as Spark 40?

---

**Status**: Research phase
**Last Updated**: 2025-12-26
