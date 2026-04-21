package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rodrigoHol/Backend-SmartBite/internal/models"
	"github.com/rodrigoHol/Backend-SmartBite/internal/services"
)

// RecetasHandler agrupa todas las funciones relacionadas con las rutas de recetas
type RecetasHandler struct {
	client *services.SpoonacularClient
}

// NewRecetasHandler es el constructor
func NewRecetasHandler(c *services.SpoonacularClient) *RecetasHandler {
	return &RecetasHandler{client: c}
}

// Buscar maneja la ruta GET /api/recetas/buscar
func (h *RecetasHandler) Buscar(w http.ResponseWriter, r *http.Request) {
	ingrediente := r.URL.Query().Get("query")

	if ingrediente == "" {
		http.Error(w, `{"error": "Debes incluir un parámetro 'query', ej: ?query=pollo"}`, http.StatusBadRequest)
		return
	}

	// 1. Buscamos en Spoonacular usando tu servicio
	resultado, err := h.client.SearchRecipes(ingrediente)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// 2. MAGIA DE GO: Pasamos el JSON crudo a nuestra estructura limpia
	var datosLimpios models.SearchResponse
	if err := json.Unmarshal(resultado.Data, &datosLimpios); err != nil {
		http.Error(w, `{"error": "Error al procesar los datos de la receta"}`, http.StatusInternalServerError)
		return
	}

	// 3. Enviamos el nuevo JSON (ya limpio) a Android
	w.Header().Set("Content-Type", "application/json")

	// json.NewEncoder toma nuestros datos limpios y los escribe directamente en la respuesta web
	json.NewEncoder(w).Encode(datosLimpios)
}

func (h *RecetasHandler) Detalle(w http.ResponseWriter, r *http.Request) {
	// Leemos el ID de la receta desde la URL
	recetaID := r.URL.Query().Get("id")

	if recetaID == "" {
		http.Error(w, `{"error": "Debes incluir el parámetro 'id', ej: ?id=647956"}`, http.StatusBadRequest)
		return
	}

	// 1. Buscamos los detalles en Spoonacular
	resultado, err := h.client.GetRecipeInformation(recetaID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// 2. Pasamos el JSON crudo a nuestra estructura limpia (RecipeDetail)
	var detalleLimpio models.RecipeDetail
	if err := json.Unmarshal(resultado.Data, &detalleLimpio); err != nil {
		http.Error(w, `{"error": "Error al procesar los detalles de la receta"}`, http.StatusInternalServerError)
		return
	}

	// 3. Enviamos el nuevo JSON (ya limpio) a Android
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detalleLimpio)
}
