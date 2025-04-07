package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

type Patients struct {
	IDUsuario      int    `json:"id_usuario"`
	Nombre         string `json:"nombre"`
	Apellido       string `json:"apellido"`
	Edad           int    `json:"edad"`
	Genero         string `json:"genero"`
	NumeroContacto string `json:"numero_contacto"`
}

type MedicalCase struct {
	IDExpediente  int       `json:"id_expediente"`
	IDUsuario     int       `json:"id_usuario"`
	Temperatura   float64   `json:"temperatura"`
	Peso          float64   `json:"peso"`
	Estatura      float64   `json:"estatura"`
	RitmoCardiaco int       `json:"ritmo_cardiaco"`
	FechaRegistro time.Time `json:"fecha_registro"`
}

var (
	patientClients    sync.Map
	expedienteClients sync.Map
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func broadCast(clients *sync.Map, message []byte) {
	clients.Range(func(key, value interface{}) bool {
		client, ok := key.(*websocket.Conn)
		if !ok {
			return true
		}
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("❌ Error al enviar mensaje: %v", err)
			client.Close()
			clients.Delete(client)
		}
		return true
	})
}

func sendMessageExpediente(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("❌ Error al conectar WebSocket /cases/: %v", err)
		return
	}

	expedienteClients.Store(conn, true)
	defer func() {
		expedienteClients.Delete(conn)
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("🔌 Desconectado /cases/: %v", err)
			return
		}
		log.Printf("📥 Mensaje recibido por WebSocket /cases/: %s", message)
		broadCast(&expedienteClients, message)
	}
}

func sendMessagePatients(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("❌ Error al conectar WebSocket /patients/: %v", err)
		return
	}

	patientClients.Store(conn, true)
	defer func() {
		patientClients.Delete(conn)
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("🔌 Desconectado /patients/: %v", err)
			return
		}
		log.Printf("📥 Mensaje recibido por WebSocket /patients/: %s", message)
		broadCast(&patientClients, message)
	}
}

func main() {
	// Cargar .env si existe
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No se pudo cargar .env (quizás no es necesario)")
	}

	engine := gin.Default()

	// Rutas WebSocket (GET)
	engine.GET("/patients/", sendMessagePatients)
	engine.GET("/cases/", sendMessageExpediente)

	// Ruta POST para casos (API Consumer puede enviar aquí)
	engine.POST("/cases/", func(ctx *gin.Context) {
		var data MedicalCase
		if err := ctx.BindJSON(&data); err != nil {
			ctx.JSON(400, gin.H{"error": "Datos inválidos"})
			return
		}

		jsonBytes, err := json.Marshal(data)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Error al convertir mensaje"})
			return
		}

		log.Printf("📨 Mensaje recibido por POST /cases/: %s", string(jsonBytes))
		broadCast(&expedienteClients, jsonBytes)

		ctx.JSON(200, gin.H{"status": "mensaje recibido"})
	})

	// Puerto desde .env o por defecto 8081
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("🚀 Servidor WebSocket corriendo en :%s", port)
	if err := engine.Run(":" + port); err != nil {
		log.Fatalf("❌ Error al iniciar servidor: %v", err)
	}
}
