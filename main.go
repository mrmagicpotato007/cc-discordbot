package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"math/rand"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Token          string
	RealChallenges Challenges
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()

	//loading challenges file
	pwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	filePath := filepath.Join(pwd, "challenges.txt")
	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Println(err)
	}
	RealChallenges = Challenges{}
	err = json.Unmarshal(file, &RealChallenges)
	if err != nil {
		log.Println(err)
	}

}

type quoter struct {
	Id     int    `json:"id"`
	Quote  string `json:"quote"`
	Author string `json:"author"`
}

type Challenges struct {
	Challengeslist []*Challenge `json:"challenges"`
}

type Challenge struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

func main() {

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreator)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func messageCreator(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Println("message:", m.Content)
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "Hello" || m.Content == "hello" {
		log.Println("Someone sent hello")
		_, err := s.ChannelMessageSend(m.ChannelID, "Hello "+m.Author.Username)
		if err != nil {
			log.Println(err)
		}
	}
	if strings.HasPrefix(m.Content, "!quote") {
		quote := GetQuote()
		_, err := s.ChannelMessageSend(m.ChannelID, quote)
		if err != nil {
			log.Println(err)
		}
		log.Println(quote)
	}
	if strings.HasPrefix(m.Content, "!challenge") {
		challenge := GetChallenge()
		_, err := s.ChannelMessageSend(m.ChannelID, challenge)
		if err != nil {
			log.Println(err)
		}
		log.Println(challenge)
	}
	if strings.HasPrefix(m.Content, "!list") {
		challenges := GetAllChallenges()
		_, err := s.ChannelMessageSend(m.ChannelID, challenges)
		if err != nil {
			log.Println(err)
		}
		log.Println(challenges)
	}
	if strings.HasPrefix(m.Content, "!add") {
		inp := strings.Split(m.Content, " ")
		url := ""
		for _, val := range inp {
			if strings.Contains(val, "codingchallenges.fyi") {
				url = val
				break
			}
		}
		if url == "" {
			_, err := s.ChannelMessageSend(m.ChannelID, "Unable to add the input url its should have domain name codingchallenges.fyi")
			if err != nil {
				log.Println(err)
			}
			return
		}
		challenges := AddChallenge(m.Author.Username+"'s challenge", url)
		_, err := s.ChannelMessageSend(m.ChannelID, challenges)
		if err != nil {
			log.Println(err)
		}
		log.Println(challenges)
	}
}

func GetQuote() string {
	resp, err := http.Get("https://dummyjson.com/quotes/random")
	if err != nil {
		log.Println(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	q1 := new(quoter)
	err = json.Unmarshal(body, q1)
	if err != nil {
		log.Println(err)
	}
	if q1.Quote != "" {
		return q1.Quote + " - " + q1.Author
	}
	return "Im the death, the destroyer of the world - Robert oppenheimer"
}

func GetChallenge() string {
	len := len(RealChallenges.Challengeslist) - 1
	index := rand.Intn(len)
	log.Println(index)
	challenge := RealChallenges.Challengeslist[index].Name + ": " + RealChallenges.Challengeslist[index].Url
	log.Println(challenge)
	return challenge
}

func GetAllChallenges() string {
	clist := ""
	for _, challenge := range RealChallenges.Challengeslist {
		clist += challenge.Name + ": " + challenge.Url + "\n"
	}
	return clist
}

func AddChallenge(name string, url string) string {
	challenge := &Challenge{Name: name, Url: url}
	RealChallenges.Challengeslist = append(RealChallenges.Challengeslist, challenge)
	return "Added " + url
}
