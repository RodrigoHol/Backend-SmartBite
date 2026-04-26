package services

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/rodrigoHol/Backend-SmartBite/internal/models"
)

const cacheFilePath = "data_cache.json"

// CacheData es la estructura que se guardará en el archivo JSON
type CacheData struct {
	SearchCache map[string]models.SearchResponse `json:"search_cache"`
	DetailCache map[string]models.RecipeDetail   `json:"detail_cache"`
}

// RecipeCache maneja el almacenamiento en memoria y en disco (ROM)
type RecipeCache struct {
	mu   sync.RWMutex
	data CacheData
}

// NewRecipeCache crea una nueva instancia de la caché y carga los datos desde el archivo
func NewRecipeCache() *RecipeCache {
	c := &RecipeCache{
		data: CacheData{
			SearchCache: make(map[string]models.SearchResponse),
			DetailCache: make(map[string]models.RecipeDetail),
		},
	}
	c.loadFromFile()
	return c
}

// loadFromFile lee el archivo JSON de disco y carga los datos en memoria
func (c *RecipeCache) loadFromFile() {
	fileData, err := os.ReadFile(cacheFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Archivo de caché no encontrado, se creará uno nuevo al guardar datos.")
		} else {
			fmt.Println("Error leyendo archivo de caché:", err)
		}
		return
	}

	if err := json.Unmarshal(fileData, &c.data); err != nil {
		fmt.Println("Error decodificando archivo de caché:", err)
	} else {
		fmt.Println("Caché cargado exitosamente desde disco. Búsquedas:", len(c.data.SearchCache), "Detalles:", len(c.data.DetailCache))
	}
}

// saveToFile escribe la caché actual al archivo JSON en disco
func (c *RecipeCache) saveToFile() {
	fileData, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		fmt.Println("Error codificando la caché para guardar:", err)
		return
	}

	if err := os.WriteFile(cacheFilePath, fileData, 0644); err != nil {
		fmt.Println("Error escribiendo archivo de caché:", err)
	}
}

// GetSearch obtiene los resultados de búsqueda de la caché
func (c *RecipeCache) GetSearch(query string) (models.SearchResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res, found := c.data.SearchCache[query]
	return res, found
}

// SetSearch guarda los resultados de búsqueda en la caché y en disco
func (c *RecipeCache) SetSearch(query string, response models.SearchResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data.SearchCache[query] = response
	c.saveToFile()
}

// GetDetail obtiene los detalles de una receta de la caché
func (c *RecipeCache) GetDetail(id string) (models.RecipeDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res, found := c.data.DetailCache[id]
	return res, found
}

// SetDetail guarda los detalles de una receta en la caché y en disco
func (c *RecipeCache) SetDetail(id string, response models.RecipeDetail) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data.DetailCache[id] = response
	c.saveToFile()
}
