package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

var C Config

type Config struct {
	Email struct {
		Username string `yaml:"username"`
		Domain   string `yaml:"domain"`
	} `yaml:"email"`

	CaptchaToken string `yaml:"captchaToken"`

	DB struct {
		Host string `yaml:"host"` // Хост
		Port string `yaml:"port"` // Порт
		User string `yaml:"user"` // Пользователь
		Pass string `yaml:"pass"` // Пароль
		Name string `yaml:"name"` // Название
	} `yaml:"db"`
}

const configFile = "config/dev.yaml"

func GetConfig() (Config, error) {
	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		return Config{}, errors.New("read config fIle err: " + err.Error())
	}

	if err = yaml.Unmarshal(yamlFile, &C); err != nil {
		return Config{}, errors.New("unmarshal config fIle err: " + err.Error())
	}

	if C.Email.Username == "" || C.Email.Domain == "" {
		return Config{}, errors.New("error config")
	}

	return C, nil
}
