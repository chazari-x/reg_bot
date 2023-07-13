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
	a := &anticaptcha.Client{APIKey: c.CaptchaToken}

	return &Controller{a: a}
}

func (c *Controller) GetCaptchaSolvedKey(key string) (string, error) {
	key, err := c.a.SendRecaptcha(
		"https://bscscan.com/register", // url that has the recaptcha
		key,                            // the recaptcha key
		time.Second*60*30,
	)
	if err != nil {
		return "", fmt.Errorf("send recapthca err: %s", err)
	}

	return key, nil
}
