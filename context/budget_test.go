package context

import "testing"

func TestBudget_Spend(t *testing.T) {
	b := NewBudget(100)

	if !b.Spend(50) {
		t.Error("expected Spend(50) to succeed")
	}
	if b.Remaining != 50 {
		t.Errorf("remaining = %d, want 50", b.Remaining)
	}

	if !b.Spend(50) {
		t.Error("expected Spend(50) to succeed (exact)")
	}
	if b.Remaining != 0 {
		t.Errorf("remaining = %d, want 0", b.Remaining)
	}

	if b.Spend(1) {
		t.Error("expected Spend(1) to fail when exhausted")
	}
}

func TestBudget_CanAfford(t *testing.T) {
	b := NewBudget(100)
	b.Spend(90)

	if !b.CanAfford(10) {
		t.Error("expected CanAfford(10) with 10 remaining")
	}
	if b.CanAfford(11) {
		t.Error("expected !CanAfford(11) with 10 remaining")
	}
}

func TestBudget_Used(t *testing.T) {
	b := NewBudget(1000)
	b.Spend(400)

	if b.Used() != 400 {
		t.Errorf("Used() = %d, want 400", b.Used())
	}
}

func TestBudget_TokenEstimate(t *testing.T) {
	b := NewBudget(2000)
	b.Spend(800)

	if b.TokenEstimate() != 200 {
		t.Errorf("TokenEstimate() = %d, want 200", b.TokenEstimate())
	}
}
