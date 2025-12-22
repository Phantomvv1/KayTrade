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
	agreementsPage
	documentsPage
	trustedContactPage
	enabledAssetsPage
)

type Contact struct {
	EmailAddress  string   `json:"email_address"`
	PhoneNumber   string   `json:"phone_number"`
	StreetAddress []string `json:"street_address"`
	Unit          string   `json:"unit"`
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
	TaxID                 string   `json:"tax_id"`
	TaxIDType             string   `json:"tax_id_type"`
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
	fundingSource         textinput.Model
}

type Disclosures struct {
	IsControlPerson             bool `json:"is_control_person"`
	IsAffiliatedExchangeOrFinra bool `json:"is_affiliated_exchange_or_finra"`
	IsPoliticallyExposed        bool `json:"is_politically_exposed"`
	ImmediateFamilyExposed      bool `json:"immediate_family_exposed"`
}

type Agreement struct {
	Agreement string `json:"agreement"`
	SignedAt  string `json:"signed_at"`
	IPAddress string `json:"ip_address"`
	Revision  string `json:"revision,omitempty"`
}

type AgreementInputs struct {
	agreement textinput.Model
	signedAt  textinput.Model
	iPAddress textinput.Model
	revision  textinput.Model
}

type Document struct {
	DocumentType    string `json:"document_type"`
	DocumentSubType string `json:"document_sub_type"`
	Content         string `json:"content"`
	MimeType        string `json:"mime_type"`
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
	Contact        Contact        `json:"contact"`
	Identity       Identity       `json:"identity"`
	Disclosures    Disclosures    `json:"disclosures"`
	Agreements     []Agreement    `json:"agreements"`
	Documents      []Document     `json:"documents"`
	TrustedContact TrustedContact `json:"trusted_contact"`
	EnabledAssets  []string       `json:"enabled_assets"`
}

type SignUpPage struct {
	BaseModel            basemodel.BaseModel
	accountInfo          AccountInfo
	contactInputs        ContactInputs
	identityInputs       IdentityInputs
	agreementInputs      AgreementInputs
	documentInputs       DocumentInputs
	trustedContactInputs TrustedContactInputs
	enabledAssets        textinput.Model
	currentPage          int
	cursor               int
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

	toggleOnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	toggleOffStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
)

func NewSignUpPage(client *http.Client) SignUpPage {
	enabledAssets := textinput.New()
	enabledAssets.Placeholder = "Enabled assets (comma-separated)"
	enabledAssets.Width = 28
	enabledAssets.CharLimit = 100

	return SignUpPage{
		BaseModel: basemodel.BaseModel{Client: client},
		accountInfo: AccountInfo{
			Disclosures: Disclosures{
				IsControlPerson:             false,
				IsAffiliatedExchangeOrFinra: false,
				IsPoliticallyExposed:        false,
				ImmediateFamilyExposed:      false,
			},
		},
		contactInputs:        newContactInputs(),
		identityInputs:       newIdentityInputs(),
		agreementInputs:      newAgreementInputs(),
		documentInputs:       newDocumentInputs(),
		trustedContactInputs: newTrustedContactInputs(),
		enabledAssets:        enabledAssets,
		currentPage:          contactPage,
		cursor:               0,
	}
}

