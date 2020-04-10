package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bmaupin/go-htmlutil"
	"golang.org/x/net/html"

	"github.com/bmaupin/go-epub"
	"github.com/fatih/color"
)

func main() {
	pageNode, _ := html.Parse(strings.NewReader(get("https://wanderinginn.com/table-of-contents/")))
	chapters := htmlutil.GetAllHtmlNodes(htmlutil.GetFirstHtmlNode(pageNode, "div", "class", "entry-content"), "a", "", "")

	color.Set(color.FgCyan)
	chdir("wandering_inn/cache")

	fmt.Println("getting", len(chapters), "chapters")
	done := make(chan bool)
	for i := 0; i < len(chapters); i++ {
		go save(chapters[i].Attr[0].Val, done)
	}
	new := 0
	for i := 0; i < len(chapters); i++ {
		if <-done {
			new++
		}
		fmt.Print("\r" + strconv.Itoa(i+1) + "/" + strconv.Itoa(len(chapters)) + " (" + strconv.Itoa(new) + " not from cache)")
	}
	fmt.Print("\n")

	color.Set(color.FgYellow)

	fmt.Println("building epub from", len(chapters), "chapters")
	book := epub.NewEpub("The Wandering Inn")
	book.SetAuthor("pirateaba")
	for i := 0; i < len(chapters); i++ {
		data, err := ioutil.ReadFile(name(chapters[i].Attr[0].Val) + ".html")
		if err != nil {
			log.Fatal(err)
		}

		pageNode, _ := html.Parse(strings.NewReader(string(data)))
		chapter := htmlutil.GetFirstHtmlNode(pageNode, "article", "", "")
		contentNode := htmlutil.GetFirstHtmlNode(chapter, "div", "class", "entry-content")

		contentNode.RemoveChild(contentNode.LastChild) //remove pesky prev/next buttons
		chapter.RemoveChild(chapter.LastChild)         //remove tags (why are they even on the site lol)

		chapterText, _ := htmlutil.HtmlNodeToString(chapter)
		title := htmlutil.GetFirstHtmlNode(chapter, "h1", "", "").FirstChild.Data

		book.AddSection(chapterText, title, name(chapters[i].Attr[0].Val), "")
		fmt.Print("\r" + strconv.Itoa(i+1) + "/" + strconv.Itoa(len(chapters)) + ": added " + title + "                                ")
	}
	fmt.Print("\r413/413: finished                 ")
	os.Chdir("..")
	book.Write("thewanderinginn.epub")
	color.Set(color.FgHiGreen)
	fmt.Println("\nepub written to ./wandering_inn/thewanderinginn.epub")
}

func save(url string, done chan bool) {
	if _, err := os.Stat(name(url) + ".html"); err != nil {
		contents := get(url)
		file, err := os.Create(name(url) + ".html")
		if err != nil {
			log.Fatal(err)
		}
		file.WriteString(contents)
		file.Close()
		done <- true
	} else {
		done <- false
	}
}

func get(url string) string {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	return string(body)
}

func chdir(path string) {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		log.Fatal(err)
	}
	os.Chdir(path)
	fmt.Println("moved to " + path)
}

func name(url string) string { //get the last part of the link e.g. tableofcontents
	l := strings.Split(url, "/")
	for i := len(l) - 1; i >= 0; i-- {
		if l[i] != "" {
			return l[i]
		}
	}
	log.Fatal("couldn't name " + url)
	return "" //idk
}
