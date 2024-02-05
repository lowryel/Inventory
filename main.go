package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	// "github.com/joho/godotenv"
	"regexp"

	"github.com/lowry/inventory-app/middleware"
	"github.com/lowry/inventory-app/storage"

	"xorm.io/xorm"
	// "github.com/golang-jwt/jwt"

	"context"
    "go.uber.org/zap"
    "github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/prometheus/client_golang/prometheus"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
)


var (
	// custom metrics
	customMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "custom_metric_total",
			Help: "Total count of custom metric events",
		},
		[]string{"status"},
	)
)

func init(){
	// register custom metric
	prometheus.MustRegister(customMetric)
}



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

type Address struct{
	Line_1 				string		`json:"line_1"`
	Line_2				string		`json:"line_2" validate:"required"`
	City 				string		`json:"city" validate:"required"`
	Town 				string		`json:"town" validate:"required"`
	Phone 				[]string	`json:"phone" validate:"required"`
	Popular_landmark 	string		`json:"popular_landmark" validate:"required, popular_landmark"`
	House_no			string		`json:"house_no" validate:"required"`
	UserID				uint		`json:"user_id"`
}

type Inventory struct{
	Name 		string		`json:"name" validate:"required"`
	Price 		float32		`json:"price" validate:"required"`
	Instock 	bool		`json:"instock" validate:"required"`
	Quantity 	int			`json:"quantity" validate:"required"`
	UserID		uint 		`json:"user_id"`
}

type Repository struct{
	DBConn *xorm.Engine
}

// OTEL LOGGER
var (
	zlog = otelzap.New(zap.NewExample())
	ctx = context.Context(context.Background())
)



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
		zlog.Ctx(ctx).Error("user not found",
		zap.Error(err))
		return err
	}
	// run some input validation checks
	VerifyEmail(newUser.Email)
	if err != nil{
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"invalid email"})
		zlog.Ctx(ctx).Error("invalid email",
		zap.Error(err))
		return err
	}
	// Create maps to index fields 
	emailMap := make(map[string]bool)
	phoneMap := make(map[string]bool)
	usernameMap := make(map[string]bool)

	// Populate maps from existing users
	for _, user := range *users {
		emailMap[user.Email] = true
		phoneMap[user.Phone] = true 
		usernameMap[user.Username] = true
	}

	// Validate new user fields against maps
	if emailMap[newUser.Email] {
		// email already exists
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"email already exists"})
		return nil
	}

	if phoneMap[newUser.Phone] {
		// phone already exists 
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"phone already exists"})
		zlog.Ctx(ctx).Error("phone already exists",
		zap.Error(err))
		return nil
	}

	if usernameMap[newUser.Username] {
		// username already exists
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"username already exists"})
		zlog.Ctx(ctx).Error("username already exists",
		zap.Error(err))
		return nil
	}
	hash_pass, err := middleware.Hash(newUser.Password)
	if err != nil{
		log.Fatal("couldn't hash password")
		return err
	}
	// insert users into DB users table
	newUser.Created = time.Now().Local()
	newUser.Password = hash_pass
	tx := r.DBConn.NewSession()
	defer tx.Close()
	tx.Begin()
	_, err = tx.Insert(&newUser)
	// insert users into login table
	// _, err = r.DBConn.Insert(&login_data)
	if err != nil {
		log.Fatal("unable to add user to login table")
		tx.Rollback()
		return nil
	}
	tx.Commit()

	login_data := LoginData{}
	login_data.Username=newUser.Username
	login_data.Email = newUser.Email
	login_data.Phone = newUser.Phone
	login_data.Password = newUser.Password

	tx = r.DBConn.NewSession()
	defer tx.Close()
	tx.Begin()
	_, err = tx.Insert(&login_data)
	// insert users into login table
	// _, err = r.DBConn.Insert(&login_data)
	if err != nil {
		log.Fatal("unable to add user to login table")
		tx.Rollback()
		return nil
	}
	tx.Commit()

	log.Printf("users created: %s\n", newUser.Username)
	c.Status(http.StatusCreated).JSON(&fiber.Map{
		"message":"user created",
		 "data":newUser,
		})
	zlog.Ctx(ctx).Info("user created")
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
	c.Status(http.StatusCreated).JSON(&fiber.Map{
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
	// user := c.Locals("claims").(*middleware.JWTClaims)
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

	// retrieve logged in user object with username
	user := &LoginData{}
	_, err = r.DBConn.SQL("select * from login_data where username = ?", loginObj.Username).Get(user)
	if err != nil {
		log.Println(err)
		return err
	}
	// match the saved hash in login table to the login input password
	if err = middleware.HashesMatch(user.Password, loginObj.Password); err != nil{
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"incorrect username or password"})
		return nil
	}
	// signed_user := StoreUsers{}
	token, err := middleware.GenerateToken(user.Username, user.Email)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"msg":"error in generating token"})
	}

	rows, err := middleware.InserToken(r.DBConn, "shops.store_users", token, user.Username)
	if err != nil{
		return nil
	}
	log.Println("Login successful")
	c.Status(http.StatusCreated).JSON(&fiber.Map{"token": token, "rows":rows})
	zlog.Ctx(ctx).Info("login successful")
    // c.Cookie(&fiber.Cookie{
    //     Name:    "token",
    //     Value:   token,
    //     Expires: time.Now().Add(time.Hour * 24),
    // })
	return err
}
 
