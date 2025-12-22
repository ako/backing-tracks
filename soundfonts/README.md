# SoundFonts Directory

Place your `.sf2` SoundFont files here for automatic detection.

## Recommended SoundFonts

### General MIDI (Full Orchestra)

| SoundFont | Size | Description | Link |
|-----------|------|-------------|------|
| **FluidR3_GM** | 140MB | Default, good quality | Pre-installed on most Linux |
| **SGM-V2.01** | 235MB | Excellent quality, realistic | [Download](https://musical-artifacts.com/artifacts/855) |
| **Timbres of Heaven** | 364MB | High quality, natural sound | [Download](https://midkar.com/soundfonts/) |
| **Arachno** | 148MB | Good all-around | [Download](https://www.arachnosoft.com/main/soundfont.php) |

### Specialized

| SoundFont | Size | Description | Link |
|-----------|------|-------------|------|
| **Nice-Keys-Ultimate** | 47MB | Great piano sounds | [Download](https://musical-artifacts.com/artifacts/1239) |
| **Real Guitar** | 13MB | Acoustic guitar focused | [Download](https://musical-artifacts.com/artifacts/1089) |
| **DSK Drumkits** | 45MB | Better drum sounds | [Download](https://www.dskmusic.com/) |

## Usage

Once you place a `.sf2` file here, it will be automatically used:

```bash
# Auto-detects soundfonts in this directory
./backing-tracks play examples/blues-full.btml

# Or specify explicitly
./backing-tracks play --soundfont soundfonts/SGM-V2.01.sf2 examples/blues-full.btml

# Set as default via environment variable
export SOUNDFONT=soundfonts/SGM-V2.01.sf2
./backing-tracks play examples/blues-full.btml
```

## Installing System SoundFonts

### Ubuntu/Debian
```bash
sudo apt install fluid-soundfont-gm      # Basic GM soundfont
sudo apt install timgm6mb-soundfont      # Smaller alternative
```

### Arch Linux
```bash
sudo pacman -S soundfont-fluid
```

### macOS (via Homebrew)
```bash
brew install fluid-synth
# Download soundfonts manually
```

## Notes

- Larger soundfonts generally sound better but use more memory
- For EDM/electronic music, the default FluidR3 is adequate
- For jazz/acoustic music, SGM or Timbres of Heaven sound more realistic
- The player will use the first `.sf2` file found in this directory
