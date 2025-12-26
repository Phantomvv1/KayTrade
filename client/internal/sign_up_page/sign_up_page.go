package signuppage

import (
	"bytes"
	"encoding/json"
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
	contactPage = iota
	identityPage
	documentsPage
	trustedContactPage
)

const (
	inputWidth = 27
)

type Contact struct {
	EmailAddress  string   `json:"email_address"`
	PhoneNumber   string   `json:"phone_number"`
	StreetAddress []string `json:"street_address"`
	Unit          string   `json:"unit,omitempty"`
	City          string   `json:"city"`
	State         string   `json:"state,omitempty"`
	PostalCode    string   `json:"postal_code,omitempty"`
}

type ContactInputs struct {
	emailAddress  textinput.Model
	phoneNumber   textinput.Model
	streetAddress textinput.Model
	unit          textinput.Model
	city          textinput.Model
	state         textinput.Model
	postalCode    textinput.Model
}

type Identity struct {
	GivenName             string   `json:"given_name"`
	FamilyName            string   `json:"family_name"`
	DateOfBirth           string   `json:"date_of_birth"`
	TaxID                 string   `json:"tax_id,omitempty"`
	TaxIDType             string   `json:"tax_id_type,omitempty"`
	CountryOfCitizenship  string   `json:"country_of_citizenship,omitempty"`
	CountryOfBirth        string   `json:"country_of_birth,omitempty"`
	CountryOfTaxResidence string   `json:"country_of_tax_residence"`
	FundingSource         []string `json:"funding_source"`
}

type IdentityInputs struct {
	givenName             textinput.Model
	familyName            textinput.Model
	dateOfBirth           textinput.Model
	taxID                 textinput.Model
	taxIDType             textinput.Model
	countryOfCitizenship  textinput.Model
	countryOfBirth        textinput.Model
	countryOfTaxResidence textinput.Model
}

type Disclosures struct {
	IsControlPerson             bool `json:"is_control_person"`
	IsAffiliatedExchangeOrFinra bool `json:"is_affiliated_exchange_or_finra"`
	IsPoliticallyExposed        bool `json:"is_politically_exposed"`
	ImmediateFamilyExposed      bool `json:"immediate_family_exposed"`
}

type Document struct {
	DocumentType    string `json:"document_type,omitempty"`
	DocumentSubType string `json:"document_sub_type,omitempty"`
	Content         string `json:"content,omitempty"`
	MimeType        string `json:"mime_type,omitempty"`
}

type DocumentInputs struct {
	documentType    textinput.Model
	documentSubType textinput.Model
	content         textinput.Model
	mimeType        textinput.Model
}

type TrustedContact struct {
	GivenName    string `json:"given_name"`
	FamilyName   string `json:"family_name"`
	EmailAddress string `json:"email_address"`
}

type TrustedContactInputs struct {
	givenName    textinput.Model
	familyName   textinput.Model
	emailAddress textinput.Model
}

type AccountInfo struct {
	Contact        Contact         `json:"contact"`
	Identity       Identity        `json:"identity"`
	Disclosures    Disclosures     `json:"disclosures"`
	Documents      []Document      `json:"documents,omitempty"`
	TrustedContact *TrustedContact `json:"trusted_contact,omitempty"`
	EnabledAssets  []string        `json:"enabled_assets"`
	Password       string          `json:"password"`
}

