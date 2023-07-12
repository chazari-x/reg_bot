package bot

import (
	"fmt"
	"time"

	"bscscan_login/connSelenium"
	"bscscan_login/loginEmail"
	"github.com/tebeka/selenium"
)

type Controller struct {
	s connSelenium.ConnSelenium
}

func GetController(s connSelenium.ConnSelenium) *Controller {
	return &Controller{s: s}
}

func (c *Controller) Registration() error {
	err := c.s.OpenURL("https://bscscan.com/register")

	err = c.s.SendKeysToElement(selenium.ByID, "ContentPlaceHolder1_txtUserName", "username")
	if err != nil {
		return fmt.Errorf("send keys to element err: %s", err)
	}

	email, err := loginEmail.GetEmail()

	err = c.s.SendKeysToElement(selenium.ByID, "ContentPlaceHolder1_txtEmail", email)
	if err != nil {
		return fmt.Errorf("send keys to element err: %s", err)
	}

	err = c.s.SendKeysToElement(selenium.ByID, "ContentPlaceHolder1_txtPassword", "password")
	if err != nil {
		return fmt.Errorf("send keys to element err: %s", err)
	}

	err = c.s.SendKeysToElement(selenium.ByID, "ContentPlaceHolder1_txtPassword2", "password")
	if err != nil {
		return fmt.Errorf("send keys to element err: %s", err)
	}

	err = c.s.SendKeysToElement(selenium.ByID, "ContentPlaceHolder1_MyCheckBox", selenium.SpaceKey)
	if err != nil {
		return fmt.Errorf("send keys to element err: %s", err)
	}

	// TODO: captcha

	// TODO: loginEmail

	// TODO: send login and password to db

	time.Sleep(time.Second * 5)

	return nil
}
