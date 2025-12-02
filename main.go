package main

import (
	"fmt"
	"net/http"
)

func main() {
	cfg, err := LoadConfig("config.json")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	InitPasswordValidator(cfg)

	db, err := InitDB(cfg)
	if err != nil {
		fmt.Printf("Database initialization failed: %v\n", err)
		return
	}
	defer db.Close()

	http.HandleFunc("/api/login", HandleLogin(cfg, db))
	http.HandleFunc("/api/users", AuthMiddlewareAdministration(cfg, db, HandleGetUsers(db)))
	http.HandleFunc("/api/register", HandleRegister(db))
	http.HandleFunc("/api/add-photo", AuthMiddleware(cfg, HandleAddPhoto(cfg, db)))
	http.HandleFunc("/api/manage-ban", AuthMiddlewareAdministration(cfg, db, HandleManageBanStatus(db)))
	http.HandleFunc("/api/toggle-public", AuthMiddleware(cfg, HandleTogglePhotoPublic(cfg, db)))
	http.HandleFunc("/api/public-gallery", HandlePublicGallery(db))
	http.HandleFunc("/api/photos/", HandleGetPhotos(cfg, db))
	http.HandleFunc("/api/delete-photo/", AuthMiddleware(cfg, HandleDeletePhoto(cfg, db)))

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Server started on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
