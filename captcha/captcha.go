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
	c *config.Config
}

func GetController(c *config.Config) *Controller {
	return &Controller{a: &anticaptcha.Client{APIKey: c.CaptchaToken}, c: c}
}

func (c *Controller) GetCaptchaSolvedKey(captchaKey string) (string, error) {
	solvedKey, err := c.a.SendRecaptcha(
		c.c.Site.URL+"register", // url that has the recaptcha
		captchaKey,              // the recaptcha key
		time.Second*60*30,
	)
	if err != nil {
		return "", fmt.Errorf("send recapthca err: %s", err)
	}

	return solvedKey, nil
}
