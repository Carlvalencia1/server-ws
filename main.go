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