func VerifyEmail(email string)error{
    if m, _ := regexp.MatchString(`^([\w\.\_]{2,10})@(\w{1,}).([a-z]{2,4})$`, email); !m {
        log.Println("no")
		return errors.New("invalid email")
    }else{
        log.Println("yes")
    }
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
	item:= &database.Inventory{}
	has, err := r.DBConn.ID(id).Get(item)
	if err !=nil{
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"item not in stock"})
		return err
	}
	if !has {
		c.Status(http.StatusNotFound).JSON(&fiber.Map{"message":"item not found", "id":id})
		return err
	}
	updatedData := &database.Inventory{}
	if err := c.BodyParser(updatedData); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
		return nil
	}
	item.Name = updatedData.Name
	item.Price = updatedData.Price
	item.Instock = updatedData.Instock
	item.Quantity = updatedData.Quantity
	_, err = r.DBConn.ID(id).Update(item)
	if err !=nil{
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"item not in stock"})
		return err
	}
	c.Status(http.StatusOK).JSON(&fiber.Map{"message":"updated", "updated_product":item})
	zlog.Ctx(ctx).Info("product update successful")
	return err
}

func (r *Repository) CreateOwnerAddress(c *fiber.Ctx) error {
	log.Println("Creating Address...")
	userID := c.Params("userID")
	var newAddress Address
	if err := c.BodyParser(&newAddress); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{"error": "Invalid request payload"})
	}
	// Set the UserID for the product
	newAddress.UserID = middleware.ParseUserID(userID)
	_, err := r.DBConn.Insert(&newAddress)
	if err != nil {
		c.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message":"Could not create address"})
			log.Printf("Error creating address: %v\n", err)
		return nil
	}
	log.Printf("address created: %s\n", newAddress.Line_1)
	c.Status(http.StatusCreated).JSON(&fiber.Map{
			"message":"address has been added",
			"data":newAddress,
		})
	return err
}

func (r *Repository) GetAddress(c *fiber.Ctx) error{
	user_id := c.Params("user_id")
	item := &Address{}
	// var name string
	has, err := r.DBConn.SQL("select * from address where user_i_d= ?", user_id).Get(item)
	if err !=nil{
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"address not added"})

		log.Printf("ID unavailable %s", err)
	}
	if !has {
		c.Status(http.StatusNotFound).JSON(&fiber.Map{"message":"item not found"})
		return nil
	}
	c.Status(http.StatusOK).JSON(&fiber.Map{"address":item})
	return err
}

func (r *Repository) UpdateOrEditAddress(c *fiber.Ctx) error {
	address_id := c.Params("user_id")
	item:= database.Address{}

	_, err := r.DBConn.SQL("select * from address where user_i_d= ? limit 1", address_id).Get(&item)
	if err !=nil{
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"address not found"})
		return nil
	}

	var updatedData database.Address
	if err = c.BodyParser(&updatedData); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
		return nil
	}
	item.Line_1 = updatedData.Line_1
	item.Line_2 = updatedData.Line_2
	item.City = updatedData.City
	item.Town = updatedData.Town
	item.Phone = updatedData.Phone
	item.Popular_landmark = updatedData.Popular_landmark
	item.House_no = updatedData.House_no
	_, err = r.DBConn.ID(address_id).Update(&item)
	if err !=nil{
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{"message":"item not in stock"})
		return nil
	}
	c.Status(http.StatusOK).JSON(&fiber.Map{"message":"update successful"})
	zlog.Ctx(ctx).Info("address updated")
	return err
}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("api")
	prometheus := fiberprometheus.New("my-awesome-api")
	prometheus.RegisterAt(app, "/metrics")
	// Register the Prometheus endpoint for monitoring
	app.Use("/metrics", prometheus.Middleware)

	// create user
	api.Post("/product/users", r.CreateUser) // check
	// Route to get all users
	api.Get("/user/:user_id", r.GetUser) // check
	// Route to create a product for a specific user
	api.Post("/users/:userID/products", r.CreateOwnerProduct) // check
	api.Post("/login", r.LoginHandler)


    // Middleware to check JWT token for protected routes
	
    // Protected route
	app.Use(middleware.JWTMiddleware())
	api.Get("/products", r.GetProducts) // check
	api.Get("/product/:id", r.GetProduct) // check
	api.Put("/product/update/:id", r.UpdateProduct)

	app.Static("/static", "./public")
	// delete a product
	api.Delete("/product/delete/:id", r.DeleteProduct)
	api.Post("/address/create/:userID", r.CreateOwnerAddress)
	api.Get("/address/:user_id", r.GetAddress)
	api.Put("/address/update/:user_id", r.UpdateOrEditAddress)
	// jwt middleware to restrict access to unauthorised users
	// ProtectedRoute is a protected route that requires a valid JWT token
	api.Get("protected/", middleware.ProtectedRoute)
}



func main(){
	// And then pass ctx to propagate the span.
	zlog.Ctx(ctx).Error("hello from zap",
		zap.Error(errors.New("hello world")),
		zap.String("foo", "bar"))

	app :=fiber.New(
		fiber.Config{
		    ServerHeader:  "Fiber",
    		AppName: "Mini Inventory v1.0.1",
		},
	)
	// prometheus := fiberprometheus.New("my-awesome-api")
	// prometheus.RegisterAt(app, "/metrics")
	// // Register the Prometheus endpoint for monitoring
	// app.Use("/metrics", prometheus.Middleware)

	engine, err := database.NewConnection()
	if err != nil{
		log.Fatal("database connection unsuccessful", err)
		zlog.Ctx(ctx).Error("database connection failed",
		zap.Error(errors.New("database connection failed")),
		)
	}
	r := Repository{
		DBConn: engine,
	}
	r.SetupRoutes(app)
	app.Listen(":5000")
}

