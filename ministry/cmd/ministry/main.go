package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glizzus/trf/ministry/internal/prompt"
	"github.com/glizzus/trf/ministry/internal/spoof"
	"github.com/sashabaranov/go-openai"
)

type FileProviderPaths struct {
	System string
	User   string
}

func getTemplatePaths() (paths *FileProviderPaths, err error) {

	// The default template paths to check
	dirs := []string{
		".",
		"./templates",
		filepath.Join(os.Getenv("HOME"), ".config/ministry/templates"),
		"/etc/ministry/templates",
	}

	// If the user has set a custom path, add it to the beginning of the list
	if customPath := os.Getenv("MINISTRY_TEMPLATE_PATH"); customPath != "" {
		dirs = append([]string{customPath}, dirs...)
	}

	var systemPath, userPath string
	var systemFound, userFound bool

	for _, dir := range dirs {
		slog.Debug("Checking for templates", "dir", dir)

		if !systemFound {
			systemPath = filepath.Join(dir, "system.tmpl")
			if _, err := os.Stat(systemPath); err == nil {
				slog.Debug("Found system template", "systemPath", systemPath)
				systemFound = true
			}
		}

		if !userFound {
			userPath = filepath.Join(dir, "user.tmpl")
			if _, err := os.Stat(userPath); err == nil {
				slog.Debug("Found user template", "userPath", userPath)
				userFound = true
			}
		}

		if systemFound && userFound {
			break
		}
	}

	if !systemFound || !userFound {
		return nil, fmt.Errorf("could not find all required templates")
	}
	return &FileProviderPaths{System: systemPath, User: userPath}, nil
}

func getPromptProvider() prompt.Provider {
	prompt_type := strings.ToLower(os.Getenv("MINISTRY_PROMPT_TYPE"))
	if prompt_type == "" {
		prompt_type = "file"
	}

	switch prompt_type {
	case "file":
		paths, err := getTemplatePaths()
		if err != nil {
			log.Fatalf("Could not find templates: %v", err)
		}
		return prompt.NewFilePromptProvider(paths.System, paths.User)
	case "static":
		return &prompt.StaticProvider{}
	default:
		log.Fatalf("Unknown prompt type: %s", prompt_type)
		return nil
	}
}

func main() {

	if os.Getenv("MINISTRY_DEBUG") != "" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	log.Println("Starting server...")

	spoofer_type := strings.ToLower(os.Getenv("MINISTRY_SPOOFER_TYPE"))
	if spoofer_type == "" {
		spoofer_type = "mock"
	}

	var spoofer spoof.Spoofer
	switch spoofer_type {
	case "mock":
		spoofer = &spoof.MockSpoofer{}
	case "openai":
		apiKey := os.Getenv("MINISTRY_OPENAI_API_KEY")
		if apiKey == "" {
			panic("OPENAI_API_KEY is not set")
		}
		client := openai.NewClient(apiKey)
		promptProvider := getPromptProvider()
		spoofer = spoof.NewOpenAI(client, promptProvider)
	default:
		log.Fatalf("Unknown spoofer type: %s", spoofer_type)
	}

	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			switch r.URL.Path {
			case "/health":
				w.WriteHeader(http.StatusOK)
				return
			}
		case http.MethodPost:
			switch r.URL.Path {
			case "/spoof":
				request := &SpoofRequest{}
				if err := json.NewDecoder(r.Body).Decode(request); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				slog.Debug("Received spoof request", "request", request)

				timeStart := time.Now()
				spoofed, err := spoofer.Spoof(r.Context(), request.Content, request.Rating)
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
				return
			}
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
}

type SpoofRequest struct {
	Content string `json:"content"`
	Rating  string `json:"rating"`
}

type SpoofResponse struct {
	Content     string  `json:"content"`
	TimeToSpoof float64 `json:"time_to_spoof"`
}
