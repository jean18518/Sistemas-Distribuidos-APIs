package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// =============================================
// 1. MODELO - TAREAS
// =============================================

type Task struct {
	ID          int       `json:"id"`
	Titulo      string    `json:"titulo"` // requerido
	Descripcion string    `json:"descripcion"`
	Estado      string    `json:"estado"`    // pendiente, en_progreso, completada
	Prioridad   string    `json:"prioridad"` // baja, media, alta
	CreatedAt   time.Time `json:"created_at"`
}

// =============================================
// 2. BASE DE DATOS EN MEMORIA
// =============================================

var tasks []Task
var ultimoID int = 0

// =============================================
// 3. FUNCIONES AUXILIARES
// =============================================

// Buscar tarea por ID
func buscarPorID(id int) (*Task, int) {
	for i, task := range tasks {
		if task.ID == id {
			return &tasks[i], i
		}
	}
	return nil, -1
}

// Responder con JSON
func responderJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Responder con Error
func responderError(w http.ResponseWriter, mensaje string, status int) {
	responderJSON(w, map[string]string{"error": mensaje}, status)
}

// =============================================
// 4. HANDLERS (CRUD)
// =============================================

// GET /health
func healthHandler(w http.ResponseWriter, r *http.Request) {
	responderJSON(w, map[string]string{
		"status":  "ok",
		"service": "api-tareas",
	}, 200)
}

// GET /api/v1/tareas - Listar todas
func listarTasks(w http.ResponseWriter, r *http.Request) {
	responderJSON(w, tasks, 200)
}

// GET /api/v1/tareas/{id} - Obtener una
func obtenerTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		responderError(w, "ID inválido", 400)
		return
	}

	task, _ := buscarPorID(id)
	if task == nil {
		responderError(w, "Tarea no encontrada", 404)
		return
	}

	responderJSON(w, task, 200)
}

// POST /api/v1/tareas - Crear
func crearTask(w http.ResponseWriter, r *http.Request) {
	var nueva Task

	err := json.NewDecoder(r.Body).Decode(&nueva)
	if err != nil {
		responderError(w, "JSON inválido", 400)
		return
	}

	// VALIDACIÓN: Campo Titulo requerido
	if nueva.Titulo == "" {
		responderError(w, "El campo 'titulo' es requerido", 400)
		return
	}

	// Asignar ID y fecha
	ultimoID++
	nueva.ID = ultimoID
	nueva.CreatedAt = time.Now()

	// Si el estado o prioridad vienen vacíos, poner defaults
	if nueva.Estado == "" {
		nueva.Estado = "pendiente"
	}
	if nueva.Prioridad == "" {
		nueva.Prioridad = "media"
	}

	// Guardar
	tasks = append(tasks, nueva)

	responderJSON(w, nueva, 201)
}

// PUT /api/v1/tareas/{id} - Actualizar
func actualizarTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		responderError(w, "ID inválido", 400)
		return
	}

	task, index := buscarPorID(id)
	if task == nil {
		responderError(w, "Tarea no encontrada", 404)
		return
	}

	var actualizado Task
	err = json.NewDecoder(r.Body).Decode(&actualizado)
	if err != nil {
		responderError(w, "JSON inválido", 400)
		return
	}

	// Validación al actualizar
	if actualizado.Titulo == "" {
		responderError(w, "El campo 'titulo' no puede estar vacío", 400)
		return
	}

	// Mantener ID y fecha original
	actualizado.ID = id
	actualizado.CreatedAt = task.CreatedAt

	// Actualizar
	tasks[index] = actualizado

	responderJSON(w, actualizado, 200)
}

// DELETE /api/v1/tareas/{id} - Eliminar
func eliminarTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		responderError(w, "ID inválido", 400)
		return
	}

	_, index := buscarPorID(id)
	if index == -1 {
		responderError(w, "Tarea no encontrada", 404)
		return
	}

	// Eliminar del slice
	tasks = append(tasks[:index], tasks[index+1:]...)

	w.WriteHeader(204) // No Content
}

// =============================================
// 5. MIDDLEWARE - CORS
// =============================================

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Permitir cualquier origen
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Permitir métodos específicos
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// Permitir headers específicos
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Manejar preflight request (OPTIONS)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// =============================================
// 6. MAIN - CONFIGURAR RUTAS
// =============================================

func main() {
	router := mux.NewRouter()

	// Rutas TAREAS
	router.HandleFunc("/health", healthHandler).Methods("GET")
	router.HandleFunc("/api/v1/tareas", listarTasks).Methods("GET")
	router.HandleFunc("/api/v1/tareas/{id}", obtenerTask).Methods("GET")
	router.HandleFunc("/api/v1/tareas", crearTask).Methods("POST")
	router.HandleFunc("/api/v1/tareas/{id}", actualizarTask).Methods("PUT")
	router.HandleFunc("/api/v1/tareas/{id}", eliminarTask).Methods("DELETE")

	// PUERTO ASIGNADO: 8001
	puerto := ":8001"

	fmt.Println("================================")
	fmt.Println("🚀 API TAREAS (con CORS) iniciada en puerto", puerto)
	fmt.Println("================================")
	fmt.Println("GET    /health")
	fmt.Println("GET    /api/v1/tareas")
	fmt.Println("POST   /api/v1/tareas")
	fmt.Println("PUT    /api/v1/tareas/{id}")
	fmt.Println("DELETE /api/v1/tareas/{id}")
	fmt.Println("================================")

	// APLICAMOS EL MIDDLEWARE CORS AQUÍ
	http.ListenAndServe(puerto, enableCORS(router))
}
