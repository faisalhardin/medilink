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
	JWTConfig        JWTConfig        `yaml:"jwt_config"`
}

type JWTConfig struct {
	DurationInHour int `json:"duration_in_hour"`
}

type Server struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type VaultData struct {
	Data Vault `json:"data"`
}

type Vault struct {
	GoogleAuth    GoogleAuth    `json:"google_auth"`
	DBMaster      DBConfig      `json:"db_master"`
	DBSlave       DBConfig      `json:"db_slave"`
	JWTCredential JWTCredential `json:"jwt_credential"`
}

type JWTCredential struct {
	Secret string `json:"secret"`
}

type DBConfig struct {
	DSN string `json:"dsn"`
}

type DBMaster struct {
	Host       string `json:"host"`
	Port       string `json:"port"`
	Password   string `json:"password"`
	User       string `json:"user"`
	DBName     string `json:"dbname"`
	SSLMode    string `json:"disable"`
	SearchPath string `json:"search_path"`
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
	dir = "D:/gol/medilink/medilink"
	filename := "files/etc/medilink/medilink.development.yaml"

	f, err := os.Open(fmt.Sprintf("%s/%s", dir, filename))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Failed to close file: %s\n", err)
		}
	}()

	var cfg Config
	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func NewVault() (VaultData, error) {
	dir, _ := os.Getwd()
	dir = "D:/gol/medilink/medilink"
	filename := "files/etc/configuration/medilink.development.json"

	f, err := os.Open(fmt.Sprintf("%s/%s", dir, filename))
	if err != nil {
		return VaultData{}, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Failed to close file: %s\n", err)
		}
	}()

	byteValue, _ := io.ReadAll(f)

	var vaultData VaultData
	err = json.Unmarshal(byteValue, &vaultData)
	if err != nil {
		return VaultData{}, err
	}
	return vaultData, nil
}
