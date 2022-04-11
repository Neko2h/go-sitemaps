package sitemaps

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	ResultChan = make(chan []Entity)
	JobsCount  int
	Timeout    int
)

type Sitemapindex struct {
	Entites []Entity `xml:"sitemap"`
	Status  int
}

type UrlSet struct {
	Entites []Entity `xml:"url"`
	Status  int
}

type Entity struct {
	Loc        string  `xml:"loc"`
	Lastmod    string  `xml:"lastmod,omitempty"`
	ChangeFreq string  `xml:"changefreq,omitempty"`
	Images     []Image `xml:"image"`
	Videos     []Video `xml:"video"`
}

type Image struct {
	Loc     string `xml:"loc"`
	Title   string `xml:"title"`
	Caption string `xml:"image:caption"`
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

func makeRequest(url string) ([]byte, error, int) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	ctx, cncl := context.WithTimeout(context.Background(), time.Second*time.Duration(Timeout))
	defer cncl()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, err, resp.StatusCode
	}

	if err != nil {
		return nil, err, resp.StatusCode
	}
	defer resp.Body.Close()

	var body []byte

	if !resp.Uncompressed && resp.Header["Content-Type"][0] == "application/gzip" {
		//If gzip compression is returned, it needs to be decompressed
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			panic(err)
		}
		defer reader.Close()
		body, err = ioutil.ReadAll(reader)
		if err != nil {
			panic(err)
		}

	} else {
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
	}

	return body, nil, resp.StatusCode
}

func ParseIndex(url string, timeout int) (Sitemapindex, int, error) {
	Timeout = timeout
	body, err, status := makeRequest(url)
	if err != nil {
		return Sitemapindex{}, 0, err
	}

	var res Sitemapindex
	err = xml.Unmarshal(body, &res)
	res.Status = status
	if err != nil {
		return Sitemapindex{}, 0, err
	}
	return res, len(res.Entites), nil

}

func ParseSitemap(url string, timeout int) (UrlSet, error) {
	Timeout = timeout

	var res UrlSet
	body, err, status := makeRequest(url)
	if err != nil {
		return UrlSet{}, err
	}
	err = xml.Unmarshal(body, &res)
	res.Status = status
	if err != nil {
		return UrlSet{}, err
	}
	return res, nil
}

func (s *Sitemapindex) Count() int {
	return len(s.Entites)
}

func (s *UrlSet) Count() int {
	return len(s.Entites)
}

func worker(id int, jobs <-chan string) {
	for j := range jobs {
		urlset, _ := ParseSitemap(j, Timeout)
		ResultChan <- urlset.Entites
	}
}

func (s *Sitemapindex) GetUrlsGreedy(workers int) (int, []Entity) {

	var urls []Entity
	JobsCount = len(s.Entites)
	jobs := make(chan string, JobsCount)

	for w := 1; w <= workers; w++ {
		go worker(w, jobs)
	}

	for j := 0; j < len(s.Entites); j++ {
		jobs <- s.Entites[j].Loc
	}

	close(jobs)

	for a := 1; a <= JobsCount; a++ {
		urls = append(urls, <-ResultChan...)
	}

	close(ResultChan)
	return len(urls), urls
}

func (s *Sitemapindex) GetUrlsLazy(workers int) {

	JobsCount = len(s.Entites)
	jobs := make(chan string, JobsCount)

	for w := 1; w <= workers; w++ {
		go worker(w, jobs)
	}
	for j := 0; j < len(s.Entites); j++ {
		jobs <- s.Entites[j].Loc
	}

	close(jobs)
}
