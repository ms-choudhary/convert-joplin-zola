package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"unicode"

	"gopkg.in/yaml.v2"
)

var (
	JoplinPath        = "/root/shared/ixdev/joplin"
	JoplinWebsitePath = JoplinPath + "/Website"
	ZolaPath          = "/root/shared/ixdev/projects/src/github.com/ms-choudhary/ms-choudhary.github.io"
	ZolaContentPath   = ZolaPath + "/content"
	SkipDirs          = []string{"books", "about"}
	ResourceRe        = regexp.MustCompile("[a-z0-9]+.png")
)

type JoplinHead struct {
	Title   string   `yaml:"title"`
	Updated string   `yaml:"updated"`
	Created string   `yaml:"created"`
	Tags    []string `yaml:"tags"`
}

type ZolaTag struct {
	Tags []string `yaml:"tags"`
}

type ZolaHead struct {
	Title      string  `yaml:"title"`
	Date       string  `yaml:"date"`
	Updated    string  `yaml:"updated"`
	Taxonomies ZolaTag `yaml:"taxonomies"`
}

func slug(in string) string {
	out := []rune{}
	for _, c := range in {
		if unicode.IsLetter(c) || unicode.IsNumber(c) || c == '.' {
			out = append(out, c)
		} else if out[len(out)-1] != '-' {
			out = append(out, '-')
		}
	}
	return string(out)
}

func copyFile(src, dst string) error {
	srcFs, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("could not open src: %v", err)
	}

	defer srcFs.Close()

	dstFs, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("could not create dst: %v", err)
	}
	defer dstFs.Close()

	_, err = io.Copy(dstFs, srcFs)
	if err != nil {
		return fmt.Errorf("failed to copy: %v", err)
	}
	return nil
}

func main() {
	resourcesToCopy := map[string]struct{}{}
	err := fs.WalkDir(os.DirFS(JoplinWebsitePath), ".", func(filepath string, file fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		for _, d := range SkipDirs {
			if strings.HasPrefix(filepath, d) {
				return nil
			}
		}

		if file.IsDir() {
			return nil
		}

		if strings.HasPrefix(file.Name(), "_") {
			return nil
		}

		f, err := os.Open(JoplinWebsitePath + "/" + filepath)
		if err != nil {
			return fmt.Errorf("could not read file: %v", err)
		}

		defer f.Close()

		scanner := bufio.NewScanner(f)
		var header, body string
		var readingHeader bool
		for scanner.Scan() {
			line := scanner.Text()
			if !readingHeader && line == "---" {
				readingHeader = true
			} else if readingHeader && line == "---" {
				readingHeader = false
			} else if readingHeader {
				header = fmt.Sprintf("%s%s\n", header, line)
			} else {
				if resource := ResourceRe.Find([]byte(line)); resource != nil {
					resourcesToCopy[string(resource)] = struct{}{}
					line = strings.ReplaceAll(line, "../../_resources/", "/images/")
				}
				body = fmt.Sprintf("%s%s\n", body, line)
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scanner err: %v", err)
		}

		var joplin JoplinHead
		if err := yaml.Unmarshal([]byte(header), &joplin); err != nil {
			return fmt.Errorf("could not unmarshal yaml: %v", err)
		}

		zola := ZolaHead{
			Title:      joplin.Title,
			Date:       joplin.Created,
			Updated:    joplin.Updated,
			Taxonomies: ZolaTag{Tags: joplin.Tags},
		}

		newheader, err := yaml.Marshal(zola)
		if err != nil {
			return fmt.Errorf("could not marshal yaml: %v", err)
		}

		newFilePath := ZolaContentPath + "/" + path.Dir(filepath) + "/" + slug(strings.ToLower(file.Name()))
		newFileContent := fmt.Sprintf("---\n%s---\n%s", string(newheader), body)

		if err := os.WriteFile(newFilePath, []byte(newFileContent), 0644); err != nil {
			return fmt.Errorf("could not write file: %v", err)
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	for file := range resourcesToCopy {
		srcPath := JoplinPath + "/_resources/" + file
		dstPath := ZolaPath + "/static/images/" + file

		if err := copyFile(srcPath, dstPath); err != nil {
			log.Fatal(err)
		}
	}

	if err := os.RemoveAll(JoplinPath); err != nil {
		log.Fatal(err)
	}
}