type SignUpPage struct {
	BaseModel            basemodel.BaseModel
	password             textinput.Model
	accountInfo          AccountInfo
	contactInputs        ContactInputs
	identityInputs       IdentityInputs
	documentInputs       DocumentInputs
	trustedContactInputs TrustedContactInputs
	fundingSourceOptions []string
	currentPage          int
	cursor               int
	typing               bool
	fundingCursor        int
	fundingSelected      map[int]bool
	err                  string
	success              string
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	formRowStyle = lipgloss.NewStyle().
			Align(lipgloss.Center)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BB88FF")).
			Bold(true).
			Align(lipgloss.Center).
			MarginBottom(1)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666666")).
			Width(32)

	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Background(lipgloss.Color("#2a2a4e")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FFFF")).
			Width(32)

	fundingIdleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF"))

	fundingFocusedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FFFF")).
				Bold(true)

	fundingSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#BB88FF")).
				Bold(true)
)

func NewSignUpPage(client *http.Client) SignUpPage {
	password := textinput.New()
	password.Placeholder = "Email address"
	password.Width = inputWidth
	password.CharLimit = 50
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '•'

	return SignUpPage{
		BaseModel: basemodel.BaseModel{Client: client},
		accountInfo: AccountInfo{
			Disclosures: Disclosures{
				IsControlPerson:             false,
				IsAffiliatedExchangeOrFinra: false,
				IsPoliticallyExposed:        false,
				ImmediateFamilyExposed:      false,
			},
			EnabledAssets: []string{"us_equity"},
		},
		password:             password,
		contactInputs:        newContactInputs(),
		identityInputs:       newIdentityInputs(),
		documentInputs:       newDocumentInputs(),
		trustedContactInputs: newTrustedContactInputs(),
		fundingSourceOptions: []string{"employment_income", "investments", "inheritance", "business_income", "savings", "family"},
		fundingCursor:        0,
		fundingSelected:      make(map[int]bool),
		currentPage:          contactPage,
		typing:               true,
		cursor:               0,
	}
}

func newContactInputs() ContactInputs {
	emailAddress := textinput.New()
	emailAddress.Placeholder = "Email address"
	emailAddress.Width = inputWidth
	emailAddress.CharLimit = 50

	phoneNumber := textinput.New()
	phoneNumber.Placeholder = "Phone number (with country code)"
	phoneNumber.Width = inputWidth
	phoneNumber.CharLimit = 20

	streetAddress := textinput.New()
	streetAddress.Placeholder = "Street address"
	streetAddress.Width = inputWidth
	streetAddress.CharLimit = 60

	unit := textinput.New()
	unit.Placeholder = "Unit (optional)"
	unit.Width = inputWidth
	unit.CharLimit = 10

	city := textinput.New()
	city.Placeholder = "City"
	city.Width = inputWidth
	city.CharLimit = 30

	state := textinput.New()
	state.Placeholder = "State (optional)"
	state.Width = inputWidth
	state.CharLimit = 20

	postalCode := textinput.New()
	postalCode.Placeholder = "Postal code (optional)"
	postalCode.Width = inputWidth
	postalCode.CharLimit = 15

	return ContactInputs{
		emailAddress:  emailAddress,
		phoneNumber:   phoneNumber,
		streetAddress: streetAddress,
		unit:          unit,
		city:          city,
		state:         state,
		postalCode:    postalCode,
	}
}

func newIdentityInputs() IdentityInputs {
	givenName := textinput.New()
	givenName.Placeholder = "First name"
	givenName.Width = inputWidth
	givenName.CharLimit = 30

	familyName := textinput.New()
	familyName.Placeholder = "Last name"
	familyName.Width = inputWidth
	familyName.CharLimit = 30

	dateOfBirth := textinput.New()
	dateOfBirth.Placeholder = "Date of birth (YYYY-MM-DD)"
	dateOfBirth.Width = inputWidth
	dateOfBirth.CharLimit = 10

	taxID := textinput.New()
	taxID.Placeholder = "Tax ID"
	taxID.Width = inputWidth
	taxID.CharLimit = 20

	taxIDType := textinput.New()
	taxIDType.Placeholder = "Tax ID type"
	taxIDType.Width = inputWidth
	taxIDType.CharLimit = 20

	countryOfCitizenship := textinput.New()
	countryOfCitizenship.Placeholder = "Country of citizenship (optional)"
	countryOfCitizenship.Width = inputWidth
	countryOfCitizenship.CharLimit = 30

	countryOfBirth := textinput.New()
	countryOfBirth.Placeholder = "Country of birth (optional)"
	countryOfBirth.Width = inputWidth
	countryOfBirth.CharLimit = 30

	countryOfTaxResidence := textinput.New()
	countryOfTaxResidence.Placeholder = "Country of tax residence"
	countryOfTaxResidence.Width = inputWidth
	countryOfTaxResidence.CharLimit = 30

	return IdentityInputs{
		givenName:             givenName,
		familyName:            familyName,
		dateOfBirth:           dateOfBirth,
		taxID:                 taxID,
		taxIDType:             taxIDType,
		countryOfCitizenship:  countryOfCitizenship,
		countryOfBirth:        countryOfBirth,
		countryOfTaxResidence: countryOfTaxResidence,
	}
}

func newDocumentInputs() DocumentInputs {
	documentType := textinput.New()
	documentType.Placeholder = "Document type"
	documentType.Width = inputWidth
	documentType.CharLimit = 30

	documentSubType := textinput.New()
	documentSubType.Placeholder = "Document sub-type"
	documentSubType.Width = inputWidth
	documentSubType.CharLimit = 30

	documentContent := textinput.New()
	documentContent.Placeholder = "Content (base64)"
	documentContent.Width = inputWidth
	documentContent.CharLimit = 80

	mimeType := textinput.New()
	mimeType.Placeholder = "MIME type (e.g. image/jpeg)"
	mimeType.Width = inputWidth
	mimeType.CharLimit = 50

	return DocumentInputs{
		documentType:    documentType,
		documentSubType: documentSubType,
		content:         documentContent,
		mimeType:        mimeType,
	}
}

func newTrustedContactInputs() TrustedContactInputs {
	givenName := textinput.New()
	givenName.Placeholder = "Trusted contact first name"
	givenName.Width = inputWidth
	givenName.CharLimit = 30

	familyName := textinput.New()
	familyName.Placeholder = "Trusted contact last name"
	familyName.Width = inputWidth
	familyName.CharLimit = 30

	email := textinput.New()
	email.Placeholder = "Trusted contact email (Optional)"
	email.Width = inputWidth
	email.CharLimit = 50

	return TrustedContactInputs{
		givenName:    givenName,
		familyName:   familyName,
		emailAddress: email,
	}
}

func (s SignUpPage) Init() tea.Cmd {
	return textinput.Blink
}

func (s SignUpPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if s.typing {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+j", "down":
				s.err = ""
				s.cursor++
				if s.cursor >= s.fieldCount() {
					s.cursor = 0
				}
				return s, nil

			case "ctrl+k", "up":
				s.err = ""
				s.cursor--
				if s.cursor < 0 {
					s.cursor = s.fieldCount() - 1
				}
				return s, nil

			case "ctrl+l", "ctrl+right":
				s.err = ""
				s.success = ""
				if err := s.validateCurrentPage(); err != nil {
					s.err = err.Error()
					return s, nil
				}
				if s.currentPage < trustedContactPage {
					s.currentPage++
					s.cursor = 0
				}
				return s, nil

			case "ctrl+h", "ctrl+left":
				s.err = ""
				s.success = ""
				if s.currentPage > contactPage {
					s.currentPage--
					s.cursor = 0
				}

				return s, nil

			case "tab":
				s.cursor++
				if s.cursor >= s.fieldCount() {
					s.cursor = 0
				}

				return s, nil

			case "h", "left":
				if s.currentPage == identityPage && s.cursor == 8 {
					s.fundingCursor--
					if s.fundingCursor < 0 {
						s.fundingCursor = len(s.fundingSourceOptions) - 1
					}

					return s, nil
				}

			case "l", "right":
				if s.currentPage == identityPage && s.cursor == 8 {
					s.fundingCursor++
					if s.fundingCursor >= len(s.fundingSourceOptions) {
						s.fundingCursor = 0
					}

					return s, nil
				}

			case "enter":
				if s.currentPage == identityPage && s.cursor == 8 {
					if s.fundingSelected[s.fundingCursor] {
						delete(s.fundingSelected, s.fundingCursor)
					} else {
						s.fundingSelected[s.fundingCursor] = true
					}

					return s, nil
				}

				if s.currentPage == trustedContactPage {
					s.err = ""
					s.success = ""
					if err := s.validateCurrentPage(); err != nil {
						s.err = err.Error()
						return s, nil
					}

					if err := s.submit(); err != nil {
						s.err = err.Error()
					} else {
						s.success = "Sign up successful!"
					}

					return s, nil
				}

				s.err = "Error, can't submit if you aren't on the last page and haven't filled all the mandatory fields!"

				return s, nil

			case "esc":
				s.typing = false
				return s, nil
			}
		}
	} else {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "ctrl+c":
				return s, tea.Quit

			case "enter":
				s.typing = true
				return s, nil

			case "esc":
				return s, func() tea.Msg {
					return messages.PageSwitchMsg{
						Page: messages.LoginPageNumber,
					}
				}
			}
		}
	}

	input := s.currentInput()
	if input != nil {
		var updatedInput textinput.Model
		input.Focus()
		updatedInput, cmd = input.Update(msg)
		input.Blur()
		s.setCurrentInput(updatedInput)
	}

	return s, cmd
}

