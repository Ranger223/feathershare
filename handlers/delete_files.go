package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"main/models"
	"main/sessionmiddleware"
	"net/http"
	"os"
)

// Request payload for batch delete
type BatchDeleteRequest struct {
	FileIDs []int `json:"file_ids"`
}

func DeleteFiles(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(sessionmiddleware.ContextUserID).(int)
	isAdmin := r.Context().Value(sessionmiddleware.ContextIsAdmin).(bool)

	var req BatchDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if len(req.FileIDs) == 0 {
		http.Error(w, `{"error": "No file IDs provided"}`, http.StatusBadRequest)
		return
	}

	var deletedFiles []string
	var failedFiles []string

	for _, fileID := range req.FileIDs {
		var filePath, fileName string
		var ownerID int
		err := models.DB.QueryRow(`
            SELECT filepath, filename, user_id 
            FROM files 
            WHERE id = ?`, fileID).
			Scan(&filePath, &fileName, &ownerID)

		if err == sql.ErrNoRows {
			failedFiles = append(failedFiles, fmt.Sprintf("File ID %d (not found)", fileID))
			continue
		} else if err != nil {
			failedFiles = append(failedFiles, fmt.Sprintf("File ID %d (DB error)", fileID))
			continue
		}

		// Ensure the user owns the file or is an admin
		if !isAdmin && ownerID != userID {
			failedFiles = append(failedFiles, fmt.Sprintf("File ID %d (no permission)", fileID))
			continue
		}

		// Delete the file from disk
		if err := os.Remove(filePath); err != nil {
			failedFiles = append(failedFiles, fmt.Sprintf("File ID %d (disk error)", fileID))
			continue
		}

		// Delete the file record from the database
		_, err = models.DB.Exec("DELETE FROM files WHERE id = ?", fileID)
		if err != nil {
			failedFiles = append(failedFiles, fmt.Sprintf("File ID %d (DB delete error)", fileID))
			continue
		}

		// Log the deletion
		go models.LogAction(userID, fileID, fileName, "delete")
		deletedFiles = append(deletedFiles, fileName)
	}

	response := map[string]interface{}{
		"deleted_files": deletedFiles,
		"failed_files":  failedFiles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
