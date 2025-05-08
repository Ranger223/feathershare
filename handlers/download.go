package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"main/models"
	"main/sessionmiddleware"
)

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(sessionmiddleware.ContextUserID).(int)
	isAdmin := r.Context().Value(sessionmiddleware.ContextIsAdmin).(bool)
	fileIDStr := r.URL.Query().Get("id")

	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid file ID"}`, http.StatusBadRequest)
		return
	}

	var filePath, fileName string
	var ownerID int
	err = models.DB.QueryRow(`SELECT filepath, filename, user_id FROM files WHERE id = ?`, fileID).Scan(&filePath, &fileName, &ownerID)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "File not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"error": "Failed to query file"}`, http.StatusInternalServerError)
		return
	}

	// Verify ownership
	if !isAdmin && ownerID != userID {
		http.Error(w, `{"error": "You do not have access to this file"}`, http.StatusForbidden)
		return
	}

	// Check if file exists on disk
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, `{"error": "File not found on server"}`, http.StatusNotFound)
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Transfer-Encoding", "binary")

	// Serve the file directly
	http.ServeFile(w, r, filePath)

	// Log download (optional)
	go models.LogAction(userID, fileID, fileName, "download")

}
