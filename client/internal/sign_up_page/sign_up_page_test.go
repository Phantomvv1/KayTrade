package signuppage

import (
	"net/http"
	"testing"

	basemodel "github.com/Phantomvv1/KayTrade/client/internal/base_model"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestPage() SignUpPage {
	return NewSignUpPage(&http.Client{}, &basemodel.TokenStore{})
}

func fillContact(page *SignUpPage) {
	page.contactInputs.emailAddress.SetValue("test@mail.com")
	page.password.SetValue("secret")
	page.contactInputs.phoneNumber.SetValue("+111")
	page.contactInputs.streetAddress.SetValue("Street 1")
	page.contactInputs.city.SetValue("Sofia")
}

func fillIdentity(page *SignUpPage) {
	page.identityInputs.givenName.SetValue("John")
	page.identityInputs.familyName.SetValue("Doe")
	page.identityInputs.dateOfBirth.SetValue("1990-01-01")
	page.identityInputs.taxID.SetValue("123")
	page.identityInputs.taxIDType.SetValue("SSN")
	page.identityInputs.countryOfTaxResidence.SetValue("BG")
}

func TestValidateContactPage_MissingFields(t *testing.T) {
	p := newTestPage()

	err := p.validateCurrentPage()
	if err == nil {
		t.Fatal("expected error for empty contact page")
	}
}

func TestValidateContactPage_Valid(t *testing.T) {
	p := newTestPage()
	fillContact(&p)

	err := p.validateCurrentPage()
	if err != nil {
		t.Fatal("unexpected error")
	}
}

func TestValidateIdentityPage_FundingRequired(t *testing.T) {
	p := newTestPage()

	p.currentPage = identityPage
	fillIdentity(&p)

	err := p.validateCurrentPage()
	if err == nil {
		t.Fatal("expected funding source validation error")
	}
}

func TestValidateIdentityPage_OK(t *testing.T) {
	p := newTestPage()

	p.currentPage = identityPage
	fillIdentity(&p)

	p.fundingSelected[0] = true

	err := p.validateCurrentPage()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateDocumentsPage_PartialDocError(t *testing.T) {
	p := newTestPage()
	p.currentPage = documentsPage

	p.documentInputs.documentType.SetValue("passport")

	err := p.validateCurrentPage()
	if err == nil {
		t.Fatal("expected document validation error")
	}
}

func TestValidateTrustedContact_OptionalOKEmpty(t *testing.T) {
	p := newTestPage()
	p.currentPage = trustedContactPage

	err := p.validateCurrentPage()
	if err != nil {
		t.Fatalf("expected no error for empty trusted contact: %v", err)
	}
}

func TestUpdate_NextPageWhenNotFilled(t *testing.T) {
	p := newTestPage()

	msg := tea.KeyMsg{Type: tea.KeyCtrlL}

	m, _ := p.Update(msg)
	np := m.(SignUpPage)

	if np.currentPage != contactPage {
		t.Fatalf("expected contact page, got %v", np.currentPage)
	}
}

func TestUpdate_NextPage(t *testing.T) {
	p := newTestPage()

	fillContact(&p)

	msg := tea.KeyMsg{Type: tea.KeyCtrlL}

	m, _ := p.Update(msg)
	np := m.(SignUpPage)

	if np.currentPage != identityPage {
		t.Fatalf("expected identity page, got %v", np.currentPage)
	}
}

func TestUpdate_PreviousPageCtrlH(t *testing.T) {
	p := newTestPage()
	p.currentPage = identityPage

	msg := tea.KeyMsg{Type: tea.KeyCtrlH}
	m, _ := p.Update(msg)

	np := m.(SignUpPage)
	if np.currentPage != contactPage {
		t.Fatalf("expected contact page, got %v", np.currentPage)
	}
}

func TestUpdate_TabCursorWrap(t *testing.T) {
	p := newTestPage()

	p.cursor = p.fieldCount() - 1

	msg := tea.KeyMsg{Type: tea.KeyTab}
	m, _ := p.Update(msg)

	np := m.(SignUpPage)
	if np.cursor != 0 {
		t.Fatalf("expected cursor wrap to 0, got %d", np.cursor)
	}
}

func TestFundingSelection_Toggle(t *testing.T) {
	p := newTestPage()
	p.currentPage = identityPage
	p.cursor = 8

	p.fundingCursor = 2

	msg := tea.KeyMsg{Type: tea.KeyEnter}

	m, _ := p.Update(msg)
	np := m.(SignUpPage)

	if !np.fundingSelected[2] {
		t.Fatal("expected funding option to be selected")
	}

	m, _ = np.Update(msg)
	np2 := m.(SignUpPage)

	if np2.fundingSelected[2] {
		t.Fatal("expected funding option to be deselected")
	}
}

func TestTogglePasswordVisibility(t *testing.T) {
	p := newTestPage()

	p.currentPage = contactPage
	p.cursor = 1

	msg := tea.KeyMsg{Type: tea.KeyCtrlE}
	m, _ := p.Update(msg)

	np := m.(SignUpPage)

	if np.password.EchoMode != textinput.EchoNormal {
		t.Fatal("expected password echo mode to toggle")
	}
}

func TestSubmitPageValidationBlocksEnter(t *testing.T) {
	p := newTestPage()
	p.currentPage = contactPage

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	m, _ := p.Update(msg)

	np := m.(SignUpPage)

	if np.err == "" {
		t.Fatal("expected error when submitting incomplete form")
	}
}

func TestSubmitSetsSuccessOrErrorState(t *testing.T) {
	p := newTestPage()

	p.currentPage = trustedContactPage

	p.identityInputs.givenName.SetValue("")
	p.identityInputs.familyName.SetValue("")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	m, _ := p.Update(msg)

	np := m.(SignUpPage)

	if np.err == "" {
		t.Fatal("expected error on invalid submit")
	}
}

func TestCurrentInputReturnsCorrectField(t *testing.T) {
	p := newTestPage()

	p.currentPage = contactPage
	p.cursor = 0

	in := p.currentInput()
	if in == nil {
		t.Fatal("expected input for contact page")
	}

	p.currentPage = identityPage
	p.cursor = 0

	in2 := p.currentInput()
	if in2 == nil {
		t.Fatal("expected input for identity page")
	}
}
