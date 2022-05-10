package main

import (
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
	"gopkg.in/gomail.v2"
	"log"
	"os"
	"strconv"
)

type Feedback struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Date   string `json:"date"`
	Source string `json:"source"`
}

// Database instance
var db *sql.DB

// Database settings
var (
	host     = os.Getenv("DATABASE_HOST")
	port, _  = strconv.Atoi(os.Getenv("DATABASE_PORT"))
	user     = os.Getenv("POSTGRES_USER")
	password = os.Getenv("POSTGRES_PASSWORD")
	dbname   = os.Getenv("POSTGRES_DB")
)

func Connect() error {
	fmt.Println(fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable", user, password, host, port, dbname))

	var err error
	db, err = sql.Open("postgres", fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable", user, password, host, port, dbname))
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	return nil
}

func main() {
	// Connect with database
	if err := Connect(); err != nil {
		log.Fatal(err)
	}

	// Create a Fiber app
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(201).JSON("Success")
	})

	// Create new Feedback
	app.Post("/add", func(c *fiber.Ctx) error {
		u := new(Feedback)

		// Parse body into struct
		if err := c.BodyParser(u); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		// Insert Feedback into database
		_, err := db.Query("INSERT INTO FEEDBACK (email, Source) VALUES ($1, $2)", u.Email, u.Source)
		if err != nil {
			return err
		}

		// Print result
		log.Println(fmt.Sprintf("Insert %s", u.Email))

		// Send email
		m := gomail.NewMessage()
		m.SetHeader("From", os.Getenv("SENDER_EMAIL"))
		m.SetHeader("To", os.Getenv("EMAIL_TO_SEND"))
		m.SetHeader("Subject", "New customer left his email!")
		m.SetBody("text/plain", fmt.Sprintf("Customer %s just left his email on %s", u.Email, u.Source))
		d := gomail.NewDialer("smtp.yandex.ru", 465, os.Getenv("SENDER_EMAIL"), os.Getenv("SENDER_PASSWORD"))

		if err := d.DialAndSend(m); err != nil {
			log.Fatal(err)
			return c.Status(500).JSON("Failed to send email")
		}

		return c.Status(201).JSON(&fiber.Map{"success": true})
	})

	app.Listen(":8080")
}
