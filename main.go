package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	// "github.com/joho/godotenv"
	"github.com/lowry/inventory-app/middleware"
	"github.com/lowry/inventory-app/storage"

	"xorm.io/xorm"
)
type StoreUsers struct{
	Name 		string	`json:"name"`
	Phone 		string	`json:"phone"`
	Email 		string	`json:"email"`
	// Inventory 	[]Inventory `json:"inventory"`
}

type Inventory struct{
	Name 		string	`json:"name"`
	Price 		float32	`json:"price"`
	Instock 	bool	`json:"instock"`
	Quantity 	int	`json:"quantity"`
	UserID		uint `json:"user_id"`
}

type Repository struct{
	DBConn *xorm.Engine
}

func (r *Repository) CreateUser(c *fiber.Ctx) error {
	var newUser StoreUsers
	err := c.BodyParser(&newUser)
	if err != nil{
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"request failed"})
		return err
	}

	// user_id := c.Params("userID")
	users := &[]StoreUsers{}
	err = r.DBConn.Find(users)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// run some input validation checks
	for _, users := range *users{
		if newUser.Email == users.Email{
			c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"email already used"})
			return err
		}
		if newUser.Phone == users.Phone{
			c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"email already used"})
			return err
		}
	}
	// insert users into DB
	_, err = r.DBConn.Insert(&newUser)
	if err != nil{
		c.Status(400).JSON(&fiber.Map{"message":"request failed"})
		return err
	}
	log.Printf("users created: %s\n", newUser.Name)
	c.Status(http.StatusOK).JSON(&fiber.Map{
		"message":"users object created",
		 "data":newUser,
		})
	return nil
}


func (r *Repository) GetUsers(c *fiber.Ctx) error {
	users := &[]StoreUsers{}
    err := r.DBConn.Find(users)
    if err != nil {
        fmt.Println(err)
        return err
    }
	c.JSON(&fiber.Map{"data":users})
	return nil
}


func (r *Repository) CreateOwnerProduct(c *fiber.Ctx) error {
	fmt.Println("Creating products...")
	userID := c.Params("userID")
	var newUser Inventory
	if err := c.BodyParser(&newUser); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{"error": "Invalid request payload"})
	}
	// Set the UserID for the product
	newUser.UserID = middleware.ParseUserID(userID)
	_, err := r.DBConn.Insert(&newUser)
	if err != nil {
		c.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message":"Could not create product"})
			log.Printf("Error creating product: %v\n", err)
		return err
	}
	log.Printf("product created: %s\n", newUser.Name)
	c.Status(http.StatusOK).JSON(&fiber.Map{
		"message":"product has been added",
		 "data":newUser,
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
	c.Status(http.StatusOK).JSON(&fiber.Map{"data":item})
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
	api.Post("/products", r.CreateOwnerProduct)
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
	app.Static("/static", "./public")

	api.Delete("/product/delete/:id", r.DeleteProduct)

	// create users
	api.Post("/product/users", r.CreateUser)

	// Route to get all users
	api.Get("/users", r.GetUsers)

	// Route to create a product for a specific users
	api.Post("/users/:userID/products", r.CreateOwnerProduct)
}

func main(){
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

	r := Repository{
		DBConn: engine,
	}
	r.SetupRoutes(app)
	app.Listen(":5000")
}

