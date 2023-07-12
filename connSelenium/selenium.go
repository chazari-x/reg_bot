package connSelenium

import (
	"fmt"

	"github.com/tebeka/selenium"
)

type ConnSelenium interface {
	SendKeysToElement(by, value, keys string) error
	OpenURL(url string) error
}

type Controller struct {
	wd selenium.WebDriver
}

func GetSelenium() (*Controller, selenium.WebDriver, error) {
	caps := selenium.Capabilities{"browserName": "chrome"}

	wd, err := selenium.NewRemote(caps, "http://localhost:4444")
	if err != nil {
		return nil, nil, err
	}

	return &Controller{wd: wd}, wd, nil
}

func (s *Controller) OpenURL(url string) error {
	if err := s.wd.Get(url); err != nil {
		return fmt.Errorf("get url err: %s", err)
	}

	return nil
}

func (s *Controller) SendKeysToElement(by, value, keys string) error {
	element, err := s.wd.FindElement(by, value)
	if err != nil {
		return fmt.Errorf("find element \"%s\" by \"%s\" err: %s", by, value, err)
	}

	if err = element.SendKeys(keys); err != nil {
		return fmt.Errorf("send keys \"%s\" err: %s", keys, err)
	}

	return nil
}
