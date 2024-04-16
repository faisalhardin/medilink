package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server           Server           `yaml:"server"`
	Vault            Vault            `yaml:"vault"`
	GoogleAuthConfig GoogleAuthConfig `yaml:"google_auth_config"`
}

type Server struct {
}

type Vault struct {
	GoogleAuth GoogleAuth `json:"google_auth"`
}

type GoogleAuth struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Key          string `json:"key"`
}

type GoogleAuthConfig struct {
	MaxAge           int    `yaml:"max_age"`
	IsProd           bool   `yaml:"is_prod"`
	CallbackURL      string `yaml:"callback_url"`
	HttpOnly         bool   `yaml:"http_only"`
	CookiePath       string `yaml:"cookie_path"`
	HomepageRedirect string `yaml:"homepage_redirect"`
}

func New(repoName string) (*Config, error) {
	dir, _ := os.Getwd()
	filename := "files/etc/auth-vessel/auth-vessel.development.yaml"

	f, err := os.Open(fmt.Sprintf("%s/%s", dir, filename))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Print("Failed to close file: %s\n", err)
		}

		return
	}()

	var cfg Config
	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func NewVault() (Vault, error) {
	dir, _ := os.Getwd()
	filename := "files/etc/configuration/auth-vessel.development.json"

	f, err := os.Open(fmt.Sprintf("%s/%s", dir, filename))
	if err != nil {
		return Vault{}, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Print("Failed to close file: %s\n", err)
		}

		return
	}()

	byteValue, _ := io.ReadAll(f)

	var vault Vault
	err = json.Unmarshal(byteValue, &vault)
	if err != nil {
		return Vault{}, err
	}
	return vault, nil
}
