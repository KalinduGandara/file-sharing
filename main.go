package main

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/*
var templatesFS embed.FS

type ServerConfig struct {
	Port        string
	SourceDir   string
	IPAddresses []string
}

func main() {
	// Default configuration
	config := ServerConfig{
		Port:      "8080",
		SourceDir: "./templates",
	}

	// Get local IP addresses
	config.IPAddresses = getLocalIPs()

	// Parse templates
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		log.Fatal("Error parsing templates:", err)
	}

	// File server handler for the source directory
	fileServer := http.FileServer(http.Dir(config.SourceDir))

	// HTTP handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			if r.Method == http.MethodPost {
				// Handle form submission for port and directory
				err := r.ParseForm()
				if err != nil {
					http.Error(w, "Error parsing form", http.StatusBadRequest)
					return
				}

				newPort := r.FormValue("port")
				newDir := r.FormValue("directory")

				if newPort != "" {
					config.Port = newPort
				}
				if newDir != "" {
					// Validate directory
					if _, err := os.Stat(newDir); !os.IsNotExist(err) {
						config.SourceDir = newDir
						fileServer = http.FileServer(http.Dir(config.SourceDir))
					} else {
						http.Error(w, "Directory does not exist", http.StatusBadRequest)
						return
					}
				}
			}

			// Render the main page
			err := tmpl.ExecuteTemplate(w, "index.html", config)
			if err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		// Serve files
		fileServer.ServeHTTP(w, r)
	})

	// Handle file uploads
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form with 32MB max memory
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error retrieving file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Sanitize filename to prevent path traversal
		filename := sanitizeFilename(handler.Filename)
		dstPath := filepath.Join(config.SourceDir, filename)

		// Create destination file
		dst, err := os.Create(dstPath)
		if err != nil {
			http.Error(w, "Error creating file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Copy file content
		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}

		// Redirect to main page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// Start server
	addr := ":" + config.Port
	fmt.Printf("Server starting on http://localhost%s\n", addr)
	for _, ip := range config.IPAddresses {
		fmt.Printf("Accessible at http://%s%s\n", ip, addr)
	}
	log.Fatal(http.ListenAndServe(addr, nil))
}

// getLocalIPs retrieves all non-loopback IP addresses of the host
func getLocalIPs() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Println("Error getting IP addresses:", err)
		return []string{"127.0.0.1"}
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	if len(ips) == 0 {
		return []string{"127.0.0.1"}
	}
	return ips
}

// sanitizeFilename removes potentially dangerous characters from filenames
func sanitizeFilename(filename string) string {
	// Replace path separators and other dangerous characters
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, "..", "_")
	return filename
}
