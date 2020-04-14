package google

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

const (
	urlSearch = "https://www.google.com/search?q=%s&num=100&hl=%s"
)

type Result struct {
	ResultRank  int
	ResultURL   string
	ResultTitle string
	ResultDesc  string
}

func buildGoogleUrl(searchTerm string) string {
	searchTerm = strings.Trim(searchTerm, " ")
	searchTerm = strings.Replace(searchTerm, " ", "+", -1)
	return fmt.Sprintf(urlSearch, searchTerm, "en")
}

func googleRequest(searchURL string) (*http.Response, error) {
	baseClient := &http.Client{}
	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")
	res, err := baseClient.Do(req)
	if err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func googleResultParser(response *http.Response) ([]Result, error) {
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}
	var results []Result
	sel := doc.Find("div.g")
	rank := 1
	for i := range sel.Nodes {
		item := sel.Eq(i)
		linkTag := item.Find("a")
		link, _ := linkTag.Attr("href")
		titleTag := item.Find("h3.r")
		descTag := item.Find("span.st")
		desc := descTag.Text()
		title := titleTag.Text()
		link = strings.Trim(link, " ")
		if link != "" && link != "#" {
			result := Result{
				rank,
				link,
				title,
				desc,
			}
			results = append(results, result)
			rank += 1
		}
	}
	return results, err
}

func Search(searchTerm string) ([]Result, error) {
	googleUrl := buildGoogleUrl(searchTerm)
	res, err := googleRequest(googleUrl)
	if err != nil {
		return nil, err
	}
	scrapes, err := googleResultParser(res)
	if err != nil {
		return nil, err
	} else {
		return scrapes, nil
	}
}
