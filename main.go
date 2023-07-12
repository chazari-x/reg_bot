package main

import (
	"fmt"
	"log"

	"bscscan_login/bot"
	"bscscan_login/connDB"
	"bscscan_login/connSelenium"
)

func main() {
	log.Print(StartBot())
}

func StartBot() error {
	s, wd, err := connSelenium.GetSelenium()
	if err != nil {
		return fmt.Errorf("connect to selenium err: %s", err)
	}

	defer func() {
		_ = wd.Quit()
	}()

	db, err := connDB.GetDB()
	if err != nil {
		return fmt.Errorf("connect to db err: %s", err)
	}

	defer func() {
		_ = db.Close()
	}()

	b := bot.GetController(s)

	// TODO: for 100 this:
	err = b.Registration()
	if err != nil {
		return fmt.Errorf("registration err err: %s", err)
	}

	// TODO: login and get api token

	return nil
}
