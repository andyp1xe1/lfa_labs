package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	videoPath = "./video.mp4"
	port      = 8080
)

//go:embed static/*
var staticFiles embed.FS

// VideoServer handles serving a single video file that may change frequently
type VideoServer struct {
	filePath string
	changes  chan string // Just send the ETag instead of a struct
}

// NewVideoServer creates a new server for the specified video file
func NewVideoServer(filePath string) *VideoServer {
	vs := &VideoServer{
		filePath: filePath,
		changes:  make(chan string, 10), // Buffer for change notifications
	}

	// Start file watcher
	go vs.watchFile()

	return vs
}

// watchFile periodically checks for file changes
func (vs *VideoServer) watchFile() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var lastModTime time.Time

	for range ticker.C {
		fileInfo, err := os.Stat(vs.filePath)
		if err != nil {
			continue // File doesn't exist yet, just continue monitoring
		}

		// Detect if file was modified
		if fileInfo.ModTime() != lastModTime {
			// Create a simple version tag based on modification time
			etag := fmt.Sprintf("v%d", fileInfo.ModTime().UnixNano())

			// Notify of change
			select {
			case vs.changes <- etag:
				// Successfully sent
			default:
				// Channel full, skip
			}

			lastModTime = fileInfo.ModTime()
		}
	}
}

// ServeVideo handles regular HTTP video requests
func (vs *VideoServer) ServeVideo(w http.ResponseWriter, r *http.Request) {
	// Check if file exists before attempting to serve
	fileInfo, err := os.Stat(vs.filePath)
	if err != nil {
		http.Error(w, "Video file not found", http.StatusNotFound)
		return
	}

	// Open the file for serving
	file, err := os.Open(vs.filePath)
	if err != nil {
		http.Error(w, "Could not open video file", http.StatusInternalServerError)
		log.Printf("Error opening video file: %v", err)
		return
	}
	defer file.Close()

	// Set appropriate headers
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Cache-Control", "no-cache")

	// Use file modification time for ETag instead of stored version
	etag := fmt.Sprintf("v%d", fileInfo.ModTime().UnixNano())
	w.Header().Set("ETag", etag)

	// Check if client has a cached version via If-None-Match header
	if clientEtag := r.Header.Get("If-None-Match"); clientEtag != "" && clientEtag == etag {
		// Client has the latest version
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Handle range requests properly with ServeContent
	// This handles large files efficiently with proper byte range support
	http.ServeContent(w, r, filepath.Base(vs.filePath), fileInfo.ModTime(), file)
}

// ServeVideoEvents provides Server-Sent Events for video changes
func (vs *VideoServer) ServeVideoEvents(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Make sure that the writer supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send current file status immediately
	fileInfo, _ := os.Stat(vs.filePath)
	if fileInfo != nil {
		etag := fmt.Sprintf("v%d", fileInfo.ModTime().UnixNano())
		fmt.Fprintf(w, "event: version\ndata: %s\n\n", etag)
		flusher.Flush()
	}

	// Listen for connection close
	notify := r.Context().Done()

	// Watch for changes and send them to this client
	for {
		select {
		case <-notify:
			// Client disconnected
			return
		case etag := <-vs.changes:
			// Send version update to client
			fmt.Fprintf(w, "event: version\ndata: %s\n\n", etag)
			flusher.Flush()
		}
	}
}

func main() {
	server := NewVideoServer(videoPath)

	// Handle the static files
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			content, err := staticFiles.ReadFile("static/index.html")
			if err != nil {
				http.Error(w, "Could not read index.html", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write(content)
			return
		}

		// Serve other static files if needed
		http.FileServer(http.FS(staticFiles)).ServeHTTP(w, r)
	})

	http.HandleFunc("/video", server.ServeVideo)
	http.HandleFunc("/events", server.ServeVideoEvents)

	fmt.Printf("Server started at http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
