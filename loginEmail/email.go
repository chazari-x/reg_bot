package loginEmail

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"bscscan_login/config"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Email interface {
	GetUsername() string
	GetPassword() string
	GetEmail() string
	GetUrl() (string, error)
}

type Controller struct {
	c   *config.Config
	srv *gmail.Service
}

func GetController(c *config.Config) (*Controller, error) {
	b, err := os.ReadFile("loginEmail/credentials.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %s", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	configFromJson, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %s", err)
	}

	client, err := getClient(configFromJson)
	if err != nil {
		return nil, fmt.Errorf("get client err: %s", err)
	}

	srv, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Gmail client: %s", err)
	}

	return &Controller{c: c, srv: srv}, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) (*http.Client, error) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "loginEmail/token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok, err = getTokenFromWeb(config)
		if err != nil {
			return nil, fmt.Errorf("get token from wev err: %s", err)
		}

		if err = saveToken(tokFile, tok); err != nil {
			return nil, fmt.Errorf("save token err: %s", err)
		}
	}
	return config.Client(context.Background(), tok), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%s\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %s", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %s", err)
	}
	return tok, nil
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("open file err: %s", err)
	}
	defer func() {
		_ = f.Close()
	}()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %s", err)
	}
	defer func() {
		_ = f.Close()
	}()

	return json.NewEncoder(f).Encode(token)
}

func (c *Controller) GetUrl() (string, error) {
	// Поиск писем с определенным заголовком
	query := "subject:\tPlease confirm your email [BscScan.com]"
	call := c.srv.Users.Messages.List("me").Q(query)
	r, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve messages: %s", err)
	}

	if len(r.Messages) == 0 {
		for i := 0; i < 60; i++ {
			// Поиск писем с определенным заголовком
			r, err = call.Do()
			if err != nil {
				return "", fmt.Errorf("unable to retrieve messages: %s", err)
			}

			if len(r.Messages) != 0 {
				break
			}

			time.Sleep(time.Second / 2)
		}
	}

	l := r.Messages[0]

	// Выполнение запрос на получение содержимого письма
	gmailMessageResponse, err := c.srv.Users.Messages.Get("me", l.Id).Format("RAW").Do()
	if err != nil {
		return "", fmt.Errorf("error when getting mail content: %s", err)
	}

	if gmailMessageResponse == nil {
		return "", fmt.Errorf("url not founded")
	}

	decodedData, _ := base64.RawURLEncoding.DecodeString(gmailMessageResponse.Raw)

	return findLink(fmt.Sprintf("%s", decodedData)), nil
}

// Функция для поиска ссылки в тексте
func findLink(text string) string {
	l1 := regexp.MustCompile(`https://BscScan\.com/confirmemail\?email=\S+`)
	l2 := regexp.MustCompile(`\S+&code=\S+`)
	m1 := l1.FindString(text)
	m := strings.ReplaceAll(m1[:len(m1)-1]+l2.FindString(text), "email=3D", "email=")
	return strings.ReplaceAll(m, "code=3D", "code=")
}

var i = rand.Intn(500)

func (c *Controller) GetUsername() string {
	i += rand.Intn(500)
	return strings.Repeat(strconv.FormatInt(int64(i), 32), 4)
}

func (c *Controller) GetPassword() string {
	return strings.Repeat(strconv.FormatInt(int64(i), 32), 6)
}

func (c *Controller) GetEmail() string {
	return fmt.Sprintf("%s+%s%s", c.c.Email.Username, strconv.FormatInt(int64(i), 32), c.c.Email.Domain)
}
