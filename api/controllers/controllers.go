package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"rest/api/database"
	"rest/api/jwtService"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Bank     int    `json:"bank"`
}

func Init() {
	var err error
	db, err = database.Init()
	if err != nil {
		fmt.Print("error initializing database")
		return
	}
	jwtService.SetSecret()
}

func HandleSignup(c *gin.Context) {

	var user User
	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to parse request body. Wrong json"})
		return
	}

	var username string
	db.QueryRow("SELECT username FROM users WHERE username = $1", user.Username).Scan(&username)
	if len(username) != 0 {
		c.JSON(400, gin.H{"error": "User already exists"})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hash)

	_, err = db.Exec("INSERT INTO users (username, password, role, bank) values ($1, $2, $3, $4)", user.Username, user.Password, "user", 0)
	if err != nil {
		c.JSON(400, gin.H{"error": "failed to insert into database"})
		fmt.Println(err)
		return
	}
	c.JSON(200, gin.H{
		"message": "User signed up successfully",
		"user":    user,
	})
}

func HandleLogin(c *gin.Context) {
	var user User
	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to parse request body. Wrong json"})
		return
	}
	var userBd User
	err = db.QueryRow("SELECT id, username, password, role, bank FROM users WHERE username = $1", user.Username).Scan(&userBd.Id, &userBd.Username, &userBd.Password, &userBd.Role, &userBd.Bank)

	if err == sql.ErrNoRows {
		c.JSON(400, gin.H{"error": "User does not exist"})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(userBd.Password), []byte(user.Password))

	if err == nil {
		claims := jwt.MapClaims{
			"userid":   userBd.Id,
			"username": userBd.Username,
			"role":     userBd.Role,
		}
		token, err := jwtService.GenerateJWT(claims)
		if err != nil {
			c.JSON(400, gin.H{"error": "problems with generating token"})
		} else {
			c.JSON(200, gin.H{"token": token})
		}
	} else {
		c.JSON(400, gin.H{"error": "incorrect password"})
	}
}

type Item struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

func AddToMenu(c *gin.Context) {
	var dto struct {
		Token string `json:"token"`
		Item  Item   `json:"item"`
	}

	if err := c.BindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
		return
	}

	token := dto.Token

	claims, _ := jwtService.ValidateJWT(string(token))
	user := parseClaims(claims)
	if user.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You are not an admin"})
		return
	}

	_, err := db.Exec("INSERT INTO items (name, price) values ($1, $2)", dto.Item.Name, dto.Item.Price)
	if err != nil {
		c.JSON(400, gin.H{"error": "Error adding item"})
	}

	c.JSON(200, gin.H{"success": "Item added"})
}

func RemoveFromMenu(c *gin.Context) {

	var dto struct {
		Token string `json:"token"`
	}
	c.BindJSON(&dto)
	token := dto.Token
	itemId := c.Param("id")

	claims, _ := jwtService.ValidateJWT(string(token))
	user := parseClaims(claims)

	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not an admin"})
		return
	}
	_, err := db.Exec("DELETE FROM items WHERE id =$1", itemId)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "can't remove the item"})
	}
	c.JSON(http.StatusAccepted, gin.H{"success": "item removed"})
}

func PlaceOrder(c *gin.Context) {
	var dto struct {
		Token string `json:"token"`
		Price int    `json:"price"`
	}
	c.BindJSON(&dto)
	token := dto.Token
	itemId := c.Param("id")

	claims, _ := jwtService.ValidateJWT(string(token))
	user := parseClaims(claims)

	if user.Bank < dto.Price {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": "you do not have enough money in your bank"})
		return
	}

	diff := user.Bank - dto.Price
	_, err := db.Exec("UPDATE users set bank = $1 where id = $2", diff, user.Id)
	if err != nil {
		c.JSON(500, gin.H{"error": "something went wrong when subtracting money"})
		return
	}
	_, err = db.Exec("INSERT INTO orders (user_id, item_id, status) values ($1, $2, $3)", user.Id, itemId, "created")

	if err != nil {
		c.JSON(500, gin.H{"error": "item does not exist"})
		return
	}

	_, err = db.Exec("UPDATE stats SET items_sold = items_sold + $1, revenue = revenue + $2 where id=1", 1, dto.Price)
	if err != nil {
		fmt.Println("error updaring stats")
	}

	c.JSON(http.StatusAccepted, gin.H{"error": "order placed successfully"})
}

func ShowMenu(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM items")
	if err != nil {
		fmt.Print("error selecting items")
	}
	var items []Item
	for rows.Next() {
		var item Item

		err := rows.Scan(&item.Id, &item.Name, &item.Price)
		if err != nil {
			fmt.Println("Failed to scan row:", err)
			continue
		}

		items = append(items, item)
	}
	c.JSON(200, items)
}

func parseClaims(claims jwt.MapClaims) User {
	user := User{}

	if userID, ok := claims["userid"].(float64); ok {
		user.Id = int(userID)
	}
	if username, ok := claims["username"].(string); ok {
		user.Username = username
	}
	if role, ok := claims["role"].(string); ok {
		user.Role = role
	}
	var bank int
	row := db.QueryRow("SELECT bank from users where username=$1", user.Username)
	row.Scan(&bank)
	user.Bank = bank
	return user
}

func GetRevenue(c *gin.Context) {

	var dto struct {
		Token string `json:"token"`
	}
	c.BindJSON(&dto)
	token := dto.Token

	claims, _ := jwtService.ValidateJWT(string(token))
	user := parseClaims(claims)

	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not an admin"})
		return
	}

	var items_sold int
	var revenue int
	row := db.QueryRow("SELECT * from stats where id=1")
	row.Scan(&items_sold, &revenue)
	c.JSON(200, gin.H{"items_sold": items_sold, "revenue": revenue})
}

func UpdateStatus(c *gin.Context) {
	var dto struct {
		Token string `json:"token"`
	}
	orderId := c.Param("id")

	fmt.Println("order id is ", orderId)
	c.BindJSON(&dto)
	token := dto.Token

	claims, _ := jwtService.ValidateJWT(string(token))
	user := parseClaims(claims)

	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not an admin"})
		return
	}

	var status string
	rows, err := db.Query("SELECT status FROM orders WHERE id = $1", orderId)
	if err != nil {
		fmt.Println("Failed to query status:", err)
		c.JSON(http.StatusInternalServerError, "")
		return
	}
	defer rows.Close()

	if !rows.Next() {
		fmt.Println("No rows found for orderId:", orderId)
		c.JSON(http.StatusNotFound, "")
		return
	}

	if err := rows.Scan(&status); err != nil {
		fmt.Println("Failed to scan status:", err)
		c.JSON(http.StatusInternalServerError, "")
		return
	}

	fmt.Println("Status:", status)

	var toSet string
	if status == "created" {
		toSet = "being made"
	} else if status == "being made" {
		toSet = "done"
	}

	if status == "done" {
		db.Exec("DELETE FROM orders where id=$1", orderId)
	} else {
		db.Exec("UPDATE orders set status=$1 where id=$2", toSet, orderId)
	}
	c.JSON(200, "")
}

func GetDocumentation(c *gin.Context) {
	htmlContent, err := os.ReadFile("index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error reading documentation file")
		return
	}
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, string(htmlContent))
}
