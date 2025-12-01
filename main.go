package main

import (
	"log"
	"net/http"
	"os"

	// Pastikan nama ini sesuai dengan yang ada di file go.mod
	"backend-go/api" 
)

// --- 1. SETUP AWAL (INIT) ---
// init() dijalankan otomatis oleh Go/Vercel saat start
func init() {
	// Panggil fungsi Init dari folder api
	api.Init()
}

// --- 2. ENTRYPOINT VERCEL ---
// Handler Vercel sekarang mendelegasikan tugas ke api.App
func Handler(w http.ResponseWriter, r *http.Request) {
	api.App.ServeHTTP(w, r)
}

// --- 3. ENTRYPOINT LOKAL ---
func main() {
	// Pastikan App sudah siap (fallback untuk jalan manual)
	if api.App == nil {
		api.Init()
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	
	log.Println("🚀 Server Golang berhasil berjalan :" + port)
	api.App.Run(":" + port)
}