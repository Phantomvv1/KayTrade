package watchlist

import (
	"reflect"
	"testing"
	"time"
)

func TestContainsSymbol(t *testing.T) {
	response := []CompanyInfo{
		{Symbol: "AAPL"},
		{Symbol: "TSLA"},
		{Symbol: "GOOG"},
	}

	tests := []struct {
		symbol   string
		expected int
	}{
		{"AAPL", 0},
		{"TSLA", 1},
		{"GOOG", 2},
		{"MSFT", -1},
	}

	for _, tt := range tests {
		index := containsSymbol(response, tt.symbol)
		if index != tt.expected {
			t.Errorf("containsSymbol(%s) = %d, expected %d", tt.symbol, index, tt.expected)
		}
	}
}

func TestDecodeAnyArray(t *testing.T) {
	input := []any{
		map[string]any{"a": 1},
		map[string]any{"b": 2},
	}

	result := decodeAnyArray(input)

	if len(result) != 2 {
		t.Fatalf("expected length 2, got %d", len(result))
	}

	if result[0]["a"] != 1 || result[1]["b"] != 2 {
		t.Errorf("decodeAnyArray returned incorrect result")
	}
}

func TestChooseLogo(t *testing.T) {
	logoData := map[string]any{
		"logos": []any{
			map[string]any{
				"theme": "light",
				"formats": []any{
					map[string]any{
						"background": "white",
						"format":     "png",
						"src":        "light.png",
					},
				},
			},
			map[string]any{
				"theme": "dark",
				"formats": []any{
					map[string]any{
						"background": "transparent",
						"format":     "png",
						"src":        "dark-transparent.png",
					},
					map[string]any{
						"background": "white",
						"format":     "png",
						"src":        "dark.png",
					},
				},
			},
		},
	}

	result := chooseLogo(logoData)

	if result != "dark-transparent.png" {
		t.Errorf("expected dark-transparent.png, got %s", result)
	}
}

func TestChooseLogoFallbackToPNG(t *testing.T) {
	logoData := map[string]any{
		"logos": []any{
			map[string]any{
				"theme": "light",
				"formats": []any{
					map[string]any{
						"background": "white",
						"format":     "png",
						"src":        "light.png",
					},
				},
			},
		},
	}

	result := chooseLogo(logoData)

	if result != "light.png" {
		t.Errorf("expected light.png, got %s", result)
	}
}

func TestCleanAssets(t *testing.T) {
	assets := []Asset{
		{Symbol: "AAPL", Name: "Apple"},
		{Symbol: "TSLA", Name: ""},
		{Symbol: "GOOG", Name: "Google"},
	}

	cleaned := cleanAssets(assets)

	if len(cleaned) != 2 {
		t.Fatalf("expected 2 assets after cleaning, got %d", len(cleaned))
	}

	for _, a := range cleaned {
		if a.Name == "" {
			t.Errorf("cleanAssets did not remove empty name asset")
		}
	}
}

func TestClearExpiration(t *testing.T) {
	exp := time.Now().Add(24 * time.Hour)

	assets := []Asset{
		{Symbol: "AAPL", Expiration: &exp},
		{Symbol: "TSLA", Expiration: &exp},
	}

	cleared := clearExpiration(assets)

	for _, asset := range cleared {
		if asset.Expiration != nil {
			t.Errorf("expected Expiration to be nil")
		}
	}

	if assets[0].Expiration == nil {
		t.Errorf("original slice should not be modified")
	}
}

func TestClearExpirationCreatesCopy(t *testing.T) {
	exp := time.Now().Add(24 * time.Hour)

	assets := []Asset{
		{Symbol: "AAPL", Expiration: &exp},
	}

	cleared := clearExpiration(assets)

	if reflect.DeepEqual(assets, cleared) {
		t.Errorf("expected different slices after clearing expiration")
	}
}
