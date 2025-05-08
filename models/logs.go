package models

import "fmt"

// Unified LogAction for uploads and downloads
func LogAction(userID int, fileID int, fileName string, actionType string) {
	_, err := DB.Exec(`
        INSERT INTO logs (user_id, file_id, file_name, action_type)
        VALUES (?, ?, ?, ?)`, userID, fileID, fileName, actionType)
	if err != nil {
		fmt.Println("Failed to log action:", err)
	}
}
