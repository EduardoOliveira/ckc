package enrichment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Define the structs that match the JSON schema for the API response.
type APIResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type CategorizationResult struct {
	Username   string     `json:"username"`
	Categories []Category `json:"categories"`
}

type Category struct {
	Name       string           `json:"name"`
	Confidence float64          `json:"confidence"`
	TechType   string           `json:"tech_type,omitempty"`
	Details    *CategoryDetails `json:"details,omitempty"`
}

type CategoryDetails struct {
	Language        string   `json:"language"`
	CommonCountries []string `json:"common_countries"`
}

// Main function to run the categorization.
func main() {
	// Replace with your actual API key
	apiKey := ""

	if apiKey == "" {
		fmt.Println("Error: API key is not set. Please provide your API key.")
		return
	}

	// Example username to categorize
	username := "jenkins"

	// Define the detailed prompt for the LLM.
	prompt := fmt.Sprintf(`
Categorize the username "%s" based on its likely origin or purpose. 
The categories should be from the following list:
1. System and Administrative Accounts (e.g., 'root', 'admin')
2. Cloud and Virtualization Defaults (e.g., 'ec2-user', 'ubuntu')
3. Application and Service Accounts (e.g., 'mysql', 'jenkins'). If this category is chosen, also classify it as either 'Corporate' or 'Home/Homelab' technology.
4. IoT and Embedded Device Defaults (e.g., 'pi', 'ubnt')
5. Generic, Test, and Simple Usernames (e.g., 'user', 'test')
6. Common Human Names (e.g., 'thomas', 'wang')
7. Other/Uncategorized

If the username is a common human name, provide the language and a few common country codes (ISO 3166-1 alpha-2) where it is used.

Return the response as a JSON object that strictly adheres to the following schema. The confidence should be a value from 0.0 to 1.0. The 'details' object should only be present if the category is 'Common Human Names'. The 'tech_type' property should only be present if the category is 'Application and Service Accounts'.
`, username)

	// Define the JSON schema for the structured response.
	// This is a direct string representation of the schema.
	generationConfigJSON := `{
		"responseMimeType": "application/json",
		"responseSchema": {
			"type": "OBJECT",
			"properties": {
				"username": { "type": "STRING" },
				"categories": {
					"type": "ARRAY",
					"items": {
						"type": "OBJECT",
						"properties": {
							"name": { "type": "STRING" },
							"confidence": { "type": "NUMBER" },
							"tech_type": { "type": "STRING" },
							"details": {
								"type": "OBJECT",
								"properties": {
									"language": { "type": "STRING" },
									"common_countries": { "type": "ARRAY", "items": { "type": "STRING" } }
								}
							}
						},
						"propertyOrdering": ["name", "confidence", "tech_type", "details"]
					}
				}
			},
			"propertyOrdering": ["username", "categories"]
		}
	}`

	// Create the request payload
	requestPayload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
		"generationConfig": json.RawMessage(generationConfigJSON),
	}

	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		fmt.Println("Error marshalling request payload:", err)
		return
	}

	// Make the API request with exponential backoff
	const maxRetries = 5
	var apiResponseBytes []byte
	apiUrl := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent?key=%s", apiKey)

	for i := 0; i < maxRetries; i++ {
		resp, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			fmt.Printf("API call failed: %v. Retrying...\n", err)
			time.Sleep(time.Duration(1<<i) * time.Second) // Exponential backoff
			continue
		}
		defer resp.Body.Close()

		apiResponseBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Failed to read response body: %v. Retrying...\n", err)
			time.Sleep(time.Duration(1<<i) * time.Second)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("API call failed with status code: %d. Response: %s. Retrying...\n", resp.StatusCode, string(apiResponseBytes))
			time.Sleep(time.Duration(1<<i) * time.Second)
			continue
		}

		break // Success, break out of the loop
	}

	if apiResponseBytes == nil {
		fmt.Println("Failed to get a successful response after multiple retries.")
		return
	}

	// Parse the main API response
	var apiResponse APIResponse
	if err := json.Unmarshal(apiResponseBytes, &apiResponse); err != nil {
		fmt.Println("Error unmarshalling API response:", err)
		return
	}

	// Extract and parse the nested JSON
	if len(apiResponse.Candidates) > 0 && len(apiResponse.Candidates[0].Content.Parts) > 0 {
		jsonString := apiResponse.Candidates[0].Content.Parts[0].Text

		var result CategorizationResult
		if err := json.Unmarshal([]byte(jsonString), &result); err != nil {
			fmt.Println("Error unmarshalling final result JSON:", err)
			return
		}

		// Print the structured result
		prettyResult, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println("--- Categorization Result ---")
		fmt.Println(string(prettyResult))
	} else {
		fmt.Println("No valid content found in the API response.")
	}
}
