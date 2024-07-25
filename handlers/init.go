package handlers

import (
  "go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection

func InitHandlers(collection *mongo.Collection) {
  userCollection = collection
}
