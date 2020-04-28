package main

// Author is the person who is attributed to the Post
type Author struct {
	Name    string  `yaml:"name"`
	Website *string `yaml:"website"`
	Email   *string `yaml:"email"`
	Image   *string `yaml:"image"`
}
