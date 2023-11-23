package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	// "github.com/joho/godotenv"
	"github.com/lowry/inventory-app/storage"

	"xorm.io/xorm"
)

type Inventory struct{
	Name 		string	`json:"name"`
	Price 		float32	`json:"price"`
	Instock 	bool	`json:"instock"`
	Quantity 	int	`json:"quantity"`
}

type Repository struct{
	DBConn *xorm.Engine
}

func (r *Repository) CreateProduct(c *fiber.Ctx) error {
	fmt.Println("Creating products...")
	item := Inventory{}
	err :=c.BodyParser(&item)
	if err != nil{
		c.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message":"request failed"})
		return err
	}
	_, err = r.DBConn.Insert(&item)
	if err != nil {
		c.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message":"Could not create product"})
			log.Printf("Error creating product: %v\n", err)
		return err
	}
	log.Printf("product created: %s\n", item.Name)
	c.Status(http.StatusOK).JSON(&fiber.Map{
		"message":"product has been added",
		 "data":item,
		})
	return nil
}



func (r *Repository) GetProducts(c *fiber.Ctx) error {
	fmt.Println("Getting all products...")
	items := &[]database.Inventory{}
	err := r.DBConn.Find(items)
	if err != nil{
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"Bad Request"})
		return err
	}
	c.Status(http.StatusOK).JSON(&fiber.Map{"data": items})
	return nil
}

func (r *Repository) GetProduct(c *fiber.Ctx) error{
	id := c.Params("id")
	fmt.Println(id)
	fmt.Println("John Robert Oppenheimer")
	var item database.Inventory
	// var name string
	has, err := r.DBConn.SQL("select * from inventory where i_d= ? limit 1", id).Get(&item)
	if err !=nil{
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"item not in stock"})
		log.Fatal(err)
	}
	fmt.Println(has)
	if !has {
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"item not available"})
		return err
	}
	c.Status(http.StatusOK).JSON(&fiber.Map{"message":"Success", "data":item})
	return nil
}


func (r *Repository) DeleteProduct(c *fiber.Ctx) error{
	id := c.Params("id")
	fmt.Println(id)
	var item database.Inventory
	// var name string
	has, err := r.DBConn.ID(id).Delete(&item) // has returned a number either 0 or 1
	if err !=nil{
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"item not in stock"})
		return err
	}
	if has < 1 {
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"item not available", "id":id})
		return err
	}
	c.Status(http.StatusOK).JSON(&fiber.Map{"message":"Delete Success"})
	return nil
}


func (r *Repository) UpdateProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	item:= database.Inventory{}
	has, err := r.DBConn.ID(id).Get(&item)
	if err !=nil{
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"item not in stock"})
		return err
	}
	if !has {
		c.Status(http.StatusNotFound).JSON(&fiber.Map{"message":"item not found", "id":id})
		return err
	}
	var updatedData database.Inventory
	if err := c.BodyParser(&updatedData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}
	item.Name = updatedData.Name
	item.Price = updatedData.Price
	item.Instock = updatedData.Instock
	item.Quantity = updatedData.Quantity
	_, err = r.DBConn.ID(id).Update(&item)
	if err !=nil{
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"item not in stock"})
		return err
	}
	c.Status(http.StatusOK).JSON(&fiber.Map{"message":"updated", "data":item})
	return nil
}


func (r *Repository) SetupRoutes(app *fiber.App) {

	api := app.Group("api")
	api.Post("/products", r.CreateProduct)
	api.Get("/products", r.GetProducts)
	
	api.Get("/product/:id", r.GetProduct)
	api.Put("/product/update/:id", r.UpdateProduct)
	// middleware to handle other functionalities on views beneath
	// you can also use the middleware on an endpoint
	app.Use("*/product", func(c *fiber.Ctx) error {
		log.Println("a middleware")
		fmt.Println("some giberisssssssssh")
		return c.Next()
	})
	app.Static("static", "./public")
	api.Delete("/product/delete/:id", r.DeleteProduct)
}

func main(){
	// err := godotenv.Load(".env")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	app :=fiber.New(
		fiber.Config{
		    ServerHeader:  "Fiber",
    		AppName: "Mini Inventory v1.0.1",
		},
	)
	engine, err := database.NewConnection()
	if err != nil{
		log.Fatal("db connection failed", err)
	}



	// config := &database.Config{
	// 	Host: os.Getenv("DB_HOST"),
	// 	Port: os.Getenv("DB_PORT"),
	// 	User: os.Getenv("DB_USER"),
	// 	DBName: os.Getenv("DB_NAME"),
	// 	Password: os.Getenv("DB_PASS"),
	// 	SSLMode: os.Getenv("DB_SSLMODE"),
	// }
	r := Repository{
		DBConn: engine,
	}
	r.SetupRoutes(app)
	app.Listen(":5000")
}