func newContactInputs() ContactInputs {
	emailAddress := textinput.New()
	emailAddress.Placeholder = "Email address"
	emailAddress.Width = 28
	emailAddress.CharLimit = 50

	phoneNumber := textinput.New()
	phoneNumber.Placeholder = "Phone number"
	phoneNumber.Width = 28
	phoneNumber.CharLimit = 20

	streetAddress := textinput.New()
	streetAddress.Placeholder = "Street address"
	streetAddress.Width = 28
	streetAddress.CharLimit = 60

	unit := textinput.New()
	unit.Placeholder = "Unit (optional)"
	unit.Width = 28
	unit.CharLimit = 10

	city := textinput.New()
	city.Placeholder = "City"
	city.Width = 28
	city.CharLimit = 30

	state := textinput.New()
	state.Placeholder = "State (optional)"
	state.Width = 28
	state.CharLimit = 20

	postalCode := textinput.New()
	postalCode.Placeholder = "Postal code (optional)"
	postalCode.Width = 28
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
	givenName.Width = 28
	givenName.CharLimit = 30

	familyName := textinput.New()
	familyName.Placeholder = "Last name"
	familyName.Width = 28
	familyName.CharLimit = 30

	dateOfBirth := textinput.New()
	dateOfBirth.Placeholder = "Date of birth (YYYY-MM-DD)"
	dateOfBirth.Width = 28
	dateOfBirth.CharLimit = 10

	taxID := textinput.New()
	taxID.Placeholder = "Tax ID"
	taxID.Width = 28
	taxID.CharLimit = 20

	taxIDType := textinput.New()
	taxIDType.Placeholder = "Tax ID type"
	taxIDType.Width = 28
	taxIDType.CharLimit = 20

	countryOfCitizenship := textinput.New()
	countryOfCitizenship.Placeholder = "Country of citizenship (optional)"
	countryOfCitizenship.Width = 28
	countryOfCitizenship.CharLimit = 30

	countryOfBirth := textinput.New()
	countryOfBirth.Placeholder = "Country of birth (optional)"
	countryOfBirth.Width = 28
	countryOfBirth.CharLimit = 30

	countryOfTaxResidence := textinput.New()
	countryOfTaxResidence.Placeholder = "Country of tax residence"
	countryOfTaxResidence.Width = 28
	countryOfTaxResidence.CharLimit = 30

	fundingSource := textinput.New()
	fundingSource.Placeholder = "Funding source (comma-separated)"
	fundingSource.Width = 28
	fundingSource.CharLimit = 100

	return IdentityInputs{
		givenName:             givenName,
		familyName:            familyName,
		dateOfBirth:           dateOfBirth,
		taxID:                 taxID,
		taxIDType:             taxIDType,
		countryOfCitizenship:  countryOfCitizenship,
		countryOfBirth:        countryOfBirth,
		countryOfTaxResidence: countryOfTaxResidence,
		fundingSource:         fundingSource,
	}
}

func newAgreementInputs() AgreementInputs {
	agreement := textinput.New()
	agreement.Placeholder = "Agreement name"
	agreement.Width = 28
	agreement.CharLimit = 40

	signedAt := textinput.New()
	signedAt.Placeholder = "Signed at (RFC3339)"
	signedAt.Width = 28
	signedAt.CharLimit = 25

	ipAddress := textinput.New()
	ipAddress.Placeholder = "IP address"
	ipAddress.Width = 28
	ipAddress.CharLimit = 45

	revision := textinput.New()
	revision.Placeholder = "Revision (optional)"
	revision.Width = 28
	revision.CharLimit = 20

	return AgreementInputs{
		agreement: agreement,
		signedAt:  signedAt,
		iPAddress: ipAddress,
		revision:  revision,
	}
}

func newDocumentInputs() DocumentInputs {
	documentType := textinput.New()
	documentType.Placeholder = "Document type"
	documentType.Width = 28
	documentType.CharLimit = 30

	documentSubType := textinput.New()
	documentSubType.Placeholder = "Document sub-type"
	documentSubType.Width = 28
	documentSubType.CharLimit = 30

	documentContent := textinput.New()
	documentContent.Placeholder = "Path to document"
	documentContent.Width = 28
	documentContent.CharLimit = 80

	mimeType := textinput.New()
	mimeType.Placeholder = "MIME type (e.g. image/jpeg)"
	mimeType.Width = 28
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
	givenName.Width = 28
	givenName.CharLimit = 30

	familyName := textinput.New()
	familyName.Placeholder = "Trusted contact last name"
	familyName.Width = 28
	familyName.CharLimit = 30

	email := textinput.New()
	email.Placeholder = "Trusted contact email"
	email.Width = 28
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit

		case "j", "down":
			s.err = ""
			s.cursor++
			if s.cursor >= s.getFieldCount() {
				s.cursor = 0
			}
			return s, nil

		case "k", "up":
			s.err = ""
			s.cursor--
			if s.cursor < 0 {
				s.cursor = s.getFieldCount() - 1
			}
			return s, nil

		case "ctrl+l", "ctrl+right":
			s.err = ""
			s.success = ""
			if err := s.validateCurrentPage(); err != nil {
				s.err = err.Error()
				return s, nil
			}
			if s.currentPage < enabledAssetsPage {
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

		case "enter":
			if s.currentPage == enabledAssetsPage {
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
			}
			return s, nil

		case "esc":
			return s, func() tea.Msg {
				return messages.PageSwitchMsg{
					Page: messages.LoginPageNumber,
				}
			}
		}
	}

	input := s.getCurrentInput()
	if input != nil {
		var updatedInput textinput.Model
		updatedInput, cmd = input.Update(msg)
		s.setCurrentInput(updatedInput)
	}

	return s, cmd
}

