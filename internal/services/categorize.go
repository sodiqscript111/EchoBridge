package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"EchoBridge/db"

	"github.com/google/uuid"
)

const (
	GeminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"
)

// CategorizePlaylist analyzes the playlist tracks and assigns a category using AI
func CategorizePlaylist(ctx context.Context, playlistID uuid.UUID) error {
	// 1. Fetch Playlist and Tracks
	var playlist db.Playlist
	if err := db.DB.Preload("Tracks").First(&playlist, "id = ?", playlistID).Error; err != nil {
		return fmt.Errorf("failed to fetch playlist: %w", err)
	}

	if len(playlist.Tracks) == 0 {
		log.Println("Playlist has no tracks, skipping categorization")
		return nil
	}

	// 2. Select up to 7 tracks
	numTracks := len(playlist.Tracks)
	if numTracks > 7 {
		numTracks = 7
	}
	selectedTracks := playlist.Tracks[:numTracks]

	// 3. Construct JSON for Prompt
	var tracksData []map[string]string
	for _, t := range selectedTracks {
		tracksData = append(tracksData, map[string]string{
			"title":  t.Title,
			"artist": t.Artist,
		})
	}
	tracksJSON, _ := json.Marshal(tracksData)

	// 4. Create Prompt
	prompt := fmt.Sprintf(`
You are a music expert. Categorize the following playlist based on these %d tracks:
%s

Choose ONE category from this list:
- Chill
- Hot songs
- Party
- Focus
- Indie
- Rock
- Hippop
- rage rap
- late night
- Everything in one

Return ONLY the category name as a JSON object like {"category": "Indie"}. Do not add any markdown formatting.
`, numTracks, string(tracksJSON))

	// 5. Call AI API
	category, err := callGeminiAPI(ctx, prompt)
	if err != nil {
		log.Printf("AI Categorization failed: %v. Using default.", err)
		// Optional: Set a default or leave empty
		return err
	}

	// 6. Update Database
	if err := db.DB.Model(&playlist).Update("category", category).Error; err != nil {
		return fmt.Errorf("failed to update playlist category: %w", err)
	}

	return nil
}

func callGeminiAPI(ctx context.Context, prompt string) (string, error) {
	apiKey := "AIzaSyDM1w0SpT712k7-gT8FYbCH7ZhCE199kIw"
	if apiKey == "" {
		// Mock response for development if no key is present
		log.Println("GEMINI_API_KEY not set. Using mock response.")
		log.Printf("Prompt sent to AI:\n%s", prompt)
		return "Indie", nil // Default mock category
	}

	requestBody, _ := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
	})

	url := fmt.Sprintf("%s?key=%s", GeminiAPIURL, apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	text := response.Candidates[0].Content.Parts[0].Text
	// Clean up response (remove markdown code blocks if present)
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")

	var result struct {
		Category string `json:"category"`
	}
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		// Fallback: try to find the category word directly if JSON parsing fails
		log.Printf("Failed to parse JSON response: %s. Trying raw text match.", text)
		validCategories := []string{"Chill", "Hot songs", "Party", "Focus", "Indie", "Rock", "Hippop", "rage rap", "late night", "Everything in one"}
		for _, cat := range validCategories {
			if strings.Contains(strings.ToLower(text), strings.ToLower(cat)) {
				return cat, nil
			}
		}
		return "", fmt.Errorf("failed to parse category from response: %s", text)
	}

	return result.Category, nil
}
