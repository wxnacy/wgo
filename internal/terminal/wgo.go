package terminal

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	prompt "github.com/wxnacy/code-prompt"
)

func NewWgo() *Wgo {
	return nil
}

type Wgo struct {
	prompt.BaseModel
}

func (m Wgo) Init() tea.Cmd {
	return textinput.Blink
}

func (m Wgo) View() string {
	return ""
}

func (m *Wgo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	return m, cmd
}
