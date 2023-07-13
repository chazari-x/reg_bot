package bot

import (
	"bscscan_login/captcha"
	"bscscan_login/connDB"
	"bscscan_login/connSelenium"
	"bscscan_login/loginEmail"
)

type Bot interface {
	Registration() error
	Authorization() error
}

type Controller struct {
	s  connSelenium.Selenium
	e  loginEmail.Email
	a  captcha.Captcha
	db connDB.DB
}

type Step struct {
	by    string
	value string
	text  string
	check Check
}

type Check struct {
	errorValue string
	errorText  string
}

func GetController(s connSelenium.Selenium, e loginEmail.Email, a captcha.Captcha, db connDB.DB) *Controller {
	return &Controller{s: s, e: e, a: a, db: db}
}
