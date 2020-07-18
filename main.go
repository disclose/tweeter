package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	bitly "github.com/retgits/bitly/client"
	"github.com/retgits/bitly/client/bitlinks"
)

const (
	account     string = "disclose_io"
	url2Shorten        = "https://github.com/disclose?utm_content=disclose_twitter&utm_medium=social&utm_source=twitter.com&utm_campaign=disclose_bot"
)

type item struct {
	program_name   string `json:"program_name"`
	policy_url     string `json:"policy_url"`
	submission_url string `json:"submission_url"`
	launch_date    string `json:"launch_date"`
	bug_bounty     bool   `json:"bug_bounty"`
	swag           bool   `json:"swag"`
	hall_of_fame   bool   `json:"hall_of_fame"`
	safe_harbor    string `json:"safee_harbor"`
}

func main() {
	resp, err := http.Get("https://raw.githubusercontent.com/disclose/disclose/master/program-list/program-list.json")

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	jsonFeed := []item{}
	err = json.Unmarshal(body, &jsonFeed)
	if err != nil {
		panic(err)
	}

	btClient := bitlinks.New(bitly.NewClient().WithAccessToken("YOUR BITLY KEY HERE"))

	request := bitlinks.ShortenRequest{
		LongURL: url2Shorten,
	}

	bitlyRes, err := btClient.ShortenLink(&request)
	if err != nil {
		panic(err)
	}

	MAIN_LINK := bitlyRes.Link
	TOTAL_PROGRAMS := float64(len(jsonFeed))
	BOUNTY := 0.0
	SWAG := 0.0
	HALL_OF_FAME := 0.0
	SAFE_HARBOUR_FULL := 0.0
	SAFE_HARBOUR_PARTIAL := 0.0

	for _, item := range jsonFeed {
		bounty := item.bug_bounty
		swag := item.swag
		hall_of_fame := item.hall_of_fame
		safe_harbor := item.safe_harbor
		if bounty {
			BOUNTY++
		}
		if swag {
			SWAG++
		}
		if hall_of_fame {
			HALL_OF_FAME++
		}
		if safe_harbor == "full" {
			SAFE_HARBOUR_FULL++
		}

		if safe_harbor == "partial" {
			SAFE_HARBOUR_PARTIAL++
		}
	}

	BOUNTY_PERCENT := (BOUNTY / TOTAL_PROGRAMS) * 100
	SAFE_HARBOUR_FULL_PERCENT := (SAFE_HARBOUR_FULL / TOTAL_PROGRAMS) * 100
	SAFE_HARBOUR_PARTIAL_PERCENT := (SAFE_HARBOUR_PARTIAL / TOTAL_PROGRAMS) * 100
	HALL_OF_FAME_PERCENT := (HALL_OF_FAME / TOTAL_PROGRAMS) * 100
	SWAG_PERCENT := (SWAG / TOTAL_PROGRAMS) * 100

	dat, err := ioutil.ReadFile("message-content.txt")
	if err != nil {
		fmt.Println("Could not read message content file")
		panic(err)
	}

	rawMessage := string(dat)

	replacements := map[string]string{
		"{{MAIN_LINK}}":                    MAIN_LINK,
		"{{BOUNTY}}":                       fmt.Sprintf("%.0fi", BOUNTY),
		"{{BOUNTY_PERCENT}}":               fmt.Sprintf("%.5f", BOUNTY_PERCENT) + "%",
		"{{TOTAL_PROGRAMS}}":               fmt.Sprintf("%.0f", TOTAL_PROGRAMS),
		"{{SAFE_HARBOUR_FULL}}":            fmt.Sprintf("%.0f", SAFE_HARBOUR_FULL),
		"{{SAFE_HARBOUR_FULL_PERCENT}}":    fmt.Sprintf("%.5f", SAFE_HARBOUR_FULL_PERCENT) + "%",
		"{{SAFE_HARBOUR_PARTIAL}}":         fmt.Sprintf("%.0f", SAFE_HARBOUR_PARTIAL),
		"{{SAFE_HARBOUR_PARTIAL_PERCENT}}": fmt.Sprintf("%.5f", SAFE_HARBOUR_PARTIAL_PERCENT) + "%",
		"{{HALL_OF_FAME}}":                 fmt.Sprintf("%.0f", HALL_OF_FAME),
		"{{HALL_OF_FAME_PERCENT}}":         fmt.Sprintf("%.5f", HALL_OF_FAME_PERCENT) + "%",
		"{{SWAG}}":                         fmt.Sprintf("%.0f", SWAG),
		"{{SWAG_PERCENT}}":                 fmt.Sprintf("%.5f", SWAG_PERCENT) + "%",
	}

	for k, v := range replacements {
		rawMessage = strings.Replace(rawMessage, k, v, -1)
	}

	tweetContent := rawMessage

	config := oauth1.NewConfig("consumerKey", "consumerSecret")
	token := oauth1.NewToken("accessToken", "accessSecret")
	httpClient := config.Client(oauth1.NoContext, token)

	twitterClient := twitter.NewClient(httpClient)

	tweet, resp, err := twitterClient.Statuses.Update(tweetContent, nil)
	if err != nil {
		fmt.Println("Error sending tweet")
		panic(err)
	}

	fmt.Println(tweet)
}
