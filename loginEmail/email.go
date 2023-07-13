package loginEmail

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
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
	GetEmail() string
	GetPassword() string
	GetUrl() (string, error)
}

type Controller struct {
	c   *config.Config
	srv *gmail.Service
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
			return nil, err
		}

		err := saveToken(tokFile, tok)
		if err != nil {
			return nil, err
		}
	}
	return config.Client(context.Background(), tok), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %v", err)
	}
	return tok, nil
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
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
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer func() {
		_ = f.Close()
	}()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return err
	}

	return nil
}

func GetController(c *config.Config) (*Controller, error) {
	ctx := context.Background()
	b, err := os.ReadFile("loginEmail/credentials.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	configFromJson, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	client, err := getClient(configFromJson)
	if err != nil {
		return nil, err
	}

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Gmail client: %v", err)
	}

	return &Controller{c: c, srv: srv}, nil
}

func (c *Controller) GetUrl() (string, error) {
	// Поиск писем с определенным заголовком
	query := "subject:\tPlease confirm your email [BscScan.com]"
	call := c.srv.Users.Messages.List("me").Q(query)
	r, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve messages: %v", err)
	}

	if len(r.Messages) == 0 {
		for i := 0; i < 60; i++ {
			// Поиск писем с определенным заголовком
			r, err = call.Do()
			if err != nil {
				return "", fmt.Errorf("unable to retrieve messages: %v", err)
			}

			if len(r.Messages) != 0 {
				break
			}

			time.Sleep(time.Second)
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

	decodedData, err := base64.RawURLEncoding.DecodeString(gmailMessageResponse.Raw)
	if err != nil {
		log.Println("error b64 decoding: ", err)
	}

	return findLink(fmt.Sprintf("%s", decodedData)), nil
}

// Функция для поиска ссылки в тексте
func findLink(text string) string {
	linkRegexp := regexp.MustCompile(`https:\/\/BscScan\.com\/confirmemail\?email=[^\s]+`)
	linkRegexp2 := regexp.MustCompile(`[^\s]+&code=[^\s]+`)
	match := linkRegexp.FindString(text)
	match2 := linkRegexp2.FindString(text)
	m := strings.ReplaceAll(match[:len(match)-1]+match2, "email=3D", "email=")
	return strings.ReplaceAll(m, "code=3D", "code=")
}

var i = rand.Intn(1000)

func (c *Controller) GetUsername() string {
	id := strconv.FormatInt(int64(i), 32)
	return strings.Repeat(id, 5)
}

func (c *Controller) GetEmail() string {
	id := strconv.FormatInt(int64(i), 32)
	i += rand.Intn(1000)

	return fmt.Sprintf("%s+%s%s", c.c.Email.Username, id, c.c.Email.Domain)
}

func (c *Controller) GetPassword() string {
	id := strconv.FormatInt(int64(i), 32)
	return strconv.FormatInt(int64(rand.Intn(1000)), 32) + strings.Repeat(id, 5) + strconv.FormatInt(int64(rand.Intn(1000)), 32)
}
