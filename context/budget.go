package context

// CharsPerToken is the rough heuristic for converting character count to token count.
// Gemini and most LLMs average ~4 characters per token for English text/code.
const CharsPerToken = 4

// Budget tracks remaining character budget for context building.
type Budget struct {
	Max       int
	Remaining int
}

// NewBudget creates a budget with the given max characters.
func NewBudget(maxChars int) Budget {
	return Budget{Max: maxChars, Remaining: maxChars}
}

// Spend deducts n characters from the budget. Returns false if insufficient.
func (b *Budget) Spend(n int) bool {
	if n > b.Remaining {
		return false
	}
	b.Remaining -= n
	return true
}

// CanAfford returns true if n characters fit in the remaining budget.
func (b *Budget) CanAfford(n int) bool {
	return n <= b.Remaining
}

// Used returns the number of characters consumed.
func (b *Budget) Used() int {
	return b.Max - b.Remaining
}

// TokenEstimate returns the approximate token count for characters consumed.
func (b *Budget) TokenEstimate() int {
	return b.Used() / CharsPerToken
}