func (s *SignUpPage) fieldCount() int {
	switch s.currentPage {
	case contactPage:
		return 8
	case identityPage:
		return 9
	case documentsPage:
		return 4
	case trustedContactPage:
		return 3
	default:
		return 0
	}
}

func (s *SignUpPage) currentInput() *textinput.Model {
	switch s.currentPage {
	case contactPage:
		inputs := []*textinput.Model{
			&s.contactInputs.emailAddress,
			&s.password,
			&s.contactInputs.phoneNumber,
			&s.contactInputs.streetAddress,
			&s.contactInputs.unit,
			&s.contactInputs.city,
			&s.contactInputs.state,
			&s.contactInputs.postalCode,
		}
		if s.cursor < len(inputs) {
			return inputs[s.cursor]
		}
	case identityPage:
		inputs := []*textinput.Model{
			&s.identityInputs.givenName,
			&s.identityInputs.familyName,
			&s.identityInputs.dateOfBirth,
			&s.identityInputs.taxID,
			&s.identityInputs.taxIDType,
			&s.identityInputs.countryOfCitizenship,
			&s.identityInputs.countryOfBirth,
			&s.identityInputs.countryOfTaxResidence,
		}
		if s.cursor < len(inputs) {
			return inputs[s.cursor]
		}
	case documentsPage:
		inputs := []*textinput.Model{
			&s.documentInputs.documentType,
			&s.documentInputs.documentSubType,
			&s.documentInputs.content,
			&s.documentInputs.mimeType,
		}
		if s.cursor < len(inputs) {
			return inputs[s.cursor]
		}
	case trustedContactPage:
		inputs := []*textinput.Model{
			&s.trustedContactInputs.givenName,
			&s.trustedContactInputs.familyName,
			&s.trustedContactInputs.emailAddress,
		}
		if s.cursor < len(inputs) {
			return inputs[s.cursor]
		}
	}
	return nil
}

