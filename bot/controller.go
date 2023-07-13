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

// Step of the action execution.
type Step struct {
	by    string // Find element by.
	value string // Search value.
	keys  string // These keys are inserted into the element.
	error Check  // error check.
}

// Check is an element that may contain an error.
type Check struct {
	by    string // Find element by.
	value string // Search value.
	error string // Text error.
}

func GetController(s connSelenium.Selenium, e loginEmail.Email, a captcha.Captcha, db connDB.DB) *Controller {
	return &Controller{s: s, e: e, a: a, db: db}
}
