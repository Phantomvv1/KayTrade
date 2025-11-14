package companypage

import (
	"net/http"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	tea "github.com/charmbracelet/bubbletea"
)

type CompanyPage struct {
	BaseModel   basemodel.BaseModel
	tabs        []string
	activeTab   int
	CompanyInfo *messages.CompanyInfo
}

func NewCompanyPage(client *http.Client) CompanyPage {
	return CompanyPage{
		BaseModel: basemodel.BaseModel{Client: client},
		tabs:      []string{"Main", "History"},
		activeTab: 0,
	}
}

func (c CompanyPage) Init() tea.Cmd {
	return nil
}

func (c CompanyPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return c, nil
}

func (c CompanyPage) View() string {
	return ""
}

func (c *CompanyPage) Reload() {
	c.activeTab = 0
	c.CompanyInfo = nil
}
