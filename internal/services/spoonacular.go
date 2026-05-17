package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
)

// SpoonacularClient maneja la configuración y estado de nuestra API externa
type SpoonacularClient struct {
	apiKeys []string
	index   int
	baseURL string
	mu      sync.Mutex // Mutex actúa como semáforo para rotar llaves de forma segura
}

// APIResponse es la estructura que devolveremos, similar al objeto de tu código JS
type APIResponse struct {
	Data      json.RawMessage // JSON flexible sin esquema estricto aún
	QuotaUsed string
	QuotaLeft string
	ActiveKey string
}

// NewSpoonacularClient es el constructor de nuestro servicio
func NewSpoonacularClient(keys []string) *SpoonacularClient {
	// Filtramos llaves vacías tal como en tu código JS
	var validKeys []string
	for _, k := range keys {
		if k != "" {
			validKeys = append(validKeys, k)
		}
	}

	return &SpoonacularClient{
		apiKeys: validKeys,
		index:   0,
		baseURL: "https://api.spoonacular.com/recipes",
	}
}

// rotateKey cambia al siguiente índice de forma segura
func (c *SpoonacularClient) rotateKey() bool {
	c.mu.Lock()         // Encendemos el semáforo rojo para otras peticiones
	defer c.mu.Unlock() // Al terminar la función, se pone en verde automáticamente

	if len(c.apiKeys) <= 1 {
		return false
	}

	c.index = (c.index + 1) % len(c.apiKeys)
	fmt.Printf("Rotando a la siguiente API key. Nuevo índice: %d\n", c.index)
	return true
}

// getCurrentKey obtiene la llave actual
func (c *SpoonacularClient) getCurrentKey() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.apiKeys) == 0 {
		return ""
	}
	return c.apiKeys[c.index]
}

// fetchWithRotation es el equivalente a tu función fetchWithRotation
func (c *SpoonacularClient) fetchWithRotation(urlBuilder func(key string) string) (*APIResponse, error) {
	maxAttempts := len(c.apiKeys)
	if maxAttempts == 0 {
		maxAttempts = 1
	}

	for attempts := 0; attempts < maxAttempts; attempts++ {
		apiKey := c.getCurrentKey()
		targetURL := urlBuilder(apiKey)

		// Hacemos la petición GET
		resp, err := http.Get(targetURL)

		if err != nil {
			// Error de red (ej. sin internet), intentamos rotar si es el primer intento
			if attempts == 0 && c.rotateKey() {
				continue
			}
			return nil, fmt.Errorf("error de red: %w", err)
		}

		// Si el status es 402 o 429, la cuota se excedió [cite: 130]
		if resp.StatusCode == 402 || resp.StatusCode == 429 {
			fmt.Printf("API Key %s excedió su cuota (Status %d).\n", apiKey, resp.StatusCode)
			resp.Body.Close() // Siempre debemos cerrar el cuerpo de la respuesta en Go
			if c.rotateKey() {
				continue
			}
		}

		// Leemos los headers de cuota
		quotaUsed := resp.Header.Get("x-api-quota-used")
		quotaLeft := resp.Header.Get("x-api-quota-left")

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("error en Spoonacular. Status code: %d", resp.StatusCode)
		}

		// Decodificamos la respuesta JSON
		var rawData json.RawMessage
		err = json.NewDecoder(resp.Body).Decode(&rawData)
		resp.Body.Close()

		if err != nil {
			return nil, fmt.Errorf("error decodificando respuesta: %w", err)
		}

		return &APIResponse{
			Data:      rawData,
			QuotaUsed: quotaUsed,
			QuotaLeft: quotaLeft,
			ActiveKey: apiKey,
		}, nil
	}

	return nil, fmt.Errorf("todas las API keys han excedido su cuota o son inválidas")
}

// SearchRecipes construye la URL para la búsqueda compleja (CU-08) [cite: 127]
func (c *SpoonacularClient) SearchRecipes(query string, diets string, intolerances string) (*APIResponse, error) {
	urlBuilder := func(key string) string {
		// url.QueryEscape equivale al encodeURIComponent de JavaScript
		safeQuery := url.QueryEscape(query)

		// Construimos la URL base
		baseURL := fmt.Sprintf("%s/complexSearch?query=%s&apiKey=%s&number=5", c.baseURL, safeQuery, key)

		// Si hay dietas, las concatenamos
		if diets != "" {
			baseURL += fmt.Sprintf("&diet=%s", url.QueryEscape(diets))
		}

		// Si hay intolerancias, las concatenamos
		if intolerances != "" {
			baseURL += fmt.Sprintf("&intolerances=%s", url.QueryEscape(intolerances))
		}

		return baseURL
	}

	return c.fetchWithRotation(urlBuilder)
}

// GetRecipeInformation construye la URL para obtener detalles (CU-09) [cite: 139]
func (c *SpoonacularClient) GetRecipeInformation(id string) (*APIResponse, error) {
	urlBuilder := func(key string) string {
		return fmt.Sprintf("%s/%s/information?apiKey=%s", c.baseURL, id, key)
	}

	return c.fetchWithRotation(urlBuilder)
}
