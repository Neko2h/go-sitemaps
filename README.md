# Go sitemaps
## Easy way to work with xml sitemaps


## Features

- Parsing sitemaps
- Parsing sitemaps Indexes
- Getting count of urls
- Getting count of files
- Concurrency
- Zero dependencies
- GZIP support



## Installation



```sh
go get github.com/Neko2h/sitemaps
```





## Usage
To simply parse sitemap
```golang
	a, err := sitemaps.ParseSitemap("https://site.com/sitemap.xml", 3)
	if err != nil {
		//error handling
	}

	for _, v := range a.Entites {
		//DO something
		fmt.Println(v.Loc, v.ChangeFreq, v.Lastmod)
	}
```
To parse sitemap Index
```golang

	a, filesCount, err := sitemaps.ParseIndex("https://somesite.com/sitemap.xml", 4)

	if err != nil {
		//error handling
	}

	for _, v := range a.Entites {
		fmt.Println(v.Loc, v.Lastmod, v.ChangeFreq)
	}

```

Concurent parsing ALL links
```golang
	sitemaps, _, _ := sitemaps.ParseIndex("https://somesite.com/index.xml", 5)
	count, links := sitemaps.GetUrls(10)
	for _, v := range links {
		fmt.Println(v.Loc, v.ChangeFreq, v.ChangeFreq)
	}
```
This library is using worker pool algorithm.
But beweare of parsing a huge amount of data. It will eat your RAM pretty fast.







## TODO

- [ ] Benchmarking
- [ ] Tests!
- [ ] Lazy concurent parsing
- [ ] Link url with sitemapIndex pointer

## License

MIT

