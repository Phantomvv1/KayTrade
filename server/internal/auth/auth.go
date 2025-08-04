package auth

import (
	"context"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
)

const (
	Admin = iota + 1
	User
)

var Domain = ""
var Secure = false

type Profile struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Type      byte      `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func GenerateJWT(id string, accountType byte, email string) (string, error) {
	claims := jwt.MapClaims{
		"id":         id,
		"type":       accountType,
		"email":      email,
		"expiration": time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey := os.Getenv("JWT_KEY")
	return token.SignedString([]byte(jwtKey))
}

func ValidateJWT(tokenString string) (string, byte, string, error) {
	claims := &jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.ErrUnsupported
		}

		return []byte(os.Getenv("JWT_KEY")), nil
	})

	if err != nil || !token.Valid {
		return "", 0, "", err
	}

	expiration, ok := (*claims)["expiration"].(float64)
	if !ok {
		return "", 0, "", errors.New("Error parsing the expiration date of the token")
	}

	if int64(expiration) < time.Now().Unix() {
		return "", 0, "", errors.New("Error token has expired")
	}

	id, ok := (*claims)["id"].(string)
	if !ok {
		return "", 0, "", errors.New("Error parsing the id")
	}

	accountType, ok := (*claims)["type"].(float64)
	if !ok {
		return "", 0, "", errors.New("Error parsing the account")
	}

	email, ok := (*claims)["email"].(string)
	if !ok {
		return "", 0, "", errors.New("Error parsing the email")
	}

	return id, byte(accountType), email, nil
}

func SHA512(text string) string {
	algorithm := sha512.New()
	algorithm.Write([]byte(text))
	result := algorithm.Sum(nil)
	return fmt.Sprintf("%x", result)
}

func CreateAuthTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists authentication (id uuid primary key default gen_random_uuid(), full_name text, "+
		"email text, password text, type int check (type in (1, 2)), created_at timestamp default current_timestamp, updated_at timestamp default current_timestamp)")
	return err
}

func CreateRefreshTokenTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists r_tokens(token uuid primary key default gen_random_uuid(), user_id uuid references authentication(id) "+
		"expiration timestamp default current_timestamp + '5 days'::interval, valid bool)")
	return err
}

func SignUp(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Databse connection failed"})
		return
	}

	var information map[string]string
	json.NewDecoder(c.Request.Body).Decode(&information) //fullName, email, password

	if err = CreateAuthTable(conn); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating a table for authentication"})
		return
	}

	if _, ok := information["email"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided email of the user"})
		return
	}

	validEmail, err := regexp.MatchString(".*@.*\\..*", information["email"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusForbidden, gin.H{"error": "Error invalid email"})
		return
	}

	if !validEmail {
		log.Println("Invalid email")
		c.JSON(http.StatusForbidden, gin.H{"error": "Error invalid email"})
		return
	}

	if _, ok := information["password"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided password of the user"})
		return
	}

	if _, ok := information["name"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided name of the user"})
		return
	}

	var check string
	err = conn.QueryRow(context.Background(), "select email from authentication where email = $1", information["email"]).Scan(&check)
	emailExists := true
	if err != nil {
		if err == pgx.ErrNoRows {
			emailExists = false
		} else {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the password from the table"})
			return
		}
	}

	if emailExists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "There is already a person with this email"})
		return
	}

	hashedPassword := SHA512(information["password"])
	_, err = conn.Exec(context.Background(), "insert into authentication (full_name, email, password, type) values ($1, $2, $3, $4)",
		information["name"], information["email"], hashedPassword, User)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting the information into the database."})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func LogIn(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	if err = CreateAuthTable(conn); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var information map[string]string
	json.NewDecoder(c.Request.Body).Decode(&information) //email, password

	var passwordCheck, email string
	var accoutType byte
	var id string
	err = conn.QueryRow(context.Background(), "select id, password, type, email from authentication a where a.email = $1;", information["email"]).Scan(
		&id, &passwordCheck, &accoutType, &email)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "There isn't anybody registered with this email!"})
			return
		} else {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while trying to log in"})
			return
		}
	}

	if SHA512(information["password"]) != passwordCheck {
		log.Println("Wrong password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error wrong password"})
		return
	}

	jwtToken, err := GenerateJWT(id, accoutType, email)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while generating your token"})
		return
	}

	refreshToken := ""
	err = conn.QueryRow(context.Background(), "insert into r_tokens (user_id, valid) values ($1, $2) returning token", id, true).Scan(&refreshToken)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to generate a refresh token"})
		return
	}

	c.SetCookie("refresh", refreshToken, int((5 * 24 * time.Hour).Seconds()), "/", Domain, Secure, true)
	c.JSON(http.StatusOK, gin.H{"token": jwtToken})
}

func GetCurrentProfile(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error couldn't connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	userID, _ := c.Get("id")
	id := userID.(string)

	accountTypeUnk, _ := c.Get("accountType")
	accountType := accountTypeUnk.(byte)

	var name, email string
	var createdAt, updatedAt time.Time
	err = conn.QueryRow(context.Background(), "select full_name, email, created_at, updated_at from authentication where id = $1", id).Scan(&name, &email, &createdAt, &updatedAt)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting information from the database"})
		return
	}

	userProfile := Profile{
		ID:        id,
		Name:      name,
		Email:     email,
		Type:      accountType,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"profile": userProfile})
}

func GetAllUsers(c *gin.Context) {
	acc, _ := c.Get("accountType")
	accountType := acc.(byte)

	if accountType != Admin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error only admins can view all of the accounts"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "select id, full_name, email, type, created_at, updated_at from authentication")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error couldn't get information from the database"})
		return
	}

	var profiles []Profile
	for rows.Next() {
		profile := Profile{}
		err = rows.Scan(&profile.ID, &profile.Name, &profile.Email, &profile.Type, &profile.CreatedAt, &profile.UpdatedAt)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error working with the data from the database"})
			return
		}

		profiles = append(profiles, profile)
	}

	if rows.Err() != nil {
		log.Println(rows.Err())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error working with the data from the database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": profiles})
}

func Refresh(c *gin.Context) {
	userId, _ := c.Get("id")
	id := userId.(string)

	refresh, err := c.Cookie("refresh")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get the refresh token"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	//Extra safety check
	ownerMatch, valid := false, false
	err = conn.QueryRow(context.Background(), "select valid, case when user_id = $1 then true else false end from r_tokens r where r.token = $2", id, refresh).
		Scan(&valid, &ownerMatch)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to check the owner of the refresh token"})
		return
	}

	if !ownerMatch {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error unathorized"})
		return
	}

	//In case of stolen refresh token
	if !valid {
		_, err = conn.Exec(context.Background(), "update r_tokens set valid = false where user_id = $1", id)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to invalidate the token"})
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error stolen refresh token"})
		return
	}

	_, err = conn.Exec(context.Background(), "update r_tokens set valid = false where user_id = $1", id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to invalidate the refresh token"})
		return
	}

	newRefresh := ""
	err = conn.QueryRow(context.Background(), "insert into r_tokens (user_id, valid) values ($1, $2) returning token", id, true).Scan(&newRefresh)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to invalidate the refresh token"})
		return
	}

	c.SetCookie("refresh", newRefresh, int((5 * 24 * time.Hour).Seconds()), "/", Domain, Secure, true)
	c.JSON(http.StatusOK, nil)
}
