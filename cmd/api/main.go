package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rodrigoHol/Backend-SmartBite/internal/handlers"
	"github.com/rodrigoHol/Backend-SmartBite/internal/services"
)

func main() {
	godotenv.Load()
	keysString := os.Getenv("SPOONACULAR_API_KEYS")
	if keysString == "" {
		log.Fatal("Error: La variable SPOONACULAR_API_KEYS no está definida en el archivo .env")
	}
	apiKeys := strings.Split(keysString, ",")

	clienteRecetas := services.NewSpoonacularClient(apiKeys)
	cacheRecetas := services.NewRecipeCache()

	controladorRecetas := handlers.NewRecetasHandler(clienteRecetas, cacheRecetas)

	//Declaramos las rutas
	http.HandleFunc("/api/recetas/buscar", controladorRecetas.Buscar)
	http.HandleFunc("/api/recetas/detalle", controladorRecetas.Detalle)

	//Encendemos el servidor
	puerto := ":8080"
	fmt.Printf("Servidor escuchando en http://localhost%s\n", puerto)
	log.Fatal(http.ListenAndServe(puerto, nil))
}
