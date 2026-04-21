package models

// Recipe representa la información limpia de una receta para Android.
// Usamos "etiquetas" (tags) para decirle a Go cómo se llama el campo en el JSON.
type Recipe struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Image string `json:"image"`
}

// SearchResponse es la "caja" principal que contiene la lista de recetas.
// Al no poner "offset" ni "totalResults" aquí, Go los eliminará automáticamente.
type SearchResponse struct {
	Results []Recipe `json:"results"`
}

type Ingredient struct {
	Original string `json:"original"` // Ej: "2 cups of flour"
}

// InstructionStep representa un paso de la receta
type InstructionStep struct {
	Number int    `json:"number"`
	Step   string `json:"step"`
}

// Instruction agrupa los pasos (Spoonacular lo envía como un arreglo de arreglos)
type Instruction struct {
	Steps []InstructionStep `json:"steps"`
}

// RecipeDetail es la "caja" principal para los detalles de UNA sola receta
type RecipeDetail struct {
	ID                   int           `json:"id"`
	Title                string        `json:"title"`
	Image                string        `json:"image"`
	ReadyInMinutes       int           `json:"readyInMinutes"`
	Servings             int           `json:"servings"`
	ExtendedIngredients  []Ingredient  `json:"extendedIngredients"`
	AnalyzedInstructions []Instruction `json:"analyzedInstructions"`
}
