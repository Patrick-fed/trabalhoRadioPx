package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/radiopx/backend/internal/auth"
	"github.com/radiopx/backend/internal/channel"
	"github.com/radiopx/backend/internal/location"
	"github.com/radiopx/backend/internal/middleware"
	"github.com/radiopx/backend/internal/user"
	"github.com/radiopx/backend/internal/voice"
	ws "github.com/radiopx/backend/pkg/websocket"
	"github.com/rs/cors"
)

type authUserServiceAdapter struct {
	svc *user.UserService
}

func (a *authUserServiceAdapter) Authenticate(email, password string) (*auth.User, error) {
	u, err := a.svc.Authenticate(email, password)
	if err != nil {
		return nil, err
	}
	return &auth.User{
		ID:       u.ID,
		Email:    u.Email,
		Name:     u.Name,
		Provider: u.Provider,
	}, nil
}

func (a *authUserServiceAdapter) GetUser(userID string) (*auth.User, error) {
	u, err := a.svc.GetUser(userID)
	if err != nil {
		return nil, err
	}
	return &auth.User{
		ID:       u.ID,
		Email:    u.Email,
		Name:     u.Name,
		Provider: u.Provider,
	}, nil
}

func (a *authUserServiceAdapter) CreateOAuthUser(email, name, provider, providerID string) (*auth.User, error) {
	u, err := a.svc.CreateOAuthUser(email, name, provider, providerID)
	if err != nil {
		return nil, err
	}
	return &auth.User{
		ID:       u.ID,
		Email:    u.Email,
		Name:     u.Name,
		Provider: u.Provider,
	}, nil
}

var hub *ws.Hub

func main() {
	loadEnvFile()

	hub = ws.NewHub()
	go hub.Run()

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-in-production"
	}

	userService := user.NewUserService()
	locationService := location.NewLocationService()
	channelService := channel.NewChannelService(locationService)
	authAdapter := &authUserServiceAdapter{svc: userService}
	authService := auth.NewAuthService(secret, authAdapter)
	voiceService := voice.NewVoiceService()

	userHandler := user.NewUserHandler(userService)
	channelHandler := channel.NewChannelHandler(channelService)
	voiceHandler := voice.NewVoiceHandler(voiceService, hub)

	setupRoutes(userHandler, channelHandler, voiceHandler, authService, userService)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(http.DefaultServeMux)
	addr := fmt.Sprintf(":%s", port)
	log.Printf("RadioPX Backend starting on port %s", port)
	log.Printf("WebSocket endpoint: ws://localhost:%s/ws/audio/{channel_id}", port)
	log.Printf("Health check: http://localhost:%s/health", port)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal("Server error:", err)
	}
}

