package middleware

import (
	// "strings"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	// "google.golang.org/appengine/user"

	"errors"
	"fmt"
	"log"

	"github.com/patrickmn/go-cache"

	"xorm.io/xorm"
)

const SecretKey = "generate-new-jwt"
// JWTClaims struct represents the claims for JWT
type JWTClaims struct {
    Email       string    	`json:"email"`
    Username string 		`json:"username"`
    jwt.StandardClaims
}


func Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func HashesMatch(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	// Compare the two hashes
	return err
}


// generate jwt token with username and email
func GenerateToken(email, username string ) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, &JWTClaims{
        Username: username,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
        },
    })
    tokenString, err := token.SignedString([]byte(SecretKey))
    if err != nil {
        return "", err
    }
    return tokenString, nil
}

// JWTMiddleware checks the JWT token in the request headers
// JWTMiddleware caches the parsed and validated token 
// for performance
func JWTMiddleware() fiber.Handler {
  tokenCache := cache.New(5*time.Minute, 10*time.Minute)
  return func(c *fiber.Ctx) error {
    tokenString := c.Get("Authorization")
    fmt.Println(tokenString[7:])
    // Check cache
    if token, ok := tokenCache.Get(tokenString[7:]); ok {
      c.Locals("user", token)
      return c.Next()
    }
    
    token, err := parseAndValidateToken(tokenString[7:])
    if err != nil {
      log.Printf("Token invalid error: %s\n", err)
      return c.Status(http.StatusBadRequest).JSON(&fiber.Map{
        "message": "Invalid token"})
    }
    // Cache valid token
    tokenCache.Set(string(tokenString[7:]), token, -1)
    // Extract and store user identity
    user := token.Claims.(jwt.MapClaims) 
    c.Locals("user", user)
    return c.Next()
  }
}

// Centralized parsing and validation
func parseAndValidateToken(tokenString string) (*jwt.Token, error) {
  token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", token.Header["Token"])
    }
    log.Println(token.Header["Token"])
    return []byte(SecretKey), nil
  })
  if err != nil {
    return nil, err 
  }

  // Additional validation checks
  if !token.Valid {
    return nil, ExpiredError
  }
  return token, nil
}
// Custom error types
var (
  ExpiredError= errors.New("token expired")
)

// ProtectedRoute is a protected route that requires a valid JWT token
func ProtectedRoute(c *fiber.Ctx) error {
    // user := c.Locals("root").(*jwt.MapClaims)
    return c.JSON(fiber.Map{"message": "Hello"})
}


func InserToken(db *xorm.Engine, tableName string, token string, username string) (int64, error){
    // SQL statement
    query := "UPDATE" + tableName + "SET token = "+ token + "WHERE username = " + username
    // Execute the SQL statement
    res, err := db.Insert(query)
    if err != nil{
        return 0, nil
    }
    return res, nil
}
