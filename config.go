package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// Config struct
type Config struct {
	Paths    Paths      `yaml:"paths"`
	Defaults Defaults   `yaml:"defaults"`
	Author   PostAuthor `yaml:"author"`
}

type Paths struct {
	InputDirectory  string `yaml:"input"`
	OutputDirectory string `yaml:"output"`
}

type Defaults struct {
	DefaultRSSTags       []string `yaml:"rssTags"`
	DefaultRSSCategories []string `yaml:"rssCategories"`
}

type PostAuthor struct {
	Name    *string `yaml:"name"`
	Email   *string `yaml:"email"`
	Website *string `yaml:"website"`
	Image   *string `yaml:"image"`
}

func getConfig() Config {
	if !FileExists("conf/config.yaml") {
		log.Fatal("ERROR: valid conf/config.yaml is required")
	}

	yamlFile, err := ioutil.ReadFile("conf/config.yaml")

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}
	return config
}
