# Go File Sharing Server

A lightweight, web-based file-sharing server written in Go. This program allows you to share files between computers on the same network with a simple UI, similar to Python's `http.server`, but with additional features like file uploading and a file listing interface.

## Features

- **Simple Web Interface**: Configure the port and source directory, view network IP addresses, and manage files.
- **File Listing**: Browse files and directories with details like name, type, size, and last modified time.
- **Bidirectional File Transfer**:
  - **Download**: Access and download files from the source directory via browser links.
  - **Upload**: Upload files to the server using a web form.
- **Network Accessibility**: Displays all non-loopback IP addresses for easy access from other devices.
- **Secure File Handling**: Sanitizes uploaded filenames to prevent path traversal attacks.
- **No External Dependencies**: Built using Go's standard library and embedded templates.

## Requirements

- **Go**: Version 1.16 or later (required for `embed` package support).
- A modern web browser to access the interface.

## Installation

1. **Clone or Download**:
   - Clone this repository or download the source files:
     ```
     git clone <repository-url>
     cd go-file-sharing
     ```

2. **Initialize Go Module**:
   - Run the following command in the project directory to create a Go module:
     ```
     go mod init file-sharing
     ```

3. **Ensure File Structure**:
   - The project requires the following structure:
     ```
     go-file-sharing/
     ├── main.go
     └── templates/
         └── index.html
     ```
   - Ensure `main.go` and `templates/index.html` are in place as provided.

## Usage

1. **Run the Server**:
   - Start the server by running:
     ```
     go run main.go
     ```
   - The server defaults to port `8080` and the current directory as the source.

2. **Access the Interface**:
   - Open a browser and navigate to `http://localhost:8080`.
   - The UI displays:
     - A form to set the port and source directory.
     - A list of IP addresses for network access (e.g., `http://192.168.x.x:8080`).
     - A file upload form.
     - A table listing files and directories in the source directory.

3. **Share Files**:
   - **Download**: Click file names in the list to download them or navigate into directories.
   - **Upload**: Use the upload form to send files to the server.
   - **Network Access**: Use the displayed IP addresses to access the server from other devices on the same network.

4. **Update Settings**:
   - Change the port or source directory via the web form.
   - Note: Changing the port requires manually restarting the server.

## Example

- Start the server:
  ```
  $ go run main.go
  Server starting on http://localhost:8080
  Accessible at http://192.168.1.100:8080
  ```
- Open `http://localhost:8080` to see the file list and upload form.
- On another device, access `http://192.168.1.100:8080` to download or upload files.

## Project Structure

- `main.go`: The main Go program implementing the server logic.
- `templates/index.html`: The HTML template for the web interface, embedded in the binary.

## Notes

- **Port Changes**: Updating the port via the UI requires restarting the server manually to apply the change.
- **Hidden Files**: Files starting with `.` are excluded from the listing for clarity.
- **Security**: Basic security measures are in place (e.g., filename sanitization). For production use, consider adding authentication and HTTPS.
- **Limitations**:
  - No breadcrumb navigation for subdirectories.
  - File sizes are displayed in bytes; formatting (e.g., KB, MB) can be added.
- **Enhancements**: Potential additions include file deletion, renaming, or maximum file size limits.

## Contributing

Feel free to submit issues or pull requests for improvements, such as:
- Adding authentication.
- Enhancing the UI with breadcrumb navigation.
- Supporting file management operations (delete, rename).
- Formatting file sizes in a human-readable way.


---

Built with ❤️ using Go.
