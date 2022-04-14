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
- Image, Video suport



## Installation



```sh
go get github.com/Neko2h/sitemaps
```





## Usage
To simply parse sitemap Index
```golang
	total, err := sitemaps.Parse("https://somesite.com/sitemap.xml", 100, true, "index", func(e sitemaps.Entity) {
		//handle results
	})
```
To parse sitemap
```golang

	total, err := sitemaps.Parse("https://somesite.com/sitemap.xml", 100, true, "sitemap", func(e sitemaps.Entity) {
		//handle results
	})

```

Concurent parsing ALL links from all sitemaps
```golang
	var results []sitemaps.Entity
	total, err := sitemaps.Parse("https://www.ebay.com/lst/VIS-0-index.xml", 100, true, "index", func(e sitemaps.Entity) {
		results = append(results, e)
	})
	if err != nil {
		//err handling
	}

	sitemaps.GetUrls(results, 6, 100, func(e sitemaps.Entity) {
		//handle results
	})
```
This library is using worker pool algorithm.
Since the library is using xml decoder it's fast and me memory efficient







## TODO

- [ ] Benchmarking
- [ ] Tests!

## License

MIT

