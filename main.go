package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var baseURL string = "https://www.thesaigontimes.vn"
var db *gorm.DB

// Article Table for save Article information
type Article struct {
	ID          uint
	URL         string
	Title       string
	Author      string
	CreatedDate string
}

func main() {
	fmt.Println("Starting application...")
	var err error

	if len(os.Args[1:]) == 0 {
		log.Fatalln("Empty argument")
	}
	firstURL := os.Args[1]

	// Get db info
	content, err := ioutil.ReadFile(".env")
	if err != nil {
		log.Fatalln("load config failed: ", err)
	}
	var config map[string]string
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatalln("wrong config format: ", err)
	}

	// Connect DB
	connectStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", config["username"], config["password"], config["host"], config["port"], config["dbname"])
	db, err = gorm.Open("mysql", connectStr)
	defer db.Close()

	if err != nil {
		log.Fatalln("open db failed: ", err)
	}

	doc, err := loadDocFromURL(firstURL)
	if err != nil {
		log.Fatal(err)
	}

	mapFunc := func(i int, s *goquery.Selection) string {
		relatedURL, _ := s.Attr("href")
		return relatedURL
	}
	mapURL := doc.Find("#ctl00_cphContent_Article_LienQuan .NOtherTitle").Map(mapFunc)

	var wg sync.WaitGroup
	for _, relatedURL := range mapURL {
		fullURL := baseURL + relatedURL
		wg.Add(1)
		go startService(&wg, fullURL)
	}

	wg.Wait()
	fmt.Println("Done")
}

func loadDocFromURL(fullURL string) (*goquery.Document, error) {
	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, errors.Wrap(err, "Error when get link:")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("status code error: %d %s", resp.StatusCode, resp.Status))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Error when load dom")
	}

	return doc, nil
}

func startService(wg *sync.WaitGroup, fullURL string) {
	defer wg.Done()

	article, err := getInfo(fullURL)
	if err != nil {
		log.Fatal(err)
	}

	err = saveInfo(article)
	if err != nil {
		log.Fatal(err)
	}
}

func getInfo(fullURL string) (Article, error) {
	article := Article{}

	doc, err := loadDocFromURL(fullURL)
	if err != nil {
		return article, errors.Wrap(err, "load dom failed")
	}

	title := doc.Find("#ctl00_cphContent_lblTitleHtml").Text()
	author := doc.Find("#ctl00_cphContent_Lbl_Author").Text()
	createdDate := doc.Find("#ctl00_cphContent_lblCreateDate").Text()

	article = Article{
		URL:         fullURL,
		Title:       title,
		Author:      author,
		CreatedDate: createdDate,
	}

	return article, nil
}

func saveInfo(article Article) error {
	return db.Save(&article).Error
}
