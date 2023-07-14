package main

import (
	"fmt"
	"log"
	"time"

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

	var b bot.Bot
	b = bot.GetController(s, e, a, d)

	allUsers, err := d.GetNumberOfAllUsers()
	if err != nil {
		return fmt.Errorf("get number of all users before err: %s", err)
	}

	invalidUsers, err := d.GetNumberOfInvalidUsers()
	if err != nil {
		return fmt.Errorf("get number of invalid users before err: %s", err)
	}

	var numsErr int
	for i := 0; i < 1001-allUsers+invalidUsers && numsErr < 15; i++ {
		if err = b.Registration(); err != nil {
			log.Printf("registration err: %s", err)
			numsErr++
			i--
		} else {
			numsErr = 0
		}
	}

	allUsers, err = d.GetNumberOfAllUsers()
	if err != nil {
		return fmt.Errorf("get number of all users after err: %s", err)
	}

	log.Printf("Зарегистрировано пользователей: %d\n", allUsers)

	nullUsers, err := d.GetNumberOfNullUsers()
	if err != nil {
		return fmt.Errorf("get number of null users err: %s", err)
	}

	numsErr = 0
	for i := 0; i < nullUsers && numsErr < 15; i++ {
		if err = b.Authorization(); err != nil {
			log.Printf("authorization err: %s", err)
			numsErr++
			i--
		} else {
			numsErr = 0
		}
	}

	nullUsers, err = d.GetNumberOfNullUsers()
	if err != nil {
		return fmt.Errorf("get number of null users err: %s", err)
	}

	invalidUsers, err = d.GetNumberOfInvalidUsers()
	if err != nil {
		return fmt.Errorf("get number of invalid users before err: %s", err)
	}

	log.Printf("Пользователей без токена: %d\n", nullUsers)

	log.Printf("Невалидных пользователей: %d\n", invalidUsers)

	allUsers, err = d.GetNumberOfAllUsers()
	if err != nil {
		return fmt.Errorf("get number of all users before err: %s", err)
	}

	log.Printf("валидных пользователей с токеном: %d", allUsers-invalidUsers-nullUsers)

	time.Sleep(time.Second * 15)

	return nil
}
