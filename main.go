package main

import (
  "context"
  "log"
  "time"

  "github.com/gofiber/fiber/v2"
  "github.com/gofiber/fiber/v2/middleware/logger"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"

  "autogpt-api/handlers"
)

var userCollection *mongo.Collection

func main() {
  // MongoDB config
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
  client, err := mongo.Connect(ctx, clientOptions)
  if err != nil {
    log.Fatal(err)
  }
  defer client.Disconnect(ctx)

  userCollection = client.Database("autogpt").Collection("users")

  // Init new fiber custom app
  app := fiber.New(fiber.Config{
    Prefork: true,
    CaseSensitive: true,
    ServerHeader: "Fiber",
    AppName: "autoGPT API v1.1.0",
  })
  
  // Middleware to log requests
  app.Use(logger.New())

  // Pass userCollection to handlers
  handlers.InitHandlers(userCollection)

  // Routes
  app.Get("/", func(c *fiber.Ctx) error {
    return c.SendString("Hello World")
  })
  
  app.Get("/brain", handlers.OpenAIBrain)
  app.Post("/openai", handlers.OpenAIHandler)
  app.Post("/google", handlers.GoogleHandler)
  app.Post("/anthropic", handlers.AnthropicHandler)
  app.Get("/whisper", handlers.WhisperHandler)

  // Init server
  port := "8080"
  err = app.Listen(":" + port)

  if err != nil {
    log.Fatal("Can't connect to port", port, ": ", err)
  }
}
