package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

// Global variable for database connection
var db *sql.DB

// Initializes the SQLite database and creates the `cotacoes` table if it doesn't exist
func init() {
	var err error
	db, err = sql.Open("sqlite3", "./database/database.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Create `cotacoes` table if it doesn't already exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS cotacoes (id INTEGER PRIMARY KEY, bid TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
}

// Cotacao struct to store the dollar exchange rate
type Cotacao struct {
	Bid string `json:"bid"`
}

// Handler function to retrieve dollar exchange rate and respond to the client
func getCotacaoHandler(w http.ResponseWriter, r *http.Request) {
	// Timeout context set to 2 seconds for the external API call
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// Create the request to the external exchange rate API
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, "Error creating request to external API", http.StatusInternalServerError)
		return
	}

	// Initialize the HTTP client with a 2-second timeout
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error calling exchange rate API: %v", err)
		http.Error(w, "Error calling exchange rate API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Log the HTTP status of the API response
	log.Printf("API response status: %d", resp.StatusCode)

	// Decode the JSON response from the external API into the `Cotacao` struct
	var data map[string]Cotacao
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Printf("Error decoding external API response: %v", err)
		http.Error(w, "Error decoding external API response", http.StatusInternalServerError)
		return
	}

	// Extract the exchange rate value from the decoded data
	bid := data["USDBRL"].Bid
	log.Printf("Exchange rate obtained: %s", bid)

	// Save the exchange rate to the database
	if err := saveCotacao(bid); err != nil {
		log.Printf("Error saving exchange rate to database: %v", err)
		http.Error(w, "Error saving exchange rate to database", http.StatusInternalServerError)
		return
	}

	// Send the exchange rate as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"bid": bid})
}

// Function to save the exchange rate to the database with a 10ms timeout
func saveCotacao(bid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := db.ExecContext(ctx, "INSERT INTO cotacoes (bid) VALUES (?)", bid)
	return err
}

func main() {
	// Set up router and route handler for `/cotacao` endpoint
	r := mux.NewRouter()
	r.HandleFunc("/cotacao", getCotacaoHandler).Methods("GET")

	log.Println("Server running on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
