package handlers

import (
	"encoding/json"
	"main/models"
	"main/sessionmiddleware"
	"net/http"
)

func ListFiles(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(sessionmiddleware.ContextUserID).(int)

	rows, err := models.DB.Query("SELECT id, filename, uploaded_at FROM files WHERE user_id = ? ORDER BY uploaded_at DESC", userID)

	if err != nil {
		http.Error(w, "Failed to query files", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var f models.File
		if err := rows.Scan(&f.ID, &f.Filename, &f.UploadedAt); err != nil {
			http.Error(w, "Failed to scan row", http.StatusInternalServerError)
			return
		}
		files = append(files, f)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}