func (s *SignUpPage) setCurrentInput(input textinput.Model) {
	switch s.currentPage {
	case contactPage:
		inputs := []*textinput.Model{
			&s.contactInputs.emailAddress,
			&s.password,
			&s.contactInputs.phoneNumber,
			&s.contactInputs.streetAddress,
			&s.contactInputs.unit,
			&s.contactInputs.city,
			&s.contactInputs.state,
			&s.contactInputs.postalCode,
		}
		if s.cursor < len(inputs) {
			*inputs[s.cursor] = input
		}
	case identityPage:
		inputs := []*textinput.Model{
			&s.identityInputs.givenName,
			&s.identityInputs.familyName,
			&s.identityInputs.dateOfBirth,
			&s.identityInputs.taxID,
			&s.identityInputs.taxIDType,
			&s.identityInputs.countryOfCitizenship,
			&s.identityInputs.countryOfBirth,
			&s.identityInputs.countryOfTaxResidence,
		}
		if s.cursor < len(inputs) {
			*inputs[s.cursor] = input
		}
	case documentsPage:
		inputs := []*textinput.Model{
			&s.documentInputs.documentType,
			&s.documentInputs.documentSubType,
			&s.documentInputs.content,
			&s.documentInputs.mimeType,
		}
		if s.cursor < len(inputs) {
			*inputs[s.cursor] = input
		}
	case trustedContactPage:
		inputs := []*textinput.Model{
			&s.trustedContactInputs.givenName,
			&s.trustedContactInputs.familyName,
			&s.trustedContactInputs.emailAddress,
		}
		if s.cursor < len(inputs) {
			*inputs[s.cursor] = input
		}
	}
}

