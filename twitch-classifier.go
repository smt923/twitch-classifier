package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	twitch "github.com/gempir/go-twitch-irc"
	"github.com/jbrukh/bayesian"
)

const (
	negative bayesian.Class = "Negative"
	neutral  bayesian.Class = "Neutral"
	positive bayesian.Class = "Positive"
)

var (
	negatives     = []string{}
	neutrals      = []string{}
	positives     = []string{}
	channelToJoin = "summit1g"
	filePath      = channelToJoin + ".json"
	ratings       = map[string]int{}
)

func main() {
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		fmt.Printf("Cleaning up, saving and exiting...\n")
		saveScores()
		os.Exit(0)
	}()

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		fmt.Printf("Found previous data for '%s', loading...\n", channelToJoin)
		c, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Could not read file: %s\n%v\n", filePath, err)
		}
		err = json.Unmarshal(c, &ratings)
		if err != nil {
			log.Fatalf("Could not unmarshal json in %s\n%v\n", filePath, err)
		}
		fmt.Printf("Done, starting up...\n")
	}

	classifier := trainClassifier("data.csv")

	client := twitch.NewClient("justinfan123123", "oauth:123123123")

	client.OnNewMessage(func(channel string, user twitch.User, message twitch.Message) {
		chatstring := strings.Split(message.Text, " ")

		_, likely, _ := classifier.LogScores(chatstring)
		adjustment := likely - 1 // scores become -1, 0 or 1

		if value, ok := ratings[user.Username]; ok {
			ratings[user.Username] = value + adjustment
		} else {
			ratings[user.Username] = 5
			ratings[user.Username] += adjustment
		}

		fmt.Printf("(%s) [#%s] %s (r:%d): %s\n",
			scoreToCategory(likely, message.Text),
			channel,
			user.Username,
			ratings[user.Username],
			message.Text,
		)

		saveScores()
	})

	client.Join(channelToJoin)

	err := client.Connect()
	if err != nil {
		panic(err)
	}
}

func scoreToCategory(category int, message string) string {
	switch category {
	case 0:
		return "Negative"
	case 1:
		return "Neutral"
	case 2:
		return "Positive"
	default:
		log.Printf("Unknown value '%v'!\nOffending message: '%s' Continuing...\n", category, message)
		return "UNKNOWN"
	}
}

func trainClassifier(path string) *bayesian.Classifier {
	// add save/load of data
	f, err := os.Open("data.csv")
	if err != nil {
		log.Fatal(err)
	}

	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		switch record[1] {
		case "0":
			negatives = append(negatives, record[0])
		case "1":
			neutrals = append(neutrals, record[0])
		case "2":
			positives = append(positives, record[0])
		}
	}

	classifier := bayesian.NewClassifier(negative, neutral, positive)
	classifier.Learn(negatives, negative)
	classifier.Learn(neutrals, neutral)
	classifier.Learn(positives, positive)

	return classifier
}

func saveScores() {
	b, err := json.Marshal(ratings)
	if err != nil {
		log.Fatalf("Could not marshal scores to json\n%v\n", err)
	}
	err = ioutil.WriteFile(filePath, b, 0644)
	if err != nil {
		log.Fatalf("Could not write file '%s' to disk\n%v\n", filePath, err)
	}
}
