package connSelenium

import (
	"fmt"

	"github.com/tebeka/selenium"
)

type Selenium interface {
	OpenURL(url string) error
	SendKeysToElement(by, value, keys string) error
	GetCaptchaKey() (string, error)
	CaptchaSolved(key string) error
	GetTitle() (string, error)
	GetElementValue(by, value string) (string, error)
	GetElementText(by, value string) (string, error)
}

type Controller struct {
	wd selenium.WebDriver
}

func GetController() (*Controller, selenium.WebDriver, error) {
	caps := selenium.Capabilities{"browserName": "chrome"}

	wd, err := selenium.NewRemote(caps, "http://localhost:4444")
	if err != nil {
		return nil, nil, err
	}

	return &Controller{wd: wd}, wd, nil
}

func (c *Controller) OpenURL(url string) error {
	if err := c.wd.Get(url); err != nil {
		return fmt.Errorf("get url err: %s", err)
	}

	return nil
}

func (c *Controller) SendKeysToElement(by, value, keys string) error {
	element, err := c.wd.FindElement(by, value)
	if err != nil {
		return fmt.Errorf("find element \"%s\" by \"%s\" err: %s", by, value, err)
	}

	if err = element.SendKeys(keys); err != nil {
		return fmt.Errorf("send keys \"%s\" err: %s", keys, err)
	}

	return nil
}

func (c *Controller) GetCaptchaKey() (string, error) {
	element, err := c.wd.FindElement(selenium.ByCSSSelector, "div[class=g-recaptcha]")
	if err != nil {
		return "", fmt.Errorf("find element err: %c", err)
	}

	attribute, err := element.GetAttribute("data-sitekey")
	if err != nil {
		return "", fmt.Errorf("get attribute err: %s", err)
	}

	return attribute, nil
}

func (c *Controller) CaptchaSolved(key string) error {
	_, err := c.wd.ExecuteScript(fmt.Sprintf(`document.getElementById("g-recaptcha-response").innerHTML="%s";`, key), nil)
	if err != nil {
		return fmt.Errorf("get attribute err: %s", err)
	}

	return nil
}

func (c *Controller) GetTitle() (string, error) {
	return c.wd.Title()
}

func (c *Controller) GetElementValue(by, value string) (string, error) {
	element, err := c.wd.FindElement(by, value)
	if err != nil {
		return "", fmt.Errorf("find element err: %s", err)
	}

	t, err := element.GetAttribute("value")
	if err != nil {
		return "", err
	}

	return t, nil
}

func (c *Controller) GetElementText(by, value string) (string, error) {
	element, err := c.wd.FindElement(by, value)
	if err != nil {
		return "", err
	}

	t, err := element.Text()
	if err != nil {
		return "", err
	}

	return t, nil
}
