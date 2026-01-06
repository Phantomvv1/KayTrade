package bankrelationshipcreationpage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	"github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	normalBankType = "bank"
	bankTypeAch    = "ach"
	inputWidth     = 27
)

type bankInputs struct {
	name            textinput.Model
	bankCode        textinput.Model
	bankCodeType    []string
	bankCodeTypeIdx int
	accountNumber   textinput.Model
	country         textinput.Model // required if BCT = "BIC"
	stateProvince   textinput.Model // required if BCT = "BIC"
	city            textinput.Model // required if BCT = "BIC"
	streetAddress   textinput.Model // required if BCT = "BIC"
}

type achInputs struct {
	accountOwnerName   textinput.Model
	bankAccountType    []string
	bankAccountTypeIdx int
	bankAccountNumber  textinput.Model
	bankRoutingNumber  textinput.Model
	nickname           textinput.Model // not required
}

type BankRelationshipCreation struct {
	BaseModel            basemodel.BaseModel
	bankRelationshipType string
	bankInputs           bankInputs
	achInputs            achInputs
	cursor               int
	totalFields          int
	typing               bool
	err                  string
	success              string
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BB88FF")).
			Width(25).
			Align(lipgloss.Center)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)
)

func NewBankInputs() bankInputs {
	name := textinput.New()
	name.Placeholder = "Name"
	name.Width = inputWidth
	name.CharLimit = 50

	bankCode := textinput.New()
	bankCode.Placeholder = "Bank code"
	bankCode.Width = inputWidth
	bankCode.CharLimit = 50

	accountNumber := textinput.New()
	accountNumber.Placeholder = "Account number"
	accountNumber.Width = inputWidth
	accountNumber.CharLimit = 50

	country := textinput.New()
	country.Placeholder = "Country"
	country.Width = inputWidth
	country.CharLimit = 50

	stateProvince := textinput.New()
	stateProvince.Placeholder = "State/Province"
	stateProvince.Width = inputWidth
	stateProvince.CharLimit = 50

	city := textinput.New()
	city.Placeholder = "City"
	city.Width = inputWidth
	city.CharLimit = 50

	streetAddress := textinput.New()
	streetAddress.Placeholder = "Street address"
	streetAddress.Width = inputWidth
	streetAddress.CharLimit = 50

	return bankInputs{
		name:            name,
		bankCode:        bankCode,
		bankCodeType:    []string{"ABA", "BIC"}, // ABA - domestic, BIC - international
		bankCodeTypeIdx: 0,
		accountNumber:   accountNumber,
		country:         country,
		stateProvince:   stateProvince,
		city:            city,
		streetAddress:   streetAddress,
	}
}

func NewAchInputs() achInputs {
	accountOwnerName := textinput.New()
	accountOwnerName.Placeholder = "Account owner name"
	accountOwnerName.Width = inputWidth
	accountOwnerName.CharLimit = 50

	bankAccountNumber := textinput.New()
	bankAccountNumber.Placeholder = "Bank account number"
	bankAccountNumber.Width = inputWidth
	bankAccountNumber.CharLimit = 50

	bankRoutingNumber := textinput.New()
	bankRoutingNumber.Placeholder = "Bank routing number"
	bankRoutingNumber.Width = inputWidth
	bankRoutingNumber.CharLimit = 50

	nickname := textinput.New()
	nickname.Placeholder = "Nickname (optional)"
	nickname.Width = inputWidth
	nickname.CharLimit = 50

	return achInputs{
		accountOwnerName:   accountOwnerName,
		bankAccountType:    []string{"CHECKING", "SAVINGS"},
		bankAccountTypeIdx: 0,
		bankAccountNumber:  bankAccountNumber,
		bankRoutingNumber:  bankRoutingNumber,
		nickname:           nickname,
	}
}

func NewBankRelationship(client *http.Client, tokenStore *basemodel.TokenStore) BankRelationshipCreation {
	return BankRelationshipCreation{
		BaseModel:            basemodel.BaseModel{Client: client, TokenStore: tokenStore},
		bankRelationshipType: normalBankType,
		bankInputs:           NewBankInputs(),
		achInputs:            NewAchInputs(),
		typing:               true,
		cursor:               0,
	}
}

func (b BankRelationshipCreation) Init() tea.Cmd {
	return textinput.Blink
}

