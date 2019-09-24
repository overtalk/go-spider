package movie

import (
	"log"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

type DouBanMovie struct {
	Rank     int
	Href     string
	Name     string
	Start    float64
	Quote    string
	Actor    []string
	Director []string
	Year     int
	Region   []string
	Category []string
}

func main() {
	c := colly.NewCollector(
		colly.Async(true),
	)

	extensions.RandomUserAgent(c) // 使用随机的UserAgent，最好能使用代理。这样就不容易被ban
	extensions.Referer(c)

	c.Limit(&colly.LimitRule{DomainGlob: "*.douban.*", Parallelism: 5})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	c.OnHTML(".item", func(e *colly.HTMLElement) {
		m := getDetail(e)
		log.Printf("%+v", m)
	})

	c.OnHTML(".paginator a", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	c.Visit("https://movie.douban.com/top250?start=0&filter=")
	c.Wait()
}

func getDetail(e *colly.HTMLElement) *DouBanMovie {
	goQuerySelection := e.DOM

	rank, _ := strconv.Atoi(goQuerySelection.Find("em").Text())

	star, _ := strconv.ParseFloat(goQuerySelection.Find(".rating_num").Text(), 64)

	detail := strings.TrimSpace(goQuerySelection.Find(".bd").Find("p").Eq(0).Text())
	temp := strings.Split(detail, "\n")
	if len(temp) != 2 {
		temp = []string{"未知", "未知"}
	}

	director, actor := getActorAndDirector(strings.TrimSpace(temp[0]))
	year, region, category := getCategoryAndYear(strings.TrimSpace(temp[1]))

	return &DouBanMovie{
		Rank:     rank,
		Href:     e.ChildAttr(".hd a", "href"),
		Name:     e.DOM.Find("span.title").Eq(0).Text(),
		Start:    star,
		Quote:    strings.TrimSpace(goQuerySelection.Find(".bd").Find("p").Eq(1).Text()),
		Actor:    actor,
		Director: director,
		Year:     year,
		Region:   region,
		Category: category,
	}
}

func getActorAndDirector(str string) ([]string, []string) {
	var (
		placeholder = "+"
		actor       []string
		director    []string
	)

	for strings.Contains(str, placeholder) {
		placeholder += placeholder
	}

	str = strings.Replace(str, "主演:", placeholder, -1)
	str = strings.Replace(str, "导演:", placeholder, -1)
	str = strings.Replace(str, "...", placeholder, -1)
	temp := strings.Split(str, placeholder)

	result := make([]string, 0)
	for _, v := range temp {
		if len(v) != 0 {
			result = append(result, v)
		}
	}

	for k, v := range result {
		if k == 0 {
			director = strings.Split(v, "/")
		} else if k == 1 {
			actor = strings.Split(v, "/")
		}
	}

	return director, actor
}

func getCategoryAndYear(str string) (int, []string, []string) {
	var (
		year     int
		region   []string
		category []string
	)
	result := strings.Split(str, "/")
	for k, v := range result {
		switch k {
		case 0:
			year, _ = strconv.Atoi(strings.TrimSpace(v))
		case 1:
			region = strings.Split(v, " ")
		case 2:
			category = strings.Split(v, " ")
		}
	}

	return year, region, category
}
