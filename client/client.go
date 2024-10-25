package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Cotacao struct to store the dollar exchange rate received from the server
type Cotacao struct {
	Bid string `json:"bid"`
}

// Function to fetch the dollar exchange rate from the server
func fetchCotacao() error {
	// Sets a 300ms timeout for the HTTP request
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Create a GET request to the server's `/cotacao` endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Make the HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error fetching exchange rate: %w", err)
	}
	defer resp.Body.Close()

	// Check if the server responded with a 200 OK status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error fetching exchange rate: status %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}
	log.Printf("Server response: %s", string(body))

	// Decode the JSON response body into the `Cotacao` struct
	var cotacao Cotacao
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&cotacao); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	// Save the exchange rate to `cotacao.txt`
	return saveCotacao(cotacao.Bid)
}

// Function to save the exchange rate in `cotacao.txt`
func saveCotacao(bid string) error {
	content := fmt.Sprintf("Dollar: %s", bid)
	return os.WriteFile("cotacao.txt", []byte(content), 0644)
}

func main() {
	// Fetch the exchange rate and save it, logging any errors that occur
	if err := fetchCotacao(); err != nil {
		log.Fatalf("Error fetching exchange rate: %v", err)
	} else {
		log.Println("Exchange rate successfully saved in cotacao.txt")
	}
}
