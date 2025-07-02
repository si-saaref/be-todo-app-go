package main

import (
	"database/sql"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	_ "github.com/jackc/pgx/v5/stdlib" // registers "pgx" and "postgres"
)

type Activity struct {
	Id int `json:"id"`
	Title string `json:"title" validate:"required"`
	Category string `json:"category" validate:"required,oneof=TASK EVENT"`
	Description string `json:"description" validate:"required"`
	ActivityDate time.Time `json:"activity_date" validate:"required"`
	Status string `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

var db *sql.DB // declare globally to reuse


func main() {
	var err error
	db, err = initDB()

	if err != nil { 
		panic(err)
	}

	defer db.Close()

	app := fiber.New()
	validate := validator.New()

	app.Get("/activities", getActivities)
	app.Post("/activities", createActivity(validate))
	app.Put("/activities/:id", updateActivity(validate))
	app.Delete("/activities/:id", deleteActivity)

	app.Listen(":8081")
}

func initDB() (*sql.DB, error) {
	// connStr := "postgresql://postgres:6imCmzMoRsAyLYbB@db.lrvfkszadiibupddigro.supabase.co:5432/postgres?sslmode=require&statement_cache_mode=describe"
	connStr := "user=postgres.lrvfkszadiibupddigro password=6imCmzMoRsAyLYbB host=aws-0-ap-southeast-1.pooler.supabase.com port=6543 dbname=postgres statement_cache_mode=describe"

	db,err := sql.Open("pgx", connStr)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	err = db.Ping()

	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func getActivities(c *fiber.Ctx) error {
	// ! If we use select * this will cause an error =>   "message": "ERROR: prepared statement \"stmtcache_6decad3076c699014b1e184da23a5c0c8679568ca37881fa\" already exists (SQLSTATE 42P05)"
	// rows, err := db.Query("SELECT * FROM activities")
	rows, err := db.Query("SELECT id, title, category, description, activity_date, status, created_at FROM activities")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	defer rows.Close()

	var activities []Activity

	for rows.Next() {
		var activity Activity
		err = rows.Scan(
			&activity.Id, 
			&activity.Title, 
			&activity.Category, 
			&activity.Description, 
			&activity.ActivityDate, 
			&activity.Status, 
			&activity.CreatedAt,
		)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError,).JSON(fiber.Map{"message": err.Error()})
		}

		activities = append(activities, activity)
	}

	return c.Status(fiber.StatusOK).JSON(activities)
}

func createActivity(validate *validator.Validate) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var activity Activity
	
		err := c.BodyParser(&activity)
	
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
	
		if err = validate.Struct(&activity); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
	
		sqlStatement := `INSERT INTO activities(title, category, description, activity_date, status) VALUES($1, $2, $3, $4, $5) RETURNING id`
	
		err = db.QueryRow(
			sqlStatement,
			activity.Title,
			activity.Category,
			activity.Description,
			activity.ActivityDate,
			activity.Status,
		).Scan(&activity.Id)
	
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
	
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Successfully create activity"})
	}
}


func updateActivity(validate * validator.Validate) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
	
		var activity Activity
		
		err := c.BodyParser(&activity)
	
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
	
		if err = validate.Struct(&activity); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}

		sqlStatement := `UPDATE activities SET title = $1, category = $2, description = $3, activity_date = $4 WHERE id = $5 RETURNING id`

		err = db.QueryRow(
			sqlStatement,
			activity.Title,
			activity.Category,
			activity.Description,
			activity.ActivityDate,
			id,
		).Scan(&activity.Id)

		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Successfully update activity"})
	}
}

func deleteActivity() error {

}