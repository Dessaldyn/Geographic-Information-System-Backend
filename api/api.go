package handler

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Global Variable agar bisa diakses main.go
var (
	App        *gin.Engine
	collection *mongo.Collection
)

// --- MODELS ---
type GeoJSON struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}

type Lokasi struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Nama      string             `json:"nama" bson:"nama"`
	Kategori  string             `json:"kategori" bson:"kategori"`
	Deskripsi string             `json:"deskripsi" bson:"deskripsi"`
	Koordinat GeoJSON            `json:"koordinat" bson:"koordinat"`
}

// --- FUNGSI UTAMA (Dipanggil oleh main.go) ---
func Handler(w http.ResponseWriter, r *http.Request) {
	if App == nil {
		Init()
	}
	App.ServeHTTP(w, r)
}
func Init() {
	// Coba load .env (hanya efek di lokal)
	_ = godotenv.Load()
	
	ConnectDB()
	SetupRouter()
}

func ConnectDB() {
	if collection != nil {
		return
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("❌ FATAL ERROR: MONGO_URI tidak ditemukan!")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Println("❌ Gagal buat client Mongo:", err)
		return
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Println("❌ Gagal ping Mongo:", err)
		return
	}

	log.Println("✅ Terhubung ke MongoDB Atlas")
	collection = client.Database("ujianSIG").Collection("lokasis")
}

func SetupRouter() {
	App = gin.New()
	App.Use(gin.Recovery())

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	App.Use(cors.New(config))

	App.GET("/api/lokasi", getLokasi)
	App.POST("/api/lokasi", createLokasi)
	App.PUT("/api/lokasi", updateLokasi)
	App.DELETE("/api/lokasi", deleteLokasi)
}

// --- CONTROLLERS ---

func getLokasi(c *gin.Context) {
	if collection == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database belum terkoneksi"})
		return
	}

	idParam := c.Query("id")
	if idParam != "" {
		objID, err := primitive.ObjectIDFromHex(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
			return
		}
		var lokasi Lokasi
		err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&lokasi)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Data tidak ditemukan"})
			return
		}
		c.JSON(http.StatusOK, lokasi)
	} else {
		cursor, err := collection.Find(context.Background(), bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var lokasis []Lokasi
		if err = cursor.All(context.Background(), &lokasis); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if lokasis == nil {
			lokasis = []Lokasi{}
		}
		c.JSON(http.StatusOK, lokasis)
	}
}

func createLokasi(c *gin.Context) {
	var lokasi Lokasi
	if err := c.ShouldBindJSON(&lokasi); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	lokasi.Koordinat.Type = "Point"
	lokasi.ID = primitive.NewObjectID()

	_, err := collection.InsertOne(context.Background(), lokasi)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, lokasi)
}

func updateLokasi(c *gin.Context) {
	idParam := c.Query("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalid"})
		return
	}
	var updateData Lokasi
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	update := bson.M{
		"$set": bson.M{
			"nama":      updateData.Nama,
			"kategori":  updateData.Kategori,
			"deskripsi": updateData.Deskripsi,
			"koordinat": updateData.Koordinat,
		},
	}
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": objID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updateData.ID = objID
	c.JSON(http.StatusOK, updateData)
}

func deleteLokasi(c *gin.Context) {
	idParam := c.Query("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalid"})
		return
	}
	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Berhasil dihapus"})
}