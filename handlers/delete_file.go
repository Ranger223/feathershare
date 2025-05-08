package handlers

import (
	"database/sql"
	"fmt"
	"main/models"
	"main/sessionmiddleware"
	"net/http"
	"os"
	"strconv"
)

func DeleteFile(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, `{"error": "You do not permission to delete this file"}`, http.StatusForbidden)
		return
	}

	// Check if file exists on disk
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, `{"error": "File not found on server"}`, http.StatusNotFound)
		return
	}

	// Delete the file from disk
	if err := os.Remove(filePath); err != nil {
		http.Error(w, `{"error": "Failed to delete file from disk"}`, http.StatusInternalServerError)
		return
	}

	_, err = models.DB.Exec(`DELETE * from files WHERE id = ?`, fileID)
	if err != nil {
		http.Error(w, `{"error": "DB file deletion failed"}`, http.StatusInternalServerError)
		return
	}

	// Log the file deletion
	go models.LogAction(userID, fileID, "delete")

	w.Header().Set("Content-Type", "application/json")
	w.Write(fmt.Appendf(nil, `{"message": "File %s deleted successfully"}`, fileName))

}
