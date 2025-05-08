package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"main/models"
	"main/sessionmiddleware"
)

type LogRecord struct {
	LogID      int    `json:"log_id"`
	UserID     int    `json:"user_id"`
	FileID     int    `json:"file_id"`
	Filename   string `json:"filename"`
	ActionType string `json:"action_type"`
	Timestamp  string `json:"timestamp"`
}

func ListAllLogs(w http.ResponseWriter, r *http.Request) {
	isAdmin, ok := r.Context().Value(sessionmiddleware.ContextIsAdmin).(bool)
	if !ok || !isAdmin {
		http.Error(w, `{"error": "Access denied"}`, http.StatusForbidden)
		return
	}

	// Pagination parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	rows, err := models.DB.Query(`
        SELECT l.id, l.user_id, l.file_id, f.filename, l.action_type, l.timestamp
        FROM logs l 
        JOIN files f ON l.file_id = f.id 
        ORDER BY l.timestamp DESC 
        LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		http.Error(w, `{"error": "Failed to query logs"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []LogRecord
	for rows.Next() {
		var record LogRecord
		if err := rows.Scan(&record.LogID, &record.UserID, &record.FileID, &record.Filename, &record.ActionType, &record.Timestamp); err != nil {
			http.Error(w, `{"error": "Failed to scan row"}`, http.StatusInternalServerError)
			return
		}
		logs = append(logs, record)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"page":      page,
		"page_size": pageSize,
		"logs":      logs,
	})
}
