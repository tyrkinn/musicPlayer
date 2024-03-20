package Songs

import (
	"fmt"
	"log"
	"main/models/messages"
	"main/playlists"
	"main/state"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/faiface/beep/mp3"
)

type Model struct {
	table       table.Model
	defaultRows []table.Row
	focused     bool
	state       *state.State
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (m Model) Init() tea.Cmd { return nil }

func (m *Model) Focus() {
	m.table.Focus()
}
func (m *Model) Blur() {
	m.table.Blur()
}
func (m *Model) Focused() bool {

	return m.table.Focused()
}

func DefaultSongs(state *state.State) (Model, error) {
	columns := []table.Column{{Title: "Songs", Width: 50}, {Title: "", Width: 5}}
	rows, err := state.SongsWithDuration()
	if err != nil {
		return Model{}, err
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	return Model{
		table:       t,
		defaultRows: rows,
		state:       state,
	}, nil
}

func (m *Model) SetFocused(b bool) {
	m.focused = b
}

func (m Model) View() string {
	if m.state.CurrentWindow == m.WindowKey() {
		baseStyle.BorderForeground(lipgloss.Color("229"))
	} else {
		baseStyle.BorderForeground(lipgloss.Color("240"))
	}
	return baseStyle.Render(m.table.View()) + "\n"
}

func (m *Model) WindowKey() state.ProgramWindow {
	return state.SONGS
}

func (m *Model) SetCurrPlaylist(newPlaylist string) {
	m.state.CurrentPlaylist = newPlaylist
	pl := playlists.P{}
	songs, _ := pl.ShowAllSongs(newPlaylist)
	var rows []table.Row
	for _, song := range songs {
		f, err := os.Open("./playlists/dir/" + m.state.CurrentPlaylist + "/" + song)
		streamer, format, err := mp3.Decode(f)
		if err != nil {
			fmt.Println(err)
		}
		rows = append(rows, table.Row{song, format.SampleRate.D(streamer.Len()).Round(time.Second).String()})
		streamer.Close()
		f.Close()
	}
	m.table.SetRows(rows)
}
func (m *Model) SetCurrentSong(song string) {
	m.state.CurrentSong = song
}
func (m *Model) NextSong() {
	for i := range m.state.SongList {
		if m.state.SongList[i] == m.state.CurrentSong {
			if i == len(m.state.SongList)-1 {
				m.table.SetCursor(0)
				m.state.CurrentSong = m.state.SongList[0]
				break
			}
			m.table.SetCursor(i + 1)
			m.state.CurrentSong = m.state.SongList[i+1]
			break
		}
	}
}
func (m *Model) PrevSong() {
	for i := range m.state.SongList {
		if m.state.SongList[i] == m.state.CurrentSong {
			if i == 0 {
				m.table.SetCursor(len(m.state.SongList) - 1)
				m.state.CurrentSong = m.state.SongList[len(m.state.SongList)-1]
				break
			}
			m.table.SetCursor(i - 1)
			m.state.CurrentSong = m.state.SongList[i-1]
			break
		}
	}
}
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case messages.SongsUpdated:
		rows, err := m.state.SongsWithDuration()
		if err != nil {
			log.Fatal(err)
		}
		m.table.SetRows(rows)
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			m.state.CurrentSong = m.table.SelectedRow()[0]
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}