func (b *BankRelationshipCreation) calculateTotalFields() int {
	if b.bankRelationshipType == normalBankType {
		count := 4 // name, bankCode, bankCodeType, accountNumber
		if b.bankInputs.bankCodeType[b.bankInputs.bankCodeTypeIdx] == "BIC" {
			count += 4 // country, stateProvince, city, streetAddress
		}
		return count
	} else {
		return 5 // accountOwnerName, bankAccountType, bankAccountNumber, bankRoutingNumber, nickname
	}
}

func (b BankRelationshipCreation) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	b.totalFields = b.calculateTotalFields()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if b.typing {
			switch msg.String() {
			case "esc":
				b.typing = false
				return b, nil

			case "ctrl+j", "down":
				b.err = ""
				b.success = ""
				b.cursor++
				if b.cursor >= b.totalFields {
					b.cursor = 0
				}
				return b, nil

			case "ctrl+k", "up":
				b.err = ""
				b.success = ""
				b.cursor--
				if b.cursor < 0 {
					b.cursor = b.totalFields - 1
				}
				return b, nil

			case "h", "left":
				b.err = ""
				b.success = ""
				if b.bankRelationshipType == normalBankType && b.cursor == 2 {
					b.bankInputs.bankCodeTypeIdx--
					if b.bankInputs.bankCodeTypeIdx < 0 {
						b.bankInputs.bankCodeTypeIdx = len(b.bankInputs.bankCodeType) - 1
					}
					b.totalFields = b.calculateTotalFields()
					return b, nil
				} else if b.bankRelationshipType == bankTypeAch && b.cursor == 1 {
					b.achInputs.bankAccountTypeIdx--
					if b.achInputs.bankAccountTypeIdx < 0 {
						b.achInputs.bankAccountTypeIdx = len(b.achInputs.bankAccountType) - 1
					}
					return b, nil
				}

			case "l", "right":
				b.err = ""
				b.success = ""
				if b.bankRelationshipType == normalBankType && b.cursor == 2 {
					b.bankInputs.bankCodeTypeIdx++
					if b.bankInputs.bankCodeTypeIdx >= len(b.bankInputs.bankCodeType) {
						b.bankInputs.bankCodeTypeIdx = 0
					}
					b.totalFields = b.calculateTotalFields()
					return b, nil
				} else if b.bankRelationshipType == bankTypeAch && b.cursor == 1 {
					b.achInputs.bankAccountTypeIdx++
					if b.achInputs.bankAccountTypeIdx >= len(b.achInputs.bankAccountType) {
						b.achInputs.bankAccountTypeIdx = 0
					}
					return b, nil
				}

			case "enter":
				b.err = ""
				b.success = ""
				if err := b.submitRelationship(); err != nil {
					b.err = err.Error()
				} else {
					b.success = "Bank relationship created successfully!"
				}
				return b, nil
			}
		} else {
			switch msg.String() {
			case "q", "ctrl+c":
				return b, func() tea.Msg {
					return messages.QuitMsg{}
				}

			case "esc":
				return b, func() tea.Msg {
					return messages.SmartPageSwitchMsg{
						Page: messages.BankRelationshipPageNumber,
					}
				}

			case "enter":
				b.typing = true
				return b, nil

			case "s", "S":
				if b.bankRelationshipType == normalBankType {
					b.bankRelationshipType = bankTypeAch
				} else {
					b.bankRelationshipType = normalBankType
				}

				return b, nil
			}
		}
	}

	// Update the focused input
	if b.bankRelationshipType == normalBankType {
		b.updateBankInput(msg, &cmd)
	} else {
		b.updateAchInput(msg, &cmd)
	}

	return b, cmd
}

