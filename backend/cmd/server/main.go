package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"

    "github.com/asmit27rai/real-time-taskboard/backend/internal/api"
    "github.com/asmit27rai/real-time-taskboard/backend/internal/db"
    "github.com/asmit27rai/real-time-taskboard/backend/internal/ws"
)

func main() {
    // MongoDB URI
    uri := os.Getenv("MONGO_URI")
    if uri == "" {
        uri = "mongodb://localhost:27017"
    }
    client, err := db.Connect(uri)
    if err != nil {
        log.Fatal("MongoDB connect error:", err)
    }

    // Start WebSocket hub
    hub := ws.NewHub()
    go hub.Run()

    // Start watching MongoDB changes
    go db.WatchTasks(context.Background(), client, hub)

    // Setup HTTP server
    router := gin.Default()
    router.Use(cors.Default())

    handler := api.NewTaskHandler(client.Database("taskdb"), hub)
    router.GET("/tasks", handler.GetTasks)
    router.POST("/tasks", handler.CreateTask)
    router.PUT("/tasks/:id", handler.UpdateTask)
    router.DELETE("/tasks/:id", handler.DeleteTask)
    router.GET("/ws", ws.ServeWs(hub))

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    srv := &http.Server{Addr: ":" + port, Handler: router, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}

    log.Println("Listening on port", port)
    if err := srv.ListenAndServe(); err != nil {
        log.Fatal("Server error:", err)
    }
}