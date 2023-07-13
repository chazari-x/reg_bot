package connDB

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"bscscan_login/config"
	_ "github.com/lib/pq"
	"golang.org/x/net/context"
)

type DB interface {
	AddUser(username, password string) error
	GetUser() (string, string, error)
	GetAllUsers() ([]Users, error)
}

type Controller struct {
	db *sql.DB
}

type Users struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token,omitempty"`
}

//goland:noinspection ALL
var (
	createTable = `CREATE TABLE IF NOT EXISTS users (
						id 			SERIAL 	PRIMARY KEY NOT NULL, 
						username 	VARCHAR UNIQUE 		NOT NULL,
						password 	VARCHAR 			NOT NULL, 
						token 		VARCHAR 			NULL)`

	selectAllUsers = `SELECT username, password, COALESCE(token, '-') FROM users`
	selectNullUser = `SELECT username, password FROM users WHERE token = null LIMIT(1)`
	insertUser     = `INSERT INTO users (username, password) VALUES ($1, $2)`
	updateToken    = `UPDATE users SET token = $2 WHERE username = $1`
)

func GetController(conf config.Config) (*Controller, *sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		conf.DB.Host, conf.DB.Port, conf.DB.User, conf.DB.Pass, conf.DB.Name)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, nil, fmt.Errorf("open db err: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, nil, fmt.Errorf("ping db err: %s", err)
	}

	_, err = db.Exec(createTable)
	if err != nil {
		return nil, nil, fmt.Errorf("create table err: %s", err)
	}

	c := &Controller{db: db}

	users, err := c.GetAllUsers()

	for _, k := range users {
		log.Printf("user in db: %s %s %s", k.Username, k.Password, k.Token)
	}

	return c, db, nil
}

func (c *Controller) AddUser(username, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := c.db.ExecContext(ctx, insertUser, username, password)
	if err != nil {
		if err != nil {
			return err
		}
	}

	log.Printf("user add: %s %s", username, password)

	return nil
}

func (c *Controller) GetAllUsers() ([]Users, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var users []Users

	rows, err := c.db.QueryContext(ctx, selectAllUsers)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var user Users
		err = rows.Scan(&user.Username, &user.Password, &user.Token)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (c *Controller) GetUser() (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var user Users

	err := c.db.QueryRowContext(ctx, selectNullUser).Scan(&user.Username, &user.Password)
	if err != nil {
		return "", "", err
	}

	return user.Username, user.Password, nil
}

func (c *Controller) UpdateToken(username, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := c.db.ExecContext(ctx, updateToken, username, token)
	if err != nil {
		if err != nil {
			return err
		}
	}

	return nil
}
