package midi

import (
	"fmt"
	"sort"
	"time"

	"backing-tracks/parser"
)

// PlaybackEvent represents a MIDI event with timing for real-time playback
type PlaybackEvent struct {
	Tick     uint32
	Channel  uint8
	Note     uint8
	Velocity uint8
	IsNoteOn bool
}

// PlaybackData contains all events needed for real-time playback
type PlaybackData struct {
	Events       []PlaybackEvent
	TicksPerBar  uint32
	TotalTicks   uint32
	TotalBars    int
	Tempo        int
	TickDuration time.Duration // Duration of one tick
}

// GeneratePlaybackData creates playback data from a track
func GeneratePlaybackData(track *parser.Track) *PlaybackData {
	ticksPerBar := uint32(1920) // 480 ticks per quarter * 4 quarters
	ticksPerQuarter := uint32(480)

	// Calculate tick duration based on tempo
	tickDuration := time.Duration(float64(time.Second) * 60.0 / float64(track.Info.Tempo) / float64(ticksPerQuarter))

	var events []PlaybackEvent
	chords := track.Progression.GetChords()

	// Calculate total ticks
	totalTicks := uint32(0)
	for _, chord := range chords {
		totalTicks += uint32(chord.Bars * float64(ticksPerBar))
	}
	totalBars := int(totalTicks / ticksPerBar)

	// Generate chord events using rhythm pattern
	chordMidiEvents := GenerateChordRhythm(chords, track.Rhythm, ticksPerBar)
	for _, evt := range chordMidiEvents {
		// Parse the MIDI message to extract note on/off
		msg := evt.message
		if len(msg) >= 3 {
			status := msg[0]
			channel := status & 0x0F
			msgType := status & 0xF0

			if msgType == 0x90 && msg[2] > 0 { // Note On with velocity > 0
				events = append(events, PlaybackEvent{
					Tick:     evt.tick,
					Channel:  channel,
					Note:     msg[1],
					Velocity: msg[2],
					IsNoteOn: true,
				})
			} else if msgType == 0x80 || (msgType == 0x90 && msg[2] == 0) { // Note Off
				events = append(events, PlaybackEvent{
					Tick:     evt.tick,
					Channel:  channel,
					Note:     msg[1],
					Velocity: 0,
					IsNoteOn: false,
				})
			}
		}
	}

	// Generate bass events
	if track.Bass != nil {
		bassNotes := GenerateBassLine(chords, track.Bass, ticksPerBar)
		for _, note := range bassNotes {
			// Note on
			events = append(events, PlaybackEvent{
				Tick:     note.Tick,
				Channel:  1, // Bass channel
				Note:     note.Note,
				Velocity: note.Velocity,
				IsNoteOn: true,
			})
			// Note off
			events = append(events, PlaybackEvent{
				Tick:     note.Tick + note.Duration,
				Channel:  1,
				Note:     note.Note,
				Velocity: 0,
				IsNoteOn: false,
			})
		}
	}

	// Generate drum events
	if track.Drums != nil {
		drumNotes := GenerateDrumPattern(totalBars, track.Drums, ticksPerBar)
		for _, note := range drumNotes {
			// Note on (drums are usually short hits)
			events = append(events, PlaybackEvent{
				Tick:     note.Tick,
				Channel:  9, // Drum channel
				Note:     note.Note,
				Velocity: note.Velocity,
				IsNoteOn: true,
			})
			// Note off (short duration for drums)
			events = append(events, PlaybackEvent{
				Tick:     note.Tick + 50, // Short drum hit
				Channel:  9,
				Note:     note.Note,
				Velocity: 0,
				IsNoteOn: false,
			})
		}
	}

	// Generate melody events
	if track.Melody != nil && track.Melody.Enabled {
		// Create melody config from track settings
		melodyConfig := &MelodyConfig{
			Density:   track.Melody.Density,
			Style:     MelodyStyle(track.Melody.Style),
		}
		if melodyConfig.Density == 0 {
			melodyConfig.Density = 0.5
		}

		melodyNotes := GenerateMelody(chords, track.Info.Key, track.Info.Style, melodyConfig, ticksPerBar)
		for _, note := range melodyNotes {
			// Note on
			events = append(events, PlaybackEvent{
				Tick:     note.Tick,
				Channel:  2, // Melody channel
				Note:     note.Note,
				Velocity: note.Velocity,
				IsNoteOn: true,
			})
			// Note off
			events = append(events, PlaybackEvent{
				Tick:     note.Tick + note.Duration,
				Channel:  2,
				Note:     note.Note,
				Velocity: 0,
				IsNoteOn: false,
			})
		}
	}

	// Sort by tick
	sort.Slice(events, func(i, j int) bool {
		return events[i].Tick < events[j].Tick
	})

	return &PlaybackData{
		Events:       events,
		TicksPerBar:  ticksPerBar,
		TotalTicks:   totalTicks,
		TotalBars:    totalBars,
		Tempo:        track.Info.Tempo,
		TickDuration: tickDuration,
	}
}

// GetEventsInRange returns events within a tick range
func (p *PlaybackData) GetEventsInRange(startTick, endTick uint32) []PlaybackEvent {
	var result []PlaybackEvent
	for _, evt := range p.Events {
		if evt.Tick >= startTick && evt.Tick < endTick {
			result = append(result, evt)
		}
		if evt.Tick >= endTick {
			break // Events are sorted, so we can stop early
		}
	}
	return result
}

// TickToTime converts a tick position to duration from start
func (p *PlaybackData) TickToTime(tick uint32) time.Duration {
	return time.Duration(tick) * p.TickDuration
}

// TimeToTick converts a duration to tick position
func (p *PlaybackData) TimeToTick(d time.Duration) uint32 {
	return uint32(d / p.TickDuration)
}

// BarToTick converts a bar number to tick position
func (p *PlaybackData) BarToTick(bar int) uint32 {
	return uint32(bar) * p.TicksPerBar
}

// FluidSynthCommand generates a FluidSynth shell command for an event
func (e *PlaybackEvent) FluidSynthCommand() string {
	if e.IsNoteOn {
		return fmt.Sprintf("noteon %d %d %d", e.Channel, e.Note, e.Velocity)
	}
	return fmt.Sprintf("noteoff %d %d", e.Channel, e.Note)
}
