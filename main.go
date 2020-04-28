package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoInstance contains MongoDB client and database
type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

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

var mg MongoInstance

// Connect to MongoDB
func Connect() error {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://customuser:customuser@heaven-uc9c7.mongodb.net"))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	db := client.Database("GoEvent")

	if err != nil {
		return err
	}

	mg = MongoInstance{
		Client: client,
		Db:     db,
	}

	return nil
}

func main() {
	if err := Connect(); err != nil {
		log.Fatal(err)
	}

	app := fiber.New()
	app.Settings.CaseSensitive = true
	app.Settings.StrictRouting = true

	app.Get("/events", func(c *fiber.Ctx) {
		cursor, err := mg.Db.Collection("events").Find(context.TODO(), bson.D{{}})
		if err != nil {
			c.Status(500).JSON(err)
			return
		}

		var events []Event = make([]Event, 0)
		if err := cursor.All(context.TODO(), &events); err != nil {
			c.Status(500).JSON(err)
			return
		}

		if err := c.JSON(events); err != nil {
			c.Status(500).JSON(err)
			return
		}
	})

	app.Post("/events", func(c *fiber.Ctx) {
		collection := mg.Db.Collection("events")

		event := new(Event)

		if err := c.BodyParser(event); err != nil {
			c.Status(400).JSON(err)
		}

		event.ID = ""

		insertedEvent, err := collection.InsertOne(context.TODO(), event)
		if err != nil {
			c.Status(500).JSON(err)
			return
		}

		filter := bson.D{{Key: "_id", Value: insertedEvent.InsertedID}}

		searchedEvent := collection.FindOne(context.TODO(), filter)

		createdEvent := &Event{}
		searchedEvent.Decode(createdEvent)

		if err := c.Status(201).JSON(createdEvent); err != nil {
			c.Status(500).JSON(err)
			return
		}
	})

	app.Listen(3000)
}
