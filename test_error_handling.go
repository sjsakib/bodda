package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	// "os"
)

func main_test() {
	baseURL := "http://localhost:8080"
	
	// Test 1: Try to access sessions without authentication
	fmt.Println("Testing unauthenticated access to /api/sessions...")
	resp, err := http.Get(baseURL + "/api/sessions")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	fmt.Printf("Status: %d\n", resp.StatusCode)
	
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("Response: %+v\n\n", result)
	
	// Test 2: Try to create session with invalid data
	fmt.Println("Testing session creation with invalid data...")
	invalidData := map[string]interface{}{
		"title": string(make([]byte, 300)), // Too long title
	}
	
	jsonData, _ := json.Marshal(invalidData)
	resp2, err := http.Post(baseURL + "/api/sessions", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp2.Body.Close()
	
	fmt.Printf("Status: %d\n", resp2.StatusCode)
	
	var result2 map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&result2)
	fmt.Printf("Response: %+v\n\n", result2)
	
	// Test 3: Health check should work
	fmt.Println("Testing health check...")
	resp3, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp3.Body.Close()
	
	fmt.Printf("Status: %d\n", resp3.StatusCode)
	
	var result3 map[string]interface{}
	json.NewDecoder(resp3.Body).Decode(&result3)
	fmt.Printf("Response: %+v\n", result3)
}