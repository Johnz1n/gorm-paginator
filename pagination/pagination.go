package pagination

import (
	"math"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
)

// Param 分页参数
type Param struct {
	DB      *gorm.DB
	Page    int
	Limit   int
	OrderBy []string
	ShowSQL bool
	Url     string
}

// Paginator 分页返回
type Paginator struct {
	Data  interface{} `json:"data"`
	Links interface{} `json:"links"`
	Meta  interface{} `json:"meta"`
}

// Links page
type Links struct {
	First string `json:"first"`
	Last  string `json:"last"`
	Prev  string `json:"prev"`
	Next  string `json:"next"`
}

// Meta page
type Meta struct {
	Current  int    `json:"current_page"`
	From     int    `json:"from"`
	LastPage int    `json:"last_page"`
	Path     string `json:"path"`
	PerPage  int    `json:"per_page"`
	To       int    `json:"to"`
	Total    int    `json:"total"`
}

// Paging 分页
func Paging(p *Param, result interface{}) *Paginator {
	db := p.DB

	if p.ShowSQL {
		db = db.Debug()
	}
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit == 0 {
		p.Limit = 10
	}
	if len(p.OrderBy) > 0 {
		for _, o := range p.OrderBy {
			db = db.Order(o)
		}
	}

	done := make(chan bool, 1)
	var paginator Paginator
	var links Links
	var meta Meta
	var count int
	var offset int
	var PrevPage int
	var NextPage int
	var lastPage int

	go countRecords(db, result, done, &count)

	if p.Page == 1 {
		offset = 0
	} else {
		offset = (p.Page - 1) * p.Limit
	}

	db.Limit(p.Limit).Offset(offset).Find(result)
	<-done
	url := strings.Split(p.Url, "?")
	p.Url = url[0]
	lastPage = int(math.Ceil(float64(count) / float64(p.Limit)))

	meta.Current = p.Page
	meta.From = offset + 1
	meta.LastPage = lastPage
	meta.Path = p.Url
	meta.PerPage = p.Limit
	meta.To = offset + p.Limit
	meta.Total = count

	if meta.To > meta.Total {
		meta.To = meta.Total
	}

	if p.Page > 1 {
		PrevPage = p.Page - 1
	} else {
		PrevPage = p.Page
	}

	if p.Page == meta.Total {
		NextPage = p.Page
	} else {
		NextPage = p.Page + 1
	}

	links.First = p.Url + "?page=1"
	links.Last = p.Url + "?page=" + strconv.Itoa(lastPage)

	if p.Page > 1 {
		links.Prev = p.Url + "?page=" + strconv.Itoa(PrevPage)
	}
	if meta.To < meta.Total {
		links.Next = p.Url + "?page=" + strconv.Itoa(NextPage)
	}

	paginator.Data = result
	paginator.Meta = meta
	paginator.Links = links

	return &paginator
}

func countRecords(db *gorm.DB, anyType interface{}, done chan bool, count *int) {
	db.Model(anyType).Count(count)
	done <- true
}
