package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	// "github.com/joho/godotenv"
	"github.com/lowry/inventory-app/middleware"
	"github.com/lowry/inventory-app/storage"

	"xorm.io/xorm"
)
type StoreUsers struct{
	First_name		string		`json:"first_name" validate:"required"`
	Username 		string		`json:"username" validate:"required, min=3, max=50"`
	Phone 			string		`json:"phone" validate:"required, min=10, max=15"`
	Email 			string		`json:"email" validate:"required, email"`
	Password		string		`json:"password" validate:"required, min=8"`
	User_type		string		`json:"user_type" validate:"required, eq=ADMIN|eq=USER"`
	Token			string		`json:"token"`
	Created			time.Time	`json:"created"`
	Refresh_token	string		`json:"refresh_token"`
}

// USer type
type LoginObject struct {
	Username string	`json:"username"`
	Password string	`json:"password"`
}

type LoginData struct{
	Username 		string		`json:"username" validate:"required, min=3, max=50"`
	Phone 			string		`json:"phone" validate:"required, min=10, max=15"`
	Email 			string		`json:"email" validate:"required, email"`
	Password		string		`json:"password" validate:"required, min=8"`
}

type Inventory struct{
	Name 		string	`json:"name" validate:"required"`
	Price 		float32	`json:"price" validate:"required"`
	Instock 	bool	`json:"instock" validate:"required"`
	Quantity 	int		`json:"quantity" validate:"required"`
	UserID		uint 	`json:"user_id"`
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

	users := &[]StoreUsers{}
	err = r.DBConn.Find(users)
	if err != nil {
		log.Println(err)
		return err
	}
	// run some input validation checks
	for _, users := range *users{
		if newUser.Email == users.Email{
			c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"email already used"})
			return err
		}
		if newUser.Phone == users.Phone{
			c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"phone number already used"})
			return err
		}
		if newUser.Username == users.Username{
			c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"username already taken"})
			return err
		}
	}
	hash_pass, err := middleware.Hash(newUser.Password)
	if err != nil{
		log.Fatal("couldn't hash password")
		return err
	}
	// insert users into DB users table
	newUser.Created = time.Now().Local()
	newUser.Password = hash_pass
	_, err = r.DBConn.Insert(&newUser)
	if err != nil{
		c.Status(400).JSON(&fiber.Map{"message":"request failed"})
		return err
	}
	login_data := LoginData{}
	login_data.Username=newUser.Username
	login_data.Email = newUser.Email
	login_data.Phone = newUser.Phone
	login_data.Password = newUser.Password

	// insert users into login table
	_, err = r.DBConn.Insert(&login_data)
	if err != nil {
		log.Fatal("unable to add user to login table")
		return nil
	}

	log.Printf("users created: %s\n", newUser.Username)
	c.Status(http.StatusOK).JSON(&fiber.Map{
		"message":"user created",
		 "data":newUser,
		})
	return nil
}


func (r *Repository) GetUser(c *fiber.Ctx) error {
	user := &database.StoreUsers{}
	id := c.Params("user_id")
    has, err := r.DBConn.SQL("select * from store_users where i_d= ? limit 1", id).Get(user)
	if !has{
		c.Status(http.StatusNotFound).JSON(&fiber.Map{"message":"data not found"})
		log.Println("invalid id")
		return nil
	}
    if err != nil {
        log.Println(err)
        return err
    }
	c.JSON(&fiber.Map{"user_data":user})
	return nil
}


func (r *Repository) CreateOwnerProduct(c *fiber.Ctx) error {
	log.Println("Creating products...")
	userID := c.Params("userID")
	var newProduct Inventory
	if err := c.BodyParser(&newProduct); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{"error": "Invalid request payload"})
	}
	// Set the UserID for the product
	newProduct.UserID = middleware.ParseUserID(userID)
	_, err := r.DBConn.Insert(&newProduct)
	if err != nil {
		c.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message":"Could not create product"})
			log.Printf("Error creating product: %v\n", err)
		return err
	}
	log.Printf("product created: %s\n", newProduct.Name)
	c.Status(http.StatusOK).JSON(&fiber.Map{
		"message":"product has been added",
		 "data":newProduct,
		})
	return nil
}


func (r *Repository) GetProducts(c *fiber.Ctx) error {
	log.Println("Getting all products...")
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
	var item database.Inventory
	// var name string
	has, err := r.DBConn.SQL("select * from inventory where i_d= ? limit 1", id).Get(&item)
	if err !=nil{
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"item not in stock"})
		log.Fatal(err)
	}
	if !has {
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"item not available"})
		return err
	}
	c.Status(http.StatusOK).JSON(&fiber.Map{"product_data":item})
	return nil
}


func (r *Repository) DeleteProduct(c *fiber.Ctx) error{
	id := c.Params("id")
	log.Println(id)
	var item database.Inventory
	// var name string
	has, err := r.DBConn.ID(id).Delete(&item) // "has" returned a number either 0 or 1
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

func (r *Repository) LoginHandler(c *fiber.Ctx) error {
	loginObj := &LoginObject{}
	err := c.BodyParser(loginObj)
	if err != nil{
		log.Fatal("invalid data")
		return nil
	}
	hash, err := middleware.Hash(loginObj.Password) // hash login password
	if err != nil {
		log.Fatal("password not hashed")
		return err
	}
	// retrieve logged in user object with username
	user := &LoginData{}
	_, err = r.DBConn.SQL("select * from login_data where username = ?", loginObj.Username).Get(user)
	if err != nil {
		log.Println(err)
		return err
	}
	// match the saved hash in login table to the login input password
	if err = middleware.HashesMatch(user.Password, loginObj.Password); err != nil{
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"incorrect password"})
		log.Println(hash, "problem with hash match")
		return nil
	}

	log.Println("Login successful")
	c.Status(200).JSON(&fiber.Map{"message":"login successful"})
	return nil
}

// func (r *Repository) insertLoginData(tableName string, user *StoreUsers) (int, error) {
// 	err := r.DBConn.SQL("insert into" + tableName + user.Username + user.Phone + user.Email + user.Password)
// 	if err != nil {
// 		log.Fatal("unable to add user to login table")
// 		return 0, nil
// 	}
// 	return 0, errors.New("")
// }

func (r *Repository) UpdateProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	item:= database.Inventory{}
	log.Printf("ID: %v", id)
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
	c.Status(http.StatusOK).JSON(&fiber.Map{"message":"updated", "updated_product":item})
	return nil
}



func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("api")

	// create user
	api.Post("/product/users", r.CreateUser) // check
	// Route to get all users
	api.Get("/user/:user_id", r.GetUser) // check
	// Route to create a product for a specific user
	api.Post("/users/:userID/products", r.CreateOwnerProduct) // check
	api.Post("/login", r.LoginHandler)

	api.Get("/products", r.GetProducts) // check
	api.Get("/product/:id", r.GetProduct) // check
	api.Put("/product/update/:id", r.UpdateProduct)

	// middleware to handle other functionalities on views beneath
	// you can also use the middleware on an endpoint
	app.Use("/api/user/:user_id", func(c *fiber.Ctx) error {
		log.Println("a middleware")
		return c.Next()
	})
	app.Static("/static", "./public")
	// delete a product
	api.Delete("/product/delete/:id", r.DeleteProduct)
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

