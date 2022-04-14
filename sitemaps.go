package sitemaps

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/xml"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	ResultChan         chan Entity
	JobsCount          int
	Timeout            int
	InsecureSkipVerify bool
	Async              bool
	wg                 sync.WaitGroup
)

type Sitemapindex struct {
	Entites Entity `xml:"sitemap"`
	Status  int
	Url     string
}
type Entity struct {
	Loc        string  `xml:"loc"`
	Changefreq string  `xml:"changefreq,omitempty"`
	Priority   float64 `xml:"priority,omitempty"`
	Images     []Image `xml:"image,omitempty"`
	Videos     []Video `xml:"video,omitempty"`
	SitemapURL string
}

type Image struct {
	Loc     string `xml:"loc"`
	Title   string `xml:"title"`
	Caption string `xml:"caption"`
	Geo     string `xml:"geo_location"`
	License string `xml:"license"`
}

type Video struct {
	ThumbnailLoc          string `xml:"thumbnail_loc"`
	Title                 string `xml:"title"`
	Description           string `xml:"description"`
	Content_loc           string `xml:"content_loc"`
	Player_loc            string `xml:"player_loc"`
	Duration              string `xml:"duration"`
	Expiration_date       string `xml:"expiration_date"`
	Rating                string `xml:"rating"`
	View_count            string `xml:"view_count"`
	Publication_date      string `xml:"publication_date"`
	Family_friendly       string `xml:"family_friendly"`
	Requires_subscription string `xml:"requires_subscription"`
	Live                  string `xml:"live"`
}

type EntityCallback func(e Entity)

func decodeXml(resp *http.Response, callback EntityCallback, entityType string) int {

	var decoder *xml.Decoder
	if !resp.Uncompressed &&
		resp.Header["Content-Type"][0] == "application/gzip" || resp.Header["Content-Type"][0] == "application/x-gzip" {
		//If gzip compression is returned, it needs to be decompressed
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			panic(err)
		}
		decoder = xml.NewDecoder(reader)
		defer reader.Close()

		if err != nil {
			panic(err)
		}

	} else {
		decoder = xml.NewDecoder(resp.Body)
	}

	var counter int
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:

			element := Entity{}

			if entityType == "index" {
				if se.Name.Local == "sitemap" {
					decoder.DecodeElement(&element, &se)
				}
			} else if entityType == "sitemap" {
				if se.Name.Local == "url" {
					decoder.DecodeElement(&element, &se)
				}
			}
			if element.Loc != "" {
				element.SitemapURL = resp.Request.URL.String()
				callback(element)
				counter++
			}
		}
	}

	return counter
}

func makeRequest(url string, callback EntityCallback, entityType string) (int, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: InsecureSkipVerify},
	}
	client := &http.Client{Transport: tr}

	ctx, cncl := context.WithTimeout(context.Background(), time.Second*time.Duration(Timeout))
	defer cncl()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Panic(url, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return 0, err
	}

	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	total := decodeXml(resp, callback, entityType)
	return total, nil
}

func Parse(url string, timeout int, sslCheck bool, scrapeType string, callback EntityCallback) (int, error) {
	InsecureSkipVerify = sslCheck
	Timeout = timeout

	total, err := makeRequest(url, callback, scrapeType)

	return total, err
}

func makeScrapeSlice(e []Entity) []string {
	var slice []string
	for _, v := range e {
		slice = append(slice, v.Loc)
	}
	return slice
}

func worker(id int, jobs <-chan string, timeout int, callback EntityCallback) {
	for url := range jobs {
		_, err := Parse(url, timeout, InsecureSkipVerify, "sitemap", callback)
		if err != nil {
			log.Println(err)
		}
	}
}

func GetUrls(e []Entity, workers int, timeout int, callback EntityCallback) {
	urls := makeScrapeSlice(e)
	JobsCount := len(urls)
	jobs := make(chan string, JobsCount)

	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go func(id int, jobs <-chan string, timeout int, callback EntityCallback) {
			worker(id, jobs, timeout, callback)

			defer wg.Done()
		}(w, jobs, timeout, callback)

	}

	for j := 0; j < len(urls); j++ {
		jobs <- urls[j]
	}

	close(jobs)
	wg.Wait()
}
