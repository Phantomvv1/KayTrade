package signuppage

import (
	"net/http"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Contact struct {
	EmailAddress  string   `json:"email_address"`
	PhoneNumber   string   `json:"phone_number"`
	StreetAddress []string `json:"street_address"`
	Unit          string   `json:"unit"`
	City          string   `json:"city"`
	State         string   `json:"state,omitempty"`       // not required
	PostalCode    string   `json:"postal_code,omitempty"` // not required
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
	CountryOfCitizenship  string   `json:"country_of_citizenship,omitempty"` // not required
	CountryOfBirth        string   `json:"country_of_birth,omitempty"`       // not required
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
	Revision  string `json:"revision,omitempty"` // not required
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

// Not required
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
}

func NewSignUpaPage(client *http.Client) SignUpPage {
	enabledAssets := textinput.New()
	enabledAssets.Placeholder = "Email address"
	enabledAssets.Width = 28
	enabledAssets.CharLimit = 50

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
	return s, nil
}

func (s SignUpPage) View() string {
	return ""
}
