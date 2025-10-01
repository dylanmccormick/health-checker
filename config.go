package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	IntervalSeconds int      `json:"check_interval_seconds"`
	TimeoutSeconds  int      `json:"timeout_seconds"`
	Urls            []string `json:"Urls"`
}

func GetConfig() (Config, error) {
	var err error

	dat, err := os.ReadFile("./config.json")
	if err != nil {
		fmt.Println(err)
	}

	conf := Config{}

	err = json.Unmarshal(dat, &conf)
	if err != nil {
		return Config{}, err
	}

	return conf, nil
}