func (b *BankRelationshipCreation) updateBankInput(msg tea.Msg, cmd *tea.Cmd) {
	idx := b.cursor
	if idx == 0 {
		b.bankInputs.name.Focus()
		b.bankInputs.name, *cmd = b.bankInputs.name.Update(msg)
		b.bankInputs.name.Blur()
	} else if idx == 1 {
		b.bankInputs.bankCode.Focus()
		b.bankInputs.bankCode, *cmd = b.bankInputs.bankCode.Update(msg)
		b.bankInputs.bankCode.Blur()
	} else if idx == 2 {
		// Slider - no update needed
	} else if idx == 3 {
		b.bankInputs.accountNumber.Focus()
		b.bankInputs.accountNumber, *cmd = b.bankInputs.accountNumber.Update(msg)
		b.bankInputs.accountNumber.Blur()
	} else if idx == 4 && b.bankInputs.bankCodeType[b.bankInputs.bankCodeTypeIdx] == "BIC" {
		b.bankInputs.country.Focus()
		b.bankInputs.country, *cmd = b.bankInputs.country.Update(msg)
		b.bankInputs.country.Blur()
	} else if idx == 5 && b.bankInputs.bankCodeType[b.bankInputs.bankCodeTypeIdx] == "BIC" {
		b.bankInputs.stateProvince.Focus()
		b.bankInputs.stateProvince, *cmd = b.bankInputs.stateProvince.Update(msg)
		b.bankInputs.stateProvince.Blur()
	} else if idx == 6 && b.bankInputs.bankCodeType[b.bankInputs.bankCodeTypeIdx] == "BIC" {
		b.bankInputs.city.Focus()
		b.bankInputs.city, *cmd = b.bankInputs.city.Update(msg)
		b.bankInputs.city.Blur()
	} else if idx == 7 && b.bankInputs.bankCodeType[b.bankInputs.bankCodeTypeIdx] == "BIC" {
		b.bankInputs.streetAddress.Focus()
		b.bankInputs.streetAddress, *cmd = b.bankInputs.streetAddress.Update(msg)
		b.bankInputs.streetAddress.Blur()
	}
}

func (b *BankRelationshipCreation) updateAchInput(msg tea.Msg, cmd *tea.Cmd) {
	switch b.cursor {
	case 0:
		b.achInputs.accountOwnerName.Focus()
		b.achInputs.accountOwnerName, *cmd = b.achInputs.accountOwnerName.Update(msg)
		b.achInputs.accountOwnerName.Blur()
	case 2:
		b.achInputs.bankAccountNumber.Focus()
		b.achInputs.bankAccountNumber, *cmd = b.achInputs.bankAccountNumber.Update(msg)
		b.achInputs.bankAccountNumber.Blur()
	case 3:
		b.achInputs.bankRoutingNumber.Focus()
		b.achInputs.bankRoutingNumber, *cmd = b.achInputs.bankRoutingNumber.Update(msg)
		b.achInputs.bankRoutingNumber.Blur()
	case 4:
		b.achInputs.nickname.Focus()
		b.achInputs.nickname, *cmd = b.achInputs.nickname.Update(msg)
		b.achInputs.nickname.Blur()

	}
}

func (b BankRelationshipCreation) View() string {
	var header string
	if b.bankRelationshipType == normalBankType {
		header = titleStyle.Render("🏦 Create Bank Relationship")
	} else {
		header = titleStyle.Render("🏦 Create ACH Relationship")
	}

	var fields []string

	if b.bankRelationshipType == normalBankType {
		fields = b.renderBankFields()
	} else {
		fields = b.renderAchFields()
	}

	content := lipgloss.JoinVertical(lipgloss.Center, fields...)

	if b.err != "" {
		content = lipgloss.JoinVertical(lipgloss.Center, content, "", errorStyle.Render("❌ "+b.err))
	}
	if b.success != "" {
		content = lipgloss.JoinVertical(lipgloss.Center, content, "", successStyle.Render("✓ "+b.success))
	}

	help := helpStyle.Render("ctrl+j/k/↑/↓: navigate | ctrl+h/l/←/→: change option | esc: stop typing/back | s: switch bank type | enter: submit/type | q: quit")

	headerHeight := lipgloss.Height(header)
	contentHeight := lipgloss.Height(content)
	helpHeight := lipgloss.Height(help)

	centeredHeader := lipgloss.Place(b.BaseModel.Width, headerHeight, lipgloss.Center, lipgloss.Top, header)
	centeredContent := lipgloss.Place(b.BaseModel.Width, contentHeight, lipgloss.Center, lipgloss.Top, content)
	centeredHelp := lipgloss.Place(b.BaseModel.Width, helpHeight, lipgloss.Center, lipgloss.Top, help)

	spaceBetween := 16 - b.calculateTotalFields()
	if spaceBetween < 0 {
		spaceBetween = 0
	}

	finalView := lipgloss.JoinVertical(
		lipgloss.Center,
		centeredHeader,
		strings.Repeat("\n", spaceBetween),
		centeredContent,
		"",
		centeredHelp,
	)

	return finalView
}

