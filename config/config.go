package config

import (
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	VegpontPatterns        []string        `yaml:"vegpont_patterns"`
	AramszunetPatterns     []string        `yaml:"aramszunet_patterns"`
	TeruletPatterns        []string        `yaml:"terulet_patterns"`
	AramszunetReplacements []*Replacements `yaml:"aramszunet_replacements"`
	VegpontReplacements    []*Replacements `yaml:"vegpont_replacements"`
}

func NewConfig(file string) *Config {

	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	conf := &Config{}
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		log.Fatal(err)
	}

	//	fmt.Printf("Result: %v\n", conf)

	return conf
}