func (s *SignUpPage) validateCurrentPage() error {
	switch s.currentPage {
	case contactPage:
		if strings.TrimSpace(s.contactInputs.emailAddress.Value()) == "" {
			return fmt.Errorf("email address is required")
		}
		if strings.TrimSpace(s.password.Value()) == "" {
			return fmt.Errorf("password is required")
		}
		if strings.TrimSpace(s.contactInputs.phoneNumber.Value()) == "" {
			return fmt.Errorf("phone number is required")
		}
		if strings.TrimSpace(s.contactInputs.streetAddress.Value()) == "" {
			return fmt.Errorf("street address is required")
		}
		if strings.TrimSpace(s.contactInputs.city.Value()) == "" {
			return fmt.Errorf("city is required")
		}
	case identityPage:
		if strings.TrimSpace(s.identityInputs.givenName.Value()) == "" {
			return fmt.Errorf("first name is required")
		}
		if strings.TrimSpace(s.identityInputs.familyName.Value()) == "" {
			return fmt.Errorf("last name is required")
		}
		if strings.TrimSpace(s.identityInputs.dateOfBirth.Value()) == "" {
			return fmt.Errorf("date of birth is required")
		}
		if strings.TrimSpace(s.identityInputs.taxID.Value()) == "" {
			return fmt.Errorf("taxID is required")
		}
		if strings.TrimSpace(s.identityInputs.taxIDType.Value()) == "" {
			return fmt.Errorf("taxIDType is required")
		}
		if strings.TrimSpace(s.identityInputs.countryOfTaxResidence.Value()) == "" {
			return fmt.Errorf("country of tax residence is required")
		}
		if len(s.fundingSelected) == 0 {
			return fmt.Errorf("at least one funding source must be selected")
		}
	case documentsPage:
		if strings.TrimSpace(s.documentInputs.documentType.Value()) != "" ||
			strings.TrimSpace(s.documentInputs.documentSubType.Value()) != "" ||
			strings.TrimSpace(s.documentInputs.content.Value()) != "" ||
			strings.TrimSpace(s.documentInputs.mimeType.Value()) != "" {

			if strings.TrimSpace(s.documentInputs.documentType.Value()) == "" {
				return fmt.Errorf("document type is required")
			}
			if strings.TrimSpace(s.documentInputs.documentSubType.Value()) == "" {
				return fmt.Errorf("document sub-type is required")
			}
			if strings.TrimSpace(s.documentInputs.content.Value()) == "" {
				return fmt.Errorf("document path is required")
			}
			if strings.TrimSpace(s.documentInputs.mimeType.Value()) == "" {
				return fmt.Errorf("MIME type is required")
			}
		}

	case trustedContactPage:
		if strings.TrimSpace(s.trustedContactInputs.givenName.Value()) != "" ||
			strings.TrimSpace(s.trustedContactInputs.familyName.Value()) != "" ||
			strings.TrimSpace(s.trustedContactInputs.emailAddress.Value()) != "" {

			if strings.TrimSpace(s.trustedContactInputs.givenName.Value()) == "" {
				return fmt.Errorf("given name is required")
			}
			if strings.TrimSpace(s.trustedContactInputs.familyName.Value()) == "" {
				return fmt.Errorf("family name is required")
			}
		}
	}
	return nil
}

