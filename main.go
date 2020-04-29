package main

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EventDate contains the year, month and day of an event
type EventDate struct {
	Year  string `json:"year"`
	Month string `json:"month"`
	Day   string `json:"day"`
}

// Event contains the event informations such as the event name, its location and its date formated
type Event struct {
	ID       string    `json:"id,omitempty" bson:"_id,omitempty"`
	Name     string    `json:"name"`
	Location string    `json:"location"`
	Date     EventDate `json:"date"`
}

func createConnection() *mongo.Database {
	connectionOptions := options.Client()
	connectionOptions.ApplyURI("mongodb://localhost:27017")

	client, err := mongo.Connect(context.TODO(), connectionOptions)
	if err != nil {
		defer client.Disconnect(context.TODO())
		panic(err)
	} else {
		fmt.Println("> Database is connected")
	}

	if err := client.Ping(context.TODO(), nil); err != nil {
		defer client.Disconnect(context.TODO())
		panic(err)
	} else {
		fmt.Println("> MongoDB client is working as desired")
	}

	return client.Database("GoEvent")
}

func main() {
	database := createConnection()

	appSettings := &fiber.Settings{
		CaseSensitive: true,
		StrictRouting: true,
	}
	app := fiber.New(appSettings)

	app.Get("/events", func(c *fiber.Ctx) {
		filter := bson.D{{}}
		cursor, err := database.Collection("events").Find(context.TODO(), filter)
		if err != nil {
			c.Status(500).JSON(fiber.Map{
				"error": err,
			})
			return
		}

		events := make([]Event, 0)
		if err := cursor.All(context.TODO(), &events); err != nil {
			c.Status(500).JSON(fiber.Map{
				"error": err,
			})
			return
		}

		if err := c.Status(200).JSON(events); err != nil {
			c.Status(500).JSON(fiber.Map{
				"error": err,
			})
			return
		}
	})

	app.Post("/events", func(c *fiber.Ctx) {
		event := &Event{}

		if err := c.BodyParser(event); err != nil {
			c.Status(400).JSON(fiber.Map{
				"error": err,
			})
			return
		}

		insertionResult, err := database.Collection("events").InsertOne(context.TODO(), &event)
		if err != nil {
			c.Status(500).JSON(fiber.Map{
				"error": err,
			})
			return
		}

		filter := bson.D{{Key: "_id", Value: insertionResult.InsertedID}}
		if err := database.Collection("events").FindOne(context.TODO(), filter).Decode(&event); err != nil {
			c.Status(400).JSON(fiber.Map{
				"error": err,
			})
			return
		}

		if err := c.Status(201).JSON(event); err != nil {
			c.Status(400).JSON(fiber.Map{
				"error": err,
			})
			return
		}
	})

	app.Put("/events/:id", func(c *fiber.Ctx) {
		event := &Event{}

		eventID, err := primitive.ObjectIDFromHex(
			c.Params("id"),
		)
		if err != nil {
			c.Status(400).JSON(fiber.Map{
				"error": err,
			})
			return
		}

		if err := c.BodyParser(event); err != nil {
			c.Status(400).JSON(fiber.Map{
				"error": err,
			})
			return
		}

		filter := bson.D{{Key: "_id", Value: eventID}}
		update := bson.D{{"$set", &event}}
		if _, err := database.Collection("events").UpdateOne(context.TODO(), filter, update); err != nil {
			c.Status(500).JSON(fiber.Map{
				"error": err,
			})
			return
		}

		if err := database.Collection("events").FindOne(context.TODO(), filter).Decode(&event); err != nil {
			c.Status(500).JSON(fiber.Map{
				"error": err,
			})
			return
		}

		if err := c.Status(200).JSON(event); err != nil {
			c.Status(400).JSON(fiber.Map{
				"error": err,
			})
			return
		}
	})

	app.Delete("/events/:id", func(c *fiber.Ctx) {
		eventID, err := primitive.ObjectIDFromHex(
			c.Params("id"),
		)
		if err != nil {
			c.Status(400).JSON(fiber.Map{
				"error": err,
			})
			return
		}

		filter := bson.D{{Key: "_id", Value: eventID}}
		database.Collection("events").FindOneAndDelete(context.TODO(), filter)

		c.Status(204)
	})

	app.Listen(3000)
}
