package domain

// Kategoria wydatku. Kategorie domyślne są słownikiem
// zarządzanym przez administratora.
type Category struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}

// Słownik kategorii zgodny ze słownikiem projektu.
func DefaultCategories() []Category {
	names := []struct{ id, name string }{
		{"streaming", "Streaming"},
		{"software", "Oprogramowanie"},
		{"cloud", "Chmura"},
		{"education", "Edukacja"},
		{"other", "Inne"},
	}
	out := make([]Category, 0, len(names))
	for _, n := range names {
		out = append(out, Category{ID: n.id, Name: n.name, IsDefault: true})
	}
	return out
}