func (s SignUpPage) View() string {
	var pageName string
	switch s.currentPage {
	case contactPage:
		pageName = "Contact Information"
	case identityPage:
		pageName = "Identity"
	case documentsPage:
		pageName = "Documents (Optional)"
	case trustedContactPage:
		pageName = "Trusted Contact (Optional)"
	}

	header := titleStyle.Render(
		fmt.Sprintf("📝 Sign Up — %s (%d/%d)",
			pageName,
			s.currentPage+1,
			trustedContactPage+1,
		),
	)

	header = lipgloss.PlaceHorizontal(s.BaseModel.Width, lipgloss.Center, header)

	var fields []string
	s.renderCurrentPageFields(&fields)

	content := lipgloss.JoinVertical(lipgloss.Center, fields...)

	if s.err != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Center,
			content,
			"",
			errorStyle.Render("❌ "+s.err),
		)
	}

	if s.success != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Center,
			content,
			"",
			successStyle.Render("✓ "+s.success),
		)
	}

	help := helpStyle.Render(
		"↑/↓: move • ctrl+h / ctrl+l: change page • enter: submit/type • esc: stop typing/back • q: quit",
	)

	finalContent := ""
	if s.currentPage != identityPage {
		finalContent = lipgloss.JoinVertical(
			lipgloss.Center,
			"",
			content,
			"\n\n",
			help,
		)
	} else {
		finalContent = lipgloss.JoinVertical(
			lipgloss.Center,
			"",
			content,
			"",
			help,
		)
	}

	return header + lipgloss.Place(
		s.BaseModel.Width,
		s.BaseModel.Height,
		lipgloss.Center,
		lipgloss.Center,
		finalContent,
	)
}

func (s SignUpPage) submit() error {
	s.accountInfo.Contact = Contact{
		EmailAddress:  s.contactInputs.emailAddress.Value(),
		PhoneNumber:   s.contactInputs.phoneNumber.Value(),
		StreetAddress: []string{s.contactInputs.streetAddress.Value()},
		Unit:          s.contactInputs.unit.Value(),
		City:          s.contactInputs.city.Value(),
		State:         s.contactInputs.state.Value(),
		PostalCode:    s.contactInputs.postalCode.Value(),
	}

	s.accountInfo.Identity = Identity{
		GivenName:             s.identityInputs.givenName.Value(),
		FamilyName:            s.identityInputs.familyName.Value(),
		DateOfBirth:           s.identityInputs.dateOfBirth.Value(),
		TaxID:                 s.identityInputs.taxID.Value(),
		TaxIDType:             s.identityInputs.taxIDType.Value(),
		CountryOfCitizenship:  s.identityInputs.countryOfCitizenship.Value(),
		CountryOfBirth:        s.identityInputs.countryOfBirth.Value(),
		CountryOfTaxResidence: s.identityInputs.countryOfTaxResidence.Value(),
	}

	var sources []string
	for i := range s.fundingSelected {
		sources = append(sources, s.fundingSourceOptions[i])
	}

	s.accountInfo.Identity.FundingSource = sources

	if s.documentInputs.documentType.Value() != "" {
		s.accountInfo.Documents = []Document{
			{
				DocumentType:    s.documentInputs.documentType.Value(),
				DocumentSubType: s.documentInputs.documentSubType.Value(),
				Content:         s.documentInputs.content.Value(),
				MimeType:        s.documentInputs.mimeType.Value(),
			},
		}
	} else {
		s.accountInfo.Documents = nil
	}

	if s.trustedContactInputs.givenName.Value() != "" {
		s.accountInfo.TrustedContact = &TrustedContact{
			GivenName:    s.trustedContactInputs.givenName.Value(),
			FamilyName:   s.trustedContactInputs.familyName.Value(),
			EmailAddress: s.trustedContactInputs.emailAddress.Value(),
		}
	} else {
		s.accountInfo.TrustedContact = nil
	}

	s.accountInfo.Password = s.password.Value()

	body, err := json.Marshal(s.accountInfo)
	if err != nil {
		return err
	}

	_, err = requests.MakeRequest(
		http.MethodPost,
		requests.BaseURL+"/sign-up",
		bytes.NewReader(body),
		s.BaseModel.Client,
		s.BaseModel.Token,
	)

	return err
}

