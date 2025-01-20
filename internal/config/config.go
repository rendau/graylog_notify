package config

import (
	"log"

	"github.com/caarlos0/env/v9"
	_ "github.com/joho/godotenv/autoload"
	"gopkg.in/yaml.v3"
)

var Conf = struct {
	Debug         bool   `env:"DEBUG" envDefault:"false"`
	HttpPort      string `env:"HTTP_PORT" envDefault:"80"`
	HttpCors      bool   `env:"HTTP_CORS" envDefault:"false"`
	TelegramToken string `env:"TELEGRAM_TOKEN"`
	// TelegramChatId   int64             `env:"TELEGRAM_CHAT_ID"`
	TelegramRulesRaw string            `env:"TELEGRAM_RULES"`
	TelegramRules    map[string]string `env:"-"`
}{}

func init() {
	if err := env.Parse(&Conf); err != nil {
		panic(err)
	}

	Conf.TelegramRules = make(map[string]string)
	if len(Conf.TelegramRulesRaw) > 0 {
		err := yaml.Unmarshal([]byte(Conf.TelegramRulesRaw), &Conf.TelegramRules)
		if err != nil {
			log.Fatal("error parsing conf.rules ", err)
		}
	}
}
