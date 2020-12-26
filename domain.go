package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"github.com/gocolly/colly"
	DB "github.com/samaita/autokata/sql"
)

const (
	DOMAIN_ACTIVE   = 1
	DOMAIN_INACTIVE = 0

	DOMAIN_FETCH_TYPE_INVALID = -1
	DOMAIN_FETCH_TYPE_URL     = 1
	DOMAIN_FETCH_TYPE_RSS     = 2
	DOMAIN_FETCH_TYPE_JS      = 3
)

type Domain struct {
	DomainID      int64     `json:"domain_id"`
	DomainName    string    `json:"domain_name"`
	DomainURL     string    `json:"domain_url"`
	FeedsURL      string    `json:"feeds_url"`
	CreateTime    time.Time `json:"create_time"`
	FetchType     int       `json:"fetch_type"`
	TitlePos      string    `json:"title_pos"`
	SummaryPos    string    `json:"summary_pos"`
	CoverImagePos string    `json:"cover_image_pos"`
	ListImagePos  string    `json:"list_image_pos"`
	ContentPos    string    `json:"content_pos"`
	URLPos        string    `json:"url_pos"`
	Status        int       `json:"status"`
}

func getAllDomain() ([]Domain, error) {
	var (
		errParse, errQuery, errScan error
		Domains                     []Domain
	)

	query := "SELECT domain_id, domain_name, domain_url, feeds_url, create_time FROM db_domain WHERE status = $1"
	rows, errQuery := DB.Collection.Main.Queryx(query, DOMAIN_ACTIVE)
	if errQuery != nil {
		log.Println(errQuery, query)
		return Domains, errQuery
	}
	defer rows.Close()

	for rows.Next() {
		var (
			d Domain
			t string
		)
		if errScan = rows.Scan(&d.DomainID, &d.DomainName, &d.DomainURL, &d.FeedsURL, &t); errScan != nil {
			log.Println(errScan)
			continue
		}
		if d.CreateTime, errParse = time.Parse(time.RFC3339, t); errParse != nil {
			log.Println(errParse, t)
			continue
		}
		Domains = append(Domains, d)
	}

	return Domains, nil
}

func getDomainWithType(fetchType int) ([]Domain, error) {
	var (
		errParse, errQuery, errScan error
		Domains                     []Domain
	)

	query := "SELECT domain_id, domain_name, domain_url, feeds_url, create_time FROM db_domain WHERE fetch_type = $1"
	rows, errQuery := DB.Collection.Main.Queryx(query, fetchType)
	if errQuery != nil {
		log.Println(errQuery, query)
		return Domains, errQuery
	}
	defer rows.Close()

	for rows.Next() {
		var (
			d Domain
			t string
		)
		if errScan = rows.Scan(&d.DomainID, &d.DomainName, &d.DomainURL, &d.FeedsURL, &t); errScan != nil {
			log.Println(errScan)
			continue
		}
		if d.CreateTime, errParse = time.Parse(time.RFC3339, t); errParse != nil {
			log.Println(errParse, t)
			continue
		}
		Domains = append(Domains, d)
	}

	return Domains, nil
}

func NewDomain() Domain {
	return Domain{}
}

func (d *Domain) validate() bool {
	if d.DomainName == "" || d.DomainURL == "" || d.FeedsURL == "" {
		return false
	}
	return true
}

func (d *Domain) validateMapping() bool {
	if d.TitlePos == "" || d.URLPos == "" || d.SummaryPos == "" {
		return false
	}
	return true
}

func (d *Domain) isRSS() bool {
	if strings.Contains(d.FeedsURL, "rss") {
		return true
	}
	return false
}

