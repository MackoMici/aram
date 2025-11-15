package config

import (
	"io"
	"os"

	"github.com/MackoMici/aram/logging"
	"gopkg.in/yaml.v3"
)

type Config struct {
	AramszunetPatterns     []string        `yaml:"aramszunet_patterns"`
	HazszamPatterns        []string        `yaml:"hazszam_patterns"`
	TeruletPatterns        []string        `yaml:"terulet_patterns"`
	VegpontPatterns        []string        `yaml:"vegpont_patterns"`
	CleanPatterns          []string        `yaml:"clean_patterns"`
	KizarPatterns          []string        `yaml:"kizar_patterns"`
	AramszunetReplacements []*Replacements `yaml:"aramszunet_replacements"`
}

func NewConfig(file string) *Config {

	f, err := os.Open(file)
	if err != nil {
		logging.Fatal("Konfig fájl megnyitás", "hiba", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		logging.Fatal("Konfig fájl olvasás", "hiba", err)
	}
	conf := &Config{}
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		logging.Fatal("Konfig fájl yaml", "hiba", err)
	}

	return conf
}
