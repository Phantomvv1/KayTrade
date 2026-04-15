package viewtransferspage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	basemodel "github.com/Phantomvv1/KayTrade/client/internal/base_model"
	"github.com/Phantomvv1/KayTrade/client/internal/messages"
	"github.com/Phantomvv1/KayTrade/client/internal/requests"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type IRA struct {
	TaxYear                string `json:"tax_year"`
	FedWithholdingPct      string `json:"fed_withholding_pct"`
	FedWithholdingAmount   string `json:"fed_withholding_amount"`
	StateWithholdingPct    string `json:"state_withholding_pct"`
	StateWithholdingAmount string `json:"state_withholding_amount"`
	DistributionReason     string `json:"distribution_reason"`
}

type Transfer struct {
	ID                    string    `json:"id"`
	RelationshipID        string    `json:"relationship_id"`
	BankID                string    `json:"bank_id"`
	Type                  string    `json:"type"`
	Status                string    `json:"status"`
	Reason                string    `json:"reason"`
	Amount                string    `json:"amount"`
	Direction             string    `json:"direction"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	ExpiresAt             time.Time `json:"expires_at"`
	AdditionalInformation string    `json:"additional_information"`
	IRA                   *IRA      `json:"ira,omitempty"`
}

func (t Transfer) FilterValue() string {
	return t.CreatedAt.Format("2006-01-02 15:04")
}

func (t Transfer) Title() string {
	return fmt.Sprintf("%s %s  %s  $%s", t.CreatedAt.Format("2006-01-02 15:04"), t.Direction[:3], t.Type, t.Amount)
}

func (t Transfer) Description() string {
	statusSymbol := ""
	switch t.Status {
	case "COMPLETE":
		statusSymbol = "✓"
	case "QUEUED":
		statusSymbol = "⋯"
	case "REJECTED":
		statusSymbol = "✗"
	default:
		statusSymbol = "•"
	}
	return fmt.Sprintf("%s %s", statusSymbol, t.Status)
}

type TransfersLoadedMsg struct {
	transfers []Transfer
	err       error
}

type ViewTransfersPage struct {
	BaseModel basemodel.BaseModel
	transfers list.Model
	titleBar  string
	loaded    bool
	spinner   spinner.Model
	err       error
	Reloaded  bool
	filtering bool
	hasFilter bool
}

func New(client *http.Client, tokenStore *basemodel.TokenStore) ViewTransfersPage {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))

	delegate := list.NewDefaultDelegate()

	cyan := lipgloss.Color("#00FFFF")
	purple := lipgloss.Color("#A020F0")

	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(cyan).
		BorderForeground(purple)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#888888")).
		BorderForeground(purple)

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.DisableQuitKeybindings()
	l.Title = ""
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(cyan).
		Bold(true).
		Padding(0, 1)
	l.Styles.PaginationStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))
	l.Styles.HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		}
	}

	return ViewTransfersPage{
		BaseModel: basemodel.BaseModel{Client: client, TokenStore: tokenStore},
		transfers: l,
		titleBar:  "TRANSFERS",
		loaded:    false,
		spinner:   s,
		Reloaded:  true,
	}
}

func (t ViewTransfersPage) Init() tea.Cmd {
	return tea.Batch(
		t.spinner.Tick,
		t.loadTransfers,
	)
}

func (t ViewTransfersPage) loadTransfers() tea.Msg {
	body, err := requests.MakeRequest(http.MethodGet, requests.BaseURL+"/transfers", nil, t.BaseModel.Client, t.BaseModel.TokenStore)
	if err != nil {
		return TransfersLoadedMsg{err: err}
	}

	var transfers []Transfer
	if err := json.Unmarshal(body, &transfers); err != nil {
		return TransfersLoadedMsg{err: err}
	}

	return TransfersLoadedMsg{transfers: transfers}
}

func (t ViewTransfersPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case TransfersLoadedMsg:
		t.loaded = true
		t.err = msg.err
		if msg.err == nil {
			items := make([]list.Item, len(msg.transfers))
			for i, transfer := range msg.transfers {
				items[i] = transfer
			}

			t.transfers.SetItems(items)
			t.transfers.SetSize(t.BaseModel.Width/3, t.BaseModel.Height/2)
		}

		return t, nil

	case spinner.TickMsg:
		if !t.loaded {
			t.spinner, cmd = t.spinner.Update(msg)
			return t, cmd
		}
		return t, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return t, func() tea.Msg {
				return messages.QuitMsg{}
			}

		case "/":
			t.filtering = true

		case "enter":
			if t.filtering {
				t.filtering = false
				t.hasFilter = true
			}

		case "esc":
			if !t.filtering && !t.hasFilter {
				return t, func() tea.Msg {
					return messages.SmartPageSwitchMsg{
						Page: messages.BankRelationshipPageNumber,
					}
				}
			} else if t.filtering {
				t.filtering = false
			} else if t.hasFilter {
				t.hasFilter = false
			}
		}
	}

	if t.loaded && t.err == nil {
		t.transfers, cmd = t.transfers.Update(msg)
	}

	return t, cmd
}

func (t ViewTransfersPage) View() string {
	cyan := lipgloss.Color("#00FFFF")
	purple := lipgloss.Color("#A020F0")
	red := lipgloss.Color("#D30000")
	gray := lipgloss.Color("#626262")

	headerStyle := lipgloss.NewStyle().
		Foreground(cyan).
		Bold(true).
		Padding(0, 2).
		MarginBottom(1).
		Align(lipgloss.Center)
	header := "\n" + headerStyle.Render(t.titleBar) + "\n\n"

	if !t.loaded {
		return lipgloss.Place(t.BaseModel.Width, t.BaseModel.Height, lipgloss.Center, lipgloss.Center, t.spinner.View())
	}

	if t.err != nil {
		errorMsg := lipgloss.NewStyle().
			Foreground(red).
			Padding(1, 2).
			Render(fmt.Sprintf("Error loading transfers: %v", t.err))
		help := lipgloss.NewStyle().
			Foreground(gray).
			Render("q: quit")
		content := lipgloss.JoinVertical(lipgloss.Left, errorMsg, "", help)
		return header + content
	}

	if len(t.transfers.Items()) == 0 {
		msg := lipgloss.NewStyle().
			Padding(1, 1).
			Render("No transfers found.\nYour transfer history will appear here.")
		content := lipgloss.JoinVertical(lipgloss.Left, msg, "")
		centerContent := lipgloss.Place(t.BaseModel.Width, t.BaseModel.Height-6, lipgloss.Center, lipgloss.Center, content)
		return header + centerContent
	}

	// Wrap the list in a border
	listView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Padding(0, 1).
		Render(t.transfers.View())

	header = lipgloss.PlaceHorizontal(t.BaseModel.Width, lipgloss.Center, header)

	centeredList := lipgloss.Place(
		t.BaseModel.Width,
		t.BaseModel.Height-6,
		lipgloss.Center,
		lipgloss.Center,
		listView,
	)

	return header + centeredList
}

func (t *ViewTransfersPage) Reload() {
	t.loaded = false
	t.err = nil
	t.transfers.SetItems([]list.Item{})
	t.Reloaded = true
}