func setupRoutes(
	userHandler *user.UserHandler,
	channelHandler *channel.ChannelHandler,
	voiceHandler *voice.VoiceHandler,
	authService *auth.AuthService,
	userService *user.UserService,
) {
	http.HandleFunc("/health", healthHandler)

	http.HandleFunc("/api/v1/auth/register", func(w http.ResponseWriter, r *http.Request) {
		handleRegister(w, r, userService, authService)
	})
	http.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		handleLogin(w, r, authService)
	})
	http.HandleFunc("/api/v1/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		handleRefresh(w, r, authService)
	})

	protected := http.NewServeMux()
	protected.HandleFunc("/api/v1/channels", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			channelHandler.HandleGetUserChannels(w, r)
		} else if r.Method == http.MethodPost {
			channelHandler.HandleCreateChannel(w, r)
		} else {
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	protected.HandleFunc("/api/v1/channels/join", func(w http.ResponseWriter, r *http.Request) {
		channelHandler.HandleJoinChannel(w, r)
	})
	protected.HandleFunc("/api/v1/channels/leave", func(w http.ResponseWriter, r *http.Request) {
		channelHandler.HandleLeaveChannel(w, r)
	})
	protected.HandleFunc("/api/v1/channels/nearby", func(w http.ResponseWriter, r *http.Request) {
		channelHandler.HandleGetNearbyChannels(w, r)
	})
	protected.HandleFunc("/api/v1/channels/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/users"):
			channelHandler.HandleGetChannelUsers(w, r)
		case strings.HasSuffix(path, "/join"):
			channelHandler.HandleJoinChannel(w, r)
		case strings.HasSuffix(path, "/leave"):
			channelHandler.HandleLeaveChannel(w, r)
		default:
			channelHandler.HandleGetChannel(w, r)
		}
	})

	protected.HandleFunc("/api/v1/user/profile", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			userHandler.HandleGetUserProfile(w, r, userID)
		case http.MethodPut:
			userHandler.HandleUpdateUserProfile(w, r, userID)
		case http.MethodDelete:
			userHandler.HandleDeleteUser(w, r)
		default:
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	protected.HandleFunc("/api/v1/user/change-password", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
			return
		}
		userHandler.HandleChangePasswordByUserID(w, r, userID)
	})

	protectedWithAuth := middleware.AuthMiddleware(protected)

	injector := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token != "" {
			claims, err := middleware.ValidateToken(token)
			if err == nil && claims.UserID != "" {
				r.Header.Set("X-User-ID", claims.UserID)
			}
		}
		protectedWithAuth.ServeHTTP(w, r)
	})

	http.Handle("/api/v1/channels", injector)
	http.Handle("/api/v1/channels/join", injector)
	http.Handle("/api/v1/channels/leave", injector)
	http.Handle("/api/v1/channels/nearby", injector)
	http.Handle("/api/v1/channels/", injector)
	http.Handle("/api/v1/user/profile", injector)
	http.Handle("/api/v1/user/change-password", injector)

	http.HandleFunc("/api/v1/location/update", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"Location updated"}`)
	})

	http.HandleFunc("/ws/audio/", func(w http.ResponseWriter, r *http.Request) {
		voiceHandler.HandleWebSocket(w, r)
	})
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

func handleRegister(w http.ResponseWriter, r *http.Request, userService *user.UserService, authService *auth.AuthService) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	u, err := userService.CreateUser(&user.CreateUserRequest{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		status := http.StatusConflict
		if err == user.ErrEmailRequired || err == user.ErrPasswordRequired {
			status = http.StatusBadRequest
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, status)
		return
	}

	tokenPair, err := authService.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, `{"error":"Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":         tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"user":          u,
	})
}

func handleLogin(w http.ResponseWriter, r *http.Request, authService *auth.AuthService) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	tokenPair, err := authService.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, `{"error":"Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	claims, _ := authService.ValidateToken(tokenPair.AccessToken)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":         tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"user": map[string]interface{}{
			"id":    claims.UserID,
			"email": claims.Email,
			"name":  claims.Name,
		},
	})
}

func handleRefresh(w http.ResponseWriter, r *http.Request, authService *auth.AuthService) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	tokenPair, err := authService.RefreshToken(req.RefreshToken)
	if err != nil {
		http.Error(w, `{"error":"Invalid refresh token"}`, http.StatusUnauthorized)
		return
	}

	claims, _ := authService.ValidateToken(tokenPair.AccessToken)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":         tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"user": map[string]interface{}{
			"id":    claims.UserID,
			"email": claims.Email,
			"name":  claims.Name,
		},
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","service":"radiopx-backend","connections":%d}`, hub.GetClientCount())
}

func loadEnvFile() {
	candidates := []string{".env", "../.env"}
	dir, err := os.Getwd()
	if err == nil {
		candidates = append(candidates,
			filepath.Join(dir, ".env"),
			filepath.Join(dir, "..", ".env"),
		)
	}

	seen := make(map[string]bool)
	for _, path := range candidates {
		abs, err := filepath.Abs(path)
		if err != nil || seen[abs] {
			continue
		}
		seen[abs] = true

		f, err := os.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(strings.Trim(parts[1], `"'`))
			if key == "" {
				continue
			}
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
		f.Close()
	}
}
