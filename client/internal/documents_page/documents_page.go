package documentspage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

type Document struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name,omitempty"`
	SubType string `json:"sub_type"`
	DateStr string `json:"date"`
	date    time.Time
}

func (d Document) FilterValue() string {
	return d.date.Format("2006-01-02 15:04")
}

func (d Document) Title() string {
	name := d.Name
	if name == "" {
		name = d.Type
	}

	return fmt.Sprintf("%s - %s", name, d.date.Format(time.DateOnly))
}

func (d Document) Description() string {
	return fmt.Sprintf("%s | %s", d.Type, d.SubType)
}

type DocumentsLoadedMsg struct {
	documents []Document
	err       error
}

type DocumentDownloadedMsg struct {
	filename string
	err      error
}

type DocumentsPage struct {
	BaseModel   basemodel.BaseModel
	documents   list.Model
	titleBar    string
	loaded      bool
	spinner     spinner.Model
	err         error
	downloading bool
	downloadMsg string
	Reloaded    bool
}

func New(client *http.Client, tokenStore *basemodel.TokenStore) DocumentsPage {
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
	l.Title = ""
	l.DisableQuitKeybindings()
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
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
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "download")),
		}
	}

	return DocumentsPage{
		BaseModel:   basemodel.BaseModel{Client: client, TokenStore: tokenStore},
		documents:   l,
		titleBar:    "DOCUMENTS",
		loaded:      false,
		spinner:     s,
		downloading: false,
		Reloaded:    true,
	}
}

func (d DocumentsPage) Init() tea.Cmd {
	return tea.Batch(
		d.spinner.Tick,
		d.loadDocuments,
	)
}

func (d DocumentsPage) loadDocuments() tea.Msg {
	body, err := requests.MakeRequest(http.MethodGet, requests.BaseURL+"/documents", nil, d.BaseModel.Client, d.BaseModel.TokenStore)
	if err != nil {
		return DocumentsLoadedMsg{err: err}
	}

	var documents []Document
	if err := json.Unmarshal(body, &documents); err != nil {
		return DocumentsLoadedMsg{err: err}
	}

	documents, err = d.parseDateToTime(documents)
	if err != nil {
		return DocumentsLoadedMsg{err: err}
	}

	return DocumentsLoadedMsg{documents: documents}
}

func (d DocumentsPage) parseDateToTime(documents []Document) ([]Document, error) {
	result := make([]Document, len(documents))
	for _, document := range documents {
		var err error
		document.date, err = time.Parse(time.DateOnly, document.DateStr)
		if err != nil {
			return nil, err
		}

		result = append(result, document)
	}

	return result, nil
}

func (d DocumentsPage) downloadDocument(document Document) tea.Cmd {
	return func() tea.Msg {
		// d.BaseModel.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// 	return http.ErrUseLastResponse
		// }

		body, err := requests.MakeRequest(http.MethodGet, requests.BaseURL+"/documents/download/"+document.ID, nil, d.BaseModel.Client, d.BaseModel.TokenStore)
		if err != nil {
			return DocumentDownloadedMsg{err: err}
		}

		filename := document.Name
		if filename == "" {
			filename = document.Type
		}

		if document.Type == "trade_confirmation_json" {
			filename = filename + ".pdf"
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return DocumentDownloadedMsg{err: err}
		}
		downloadsDir := filepath.Join(homeDir, "Downloads")

		if err := os.MkdirAll(downloadsDir, 0755); err != nil {
			return DocumentDownloadedMsg{err: err}
		}

		filePath := filepath.Join(downloadsDir, filename)
		file, err := os.Create(filePath)
		if err != nil {
			return DocumentDownloadedMsg{err: err}
		}
		defer file.Close()

		_, err = file.Write(body)
		if err != nil {
			return DocumentDownloadedMsg{err: err}
		}

		return DocumentDownloadedMsg{filename: filename}
	}
}

