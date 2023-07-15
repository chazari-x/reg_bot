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
	AddUser(username, password string, t, w float64) error
	GetNullUser() (string, string, error)
	GetAllUsers() ([]Users, error)
	UpdateToken(username, token string, t, w float64) error
	UpdateInvalidUser(username string) error
}

type Controller struct {
	db *sql.DB
	t  Table
}

type Users struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token,omitempty"`
}

type Tables struct {
	Bscscan   Table
	Etherscan Table
}

type Table struct {
	createTable                string
	selectNumberOfAllUsers     string
	selectNumberOfNullUsers    string
	selectNumberOfInvalidUsers string
	selectAllUsers             string
	selectNullUser             string
	insertUser                 string
	updateToken                string
	updateInvalidUser          string
}

//goland:noinspection ALL
const (
	cTable = `CREATE TABLE IF NOT EXISTS %s (
						id 			SERIAL 	PRIMARY KEY NOT NULL, 
						username 	VARCHAR UNIQUE 		NOT NULL,
						password 	VARCHAR 			NOT NULL, 
						token 		VARCHAR 			NULL)`

	sNumberOfAllUsers     = `SELECT COUNT(*) FROM %s`
	sNumberOfNullUsers    = `SELECT COUNT(*) FROM %s WHERE token IS NULL`
	sNumverOfInvalidUsers = `SELECT COUNT(*) FROM %s WHERE token = 'INVALID'`

	sAllUsers = `SELECT username, password, COALESCE(token, '-') FROM %s`
	sNullUser = `SELECT username, password FROM %s WHERE token IS NULL LIMIT(1)`

	iUser = `INSERT INTO %s (username, password) VALUES ($1, $2) RETURNING id`

	uToken       = `UPDATE %s SET token = $2 WHERE username = $1`
	uInvalidUser = `UPDATE %s SET token = 'INVALID' WHERE username = $1`
)

func GetController(conf *config.Config) (*Controller, *sql.DB, error) {
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

	var t = Table{
		createTable: fmt.Sprintf(cTable, conf.Site.Name),

		selectNumberOfAllUsers:     fmt.Sprintf(sNumberOfAllUsers, conf.Site.Name),
		selectNumberOfNullUsers:    fmt.Sprintf(sNumberOfNullUsers, conf.Site.Name),
		selectNumberOfInvalidUsers: fmt.Sprintf(sNumverOfInvalidUsers, conf.Site.Name),

		selectAllUsers: fmt.Sprintf(sAllUsers, conf.Site.Name),
		selectNullUser: fmt.Sprintf(sNullUser, conf.Site.Name),

		insertUser: fmt.Sprintf(iUser, conf.Site.Name),

		updateToken:       fmt.Sprintf(uToken, conf.Site.Name),
		updateInvalidUser: fmt.Sprintf(uInvalidUser, conf.Site.Name),
	}

	c := &Controller{db: db, t: t}

	if _, err = db.Exec(c.t.createTable); err != nil {
		return nil, nil, fmt.Errorf("create table err: %s", err)
	}

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
	if err := c.db.QueryRowContext(ctx, c.t.selectNumberOfAllUsers).Scan(&users); err != nil {
		return 0, err
	}

	return users, nil
}

func (c *Controller) GetNumberOfNullUsers() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var users int
	if err := c.db.QueryRowContext(ctx, c.t.selectNumberOfNullUsers).Scan(&users); err != nil {
		return 0, err
	}

	return users, nil
}

func (c *Controller) GetNumberOfInvalidUsers() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var users int
	if err := c.db.QueryRowContext(ctx, c.t.selectNumberOfInvalidUsers).Scan(&users); err != nil {
		return 0, err
	}

	return users, nil
}

func (c *Controller) AddUser(username, password string, t, w float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var i int
	if err := c.db.QueryRowContext(ctx, c.t.insertUser, username, password).Scan(&i); err != nil {
		if err != nil {
			return err
		}
	}

	log.Printf("%d user add: %s %s. Execution time: %f sec, avg: %f sec", i, username, password, w, t)
	return nil
}

func (c *Controller) GetAllUsers() ([]Users, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var users []Users
	rows, err := c.db.QueryContext(ctx, c.t.selectAllUsers)
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
	if err := c.db.QueryRowContext(ctx, c.t.selectNullUser).Scan(&user.Username, &user.Password); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", "", err
		}
	}

	return user.Username, user.Password, nil
}

func (c *Controller) UpdateToken(username, token string, t, w float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := c.db.ExecContext(ctx, c.t.updateToken, username, token); err != nil {
		if err != nil {
			return err
		}
	}

	log.Printf("update token: %s %s. Execution time: %f sec, avg: %f sec", username, token, w, t)

	return nil
}

func (c *Controller) UpdateInvalidUser(username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := c.db.ExecContext(ctx, c.t.updateInvalidUser, username); err != nil {
		if err != nil {
			return err
		}
	}

	log.Printf("update invalid user: %s", username)

	return nil
}
