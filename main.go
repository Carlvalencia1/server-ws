package main


import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
	clients = make(map[*websocket.Conn]bool)
	mu      sync.Mutex
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func broadCast(message []byte) {
	mu.Lock()
	defer mu.Unlock() // Corregido: Unlock en lugar de Lock

	for client := range clients {
		errSenMessage := client.WriteMessage(websocket.TextMessage, message)
		if errSenMessage != nil {
			log.Printf("error to send message: %v", errSenMessage)
			client.Close()
			delete(clients, client)
		}
	}
}

func sendMessageExpediente(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("error to connect: %v", err)
		return
	}

	// Registrar cliente
	mu.Lock()
	clients[conn] = true
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(clients, conn)
		mu.Unlock()
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error reading message: %v", err)
			return
		}
		broadCast(message)
	}
}

func sendMessagePatients(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("error to connect: %v", err)
		return
	}

	// Registrar cliente
	mu.Lock()
	clients[conn] = true
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(clients, conn)
		mu.Unlock()
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage() // Corregido
		if err != nil {
			log.Printf("error reading message: %v", err)
			return
		}
		log.Printf("message received: %s", message)
		broadCast(message)
	}
}

func main() {
	engine := gin.Default()
	engine.GET("/expediente", sendMessageExpediente)
	engine.GET("/pacientes", sendMessagePatients)
	engine.Run(":8081") // Corregido: puerto como string
}