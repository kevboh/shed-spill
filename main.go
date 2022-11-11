package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")
	tokenKey := os.Getenv("TOKEN_KEY")
	tokenSecret := os.Getenv("TOKEN_SECRET")

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(tokenKey, tokenSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	name := os.Args[1]
	file := fmt.Sprintf("%s.csv", name)
	f, err := os.Create(file)
	if err != nil {
		panic(fmt.Errorf("Could not create %s!", file))
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	header := []string{
		"ID",
		"Text",
		"Date",
		"Favorites",
		"Replies",
		"Retweets",
		"Is Retweet",
	}
	w.Write(header)

	trimUser := true
	excludeReplies := false
	includeRetweets := true
	maxID := int64(0)
	lastWrittenID := int64(0)
	page := 1
	for {
		fmt.Printf("Fetching page %d of tweets…\n", page)
		params := &twitter.UserTimelineParams{
			ScreenName:      name,
			TrimUser:        &trimUser,
			ExcludeReplies:  &excludeReplies,
			IncludeRetweets: &includeRetweets,
			TweetMode:       "extended",
			MaxID:           maxID,
		}
		tweets, _, _ := client.Timelines.UserTimeline(params)

		if tweets == nil || len(tweets) == 0 || (len(tweets) == 1 && tweets[0].ID == lastWrittenID) {
			return
		}

		for i := 0; i < len(tweets); i++ {
			tweet := tweets[i]
			if tweet.ID == lastWrittenID {
				continue
			}

			row := []string{
				tweet.IDStr,
				tweet.FullText,
				tweet.CreatedAt,
				strconv.Itoa(tweet.FavoriteCount),
				strconv.Itoa(tweet.ReplyCount),
				strconv.Itoa(tweet.RetweetCount),
				strconv.FormatBool(tweet.Retweeted),
			}

			w.Write(row)

			maxID = tweet.ID
			lastWrittenID = tweet.ID
		}

		page += 1
		// Sleep a bit per loop to avoid making Twitter even more unstable.
		time.Sleep(time.Second)
	}
}