func (s SignUpPage) renderCurrentPageFields(fields *[]string) {
	switch s.currentPage {

	case contactPage:
		s.addInput(fields, "Email", s.contactInputs.emailAddress, 0)
		s.addInput(fields, "Password", s.password, 1)
		s.addInput(fields, "Phone", s.contactInputs.phoneNumber, 2)
		s.addInput(fields, "Street", s.contactInputs.streetAddress, 3)
		s.addInput(fields, "Unit", s.contactInputs.unit, 4)
		s.addInput(fields, "City", s.contactInputs.city, 5)
		s.addInput(fields, "State", s.contactInputs.state, 6)
		s.addInput(fields, "Postal Code", s.contactInputs.postalCode, 7)

	case identityPage:
		s.addInput(fields, "First Name", s.identityInputs.givenName, 0)
		s.addInput(fields, "Last Name", s.identityInputs.familyName, 1)
		s.addInput(fields, "DOB", s.identityInputs.dateOfBirth, 2)
		s.addInput(fields, "Tax ID", s.identityInputs.taxID, 3)
		s.addInput(fields, "Tax ID Type", s.identityInputs.taxIDType, 4)
		s.addInput(fields, "Citizenship", s.identityInputs.countryOfCitizenship, 5)
		s.addInput(fields, "Birth Country", s.identityInputs.countryOfBirth, 6)
		s.addInput(fields, "Tax Residence", s.identityInputs.countryOfTaxResidence, 7)

		*fields = append(*fields, s.renderFundingSources())

	case documentsPage:
		s.addInput(fields, "Doc Type", s.documentInputs.documentType, 0)
		s.addInput(fields, "Sub-Type", s.documentInputs.documentSubType, 1)
		s.addInput(fields, "Content", s.documentInputs.content, 2)
		s.addInput(fields, "MIME", s.documentInputs.mimeType, 3)

	case trustedContactPage:
		s.addInput(fields, "First Name", s.trustedContactInputs.givenName, 0)
		s.addInput(fields, "Last Name", s.trustedContactInputs.familyName, 1)
		s.addInput(fields, "Email", s.trustedContactInputs.emailAddress, 2)
	}
}
func renderInput(label string, input textinput.Model, focused, typing bool) string {
	style := lipgloss.Style{}
	if typing {
		if focused {
			input.Focus()
			style = focusedStyle
		} else {
			input.Blur()
			style = inputStyle
		}
	} else {
		style = inputStyle
	}

	block := lipgloss.JoinVertical(
		lipgloss.Center,
		labelStyle.Render(label),
		style.Render(input.View()),
	)

	return formRowStyle.Render(block)
}

func (s SignUpPage) addInput(fields *[]string, label string, input textinput.Model, index int) {
	*fields = append(*fields,
		renderInput(label, input, s.cursor == index, s.typing),
	)
}
func (s SignUpPage) renderFundingSources() string {
	var rows []string

	for i, option := range s.fundingSourceOptions {
		style := fundingIdleStyle

		if s.typing {
			if s.fundingSelected[i] {
				style = fundingSelectedStyle
			}
			if s.cursor == 8 && s.fundingCursor == i {
				style = fundingFocusedStyle
			}
		} else {
			style = fundingIdleStyle
		}

		prefix := "  "
		if s.fundingCursor == i && s.cursor == 8 {
			prefix = " ▸ "
		}

		rows = append(rows, style.Render(prefix+option))
	}

	block := lipgloss.JoinVertical(
		lipgloss.Center,
		labelStyle.Render("Funding Sources (select one or more)"),
		lipgloss.JoinHorizontal(lipgloss.Center, rows...),
	)

	return formRowStyle.Render(block)
}
