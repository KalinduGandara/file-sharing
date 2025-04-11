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
	"sort"
	"strings"
)

//go:embed templates/*
var templatesFS embed.FS

type ServerConfig struct {
	Port        string
	SourceDir   string
	IPAddresses []string
	Files       []FileInfo
}

type FileInfo struct {
	Name    string
	IsDir   bool
	Size    int64
	ModTime string
	RelPath string
}

func main() {
	// Default configuration
	config := ServerConfig{
		Port:      "8080",
		SourceDir: ".",
	}

	// Get local IP addresses
	config.IPAddresses = getLocalIPs()

	// Parse templates
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		log.Fatal("Error parsing templates:", err)
	}

	// HTTP handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get relative path from URL (e.g., /subfolder)
		relPath := strings.TrimPrefix(r.URL.Path, "/")
		if relPath == "" {
			relPath = "."
		}

		// Resolve absolute path
		absPath := filepath.Join(config.SourceDir, relPath)

		// Handle form submission for port and directory
		if r.Method == http.MethodPost && r.URL.Path == "/" {
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
					absPath = config.SourceDir
					relPath = "."
				} else {
					http.Error(w, "Directory does not exist", http.StatusBadRequest)
					return
				}
			}
		}

		// Check if path is a file
		if info, err := os.Stat(absPath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, absPath)
			return
		}

		// List files in the current directory
		config.Files, err = listFiles(config.SourceDir, relPath)
		if err != nil {
			http.Error(w, "Error listing files", http.StatusInternalServerError)
			return
		}

		// Render the main page
		err = tmpl.ExecuteTemplate(w, "index.html", config)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
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
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, "..", "_")
	return filename
}

// listFiles returns a sorted list of files and directories in the given path
func listFiles(rootDir, relPath string) ([]FileInfo, error) {
	absPath := filepath.Join(rootDir, relPath)
	var files []FileInfo

	dir, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	infos, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			continue
		}

		relFilePath := filepath.Join(relPath, info.Name())
		if relPath == "." {
			relFilePath = info.Name()
		}

		file := FileInfo{
			Name:    info.Name(),
			IsDir:   info.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
			RelPath: relFilePath,
		}
		files = append(files, file)
	}

	// Sort files: directories first, then by name
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].Name < files[j].Name
	})

	return files, nil
}
