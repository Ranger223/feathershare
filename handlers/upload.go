package handlers

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"main/models"
	"main/sessionmiddleware"

	"github.com/google/uuid"
)

func UploadFile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(sessionmiddleware.ContextUserID).(int)

	// Get boundary from Content-Type header
	contentType := r.Header.Get("Content-Type")
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil || params["boundary"] == "" {
		http.Error(w, `{"error": "Invalid Content-Type"}`, http.StatusBadRequest)
		return
	}

	reader := multipart.NewReader(r.Body, params["boundary"])

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, `{"error": "Error reading part"}`, http.StatusBadRequest)
			return
		}

		// Skip non-file fields
		if part.FileName() == "" {
			continue
		}

		if err := ensureUploadsFolderExists("uploads"); err != nil {
			fmt.Println("Error:", err)
		}

		// Generate safe filename
		timestamp := time.Now().Unix()
		ext := filepath.Ext(part.FileName())
		base := strings.TrimSuffix(part.FileName(), ext)
		safeName := fmt.Sprintf("%d_%s_%s%s", timestamp, base, uuid.New().String(), ext)
		savePath := filepath.Join("uploads", safeName)

		// Create destination file
		dst, err := os.Create(savePath)
		if err != nil {
			http.Error(w, `{"error": "Could not save file"}`, http.StatusInternalServerError)
			return
		}

		// Stream file data directly to disk
		buf := bufio.NewWriter(dst)
		_, err = io.Copy(buf, part)
		dst.Close()
		if err != nil {
			http.Error(w, `{"error": "Failed to write file"}`, http.StatusInternalServerError)
			return
		}

		// Store file metadata in DB
		result, err := models.DB.Exec(`
            INSERT INTO files (user_id, filename, filepath)
            VALUES (?, ?, ?)`, userID, part.FileName(), savePath)
		if err != nil {
			http.Error(w, `{"error": "DB insert failed"}`, http.StatusInternalServerError)
			return
		}

		fileID, _ := result.LastInsertId()

		// Log the upload
		go models.LogAction(userID, int(fileID), part.FileName(), "upload")

		w.Write(fmt.Appendf(nil, `{"message": "uploaded: %s"}`, part.FileName()))
		return // only handling one file per request
	}

	http.Error(w, `{"error": "No filed uploaded"}`, http.StatusBadRequest)
}

func ensureUploadsFolderExists(path string) error {
	// Check if the folder exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// Create the folder (and any necessary parent directories)
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create folder: %w", err)
		}
		fmt.Println("Uploads folder created.")
	} else if err != nil {
		return fmt.Errorf("error checking folder: %w", err)
	} else {
		fmt.Println("Uploads folder already exists.")
	}
	return nil
}
