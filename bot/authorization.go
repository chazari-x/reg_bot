package bot

import (
	"fmt"
	"strings"

	"github.com/tebeka/selenium"
)

func (c *Controller) Authorization() error {
	err := c.s.OpenURL("https://bscscan.com/register")
	if err != nil {
		return fmt.Errorf("open url err: %s", err)
	}

	username := c.e.GetUsername()
	password := c.e.GetPassword()
	email := c.e.GetEmail()

	steps := []Step{
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_txtUserName",
			text:  username,
			check: Check{
				errorValue: "ContentPlaceHolder1_txtUserName-error",
				errorText:  "Username is invalid.",
			},
		},
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_txtEmail",
			text:  email,
			check: Check{
				errorValue: "ContentPlaceHolder1_txtEmail-error",
				errorText:  "Please enter a valid email address.",
			},
		},
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_txtPassword",
			text:  password,
			check: Check{
				errorValue: "ContentPlaceHolder1_txtPassword-error",
				errorText:  "Your password must be at least 5 characters long.",
			},
		},
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_txtPassword2",
			text:  password,
			check: Check{
				errorValue: "",
				errorText:  "",
			},
		},
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_MyCheckBox",
			text:  selenium.SpaceKey,
			check: Check{
				errorValue: "ctl00$ContentPlaceHolder1$MyCheckBox-error",
				errorText:  "Please accept our Terms and Conditions.",
			},
		},
	}

	for _, step := range steps {
		err = c.s.SendKeysToElement(step.by, step.value, step.text)
		if err != nil {
			return fmt.Errorf("send keys to element err: %s", err)
		}

		if step.check.errorText != "" {
			text, _ := c.s.GetElementText(selenium.ByID, step.check.errorText)
			if strings.Contains(text, step.check.errorValue) {
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

	err = c.s.CaptchaSolved(solvedKey)
	if err != nil {
		return fmt.Errorf("captcha solved err: %s", err)
	}

	err = c.s.SendKeysToElement(selenium.ByID, "ContentPlaceHolder1_btnRegister", selenium.SpaceKey)
	if err != nil {
		return fmt.Errorf("send keys to element err: %s", err)
	}

	text, _ := c.s.GetElementText(selenium.ByCSSSelector, "div[role=alert]")
	if !strings.Contains(text, "Your account registration has been") {
		return fmt.Errorf(text)
	}

	url, err := c.e.GetUrl()
	if err != nil {
		return fmt.Errorf("get url err: %s", err)
	}

	err = c.s.OpenURL(url)
	if err != nil {
		return fmt.Errorf("open url err: %s", err)
	}

	text, _ = c.s.GetElementValue(selenium.ByCSSSelector, "input[type=submit]")
	if !strings.Contains(text, "Click to Login") {
		return fmt.Errorf("invalid account confirmation: %s != %s", "Click to Login", text)
	}

	return c.db.AddUser(username, password)
}
