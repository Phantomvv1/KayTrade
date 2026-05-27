package bankrelationshipcreationpage

import (
	"testing"
)

func setupBankPage() BankRelationshipCreationPage {
	p := NewBankRelationship(nil, nil)
	return p
}

func TestCalculateTotalFields_Bank_ABA(t *testing.T) {
	p := setupBankPage()

	got := p.calculateTotalFields()
	want := 4

	if got != want {
		t.Fatalf("expected %d fields, got %d", want, got)
	}
}

func TestCalculateTotalFields_Bank_BIC(t *testing.T) {
	p := setupBankPage()
	p.bankInputs.bankCodeTypeIdx = 1 // BIC

	got := p.calculateTotalFields()
	want := 8

	if got != want {
		t.Fatalf("expected %d fields for BIC, got %d", want, got)
	}
}

func TestCalculateTotalFields_ACH(t *testing.T) {
	p := setupBankPage()
	p.bankRelationshipType = bankTypeAch

	got := p.calculateTotalFields()
	want := 5

	if got != want {
		t.Fatalf("expected %d ACH fields, got %d", want, got)
	}
}

func TestSubmitBankRelationship_MissingName(t *testing.T) {
	p := setupBankPage()

	p.bankInputs.bankCode.SetValue("ABC")
	p.bankInputs.accountNumber.SetValue("123")

	err := p.submitBankRelationship(make(map[string]any))
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestSubmitBankRelationship_MissingBankCode(t *testing.T) {
	p := setupBankPage()

	p.bankInputs.name.SetValue("Test")
	p.bankInputs.accountNumber.SetValue("123")

	err := p.submitBankRelationship(make(map[string]any))
	if err == nil {
		t.Fatal("expected error for missing bank code")
	}
}

func TestSubmitBankRelationship_MissingAccountNumber(t *testing.T) {
	p := setupBankPage()

	p.bankInputs.name.SetValue("Test")
	p.bankInputs.bankCode.SetValue("ABC")

	err := p.submitBankRelationship(make(map[string]any))
	if err == nil {
		t.Fatal("expected error for missing account number")
	}
}

func TestSubmitBankRelationship_BIC_MissingCountry(t *testing.T) {
	p := setupBankPage()
	p.bankInputs.bankCodeTypeIdx = 1 // BIC

	p.bankInputs.name.SetValue("Test")
	p.bankInputs.bankCode.SetValue("ABC")
	p.bankInputs.accountNumber.SetValue("123")

	p.bankInputs.country.SetValue("")
	p.bankInputs.stateProvince.SetValue("State")
	p.bankInputs.city.SetValue("City")
	p.bankInputs.streetAddress.SetValue("Street")

	err := p.submitBankRelationship(make(map[string]any))
	if err == nil {
		t.Fatal("expected error for missing country in BIC mode")
	}
}

func TestSubmitACH_MissingOwnerName(t *testing.T) {
	p := setupBankPage()
	p.bankRelationshipType = bankTypeAch

	p.achInputs.bankAccountNumber.SetValue("123")
	p.achInputs.bankRoutingNumber.SetValue("021000021")

	err := p.submitAchRelationship(make(map[string]any))
	if err == nil {
		t.Fatal("expected error for missing account owner name")
	}
}

func TestSubmitACH_InvalidRoutingNumberLength(t *testing.T) {
	p := setupBankPage()
	p.bankRelationshipType = bankTypeAch

	p.achInputs.accountOwnerName.SetValue("John")
	p.achInputs.bankAccountNumber.SetValue("123")
	p.achInputs.bankRoutingNumber.SetValue("123") // invalid

	err := p.submitAchRelationship(make(map[string]any))
	if err == nil {
		t.Fatal("expected routing number length error")
	}
}

func TestValidateRoutingNumber_InvalidLength(t *testing.T) {
	err := validateRoutingNumber("123")
	if err == nil {
		t.Fatal("expected error for invalid length")
	}
}

func TestValidateRoutingNumber_InvalidChecksum(t *testing.T) {
	err := validateRoutingNumber("000000001")
	if err == nil {
		t.Fatal("expected checksum validation failure")
	}
}
