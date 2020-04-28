package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/dhowden/tag"
	colorful "github.com/lucasb-eyer/go-colorful"
	"gopkg.in/yaml.v2"
)

var config = getConfig()

func main() {
	files, err := filepath.Glob(config.Paths.InputDirectory + "/*.mp3")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		post := generatePostForFile(f)

		outputDirectory := config.Paths.OutputDirectory + "/" + post.Slug
		_ = os.Mkdir(outputDirectory, os.ModePerm)

		newMarkdownFile, _ := os.Create(outputDirectory + "/" + post.Slug + ".md")
		defer newMarkdownFile.Close()

		postString := generateOutputStringForPost(post)
		newMarkdownFile.WriteString(postString)
	}
}

func generatePostForFile(filename string) Post {
	file, err := os.Open(filename)
	defer file.Close()

	metadata, err := tag.ReadFrom(file)
	if err != nil {
		log.Println(err)
	}

	title := metadata.Title()
	slug := generateSlugFromTitle(title)
	newFilename := "/mixes/" + slug + "/" + slug + ".mp3"
	outputDirectory := config.Paths.OutputDirectory + "/" + slug
	_ = os.Mkdir(outputDirectory, os.ModePerm)

	filesystemLocationOfAudio := outputDirectory + "/" + slug + ".mp3"
	os.Rename(filename, filesystemLocationOfAudio)
	info, err := os.Stat(filesystemLocationOfAudio)

	post := Post{}
	post.Title = metadata.Title()
	description := metadata.Comment()
	post.Description = &description
	post.Date = string(info.ModTime().Format("2006-01-02"))
	post.Type = "post"
	post.Enclosure = &newFilename
	post.Slug = slug

	author := Author{}
	if config.Author.Name == nil {
		name := metadata.Artist()
		author.Name = name
	} else {
		author.Name = *config.Author.Name
	}

	author.Email = config.Author.Email
	author.Website = config.Author.Website

	post.Author = &author

	// Images
	image := metadata.Picture()
	if image != nil {
		imageFilename := slug + ".png"
		err := ioutil.WriteFile(outputDirectory+"/"+imageFilename, image.Data, 0644)
		if image != nil {
			log.Println(err)
		}
		imagePath := "/mixes/" + slug + "/" + imageFilename

		post.CardHeaderImage = &imagePath
		post.CardThumbImage = &imagePath
		hexColor := colorful.WarmColor().Hex()
		post.CardBackgroundColor = &hexColor
		post.Images = append(post.Categories, imagePath)
	}

	if metadata.Genre() != "" {
		post.Categories = append(post.Categories, metadata.Genre())
		post.Tags = append(post.Tags, metadata.Genre())
	}

	post.Tags = append(post.Tags, config.Defaults.DefaultRSSTags...)
	post.Categories = append(post.Categories, config.Defaults.DefaultRSSCategories...)

	post.Content = metadata.Comment()

	os.Rename(newFilename, outputDirectory+"/"+newFilename)

	return post
}

func generateSlugFromTitle(title string) string {
	var re = regexp.MustCompile("[^a-z0-9]+")
	return strings.Trim(re.ReplaceAllString(strings.ToLower(title), "-"), "-")
}

func generateEmbedHTMLForFilename(filename string) string {
	html := `<audio controls preload="metadata" style=" width:300px;">
	<source src="{{.Filename}}" type="audio/mpeg">
	Your browser does not support the audio element.
</audio>
	`

	type DataValues struct {
		Filename string
	}
	data := DataValues{filename}

	tmpl := template.New("html")
	var err error
	if tmpl, err = tmpl.Parse(html); err != nil {
		fmt.Println(err)
	}

	var output bytes.Buffer
	tmpl.Execute(&output, data)

	return output.String()
}

func generateOutputStringForPost(post Post) string {
	markdown := `---
{{.FrontMatter}}
---
<center><img src="{{.Image}}" width="30%"></center>
<center>{{.PlayerHTML}}</center>

<br>
<h4>{{.Content}}</h4>
`

	type DataValues struct {
		Content     string
		FrontMatter string
		PlayerHTML  string
		Image       string
	}

	frontMatterData, err := yaml.Marshal(&post)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	data := DataValues{}
	data.FrontMatter = string(frontMatterData)
	data.Content = post.Content
	data.PlayerHTML = generateEmbedHTMLForFilename(*post.Enclosure)
	data.Image = *post.CardHeaderImage

	tmpl := template.New("markdown")

	if tmpl, err = tmpl.Parse(markdown); err != nil {
		fmt.Println(err)
	}

	var output bytes.Buffer
	tmpl.Execute(&output, data)

	return output.String()
}
