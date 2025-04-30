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
		http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	reader := multipart.NewReader(r.Body, params["boundary"])

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "Error reading part", http.StatusBadRequest)
			return
		}

		// Skip non-file fields
		if part.FileName() == "" {
			continue
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
			http.Error(w, "Could not save file", http.StatusInternalServerError)
			return
		}

		// Stream file data directly to disk
		buf := bufio.NewWriter(dst)
		_, err = io.Copy(buf, part)
		dst.Close()
		if err != nil {
			http.Error(w, "Failed to write file", http.StatusInternalServerError)
			return
		}

		// Store file metadata in DB
		_, err = models.DB.Exec(`
            INSERT INTO files (user_id, filename, filepath)
            VALUES (?, ?, ?)`, userID, part.FileName(), savePath)
		if err != nil {
			http.Error(w, "DB insert failed", http.StatusInternalServerError)
			return
		}

		w.Write([]byte(fmt.Sprintf(`{"message": "uploaded: %s"}`, part.FileName())))
		return // only handling one file per request
	}

	http.Error(w, "No file uploaded", http.StatusBadRequest)
}