func (s *SignUpPage) getFieldCount() int {
	switch s.currentPage {
	case contactPage:
		return 8
	case identityPage:
		return 9
	case agreementsPage:
		return 4
	case documentsPage:
		return 4
	case trustedContactPage:
		return 3
	case enabledAssetsPage:
		return 1
	default:
		return 0
	}
}

func (s *SignUpPage) getCurrentInput() *textinput.Model {
	switch s.currentPage {
	case contactPage:
		inputs := []*textinput.Model{
			&s.contactInputs.emailAddress,
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
			&s.identityInputs.fundingSource,
		}
		if s.cursor < len(inputs) {
			return inputs[s.cursor]
		}
	case agreementsPage:
		inputs := []*textinput.Model{
			&s.agreementInputs.agreement,
			&s.agreementInputs.signedAt,
			&s.agreementInputs.iPAddress,
			&s.agreementInputs.revision,
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
	case enabledAssetsPage:
		return &s.enabledAssets
	}
	return nil
}

func (s *SignUpPage) setCurrentInput(input textinput.Model) {
	switch s.currentPage {
	case contactPage:
		inputs := []*textinput.Model{
			&s.contactInputs.emailAddress,
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
			&s.identityInputs.fundingSource,
		}
		if s.cursor < len(inputs) {
			*inputs[s.cursor] = input
		}
	case agreementsPage:
		inputs := []*textinput.Model{
			&s.agreementInputs.agreement,
			&s.agreementInputs.signedAt,
			&s.agreementInputs.iPAddress,
			&s.agreementInputs.revision,
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
	case enabledAssetsPage:
		s.enabledAssets = input
	}
}

func (s *SignUpPage) validateCurrentPage() error {
	switch s.currentPage {
	case contactPage:
		if strings.TrimSpace(s.contactInputs.emailAddress.Value()) == "" {
			return fmt.Errorf("email address is required")
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
			return fmt.Errorf("tax ID is required")
		}
		if strings.TrimSpace(s.identityInputs.taxIDType.Value()) == "" {
			return fmt.Errorf("tax ID type is required")
		}
		if strings.TrimSpace(s.identityInputs.countryOfTaxResidence.Value()) == "" {
			return fmt.Errorf("country of tax residence is required")
		}
		if strings.TrimSpace(s.identityInputs.fundingSource.Value()) == "" {
			return fmt.Errorf("funding source is required")
		}
	case agreementsPage:
		if strings.TrimSpace(s.agreementInputs.agreement.Value()) == "" {
			return fmt.Errorf("agreement name is required")
		}
		if strings.TrimSpace(s.agreementInputs.signedAt.Value()) == "" {
			return fmt.Errorf("signed at is required")
		}
		if strings.TrimSpace(s.agreementInputs.iPAddress.Value()) == "" {
			return fmt.Errorf("IP address is required")
		}
	case documentsPage:
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
	case enabledAssetsPage:
		if strings.TrimSpace(s.enabledAssets.Value()) == "" {
			return fmt.Errorf("enabled assets is required")
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
	case agreementsPage:
		pageName = "Agreements"
	case documentsPage:
		pageName = "Documents"
	case trustedContactPage:
		pageName = "Trusted Contact (Optional)"
	case enabledAssetsPage:
		pageName = "Enabled Assets"
	}

	header := titleStyle.Render(
		fmt.Sprintf("📝 Sign Up — %s (%d/%d)",
			pageName,
			s.currentPage+1,
			enabledAssetsPage+1,
		),
	)

	var fields []string
	s.renderCurrentPageFields(&fields)

	content := lipgloss.JoinVertical(lipgloss.Center, fields...)

	if s.err != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			"",
			errorStyle.Render("❌ "+s.err),
		)
	}

	if s.success != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			"",
			successStyle.Render("✓ "+s.success),
		)
	}

	help := helpStyle.Render(
		"↑/↓ move • ctrl+h / ctrl+l change page • enter submit • q quit",
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		content,
		"",
		help,
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
		FundingSource: strings.Split(
			s.identityInputs.fundingSource.Value(), ",",
		),
	}

	s.accountInfo.Agreements = []Agreement{
		{
			Agreement: s.agreementInputs.agreement.Value(),
			SignedAt:  s.agreementInputs.signedAt.Value(),
			IPAddress: s.agreementInputs.iPAddress.Value(),
			Revision:  s.agreementInputs.revision.Value(),
		},
	}

	s.accountInfo.Documents = []Document{
		{
			DocumentType:    s.documentInputs.documentType.Value(),
			DocumentSubType: s.documentInputs.documentSubType.Value(),
			Content:         s.documentInputs.content.Value(),
			MimeType:        s.documentInputs.mimeType.Value(),
		},
	}

	s.accountInfo.TrustedContact = TrustedContact{
		GivenName:    s.trustedContactInputs.givenName.Value(),
		FamilyName:   s.trustedContactInputs.familyName.Value(),
		EmailAddress: s.trustedContactInputs.emailAddress.Value(),
	}

	s.accountInfo.EnabledAssets = strings.Split(
		s.enabledAssets.Value(), ",",
	)

	body, err := json.Marshal(s.accountInfo)
	if err != nil {
		return err
	}

	_, err = requests.MakeRequest(
		http.MethodPost,
		requests.BaseURL+"/users/signup",
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
		s.addInput(fields, "Phone", s.contactInputs.phoneNumber, 1)
		s.addInput(fields, "Street", s.contactInputs.streetAddress, 2)
		s.addInput(fields, "Unit", s.contactInputs.unit, 3)
		s.addInput(fields, "City", s.contactInputs.city, 4)
		s.addInput(fields, "State", s.contactInputs.state, 5)
		s.addInput(fields, "Postal Code", s.contactInputs.postalCode, 6)

	case identityPage:
		s.addInput(fields, "First Name", s.identityInputs.givenName, 0)
		s.addInput(fields, "Last Name", s.identityInputs.familyName, 1)
		s.addInput(fields, "DOB", s.identityInputs.dateOfBirth, 2)
		s.addInput(fields, "Tax ID", s.identityInputs.taxID, 3)
		s.addInput(fields, "Tax ID Type", s.identityInputs.taxIDType, 4)
		s.addInput(fields, "Citizenship", s.identityInputs.countryOfCitizenship, 5)
		s.addInput(fields, "Birth Country", s.identityInputs.countryOfBirth, 6)
		s.addInput(fields, "Tax Residence", s.identityInputs.countryOfTaxResidence, 7)
		s.addInput(fields, "Funding Source", s.identityInputs.fundingSource, 8)

	case agreementsPage:
		s.addInput(fields, "Agreement", s.agreementInputs.agreement, 0)
		s.addInput(fields, "Signed At", s.agreementInputs.signedAt, 1)
		s.addInput(fields, "IP Address", s.agreementInputs.iPAddress, 2)
		s.addInput(fields, "Revision", s.agreementInputs.revision, 3)

	case documentsPage:
		s.addInput(fields, "Doc Type", s.documentInputs.documentType, 0)
		s.addInput(fields, "Sub-Type", s.documentInputs.documentSubType, 1)
		s.addInput(fields, "Path", s.documentInputs.content, 2)
		s.addInput(fields, "MIME", s.documentInputs.mimeType, 3)

	case trustedContactPage:
		s.addInput(fields, "First Name", s.trustedContactInputs.givenName, 0)
		s.addInput(fields, "Last Name", s.trustedContactInputs.familyName, 1)
		s.addInput(fields, "Email", s.trustedContactInputs.emailAddress, 2)

	case enabledAssetsPage:
		s.addInput(fields, "Assets", s.enabledAssets, 0)
	}
}
func renderInput(label string, input textinput.Model, focused bool) string {
	if focused {
		input.Focus()
	} else {
		input.Blur()
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render(label),
		input.View(),
	)
}

func (s SignUpPage) addInput(fields *[]string, label string, input textinput.Model, index int) {
	*fields = append(*fields,
		renderInput(label, input, s.cursor == index),
	)
}
