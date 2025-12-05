package buypage

import (
	"net/http"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type TakeProfit struct {
	limitPrice textinput.Model
}

type StopLoss struct {
	stopPrice  textinput.Model
	limitPrice textinput.Model
}

type BuyPage struct {
	BaseModel        basemodel.BaseModel
	Symbol           string
	quantity         float64 // only when purchaseType == "market" && timeInForce == "day"
	side             string
	purchaseType     []string
	timeInForce      []string
	additionalFields map[string][]textinput.Model
	takeProfit       TakeProfit
	stopLoss         StopLoss
}

func NewBuyPage(client *http.Client) BuyPage {
	limitPrice := textinput.New()
	limitPrice.Placeholder = "Limit price"
	limitPrice.Width = 12

	stopPrice := textinput.New()
	stopPrice.Placeholder = "Stop price"
	stopPrice.Width = 12

	trailPrice := textinput.New()
	trailPrice.Placeholder = "Trail price"
	trailPrice.Width = 12

	trailPercent := textinput.New()
	trailPercent.Placeholder = "Trail percent"
	trailPercent.Width = 12

	takeProfit := textinput.New()
	takeProfit.Placeholder = "Limit price"
	takeProfit.Width = 12

	stopLossStopPrice := textinput.New()
	takeProfit.Placeholder = "Stop price"
	takeProfit.Width = 12

	stopLossLimitPrice := textinput.New()
	stopLossLimitPrice.Placeholder = "Limit price"
	stopLossLimitPrice.Width = 12

	stopLoss := StopLoss{
		stopPrice:  stopLossStopPrice,
		limitPrice: stopLossLimitPrice,
	}

	return BuyPage{
		BaseModel:    basemodel.BaseModel{Client: client},
		quantity:     1,
		side:         "buy",
		purchaseType: []string{"market", "limit", "stop", "stop_limit", "trailing_stop"},
		timeInForce:  []string{"day", "gtc", "opg", "cls", "ioc", "fok"},
		additionalFields: map[string][]textinput.Model{
			"limit": []textinput.Model{
				limitPrice,
			},

			"stop_limit": []textinput.Model{
				limitPrice,
				stopPrice,
			},

			"stop": []textinput.Model{
				stopPrice,
			},

			"trailing_stop": []textinput.Model{
				trailPrice,
				trailPercent,
			},
		},

		takeProfit: TakeProfit{
			limitPrice: takeProfit,
		},

		stopLoss: stopLoss,
	}
}

func (b BuyPage) Init() tea.Cmd {
	return nil
}

func (b BuyPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return b, nil
}

func (b BuyPage) View() string {
	return ""
}
