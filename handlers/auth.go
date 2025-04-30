package handlers

import (
	"encoding/json"
	"main/models"

	// "main/utils"
	"net/http"
	"time"

	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"
)

const sessionCookieName = "session_id"

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Signup(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error creating account", http.StatusInternalServerError)
		return
	}

	_, err = models.DB.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", creds.Username, string(hashedPassword))
	if err != nil {
		http.Error(w, "Username already exists", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "User created"}`))
}

func Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var user models.User
	row := models.DB.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", creds.Username)
	err = row.Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = models.DB.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, user.ID, expiresAt)
	if err != nil {
		http.Error(w, "Could not create session", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   false, // Set to true in production (HTTPS)
		Path:     "/",
	})

	// token, err := utils.GenerateJWT(user.ID)
	// if err != nil {
	// 	http.Error(w, "Could not generate token", http.StatusInternalServerError)
	// 	return
	// }

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "logged in"}`))
	// w.Write([]byte(`{"token": "` + token + `"}`))
}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		models.DB.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	})

	w.Write([]byte(`{"message": "logged out"}`))
}
