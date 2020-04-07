package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"path/filepath"

	"strings"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v2"
)

// Variables used for command line parameters
var (
	Token string
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

type ResourceList struct {
	Resources map[string]guides `yaml:"Resources"`
	Tokenkey  string            `yaml:"DiscordBotToken"`
}
type guides []Resource
type Resource struct {
	Name  string `yaml:"title"`
	Links string `yaml:"link"`
}

// Unmarshal and get resources files
var a ResourceList
var keys []string

func main() {
	//Get the file bytes
	filename, _ := filepath.Abs("./resources.yml")
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("cannot Read data: %v", err)
	}

	// Unmarshal the resource file to object
	err = yaml.Unmarshal(yamlFile, &a)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}
	fmt.Println(a.Resources)

	// Create a new Discord session using the provided bot token.

	if a.Tokenkey != "" {
		Token = a.Tokenkey
	}

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	keys = make([]string, len(a.Resources))
	i := 0
	for key := range a.Resources {
		keys[i] = key // explicit array element assignment instead of append function
		i++
	}
	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	message := strings.Split(m.Content, " ")
	// Provides users with a guide, if availible
	if message[0] == "!guides" {
		s.ChannelMessageSend(m.ChannelID, getGuides(message[1]))
	}

	// Provides users of the bot with the guides supplied in the yml file
	if m.Content == "!help" {
		s.ChannelMessageSend(m.ChannelID, "`\n Usage is !guides \"GuideName\" \n Availible Guides:"+strings.Join(keys, ", ")+"`")
	}
}

//Returns the gudes passed for a given string
func getGuides(s string) string {
	var line = ""
	if s != "" {
		if a.Resources[s] != nil {
			line = "Here are the guides for " + s + "!\n"
			for _, r := range a.Resources[s] {
				line += "`" + r.Name + "`\n" + r.Links + "\n"
			}
			line += ""
		} else {
			line = "No guides found for " + s
		}
	}
	return line
}
