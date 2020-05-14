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
		// Don't overwrite files that have been previously generated
		if FileExists(outputDirectory + "/" + post.Slug + ".md") {
			log.Println(outputDirectory + " already exists.  Ignoring.")
			continue
		}

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
	info, err := os.Stat(filename)

	filesystemLocationOfAudio := outputDirectory + "/" + slug + ".mp3"
	os.Rename(filename, filesystemLocationOfAudio)

	post := Post{}
	post.Title = metadata.Title()
	description := metadata.Comment()
	post.Description = &description
	post.Date = string(info.ModTime().Format("2006-01-02"))
	post.Enclosure = &newFilename
	post.Slug = slug
	post.Type = "post"

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
	post.Tracklist = generateTracklistForMP3(filename)

	return post
}

func generateSlugFromTitle(title string) string {
	var re = regexp.MustCompile("[^a-z0-9]+")
	return strings.Trim(re.ReplaceAllString(strings.ToLower(title), "-"), "-")
}

func generateTracklistForMP3(filepath string) string {
	cuefile := strings.Replace(filepath, ".mp3", ".cue", -1)
	if !FileExists(cuefile) {
		return ""
	}
	return generateTracklistFromCue(cuefile)
}

func generateEmbedHTMLForFilename(filename string) string {
	html := `<audio controls preload="metadata" style=" width:300px;">
	<source src="{{.Filename}}" type="audio/mpeg">
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

{{.Content}}

<br>
<h4>Tracklist</h4>
{{.Tracklist}}
`

	type DataValues struct {
		Content     string
		Tracklist   string
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
	data.Tracklist = post.Tracklist
	data.PlayerHTML = generateEmbedHTMLForFilename(*post.Enclosure)
	if post.CardHeaderImage != nil {
		data.Image = *post.CardHeaderImage
	}

	tmpl := template.New("markdown")

	if tmpl, err = tmpl.Parse(markdown); err != nil {
		fmt.Println(err)
	}

	var output bytes.Buffer
	tmpl.Execute(&output, data)

	return output.String()
}
