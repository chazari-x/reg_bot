package bot

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tebeka/selenium"
)

func (c *Controller) Authorization() error {
	if err := c.s.OpenURL("https://bscscan.com/login"); err != nil {
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
				error: "Username is required",
			},
		},
		{
			by:    selenium.ByID,
			value: "ContentPlaceHolder1_txtPassword",
			keys:  password,
			error: Check{
				by:    selenium.ByID,
				value: "ContentPlaceHolder1_txtPassword-error",
				error: "Your password is invalid",
			},
		},
	}

	for _, step := range steps {
		if err := c.s.SendKeysToElement(step.by, step.value, step.keys); err != nil {
			return fmt.Errorf("send keys to element err: %s", err)
		}

		if step.error.by != "" && step.error.value != "" && step.error.error != "" {
			if keys, _ := c.s.GetElementText(selenium.ByID, step.error.value); strings.Contains(keys, step.error.value) {
				return fmt.Errorf(keys)
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

	text, _ := c.s.GetElementText(selenium.ByXPATH, "//*[@id=\"form1\"]/div[4]")
	if strings.Contains(text, "Invalid login information") ||
		strings.Contains(text, "Please verify your email address first. ") {
		if err = c.db.UpdateInvalidUser(username); err != nil {
			return fmt.Errorf("invalid login information: %s %s (the db has not meen update: %s)", username, password, err)
		}

		return fmt.Errorf("invalid login information: %s %s", username, password)
	}

	err = c.s.OpenURL("https://bscscan.com/myapikey")
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

	//if err = c.s.SendKeysToElement(selenium.ByID, "ContentPlaceHolder1_addnew", selenium.EnterKey); err != nil {
	//	return fmt.Errorf("send keys to element err: %s", err)
	//}
	//
	//time.Sleep(time.Second * 5)
	//
	//if err = c.s.SendKeysToElement(selenium.ByID, "ContentPlaceHolder1_txtAppName", username); err != nil {
	//	return fmt.Errorf("send keys to element err: %s", err)
	//}
	//
	//if err = c.s.SendKeysToElement(selenium.ByID, "ContentPlaceHolder1_btnSubmit", selenium.EnterKey); err != nil {
	//	return fmt.Errorf("send keys to element err: %s", err)
	//}

	text, _ = c.s.GetElementText(selenium.ByCSSSelector, "div[class=alert]")
	if !strings.Contains(text, "Successfully Created") && text != "" {
		return fmt.Errorf(text)
	}

	if err = c.s.SendKeysToElement(selenium.ByXPATH, "//*[@id=\"SVGdataReport1\"]/table/tbody/tr/td[1]/a[2]", selenium.EnterKey); err != nil {
		return fmt.Errorf("send keys to element err: %s", err)
	}

	url, err := c.s.GetURL()
	if err != nil {
		return fmt.Errorf("get url err: %s", err)
	}

	return c.db.UpdateToken(username, regexp.MustCompile(`[A-Za-z0-9]{10,}`).FindString(url))
}
