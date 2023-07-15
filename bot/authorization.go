package bot

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

var authorizationAverageTime = 5.0
var authNums = 1.0

func (c *Controller) Authorization() error {
	start := time.Now()

	if err := c.s.OpenURL(c.c.Site.URL + "login"); err != nil {
		return fmt.Errorf("open url err: %s", err)
	}

	username, password, err := c.db.GetNullUser()
	if err != nil {
		return fmt.Errorf("get null user err: %s", err)
	}

	if username == "" || password == "" {
		return nil
	}

	steps := []Step{
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_txtUserName",
			keys:  username,
			error: Check{
				by:    selenium.ByID,
				value: "ContentPlaceHolder1_txtUserName-error",
				error: []string{"Username is required", "Please enter your"},
			},
		},
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_txtPassword",
			keys:  password,
			error: Check{
				by:    selenium.ByID,
				value: "ContentPlaceHolder1_txtPassword-error",
				error: []string{"Your password is invalid", "Please enter your"},
			},
		},
	}

	for _, step := range steps {
		if err := c.s.SendKeysToElement(step.by, step.value, step.keys); err != nil {
			text, _ := c.s.GetElementText(selenium.ByXPATH, "//*[@id=\"content\"]/div/div/div/div/div/h1")
			if strings.Contains(text, "Sorry, our servers are currently busy") {
				time.Sleep(time.Second * 5)

				return fmt.Errorf("timeout: %s", text)
			}

			return fmt.Errorf("send keys to element err: %s", err)
		}

		if step.error.by != "" && step.error.value != "" && step.error.error != nil {
			for _, err := range step.error.error {
				if text, _ := c.s.GetElementText(selenium.ByID, step.error.value); strings.Contains(text, err) {
					return fmt.Errorf(text)
				}
			}
		}
	}

	captchaKey, err := c.s.GetCaptchaKey()
	if err != nil {
		if !strings.Contains(err.Error(), "find element err") {
			return fmt.Errorf("get captcha key err: %s", err)
		}
	} else {
		solvedKey, err := c.a.GetCaptchaSolvedKey(captchaKey)
		if err != nil {
			return fmt.Errorf("get captcha solved key err: %s", err)
		}

		if err = c.s.CaptchaSolved(solvedKey); err != nil {
			return fmt.Errorf("captcha solved err: %s", err)
		}
	}

	if err = c.s.SendKeysToElement(selenium.ByID, "ContentPlaceHolder1_btnLogin", selenium.SpaceKey); err != nil {
		return fmt.Errorf("send keys to element err: %s", err)
	}

	var value string
	switch c.c.Site.Name {
	case "bscscan":
		value = "//*[@id=\"form1\"]/div[4]"
	case "etherscan":
		value = "//*[@id=\"ContentPlaceHolder1_divLogin\"]/div[2]"
	}

	text, _ := c.s.GetElementText(selenium.ByXPATH, value)
	if strings.Contains(text, "Invalid login information") ||
		strings.Contains(text, "Please verify your email") {
		if err = c.db.UpdateInvalidUser(username); err != nil {
			return fmt.Errorf("invalid login information: %s %s (the db has not meen update: %s)", username, password, err)
		}

		return fmt.Errorf("invalid login information: %s %s", username, password)
	} else if strings.Contains(text, "Invalid captcha") {
		return fmt.Errorf(text)
	}

	err = c.s.OpenURL(c.c.Site.URL + "myapikey")
	if err != nil {
		return fmt.Errorf("open url err: %s", err)
	}

	steps = []Step{
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_addnew",
			keys:  selenium.EnterKey,
		},
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_txtAppName",
			keys:  username,
		},
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_btnSubmit",
			keys:  selenium.EnterKey,
		},
	}

	for _, step := range steps {
		if err = c.s.SendKeysToElement(step.by, step.value, step.keys); err != nil {
			return fmt.Errorf("send keys to element err: %s", err)
		}
	}

	switch c.c.Site.Name {
	case "bscscan":
		text, _ = c.s.GetElementText(selenium.ByCSSSelector, "div[class=alert]")
		if !strings.Contains(text, "Successfully Created") && text != "" {
			return fmt.Errorf(text)
		}

		value = "//*[@id=\"SVGdataReport1\"]/table/tbody/tr[1]/td[1]/a[1]"
	case "etherscan":
		text, _ = c.s.GetElementText(selenium.ByXPATH, "//*[@id=\"content\"]/div/div/div[2]/div[1]/div[1]/div/span")
		if !strings.Contains(text, "Successfully created") && text != "" {
			return fmt.Errorf(text)
		}

		value = "//*[@id=\"content\"]/div/div/div[2]/div[1]/div[4]/table/tbody/tr/td[3]/a"
	}

	if err = c.s.SendKeysToElement(selenium.ByXPATH, value, selenium.EnterKey); err != nil {
		return fmt.Errorf("send keys to element err: %s", err)
	}

	url, err := c.s.GetURL()
	if err != nil {
		return fmt.Errorf("get url err: %s", err)
	}

	authNums++
	authorizationAverageTime = authorizationAverageTime + time.Since(start).Seconds()

	return c.db.UpdateToken(username, regexp.MustCompile(`[A-Za-z0-9]{10,}`).FindString(url), authorizationAverageTime/authNums, time.Since(start).Seconds())
}