func (b BankRelationshipCreation) renderBankFields() []string {
	var fields []string
	idx := 0

	// Name
	if b.cursor == idx {
		b.bankInputs.name.Focus()
	} else {
		b.bankInputs.name.Blur()
	}
	fields = append(fields, b.renderField("Name", b.bankInputs.name.View(), b.cursor == idx))
	idx++

	// Bank Code
	if b.cursor == idx {
		b.bankInputs.bankCode.Focus()
	} else {
		b.bankInputs.bankCode.Blur()
	}
	fields = append(fields, b.renderField("Bank Code", b.bankInputs.bankCode.View(), b.cursor == idx))
	idx++

	// Bank Code Type (slider)
	fields = append(fields, b.renderField("Bank Code Type", b.renderSlider(b.bankInputs.bankCodeType, b.bankInputs.bankCodeTypeIdx, b.cursor == idx), b.cursor == idx))
	idx++

	// Account Number
	if b.cursor == idx {
		b.bankInputs.accountNumber.Focus()
	} else {
		b.bankInputs.accountNumber.Blur()
	}
	fields = append(fields, b.renderField("Account Number", b.bankInputs.accountNumber.View(), b.cursor == idx))
	idx++

	// BIC-specific fields
	if b.bankInputs.bankCodeType[b.bankInputs.bankCodeTypeIdx] == "BIC" {
		// Country
		if b.cursor == idx {
			b.bankInputs.country.Focus()
		} else {
			b.bankInputs.country.Blur()
		}
		fields = append(fields, b.renderField("Country", b.bankInputs.country.View(), b.cursor == idx))
		idx++

		// State/Province
		if b.cursor == idx {
			b.bankInputs.stateProvince.Focus()
		} else {
			b.bankInputs.stateProvince.Blur()
		}
		fields = append(fields, b.renderField("State/Province", b.bankInputs.stateProvince.View(), b.cursor == idx))
		idx++

		// City
		if b.cursor == idx {
			b.bankInputs.city.Focus()
		} else {
			b.bankInputs.city.Blur()
		}
		fields = append(fields, b.renderField("City", b.bankInputs.city.View(), b.cursor == idx))
		idx++

		// Street Address
		if b.cursor == idx {
			b.bankInputs.streetAddress.Focus()
		} else {
			b.bankInputs.streetAddress.Blur()
		}
		fields = append(fields, b.renderField("Street Address", b.bankInputs.streetAddress.View(), b.cursor == idx))
	}

	return fields
}

func (b BankRelationshipCreation) renderAchFields() []string {
	var fields []string
	idx := 0

	// Account Owner Name
	if b.cursor == idx {
		b.achInputs.accountOwnerName.Focus()
	} else {
		b.achInputs.accountOwnerName.Blur()
	}
	fields = append(fields, b.renderField("Account Owner Name", b.achInputs.accountOwnerName.View(), b.cursor == idx))
	idx++

	// Bank Account Type (slider)
	fields = append(fields, b.renderField("Bank Account Type", b.renderSlider(b.achInputs.bankAccountType, b.achInputs.bankAccountTypeIdx, b.cursor == idx), b.cursor == idx))
	idx++

	// Bank Account Number
	if b.cursor == idx {
		b.achInputs.bankAccountNumber.Focus()
	} else {
		b.achInputs.bankAccountNumber.Blur()
	}
	fields = append(fields, b.renderField("Bank Account Number", b.achInputs.bankAccountNumber.View(), b.cursor == idx))
	idx++

	// Bank Routing Number
	if b.cursor == idx {
		b.achInputs.bankRoutingNumber.Focus()
	} else {
		b.achInputs.bankRoutingNumber.Blur()
	}
	fields = append(fields, b.renderField("Bank Routing Number", b.achInputs.bankRoutingNumber.View(), b.cursor == idx))
	idx++

	// Nickname (optional)
	if b.cursor == idx {
		b.achInputs.nickname.Focus()
	} else {
		b.achInputs.nickname.Blur()
	}
	fields = append(fields, b.renderField("Nickname (optional)", b.achInputs.nickname.View(), b.cursor == idx))

	return fields
}

