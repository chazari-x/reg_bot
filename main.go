package main

import (
	"fmt"
	"log"

	"bscscan_login/bot"
	"bscscan_login/captcha"
	"bscscan_login/config"
	"bscscan_login/connDB"
	"bscscan_login/connSelenium"
	"bscscan_login/loginEmail"
)

func main() {
	if err := StartBot(); err != nil {
		log.Print(err)
	}
}

func StartBot() error {
	c, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("connect to selenium err: %s", err)
	}

	s, wd, err := connSelenium.GetController()
	if err != nil {
		return fmt.Errorf("connect to selenium err: %s", err)
	}

	defer func() {
		_ = wd.Quit()
	}()

	d, db, err := connDB.GetController(c)
	if err != nil {
		return fmt.Errorf("connect to db err: %s", err)
	}

	defer func() {
		_ = db.Close()
	}()

	e, err := loginEmail.GetController(&c)
	if err != nil {
		return fmt.Errorf("get email controller err: %s", err)
	}

	a := captcha.GetController(&c)

	b := bot.GetController(s, e, a, d)

	// TODO: for 1000 this:
	//for i := 0; i < 10; i++ {
	if err = b.Registration(); err != nil {
		log.Printf("registration err: %s", err)
	}
	//}

	// TODO: login and get api token

	return nil
}