func (d *Domain) add() error {
	query := `
		INSERT INTO db_domain (
			domain_name,
			domain_url,
			feeds_url,
			create_time,
			fetch_type,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	tx, errQuery := DB.Collection.Main.Beginx()
	if errQuery != nil {
		log.Println(errQuery, query)
		return errQuery
	}
	result, err := tx.Exec(query, d.DomainName, d.DomainURL, d.FeedsURL, time.Now().UTC().Format(time.RFC3339), d.FetchType, d.Status)
	if err != nil {
		return err
	}
	if d.DomainID, err = result.LastInsertId(); err != nil {
		log.Println(errQuery, query)
		return errQuery
	}
	tx.Commit()
	return nil
}

func (d *Domain) testRSS() error {
	var (
		feeds         []Feed
		errGetURLFeed error
	)

	d.FetchType = DOMAIN_FETCH_TYPE_RSS
	if feeds, errGetURLFeed = getRSSFeed(context.Background(), d.FeedsURL); errGetURLFeed != nil {
		log.Println(errGetURLFeed, d.FeedsURL, "Invalid RSS")
	}

	if len(feeds) > 0 {
		d.Status = DOMAIN_ACTIVE
	}

	return errGetURLFeed
}

func (d *Domain) testMapping() error {
	var (
		feeds         []Feed
		errGetURLFeed error
	)

	d.FetchType = DOMAIN_FETCH_TYPE_URL
	if feeds, errGetURLFeed = d.getURLFeedV2(); errGetURLFeed != nil {
		log.Println(errGetURLFeed, d.DomainURL, "Retrying..")
		if feeds, errGetURLFeed = d.getURLFeed(); errGetURLFeed != nil {
			log.Println(errGetURLFeed, d.DomainURL, "Give Up..")
			d.FetchType = DOMAIN_FETCH_TYPE_INVALID
		}
	}

	if len(feeds) > 0 {
		d.Status = DOMAIN_ACTIVE
	}

	return errGetURLFeed
}

func (d *Domain) addMapping() error {
	query := `
		INSERT INTO db_domain_mapping (
			domain_id,
			title_pos,
			summary_pos,
			cover_image_pos,
			url_pos,
			create_time
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	tx, errQuery := DB.Collection.Main.Beginx()
	if errQuery != nil {
		log.Println(errQuery, query)
		return errQuery
	}
	result, err := tx.Exec(query, d.DomainID, d.TitlePos, d.SummaryPos, d.CoverImagePos, d.URLPos, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return err
	}
	if _, err = result.LastInsertId(); err != nil {
		log.Println(errQuery, query)
		return errQuery
	}
	tx.Commit()
	return nil
}

func (d *Domain) isExist() (bool, error) {
	var (
		b        bool
		s        int
		errQuery error
	)

	query := `SELECT 1 FROM db_domain WHERE domain_url = $1`
	errQuery = DB.Collection.Main.QueryRowx(query, d.DomainURL).Scan(&s)
	if errQuery != nil && errQuery != sql.ErrNoRows {
		log.Println(errQuery, query)
		return b, errQuery
	}
	b = s == 1
	return b, nil
}

func (d *Domain) remove() error {
	query := `UPDATE db_domain SET status = $1, update_time = $2 WHERE domain_url = $3`
	tx, errQuery := DB.Collection.Main.Beginx()
	if errQuery != nil {
		log.Println(errQuery, query)
		return errQuery
	}
	_, err := tx.Exec(query, DOMAIN_INACTIVE, time.Now().UTC().Format(time.RFC3339), d.DomainURL)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (d *Domain) getAllPos() error {
	var (
		errQuery error
	)

	query := `
	SELECT
		title_pos,
		summary_pos,
		cover_image_pos,
		list_image_pos,
		content_pos,
		url_pos 
	FROM db_domain_mapping
	WHERE domain_id = $1`
	errQuery = DB.Collection.Main.QueryRowx(query, d.DomainID).Scan(&d.TitlePos, &d.SummaryPos, &d.CoverImagePos, &d.ListImagePos, &d.ContentPos, &d.URLPos)
	if errQuery != nil && errQuery != sql.ErrNoRows {
		log.Println(errQuery, query)
		return errQuery
	}

	return errQuery
}

func (d *Domain) getURLFeed() ([]Feed, error) {
	var (
		arrFeed                                     []Feed
		arrTitle, arrURL, arrCoverImage, arrSummary []string
		err                                         error
	)

	c := colly.NewCollector()

	c.OnHTML(d.TitlePos, func(e *colly.HTMLElement) {
		arrTitle = append(arrTitle, e.Text)
	})
	c.OnHTML(d.URLPos, func(e *colly.HTMLElement) {
		if strings.Contains(e.Attr("href"), "http://") || strings.Contains(e.Attr("href"), "https://") {
			arrURL = append(arrURL, e.Attr("href"))
		} else {
			arrURL = append(arrURL, d.DomainURL+e.Attr("href"))
		}
	})
	c.OnHTML(d.CoverImagePos, func(e *colly.HTMLElement) {
		if strings.Contains(e.Attr("src"), "http://") || strings.Contains(e.Attr("src"), "https://") {
			arrCoverImage = append(arrCoverImage, e.Attr("src"))
		} else {
			arrCoverImage = append(arrCoverImage, e.Attr("data-src"))
		}
	})
	c.OnHTML(d.SummaryPos, func(e *colly.HTMLElement) {
		arrSummary = append(arrSummary, e.Text)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", d.FeedsURL)
	})
	c.Visit(d.FeedsURL)

	if len(arrTitle) == len(arrURL) && len(arrCoverImage) == len(arrSummary) && len(arrTitle) == len(arrCoverImage) {
		for i, title := range arrTitle {
			f := Feed{
				ArticleTitle:      title,
				ArticleURL:        arrURL[i],
				ArticleSummary:    arrSummary[i],
				ArticleCoverImage: arrCoverImage[i],
			}
			arrFeed = append(arrFeed, f)
		}
	}
	return arrFeed, err
}

func (d *Domain) getURLFeedV2() ([]Feed, error) {
	var (
		arrFeed                                     []Feed
		arrTitle, arrURL, arrCoverImage, arrSummary []string
		err                                         error
		isHTML                                      bool
	)

	geziyor.NewGeziyor(&geziyor.Options{
		StartRequestsFunc: func(g *geziyor.Geziyor) {
			g.GetRendered(d.FeedsURL, g.Opt.ParseFunc)
		},
		ParseFunc: func(g *geziyor.Geziyor, r *client.Response) {
			isHTML = r.IsHTML()
			if isHTML {
				r.HTMLDoc.Find(d.TitlePos).Each(func(_ int, s *goquery.Selection) {
					arrTitle = append(arrTitle, s.Text())
				})
				r.HTMLDoc.Find(d.URLPos).Each(func(_ int, s *goquery.Selection) {
					if href, ok := s.Attr("href"); ok {
						if strings.Contains(href, "http://") || strings.Contains(href, "https://") {
							arrURL = append(arrURL, href)
						} else {
							arrURL = append(arrURL, d.DomainURL+href)
						}
					}
				})
				r.HTMLDoc.Find(d.CoverImagePos).Each(func(_ int, s *goquery.Selection) {
					if href, ok := s.Attr("src"); ok {
						arrCoverImage = append(arrCoverImage, href)
					}
				})
				r.HTMLDoc.Find(d.SummaryPos).Each(func(_ int, s *goquery.Selection) {
					arrSummary = append(arrSummary, s.Text())
				})
			}
		},
	}).Start()

	if !isHTML {
		return arrFeed, fmt.Errorf("Not an HTML")
	}

	if len(arrTitle) == len(arrURL) && len(arrCoverImage) == len(arrSummary) && len(arrTitle) == len(arrCoverImage) {
		for i, title := range arrTitle {
			f := Feed{
				ArticleTitle:      title,
				ArticleURL:        arrURL[i],
				ArticleSummary:    arrSummary[i],
				ArticleCoverImage: arrCoverImage[i],
			}
			arrFeed = append(arrFeed, f)
		}
	}
	return arrFeed, err
}