func (b BankRelationshipCreation) renderField(label, value string, focused bool) string {
	styledLabel := labelStyle.Render(label)

	var fieldStyle lipgloss.Style
	if focused {
		if b.typing {
			fieldStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FFFF")).
				Background(lipgloss.Color("#2a2a4e")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00FFFF")).
				Width(32).
				Align(lipgloss.Center)
		} else {
			fieldStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#666666")).
				Width(32).
				Align(lipgloss.Center)
		}
	} else {
		fieldStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666666")).
			Width(32).
			Align(lipgloss.Center)
	}

	styledValue := fieldStyle.Render(value)

	return lipgloss.JoinVertical(lipgloss.Center, styledLabel, styledValue)
}

func (b BankRelationshipCreation) renderSlider(options []string, selectedIdx int, focused bool) string {
	selected := strings.ToUpper(options[selectedIdx])

	if focused {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true).
			Render("◀ " + selected + " ▶")
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Render("◀ " + selected + " ▶")
}

func (b *BankRelationshipCreation) submitRelationship() error {
	data := make(map[string]any)

	if b.bankRelationshipType == normalBankType {
		return b.submitBankRelationship(data)
	} else {
		return b.submitAchRelationship(data)
	}
}

func (b *BankRelationshipCreation) submitBankRelationship(data map[string]any) error {
	name := strings.TrimSpace(b.bankInputs.name.Value())
	if name == "" {
		return errors.New("name is required")
	}
	data["name"] = name

	bankCode := strings.TrimSpace(b.bankInputs.bankCode.Value())
	if bankCode == "" {
		return errors.New("bank code is required")
	}
	data["bank_code"] = bankCode

	data["bank_code_type"] = b.bankInputs.bankCodeType[b.bankInputs.bankCodeTypeIdx]

	accountNumber := strings.TrimSpace(b.bankInputs.accountNumber.Value())
	if accountNumber == "" {
		return errors.New("account number is required")
	}
	data["account_number"] = accountNumber

	// BIC-specific fields
	if b.bankInputs.bankCodeType[b.bankInputs.bankCodeTypeIdx] == "BIC" {
		country := strings.TrimSpace(b.bankInputs.country.Value())
		if country == "" {
			return errors.New("country is required for BIC")
		}
		data["country"] = country

		stateProvince := strings.TrimSpace(b.bankInputs.stateProvince.Value())
		if stateProvince == "" {
			return errors.New("state/province is required for BIC")
		}
		data["state_province"] = stateProvince

		city := strings.TrimSpace(b.bankInputs.city.Value())
		if city == "" {
			return errors.New("city is required for BIC")
		}
		data["city"] = city

		streetAddress := strings.TrimSpace(b.bankInputs.streetAddress.Value())
		if streetAddress == "" {
			return errors.New("street address is required for BIC")
		}
		data["street_address"] = streetAddress
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	_, err = requests.MakeRequest(http.MethodPost, requests.BaseURL+"/funding", bytes.NewReader(jsonData), b.BaseModel.Client, b.BaseModel.TokenStore)
	if err != nil {
		return fmt.Errorf("failed to create bank relationship: %w", err)
	}

	return nil
}

func (b *BankRelationshipCreation) submitAchRelationship(data map[string]any) error {
	accountOwnerName := strings.TrimSpace(b.achInputs.accountOwnerName.Value())
	if accountOwnerName == "" {
		return errors.New("account owner name is required")
	}
	data["account_owner_name"] = accountOwnerName

	data["bank_account_type"] = b.achInputs.bankAccountType[b.achInputs.bankAccountTypeIdx]

	bankAccountNumber := strings.TrimSpace(b.achInputs.bankAccountNumber.Value())
	if bankAccountNumber == "" {
		return errors.New("bank account number is required")
	}
	data["bank_account_number"] = bankAccountNumber

	bankRoutingNumber := strings.TrimSpace(b.achInputs.bankRoutingNumber.Value())
	if bankRoutingNumber == "" {
		return errors.New("bank routing number is required")
	}
	data["bank_routing_number"] = bankRoutingNumber

	// Nickname is optional
	nickname := strings.TrimSpace(b.achInputs.nickname.Value())
	if nickname != "" {
		data["nickname"] = nickname
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	_, err = requests.MakeRequest(http.MethodPost, requests.BaseURL+"/funding/ach", bytes.NewReader(jsonData), b.BaseModel.Client, b.BaseModel.TokenStore)
	if err != nil {
		return fmt.Errorf("failed to create ACH relationship: %w", err)
	}

	return nil
}
