package main

import (
	"fmt"
	"log"
	"net/http"
	"os"      // Importado para leer variables del sistema (.env)
	"strings" // Importado para separar nuestra lista de llaves

	"github.com/joho/godotenv" // Importado para cargar el archivo .env
	// IMPORTANTE: Mantén la ruta correcta de tu proyecto
	"github.com/rodrigoHol/Backend-SmartBite/internal/services"
)

func main() {
	// 1. Cargar el archivo .env
	// godotenv.Load() busca un archivo llamado ".env" en la raíz y carga sus variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Advertencia: No se encontró el archivo .env o no se pudo cargar.")
	}

	// 2. Leer la variable desde el entorno
	// Usamos VITE_SPOONACULAR_API_KEYS tal como la definiste
	keysString := os.Getenv("SPOONACULAR_API_KEYS")

	// Validamos que no esté vacía para evitar errores más adelante
	if keysString == "" {
		log.Fatal("Error: La variable VITE_SPOONACULAR_API_KEYS no está definida en el archivo .env")
	}

	// 3. Separar las llaves
	// Si en tu .env pones: VITE_SPOONACULAR_API_KEYS=llave1,llave2
	// strings.Split lo convierte en un arreglo de Go: []string{"llave1", "llave2"}
	apiKeys := strings.Split(keysString, ",")

	// 4. Configurar las llaves de la API en el cliente
	clienteRecetas := services.NewSpoonacularClient(apiKeys)

	// 5. Crear la ruta para buscar recetas (CU-08)
	http.HandleFunc("/api/recetas/buscar", func(w http.ResponseWriter, r *http.Request) {
		// Leemos lo que el usuario quiere buscar desde la URL (ej. ?query=pollo)
		ingrediente := r.URL.Query().Get("query")

		if ingrediente == "" {
			http.Error(w, `{"error": "Debes incluir un parámetro 'query', ej: ?query=pollo"}`, http.StatusBadRequest)
			return
		}

		// Usamos nuestro servicio para ir a Spoonacular
		resultado, err := clienteRecetas.SearchRecipes(ingrediente)

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		// Configuramos la respuesta para que Android sepa que es un JSON puro
		w.Header().Set("Content-Type", "application/json")

		// Escribimos el JSON crudo (RawMessage) en la respuesta
		w.Write(resultado.Data)
	})

	// 6. Encender el servidor
	puerto := ":8080"
	fmt.Printf("Servidor escuchando en http://localhost%s\n", puerto)
	log.Fatal(http.ListenAndServe(puerto, nil))
}
