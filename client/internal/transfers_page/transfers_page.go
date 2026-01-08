package transferspage

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	"github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type TransfersPage struct {
	BaseModel          basemodel.BaseModel
	amount             textinput.Model
	FundingInformation *messages.FundingInformation
	direction          []string
	directionIdx       int
	cursor             int
	typing             bool
}

func NewTransfersPage(client *http.Client, tokenStore *basemodel.TokenStore) TransfersPage {
	amount := textinput.New()
	amount.Focus()
	amount.Placeholder = "Amount"
	amount.Width = 27
	amount.CharLimit = 20

	return TransfersPage{
		BaseModel: basemodel.BaseModel{Client: client, TokenStore: tokenStore},
		amount:    amount,
		direction: []string{"INCOMING", "OUTGOING"},
	}
}

func (t TransfersPage) Init() tea.Cmd {
	return textinput.Blink
}

func (t TransfersPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return t, nil
}

func (t TransfersPage) View() string {
	return ""
}

func (t *TransfersPage) Submit() error {
	data := make(map[string]any)
	data["transfer_type"] = t.FundingInformation.TransferType
	data["direction"] = t.direction[t.directionIdx]
	data["timing"] = "immediate"

	if t.FundingInformation.TransferType == "ach" && t.FundingInformation.RelationshipId != "" {
		data["relationship_id"] = t.FundingInformation.RelationshipId
	} else if t.FundingInformation.TransferType == "wire" && t.FundingInformation.BankId != "" {
		data["bank_id"] = t.FundingInformation.BankId
	} else {
		return errors.New("Error no relationship or bank found to do this transfer with")
	}

	amountString := strings.TrimSpace(t.amount.Value())
	amount, err := strconv.Atoi(amountString)
	if err != nil {
		return err
	}

	data["amount"] = amount

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(jsonData)

	_, err = requests.MakeRequest(http.MethodPost, requests.BaseURL+"transfers", reader, t.BaseModel.Client, t.BaseModel.TokenStore)

	return err
}
