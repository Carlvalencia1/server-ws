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

// Mapas concurrentes para clientes WebSocket
var (
	patientClients   sync.Map 
	expedienteClients sync.Map 
)

// Configuración del WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Función para enviar mensajes a todos los clientes conectados
func broadCast(clients *sync.Map, message []byte) {
	clients.Range(func(key, value interface{}) bool {
		client, ok := key.(*websocket.Conn)
		if !ok {
			return true
		}
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Error al enviar mensaje: %v", err)
			client.Close()
			clients.Delete(client)
		}
		return true
	})
}

// Manejo de WebSocket para expedientes médicos
func sendMessageExpediente(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("Error al conectar WebSocket: %v", err)
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
			log.Printf("Error al leer mensaje: %v", err)
			return
		}
		broadCast(&expedienteClients, message)
	}
}

// Manejo de WebSocket para pacientes
func sendMessagePatients(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("Error al conectar WebSocket: %v", err)
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
			log.Printf("Error al leer mensaje: %v", err)
			return
		}
		log.Printf("Mensaje recibido: %s", message)
		broadCast(&patientClients, message)
	}
}

func main() {
	engine := gin.Default()
	engine.GET("/expediente", sendMessageExpediente)
	engine.GET("/pacientes", sendMessagePatients)
	engine.Run(":8081")
}
