package entities_test

import (
	"testing"

	"warehouse.local/core/entities"
)

func TestSKUFactoryRejectsBadInputAndComparesByValue(t *testing.T) {
	t.Parallel()

	valid, err := entities.NewSKU("HW-001")
	if err != nil {
		t.Fatalf("expected valid sku: %v", err)
	}
	same, err := entities.NewSKU("HW-001")
	if err != nil {
		t.Fatalf("expected same sku to be valid: %v", err)
	}
	if !valid.Equal(same) {
		t.Fatal("expected sku value equality")
	}

	for _, candidate := range []string{"", "ab", "lower-123", "BAD_SKU", "TOO-LONG-ABCDEFGHIJKLMNOPQRSTUVWXYZ"} {
		if _, err := entities.NewSKU(candidate); err == nil {
			t.Fatalf("expected invalid sku %q to be rejected", candidate)
		}
	}
}

func TestMoneyFactoryRejectsBadInputAndComparesByValue(t *testing.T) {
	t.Parallel()

	zero, err := entities.NewMoney(0, "EUR")
	if err != nil {
		t.Fatalf("zero money should be valid as value object: %v", err)
	}
	same, err := entities.NewMoney(0, "EUR")
	if err != nil {
		t.Fatalf("expected same money to be valid: %v", err)
	}
	if !zero.Equal(same) {
		t.Fatal("expected money value equality")
	}

	for _, tc := range []struct {
		cents    int64
		currency string
	}{
		{-1, "EUR"},
		{100, ""},
		{100, "EU"},
		{100, "EURO"},
		{100, "eur"},
	} {
		if _, err := entities.NewMoney(tc.cents, tc.currency); err == nil {
			t.Fatalf("expected invalid money %+v to be rejected", tc)
		}
	}
}
