package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type AskRequest struct {
	Question string `json:"question" binding:"required"`
}

type AskResponse struct {
	Answer string `json:"answer"`
}

type EmbedRequest struct {
	Text string `json:"text"`
}

type EmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type StreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func main() {
	r := gin.Default()
	connStr := os.Getenv("POSTGRES_DSN")
	if connStr == "" {
		connStr = "postgres://postgres:jCfAaeEOZJ@localhost:5432/ragdb?sslmode=disable" // fallback
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	r.POST("/api/ask", func(c *gin.Context) {
		var req AskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid question field"})
			return
		}

		// Marshal the request to JSON
		embedPayload, err := json.Marshal(EmbedRequest{Text: req.Question})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal embed request"})
			return
		}

		// Send POST to Python service
		embeddingURL := os.Getenv("EMBEDDING_SERVICE_URL")
		if embeddingURL == "" {
			embeddingURL = "http://localhost:5001/embed" // fallback for dev
		}

		fmt.Printf("Embedding service URL: %s\n", embeddingURL)

		resp, err := http.Post(embeddingURL, "application/json", bytes.NewBuffer(embedPayload))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact embedding service"})
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read embedding response"})
			return
		}

		var embedResp EmbedResponse
		if err := json.Unmarshal(body, &embedResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid embedding response format"})
			return
		}

		fmt.Printf("Received embedding: %v\n", embedResp.Embedding)

		// Static  answer
		results, err := querySimilarChunks(db, embedResp.Embedding, 3)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed vector search", "details": err.Error()})
			return
		}

		streamLLMResponse(c, req.Question, results)
	})

	r.Run(":8080")
}
func querySimilarChunks(db *sql.DB, embedding []float64, limit int) ([]string, error) {
	if len(embedding) != 384 {
		return nil, fmt.Errorf("embedding length is %d, expected 384", len(embedding))
	}

	// Format the Go slice as a Postgres-compatible vector string
	var sb strings.Builder
	sb.WriteString("[")
	for i, val := range embedding {
		sb.WriteString(fmt.Sprintf("%f", val))
		if i < len(embedding)-1 {
			sb.WriteString(",")
		}
	}
	sb.WriteString("]")
	vectorStr := sb.String()

	// Parameterized query: vectorStr becomes $2
	query := `
		SELECT chunk 
		FROM embeddings 
		ORDER BY embedding <-> $2::vector 
		LIMIT $1;
	`

	rows, err := db.Query(query, limit, vectorStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var chunk string
		if err := rows.Scan(&chunk); err != nil {
			return nil, err
		}
		results = append(results, chunk)
	}
	return results, nil
}
func streamLLMResponse(c *gin.Context, question string, contextChunks []string) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Missing OPENAI_API_KEY"})
		return
	}

	// Construct system + user prompt
	systemPrompt := "You are a helpful assistant. Use the following context to answer the user's question."
	contextText := strings.Join(contextChunks, "\n\n")

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("Context:\n%s\n\nQuestion:\n%s", contextText, question)},
	}

	reqBody := ChatRequest{
		Model:    "gpt-3.5-turbo",
		Messages: messages,
		Stream:   true,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal OpenAI request"})
		return
	}

	// Create streaming POST request
	req, err := http.NewRequestWithContext(context.Background(), "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create OpenAI request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact OpenAI API"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, gin.H{"error": "OpenAI error", "details": string(body)})
		return
	}

	// Set response headers for streaming
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(http.StatusOK)

	writer := c.Writer
	flusher, ok := writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			raw := strings.TrimPrefix(line, "data: ")
			if raw == "[DONE]" {
				break
			}

			var chunk StreamChunk
			if err := json.Unmarshal([]byte(raw), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 {
				token := chunk.Choices[0].Delta.Content
				if token != "" {
					fmt.Fprintf(writer, "%s", token)
					flusher.Flush()
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Stream read error: %v", err)
	}
}
