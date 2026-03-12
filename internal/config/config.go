package config

import (
  "os"
  "encoding/json"
  "path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct{
  DBURL  string `json:"db_url"`
  CurrentUserName string `json:"current_user_name"`
}

func getConfigFilePath() (string, error) {
  home, err := os.UserHomeDir()
  if err != nil {
    return "", err
  }
  return filepath.Join(home, configFileName), nil
}

func write(cfg Config) error {
  data, err := json.Marshal(cfg)
  if err != nil {
    return err
  }
  configPath, err := getConfigFilePath()
  if err != nil {
    return err
  }

  err = os.WriteFile(configPath, data, 0644)
  if err != nil {
    return err
  }
  return nil
}

func Read() (Config, error) {
  configPath, err := getConfigFilePath()
  if err != nil {
    return Config{}, err
  }

  data, err := os.ReadFile(configPath)
  if err != nil {
    return Config{}, err
  }
  var cfg Config
  err = json.Unmarshal(data, &cfg)
  if err != nil {
    return Config{}, err
  }

  return cfg, nil
}


func (cfg Config) SetUser(userName string) error{
  cfg.CurrentUserName = userName
  err := write(cfg)
  if err != nil {
    return err
  }
  return nil
}
