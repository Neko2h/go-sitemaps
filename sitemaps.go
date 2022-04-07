package sitemaps

import (
	"compress/gzip"
	"context"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"time"
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
	Loc        string `xml:"loc"`
	Lastmod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
}

func (s *Sitemapindex) Count() int {
	return len(s.Entites)
}

func (s *UrlSet) Count() int {
	return len(s.Entites)
}

func worker(id int, jobs <-chan string, results chan<- []Entity, timeout int) {
	for j := range jobs {
		urlset, _ := ParseSitemap(j, timeout)
		results <- urlset.Entites
	}
}

func (s *Sitemapindex) GetUrls(workers int, timeout int) (int, []Entity) {

	var urls []Entity
	numJobs := len(s.Entites)
	jobs := make(chan string, numJobs)
	results := make(chan []Entity, numJobs)

	for w := 1; w <= workers; w++ {
		go worker(w, jobs, results, timeout)
	}

	for j := 0; j < len(s.Entites); j++ {
		jobs <- s.Entites[j].Loc
	}

	close(jobs)

	for a := 1; a <= numJobs; a++ {
		urls = append(urls, <-results...)
	}

	close(results)
	return len(urls), urls
}

func makeRequest(url string, timeout int) ([]byte, error, int) {

	ctx, cncl := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
	defer cncl()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	resp, err := http.DefaultClient.Do(req)

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
	body, err, status := makeRequest(url, timeout)
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

	var res UrlSet
	body, err, status := makeRequest(url, timeout)
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
