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

	if err := c.s.OpenURL(c.c.Site.URL + "register"); err != nil {
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
				error: []string{"Username is invalid.", "Please enter at least 5 characters."},
			},
		},
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_txtEmail",
			keys:  email,
			error: Check{
				by:    selenium.ByID,
				value: "ContentPlaceHolder1_txtEmail-error",
				error: []string{"Please enter a valid email address."},
			},
		},
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_txtConfirmEmail",
			keys:  email,
			error: Check{
				by:    "",
				value: "",
				error: nil,
			},
		},
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_txtPassword",
			keys:  password,
			error: Check{
				by:    selenium.ByID,
				value: "ContentPlaceHolder1_txtPassword-error",
				error: []string{"Your password must be at least"},
			},
		},
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_txtPassword2",
			keys:  password,
			error: Check{
				by:    "",
				value: "",
				error: nil,
			},
		},
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_MyCheckBox",
			keys:  selenium.SpaceKey,
			error: Check{
				by:    selenium.ByID,
				value: "ctl00$ContentPlaceHolder1$MyCheckBox-error",
				error: []string{"Please accept our Terms and Conditions."},
			},
		},
	}

	for _, step := range steps {
		if err := c.s.SendKeysToElement(step.by, step.value, step.keys); err != nil {
			if step.value != "ContentPlaceHolder1_txtConfirmEmail" {
				return fmt.Errorf("send keys to element err: %s", err)
			}
		}

		if step.error.by != "" && step.error.value != "" && step.error.error != nil {
			text, _ := c.s.GetElementText(selenium.ByID, step.error.value)
			for _, err := range step.error.error {
				if strings.Contains(text, err) {
					return fmt.Errorf(text)
				}
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

	var value string
	switch c.c.Site.Name {
	case "bscscan":
		value = "//*[@id=\"ctl00\"]/div[4]"
	case "etherscan":
		value = "//*[@id=\"ctl00\"]/div[3]"
	}

	text, _ := c.s.GetElementText(selenium.ByXPATH, value)
	if !strings.Contains(text, "Your account registration has been") {
		return fmt.Errorf(text)
	}

	url, err := c.e.GetUrl()
	if err != nil {
		return fmt.Errorf("get verify url err: %s", err)
	}

	if url == "" {
		return fmt.Errorf("verify url is nil")
	}

	if err = c.s.OpenURL(url); err != nil {
		return fmt.Errorf("open verify url err: %s", err)
	}

	switch c.c.Site.Name {
	case "bscscan":
		text, _ = c.s.GetElementValue(selenium.ByCSSSelector, "input[type=submit]")
		if !strings.Contains(text, "Click to Login") {
			return fmt.Errorf("invalid account confirmation: %s != %s", "Click to Login", text)
		}
	case "etherscan":
		text, _ = c.s.GetElementText(selenium.ByXPATH, "//*[@id=\"form1\"]/div[3]/div/p/strong")
		if !strings.Contains(text, "Congratulations!") {
			return fmt.Errorf("invalid account confirmation: %s != %s", "Congratulations!", text)
		}
	}

	regNums++
	registrationAverageTime = registrationAverageTime + time.Since(start).Seconds()

	return c.db.AddUser(username, password, registrationAverageTime/regNums, time.Since(start).Seconds())
}
