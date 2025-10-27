package terminal

import (
	"context"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	prompt "github.com/wxnacy/code-prompt"
	"github.com/wxnacy/code-prompt/pkg/lsp"
)

func NewWgo(ctx context.Context) *Wgo {
	m := &Wgo{
		ctx: ctx,
	}
	p := prompt.NewPrompt(
		prompt.WithHistoryFile("~/.wgo_history"),
		prompt.WithOutFunc(outFunc),
		prompt.WithCompletionSelectFunc(prompt.DefaultCompletionLSPSelectFunc),
	)
	m.prompt = p
	return m
}

type Wgo struct {
	prompt.BaseModel

	ctx       context.Context
	lspClient *lsp.LSPClient

	prompt *prompt.Prompt
}

func (m Wgo) Init() tea.Cmd {
	return textinput.Blink
}

func (m Wgo) View() string {
	return m.prompt.View()
}

func (m *Wgo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	// lsp 启动后，设置补全方法
	if m.lspClient != nil {
		m.prompt.CompletionFunc(func(input string, cursor int) []prompt.CompletionItem {
			return completionFunc(input, cursor, m.lspClient, m.ctx)
		})
	}
	model, cmd := m.prompt.Update(msg)
	m.prompt = model.(*prompt.Prompt)
	return m, cmd
}

func (m *Wgo) LspClient(client *lsp.LSPClient) {
	m.lspClient = client
}
