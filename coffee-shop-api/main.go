package main

import (
	"coffee-shop-api/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Base de données en mémoire
var drinks []models.Drink
var orders []models.Order
var orderCounter int = 1

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GET /menu - Récupérer le menu complet
func getMenu(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(drinks)
}

// fonction 2 getdrink recupérer une boisson spécifique
func getDrink(w http.ResponseWriter, r *http.Request) {
	// definir le header content type
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["id"]
	// parcourir le slice drinks si trouver encoder la boisson
	for _, drink := range drinks {
		if drink.ID == id {
			json.NewEncoder(w).Encode(drink)
			return
		}
	}
	http.Error(w, "la boisson n'est pas là", http.StatusNotFound)
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var order models.Order
	err := json.NewDecoder(r.Body).Decode(&order)

	// gérer l'erreur de décodage 400 Bad request
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// verifier que drinkID existe dans drinks
	var drinkFound *models.Drink
	for _, drink := range drinks {
		if drink.ID == order.DrinkID {
			drinkFound = &drink
			break
		}
	}

	if drinkFound == nil {
		http.Error(w, "Drink not found", http.StatusNotFound)
		return
	}

	// Remplir les champs de la commande AVANT d'ajouter au slice
	order.ID = fmt.Sprintf("ORD-%03d", orderCounter)
	orderCounter++
	order.DrinkName = drinkFound.Name
	order.Status = models.StatusPending
	order.OrderedAt = time.Now()
	order.TotalPrice = calculatePrice(drinkFound.BasePrice, order.Size, order.Extras)

	// ajouter la commande au slice UNE SEULE FOIS
	orders = append(orders, order)

	// retourner 201 Created avec la commande en JSON
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}
func calculatePrice(basePrice float64, size string, extras []string) float64 {
	//partir du base price
	var total float64
	total = basePrice
	//ajuster selon la taille
	switch size {
	case "medium":
		total += 0.50
	case "large":
		total += 1.00
	}
	//ajouter 0.50 pour chaque extra
	total += 0.50 * float64(len(extras))
	return total
}

// GET /orders - recupérer toutes les commandes
func getOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// GET /orders/{id} - récupérer une commande spécifique
func getOrder(w http.ResponseWriter, r *http.Request) {
	//header content type
	w.Header().Set("Content-Type", "application/json")
	// recupérer l'id depuis les routes
	vars := mux.Vars(r)
	id := vars["id"]
	// parcourir le slice
	for _, order := range orders {
		if order.ID == id {
			json.NewEncoder(w).Encode(order)
			return
		}

	}
	http.Error(w, "erreur 404", http.StatusNotFound)
}

func updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["id"]
	//creer une struct temporaire avec un champs status parcourir les orders et trouver la commande
	var statusUpdate struct {
		Status string `json:"status"`
	}
	err := json.NewDecoder(r.Body).Decode(&statusUpdate)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	for i, order := range orders {
		if order.ID == id {
			orders[i].Status = models.OrderStatus(statusUpdate.Status)
			json.NewEncoder(w).Encode(orders[i])
			return
		}
	}
	http.Error(w, "erreur 404", http.StatusNotFound)
}

// delete /orders/{id} - annuler une commande
func deleteOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["id"]
	// parcourir l'orders avec l'index si la commande est trouvée verifier que le statut n'est pas picked-up si oui retouner 404 bad request sinon supprimer la commande du slice
	for i, order := range orders {
		if order.ID == id {
			if order.Status == models.StatusPickedUp {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			orders = append(orders[:i], orders[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}

	}
	http.Error(w, "order not found", http.StatusNotFound)
}

func main() {
	drinks = []models.Drink{
		{ID: "1", Name: "Espresso", Category: "coffee", BasePrice: 2.00},
		{ID: "2", Name: "Cappuccino", Category: "coffee", BasePrice: 3.00},
		{ID: "3", Name: "Latte", Category: "coffee", BasePrice: 3.50},
		{ID: "4", Name: "Green Tea", Category: "tea", BasePrice: 2.50},
		{ID: "5", Name: "Matcha latte", Category: "tea", BasePrice: 4.00},
		{ID: "6", Name: "Iced Coffee", Category: "cold", BasePrice: 3.00},
		{ID: "7", Name: "Iced Latte", Category: "cold", BasePrice: 3.50},
	}
	// routeur
	r := mux.NewRouter()
	// definir les routes
	r.HandleFunc("/menu", getMenu).Methods("GET")
	r.HandleFunc("/menu/{id}", getDrink).Methods("GET")
	r.HandleFunc("/orders", createOrder).Methods("POST")
	r.HandleFunc("/orders/{id}", getOrder).Methods("GET")
	r.HandleFunc("/orders", getOrders).Methods("GET")
	r.HandleFunc("/orders/{id}", updateOrderStatus).Methods("PATCH")
	r.HandleFunc("/orders/{id}", deleteOrder).Methods("DELETE")
	// ajouter une route get pour un message de bienvenue
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Bienvenue au COFFEE SHOP"}`))
	}).Methods("GET")
	// afficher un message au demarrage du serveur
	fmt.Println("Le serveur démarre...")
	// demarrer le serveur
	err := http.ListenAndServe(":8080", corsMiddleware(r))
	if err != nil {
		log.Fatalf("Le serveur a échoué à démarrer: %v", err)
	}

}
