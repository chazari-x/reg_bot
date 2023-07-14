package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

var registrationAverageTime = 20.0
var regNums = 1.0

func (c *Controller) Registration() error {
	start := time.Now()

	if err := c.s.OpenURL("https://bscscan.com/register"); err != nil {
		return fmt.Errorf("open url err: %s", err)
	}

	username, password, email := c.e.GetUsername(), c.e.GetPassword(), c.e.GetEmail()

	steps := []Step{
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_txtUserName",
			keys:  username,
			error: Check{
				by:    selenium.ByID,
				value: "ContentPlaceHolder1_txtUserName-error",
				error: "Username is invalid.",
			},
		},
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_txtEmail",
			keys:  email,
			error: Check{
				by:    selenium.ByID,
				value: "ContentPlaceHolder1_txtEmail-error",
				error: "Please enter a valid email address.",
			},
		},
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_txtPassword",
			keys:  password,
			error: Check{
				by:    selenium.ByID,
				value: "ContentPlaceHolder1_txtPassword-error",
				error: "Your password must be at least 5 characters long.",
			},
		},
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_txtPassword2",
			keys:  password,
			error: Check{
				by:    "",
				value: "",
				error: "",
			},
		},
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_MyCheckBox",
			keys:  selenium.SpaceKey,
			error: Check{
				by:    selenium.ByID,
				value: "ctl00$ContentPlaceHolder1$MyCheckBox-error",
				error: "Please accept our Terms and Conditions.",
			},
		},
	}

	for _, step := range steps {
		if err := c.s.SendKeysToElement(step.by, step.value, step.keys); err != nil {
			return fmt.Errorf("send keys to element err: %s", err)
		}

		if step.error.by != "" && step.error.value != "" && step.error.error != "" {
			text, _ := c.s.GetElementText(selenium.ByID, step.error.error)
			if strings.Contains(text, step.error.value) {
				return fmt.Errorf(text)
			}
		}
	}

	captchaKey, err := c.s.GetCaptchaKey()
	if err != nil {
		return fmt.Errorf("get captcha key err: %s", err)
	}

	solvedKey, err := c.a.GetCaptchaSolvedKey(captchaKey)
	if err != nil {
		return fmt.Errorf("get captcha solved key err: %s", err)
	}

	if err = c.s.CaptchaSolved(solvedKey); err != nil {
		return fmt.Errorf("captcha solved err: %s", err)
	}

	if err = c.s.SendKeysToElement(selenium.ByID, "ContentPlaceHolder1_btnRegister", selenium.SpaceKey); err != nil {
		return fmt.Errorf("send keys to element err: %s", err)
	}

	text, _ := c.s.GetElementText(selenium.ByXPATH, "//*[@id=\"ctl00\"]/div[4]")
	if !strings.Contains(text, "Your account registration has been") {
		return fmt.Errorf(text)
	}

	url, err := c.e.GetUrl()
	if err != nil {
		return fmt.Errorf("get url err: %s", err)
	}

	if err = c.s.OpenURL(url); err != nil {
		return fmt.Errorf("open url err: %s", err)
	}

	text, _ = c.s.GetElementValue(selenium.ByCSSSelector, "input[type=submit]")
	if !strings.Contains(text, "Click to Login") {
		return fmt.Errorf("invalid account confirmation: %s != %s", "Click to Login", text)
	}

	regNums++
	registrationAverageTime = registrationAverageTime + time.Since(start).Seconds()

	return c.db.AddUser(username, password, registrationAverageTime/regNums)
}
