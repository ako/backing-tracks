package display

import (
	"time"

	"backing-tracks/midi"
	"backing-tracks/parser"
	"backing-tracks/theory"

	tea "github.com/charmbracelet/bubbletea"
)

// TickMsg is sent on each tick for time updates
type TickMsg time.Time

// PlayerController interface for controlling audio playback
type PlayerController interface {
	TogglePause()
	SeekRelative(bars int)
	GetPlaybackState() (bar int, beat int, strum int, paused bool)
	IsPaused() bool
	Transpose(semitones int)
	GetTranspose() int
	SetCapo(fret int)
	GetCapo() int
	ToggleTrackMute(track int) // 0=drums, 1=bass, 2=chords, 3=melody, 4=fingerstyle
	IsTrackMuted(track int) bool
	SetFingerstylePattern(pattern midi.PatternType)
	GetFingerstylePattern() midi.PatternType
	ToggleLoop(length int)                                 // Toggle loop of N bars from current position
	GetLoop() (enabled bool, startBar, endBar, length int) // Get loop state
	AdjustTempo(deltaBPM int)                              // Adjust playback tempo by delta BPM
	GetTempo() (effectiveBPM int, offset int)              // Get current effective tempo and offset
	GetCurrentSection() (name string, startBar, endBar int) // Get current section info
	LoopCurrentSection()                                    // Toggle loop for current section
	GetCurrentLyrics() (text string, chords []string)       // Get lyrics at current position
	GetLyricsForBar(bar int) (text string, chords []string) // Get lyrics for specific bar
	GetBeatLyricsForBar(bar int) []struct {                 // Get lyrics with beat positions
		Beat   int
		Lyrics string
		Chord  string
	}
	HasLyrics() bool                         // Check if track has any lyrics
	UpdateLyrics(lyrics []parser.BeatLyric) // Update lyrics after editing
}

// TUIModel is the Bubbletea model for live display
type TUIModel struct {
	track        *parser.Track
	bars         []Bar
	chords       []parser.Chord
	tempo        int
	timePerBeat  time.Duration
	startTime    time.Time
	currentBar   int
	currentBeat  int
	currentStrum int

	// Display components
	fretboard    *FretboardDisplay
	chordChart   *ChordChart
	tablature    *TablatureDisplay
	currentScale *theory.Scale
	tuning       theory.Tuning
	tuningIndex  int    // Index into theory.TuningNames
	tuningName   string // Current tuning name for display

	// Layout
	width  int
	height int

	// State
	playing         bool
	paused          bool
	pausedAt        time.Time
	pausedTotal     time.Duration
	seekOffset      time.Duration // For seeking forward/backward
	transposeOffset int           // Semitones to transpose (+/-)
	capoPosition    int           // Capo fret position (0 = no capo)
	quitting        bool

	// Display toggles (for left column components)
	showLyrics       bool // Show lyrics display
	showMetronome    bool // Show beat counter/metronome
	showStrumPattern bool // Show strum pattern visualization
	showTablature    bool // Show inline tablature
	showChordNames   bool // Show chord names above lyrics

	// Edit mode state (for lyrics timing adjustment)
	editMode     bool               // Whether in edit mode
	editBar      int                // Selected bar number
	editBeat     int                // Selected beat within bar (0-3)
	editLyrics   []parser.BeatLyric // Working copy of lyrics for editing
	editDirty    bool               // Whether changes have been made
	editFilename string             // BTML filename for saving

	// Edit focus: "lyrics", "form", "sections", "track"
	editFocus         string   // Current edit focus area
	editFormIndex     int      // Selected form entry index
	editSectionIndex  int      // Selected section index
	editChordIndex    int      // Selected chord index within section (-1 = adding new)
	editForm          []string // Working copy of form for editing
	editChordInput    string   // Current chord input buffer
	editSectionRename bool     // Whether renaming a section
	editSectionName   string   // New section name being typed

	// Track property editing
	editTrackField int    // Selected track field (0=title, 1=key, 2=tempo, etc.)
	editTrackInput string // Current input buffer for track field

	// Audio player (optional - for synced playback)
	player PlayerController
}

// NewTUIModel creates a new TUI model
func NewTUIModel(track *parser.Track) *TUIModel {
	beatsPerSecond := float64(track.Info.Tempo) / 60.0
	timePerBeat := time.Duration(float64(time.Second) / beatsPerSecond)

	bars := processChordsIntoBars(track)
	scale := theory.GetScaleForStyle(track.Info.Key, track.Info.Style, "")
	tuningName := track.Info.Tuning
	if tuningName == "" {
		tuningName = "standard"
	}
	tuning := theory.GetTuning(tuningName)
	tuningIndex := theory.GetTuningIndex(tuningName)
	fretboard := NewFretboardDisplayWithTuning(scale, 15, tuning)
	fretboard.SetCompactMode(true)
	chordChart := NewChordChart()
	tablature := NewTablatureDisplay(track, tuning, track.Info.Capo)

	// Check if track has lyrics (in sections or per-bar)
	hasLyrics := len(track.Lyrics) > 0
	for _, section := range track.Sections {
		if section.Lyrics != "" {
			hasLyrics = true
			break
		}
	}

	return &TUIModel{
		track:            track,
		bars:             bars,
		chords:           track.Progression.GetChords(),
		tempo:            track.Info.Tempo,
		timePerBeat:      timePerBeat,
		fretboard:        fretboard,
		chordChart:       chordChart,
		tablature:        tablature,
		currentScale:     scale,
		tuning:           tuning,
		tuningIndex:      tuningIndex,
		tuningName:       tuningName,
		capoPosition:     track.Info.Capo, // Initialize from track
		playing:          true,
		width:            120,
		height:           30,
		// Display toggles - sensible defaults
		showLyrics:       hasLyrics, // Enable by default if track has lyrics
		showMetronome:    true,      // Show beat counter by default
		showStrumPattern: true,      // Show strum pattern by default
		showTablature:    false,     // Inline tablature disabled by default
		showChordNames:   true,      // Show chord names by default
	}
}

// SetPlayer sets the audio player controller for synced playback
func (m *TUIModel) SetPlayer(p PlayerController) {
	m.player = p
}

// SetFilename sets the BTML filename for saving edits
func (m *TUIModel) SetFilename(filename string) {
	m.editFilename = filename
}

// Init initializes the model
func (m *TUIModel) Init() tea.Cmd {
	m.startTime = time.Now()
	return tea.Batch(
		tickCmd(),
		tea.EnterAltScreen,
	)
}

// tickCmd returns a command that ticks every 50ms
func tickCmd() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Update handles messages
func (m *TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle edit mode keys first
		if m.editMode {
			return m.handleEditModeKeys(msg)
		}
		// Handle playback mode keys
		return m.handlePlaybackKeys(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case TickMsg:
		if m.playing {
			// Always update when we have a player (it controls pause state)
			// Otherwise check local pause state
			if m.player != nil || !m.paused {
				m.updatePosition()
			}
			return m, tickCmd()
		}
	}

	return m, nil
}
