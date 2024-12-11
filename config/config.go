package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ExcludeDatabases []string            `yaml:"exclude_databases"`
	ExcludeTables    map[string][]string `yaml:"exclude_tables"`
	Keep             int                 `yaml:"keep"`
	Cron             string              `yaml:"cron"`
	RemoteBackup     struct {
		AzureBlobStorage struct {
			Enable bool   `yaml:"enable"`
			SAS    string `yaml:"sas"`
		} `yaml:"azure_blob_storage"`
	} `yaml:"remote_backup"`
}

func ReadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
