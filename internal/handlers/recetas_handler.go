package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/rodrigoHol/Backend-SmartBite/internal/models"
	"github.com/rodrigoHol/Backend-SmartBite/internal/services"
)

// RecetasHandler agrupa todas las funciones relacionadas con las rutas de recetas
type RecetasHandler struct {
	client *services.SpoonacularClient
	cache  *services.RecipeCache
}

// NewRecetasHandler es el constructor
func NewRecetasHandler(c *services.SpoonacularClient, cache *services.RecipeCache) *RecetasHandler {
	return &RecetasHandler{client: c, cache: cache}
}

// Buscar maneja la ruta GET /api/recetas/buscar
func (h *RecetasHandler) Buscar(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	ingrediente := queryParams.Get("query")
	ingrediente = strings.ToLower(strings.TrimSpace(ingrediente))

	if ingrediente == "" {
		http.Error(w, `{"error": "Debes incluir un parámetro 'query', ej: ?query=pollo"}`, http.StatusBadRequest)
		return
	}

	// 1. Extraer dietas e intolerancias (devuelve []string)
	dietsList := queryParams["diet"]
	intolerancesList := queryParams["intolerance"]

	// Unir arreglos en un string separado por comas para Spoonacular ("vegan,vegetarian")
	dietsStr := strings.Join(dietsList, ",")
	intolerancesStr := strings.Join(intolerancesList, ",")

	// 2. Crear una llave de caché única que incluya las preferencias
	// Ej: "chicken|vegan,vegetarian|gluten"
	cacheKey := fmt.Sprintf("%s|%s|%s", ingrediente, dietsStr, intolerancesStr)

	// 3. Intentamos obtener de la caché primero usando la nueva llave
	if cachedResult, found := h.cache.GetSearch(cacheKey); found {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cachedResult)
		return
	}

	// 4. Buscamos en Spoonacular pasando las preferencias
	resultado, err := h.client.SearchRecipes(ingrediente, dietsStr, intolerancesStr)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// 5. MAGIA DE GO: Pasamos el JSON crudo a nuestra estructura limpia
	var datosLimpios models.SearchResponse
	if err := json.Unmarshal(resultado.Data, &datosLimpios); err != nil {
		http.Error(w, `{"error": "Error al procesar los datos de la receta"}`, http.StatusInternalServerError)
		return
	}

	// 6. Guardamos en la caché usando la llave compuesta
	h.cache.SetSearch(cacheKey, datosLimpios)

	// 7. Enviamos el nuevo JSON (ya limpio) a Android
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(datosLimpios)
}

func (h *RecetasHandler) Detalle(w http.ResponseWriter, r *http.Request) {
	// Leemos el ID de la receta desde la URL
	recetaID := r.URL.Query().Get("id")
	recetaID = strings.TrimSpace(recetaID)

	if recetaID == "" {
		http.Error(w, `{"error": "Debes incluir el parámetro 'id', ej: ?id=647956"}`, http.StatusBadRequest)
		return
	}

	// 0. Intentamos obtener de la caché primero
	if cachedResult, found := h.cache.GetDetail(recetaID); found {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cachedResult)
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

	// 3. Guardamos en la caché para futuras peticiones
	h.cache.SetDetail(recetaID, detalleLimpio)

	// 4. Enviamos el nuevo JSON (ya limpio) a Android
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detalleLimpio)
}