func (d DocumentsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case DocumentsLoadedMsg:
		d.loaded = true
		d.err = msg.err
		if msg.err == nil {
			slices.SortFunc(msg.documents, func(a Document, b Document) int {
				return -a.date.Compare(b.date)
			})

			items := make([]list.Item, len(msg.documents))
			for i, document := range msg.documents {
				items[i] = document
			}

			d.documents.SetItems(items)
			d.documents.SetSize(d.BaseModel.Width/4, (d.BaseModel.Height*3)/4)
		}

		return d, nil

	case DocumentDownloadedMsg:
		d.downloading = false
		if msg.err != nil {
			d.downloadMsg = fmt.Sprintf("Download failed: %v", msg.err)
		} else {
			d.downloadMsg = fmt.Sprintf("Downloaded %s to Downloads folder", msg.filename)
		}
		return d, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearDownloadMsg{}
		})

	case clearDownloadMsg:
		d.downloadMsg = ""
		return d, nil

	case spinner.TickMsg:
		if !d.loaded || d.downloading {
			d.spinner, cmd = d.spinner.Update(msg)
			return d, cmd
		}
		return d, nil

	case tea.KeyMsg:
		if d.downloading {
			return d, nil
		}

		switch msg.String() {
		case "q":
			return d, func() tea.Msg {
				return messages.QuitMsg{}
			}

		case "esc":
			return d, func() tea.Msg {
				return messages.SmartPageSwitchMsg{
					Page: messages.ProfilePageNumber,
				}
			}

		case "enter":
			if d.loaded && d.err == nil && len(d.documents.Items()) > 0 {
				selectedItem := d.documents.SelectedItem()
				if doc, ok := selectedItem.(Document); ok {
					d.downloading = true
					d.downloadMsg = fmt.Sprintf("Downloading %s...", doc.Title())
					return d, tea.Batch(
						d.spinner.Tick,
						d.downloadDocument(doc),
					)
				}
			}
		}
	}

	if d.loaded && d.err == nil && !d.downloading {
		d.documents, cmd = d.documents.Update(msg)
	}

	return d, cmd
}

type clearDownloadMsg struct{}

func (d DocumentsPage) View() string {
	cyan := lipgloss.Color("#00FFFF")
	purple := lipgloss.Color("#A020F0")
	red := lipgloss.Color("#D30000")
	green := lipgloss.Color("#0B6623")
	gray := lipgloss.Color("#626262")

	headerStyle := lipgloss.NewStyle().
		Foreground(cyan).
		Bold(true).
		Padding(0, 2).
		MarginBottom(1).
		Align(lipgloss.Center)
	header := "\n" + headerStyle.Render(d.titleBar) + "\n\n"

	if !d.loaded {
		return lipgloss.Place(d.BaseModel.Width, d.BaseModel.Height, lipgloss.Center, lipgloss.Center, d.spinner.View())
	}

	if d.err != nil {
		errorMsg := lipgloss.NewStyle().
			Foreground(red).
			Padding(1, 2).
			Render(fmt.Sprintf("Error loading documents: %v", d.err))
		help := lipgloss.NewStyle().
			Foreground(gray).
			Render("q: quit • esc: back")
		content := lipgloss.JoinVertical(lipgloss.Left, errorMsg, "", help)
		return header + content
	}

	if len(d.documents.Items()) == 0 {
		msg := lipgloss.NewStyle().
			Padding(1, 1).
			Render("No documents found.\nYour documents will appear here.")
		help := lipgloss.NewStyle().
			Foreground(gray).
			Render("q: quit • esc: back")
		content := lipgloss.JoinVertical(lipgloss.Left, msg, "", help)
		centerContent := lipgloss.Place(d.BaseModel.Width, d.BaseModel.Height-6, lipgloss.Center, lipgloss.Center, content)
		return header + centerContent
	}

	var statusBar string
	if d.downloading {
		statusBar = lipgloss.NewStyle().
			Foreground(cyan).
			Render(d.spinner.View() + " " + d.downloadMsg)
	} else if d.downloadMsg != "" {
		color := green
		if len(d.downloadMsg) > 0 && strings.SplitAfter(d.downloadMsg, " ")[0] == "Download" {
			color = red
		}

		statusBar = lipgloss.NewStyle().
			Foreground(color).
			Render(d.downloadMsg)
	}

	listView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Padding(0, 1).
		Render(d.documents.View())

	header = lipgloss.PlaceHorizontal(d.BaseModel.Width, lipgloss.Center, header)
	headerHeight := lipgloss.Height(header)

	availableHeight := d.BaseModel.Height - headerHeight

	content := lipgloss.Place(
		d.BaseModel.Width,
		availableHeight,
		lipgloss.Center,
		lipgloss.Center,
		listView,
	)

	if statusBar != "" {
		content = lipgloss.JoinVertical(lipgloss.Center, statusBar, "", content)
	}

	return header + content
}

func (d *DocumentsPage) Reload() {
	d.Reloaded = true
	d.loaded = false
	d.err = nil
	d.downloading = false
	d.downloadMsg = ""
	d.documents.SetItems([]list.Item{})
}
