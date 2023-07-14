package captcha

import (
	"fmt"
	"time"

	"bscscan_login/config"
	"github.com/nuveo/anticaptcha"
)

type Captcha interface {
	GetCaptchaSolvedKey(key string) (string, error)
}

type Controller struct {
	a *anticaptcha.Client
}

func GetController(c *config.Config) *Controller {
	return &Controller{a: &anticaptcha.Client{APIKey: c.CaptchaToken}}
}

func (c *Controller) GetCaptchaSolvedKey(captchaKey string) (string, error) {
	//log.Printf("getting the solved key... ")

	solvedKey, err := c.a.SendRecaptcha(
		"https://bscscan.com/register", // url that has the recaptcha
		captchaKey,                     // the recaptcha key
		time.Second*60*30,
	)
	if err != nil {
		return "", fmt.Errorf("send recapthca err: %s", err)
	}

	//log.Printf("solved key received\n")

	return solvedKey, nil
}
