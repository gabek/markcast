package main

// Post is a Single markdown file generated
type Post struct {
	Title       string  `yaml:"title"`
	Description *string `yaml:"description"`
	Date        string  `yaml:"date"`
	Type        string  `yaml:"type"`
	Enclosure   *string `yaml:"enclosure"`
	Author      *Author `yaml:"author"`
	Content     string  `yaml:"-"`

	Tags       []string `yaml:"tags"`
	Categories []string `yaml:"categories"`

	Slug  string `yaml:"slug"`
	Draft bool   `yaml:"draft"`

	// Optional images used by theme
	Images              []string `yaml:"images"`
	CardThumbImage      *string  `yaml:"cardthumbimage"`
	CardHeaderImage     *string  `yaml:"cardheaderimage"`
	CardBackgroundColor *string  `yaml:"cardbackground"`
}
