package models

import "fmt"

// Unified LogAction for uploads and downloads
func LogAction(userID int, fileID int, actionType string) {
	_, err := DB.Exec(`
        INSERT INTO logs (user_id, file_id, action_type)
        VALUES (?, ?, ?)`, userID, fileID, actionType)
	if err != nil {
		fmt.Println("Failed to log action:", err)
	}
}
