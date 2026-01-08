package messages

const (
	LandingPageNumber = iota
	WatchlistPageNumber
	LoginPageNumber
	SearchPageNumber
	CompanyPageNumber
	BuyPageNumber
	TradingInfoPageNumber
	ProfilePageNumber
	SellPageNumber
	SignUpPageNumber
	OrderPageNumber
	PositionPageNumber
	BankRelationshipPageNumber
	BankRelationshipCreationPageNumber
	TransfersPageNumber
	ErrorPageNumber
)

type PageSwitchMsg struct {
	Page               int
	Err                error
	Company            *CompanyInfo
	Symbol             string
	MaxQuantity        float64
	Order              *Order
	Position           *Position
	FundingInformation *FundingInformation
}

type LoginSuccessMsg struct {
	Token string
	Page  int
}

type PageSwitchWithoutInitMsg struct {
	Page int
}

type ReloadMsg struct {
	Page int
}

type SmartPageSwitchMsg struct {
	Page int
}

type QuitMsg struct{}

type CompanyInfo struct {
	Symbol       string  `json:"symbol"`
	OpeningPrice float64 `json:"opening_price,omitempty"`
	ClosingPrice float64 `json:"closing_price,omitempty"`
	Logo         string  `json:"logo"`
	Name         string  `json:"name"`
	History      string  `json:"history"`
	IsNSFW       bool    `json:"isNsfw"`
	Description  string  `json:"description"`
	FoundedYear  int     `json:"founded_year"`
	Domain       string  `json:"domain"`
}

type Order struct {
	AssetClass     string  `json:"asset_class"`
	AssetID        string  `json:"asset_id"`
	CanceledAt     string  `json:"canceled_at"`
	CreatedAt      string  `json:"created_at"`
	ExpiredAt      string  `json:"expired_at"`
	ExpiresAt      string  `json:"expires_at"`
	FailedAt       string  `json:"failed_at"`
	FilledAt       string  `json:"filled_at"`
	FilledAvgPrice string  `json:"filled_avg_price"`
	FilledQty      string  `json:"filled_qty"`
	ID             string  `json:"id"`
	LimitPrice     *string `json:"limit_price"`
	Notional       *string `json:"notional"`
	OrderType      string  `json:"order_type"`
	PositionIntent string  `json:"position_intent"`
	Quantity       string  `json:"qty"`
	Side           string  `json:"side"`
	Status         string  `json:"status"`
	StopPrice      *string `json:"stop_price"`
	SubmittedAt    string  `json:"submitted_at"`
	Symbol         string  `json:"symbol"`
	TimeInForce    string  `json:"time_in_force"`
	TrailPercent   *string `json:"trail_percent"`
	TrailPrice     *string `json:"trail_price"`
	Type           string  `json:"type"`
	UpdatedAt      string  `json:"updated_at"`
}

type Position struct {
	AssetClass             string `json:"asset_class"`
	AssetID                string `json:"asset_id"`
	AssetMarginable        bool   `json:"asset_marginable"`
	AvgEntryPrice          string `json:"avg_entry_price"`
	ChangeToday            string `json:"change_today"`
	CostBasis              string `json:"cost_basis"`
	CurrentPrice           string `json:"current_price"`
	Exchange               string `json:"exchange"`
	LastdayPrice           string `json:"lastday_price"`
	MarketValue            string `json:"market_value"`
	Qty                    string `json:"qty"`
	QtyAvailable           string `json:"qty_available"`
	Side                   string `json:"side"`
	Symbol                 string `json:"symbol"`
	UnrealizedIntradayPL   string `json:"unrealized_intraday_pl"`
	UnrealizedIntradayPLPC string `json:"unrealized_intraday_plpc"`
	UnrealizedPL           string `json:"unrealized_pl"`
	UnrealizedPLPC         string `json:"unrealized_plpc"`
}

type FundingInformation struct {
	TransferType string

	// Only 1 of RelationshipId and BankId will be fileld
	RelationshipId string
	BankId         string
}
