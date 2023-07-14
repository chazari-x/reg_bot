package connDB

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"bscscan_login/config"
	_ "github.com/lib/pq"
	"golang.org/x/net/context"
)

type DB interface {
	GetNumberOfAllUsers() (int, error)
	GetNumberOfNullUsers() (int, error)
	GetNumberOfInvalidUsers() (int, error)
	AddUser(username, password string, t float64) error
	GetNullUser() (string, string, error)
	GetAllUsers() ([]Users, error)
	UpdateToken(username, token string, t float64) error
	UpdateInvalidUser(username string) error
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

	selectNumberOfAllUsers     = `SELECT COUNT(*) FROM users`
	selectNumberOfNullUsers    = `SELECT COUNT(*) FROM users WHERE token IS NULL`
	selectNumverOfInvalidUsers = `SELECT COUNT(*) FROM users WHERE token = 'INVALID'`

	selectAllUsers = `SELECT username, password, COALESCE(token, '-') FROM users`
	selectNullUser = `SELECT username, password FROM users WHERE token IS NULL LIMIT(1)`

	insertUser = `INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id`

	updateToken       = `UPDATE users SET token = $2 WHERE username = $1`
	updateInvalidUser = `UPDATE users SET token = 'INVALID' WHERE username = $1`
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

	if _, err = db.Exec(createTable); err != nil {
		return nil, nil, fmt.Errorf("create table err: %s", err)
	}

	c := &Controller{db: db}

	users, err := c.GetAllUsers()
	for _, k := range users {
		log.Printf("user in db: %s %s %s", k.Username, k.Password, k.Token)
	}

	return c, db, nil
}

func (c *Controller) GetNumberOfAllUsers() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var users int
	if err := c.db.QueryRowContext(ctx, selectNumberOfAllUsers).Scan(&users); err != nil {
		return 0, err
	}

	return users, nil
}

func (c *Controller) GetNumberOfNullUsers() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var users int
	if err := c.db.QueryRowContext(ctx, selectNumberOfNullUsers).Scan(&users); err != nil {
		return 0, err
	}

	return users, nil
}

func (c *Controller) GetNumberOfInvalidUsers() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var users int
	if err := c.db.QueryRowContext(ctx, selectNumverOfInvalidUsers).Scan(&users); err != nil {
		return 0, err
	}

	return users, nil
}

func (c *Controller) AddUser(username, password string, t float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var i int
	if err := c.db.QueryRowContext(ctx, insertUser, username, password).Scan(&i); err != nil {
		if err != nil {
			return err
		}
	}

	log.Printf("%d user add: %s %s. Average time: %f", i, username, password, t)
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
		if err = rows.Scan(&user.Username, &user.Password, &user.Token); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (c *Controller) GetNullUser() (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var user Users
	if err := c.db.QueryRowContext(ctx, selectNullUser).Scan(&user.Username, &user.Password); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", "", err
		}
	}

	return user.Username, user.Password, nil
}

func (c *Controller) UpdateToken(username, token string, t float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := c.db.ExecContext(ctx, updateToken, username, token); err != nil {
		if err != nil {
			return err
		}
	}

	log.Printf("update token: %s %s. Average time: %f", username, token, t)

	return nil
}

func (c *Controller) UpdateInvalidUser(username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := c.db.ExecContext(ctx, updateInvalidUser, username); err != nil {
		if err != nil {
			return err
		}
	}

	log.Printf("update invalid user: %s", username)

	return nil
}
