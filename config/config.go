package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Server struct {
	Port int `json:"port"`
}
type Wx struct {
	Appid     string `json:"appid"`
	Appsecret string `json:"appsecret"`
	SubSucee  string `json:"subSucee"`
	SubFailed string `json:"subFailed"`
}
type MyConfig struct {
	Server Server `json:"server"`
	Wx     Wx     `json:"wx"`
}

func ReadConfig() MyConfig {
	// Open our jsonFile
	jsonFile, err := os.Open("config/config_example.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	// read our opened jsonFile as a byte array.
	byteValue, _ := io.ReadAll(jsonFile)

	// we initialize our Users array
	var config MyConfig

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &config)
	fmt.Println(config)
	return config
}
