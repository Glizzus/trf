package main

import (
	"os"

	"github.com/sashabaranov/go-openai"

	"github.com/glizzus/trf/ministry/internal"
)

func main() {

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		panic("OPENAI_API_KEY is not set")
	}

	client := openai.NewClient(apiKey)
	promptProvider := NewFilePromptProvider("system.tmpl", "user.tmpl")

	spoofer := NewOpenAISpoofer(client, promptProvider)

	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content := r.FormValue("content")
		if content == "" {
			http.Error(w, "content is required", http.StatusBadRequest)
			return
		}

		rating := r.FormValue("rating")
		if rating == "" {
			http.Error(w, "rating is required", http.StatusBadRequest)
			return
		}

		timeStart := time.Now()
		spoofed, err := spoofer.Spoof(r.Context(), content, rating)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		timeToSpoof := time.Since(timeStart).Seconds()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SpoofResponse{
			Content:     spoofed,
			TimeToSpoof: timeToSpoof,
		})
	}))
}

type SpoofResponse struct {
	Content     string  `json:"content"`
	TimeToSpoof float64 `json:"time_to_spoof"`
}
