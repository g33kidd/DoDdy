package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	botcommands "github.com/Devs-On-Discord/DoDdy/botcommands"
	"github.com/bwmarrin/discordgo"
	bolt "go.etcd.io/bbolt"
)

const version = "0.0.1"

// The store is global for access in goroutines, this might create race conditions and lead to loss of data
// Miyoyo: I can't find anything saying this is goroutine safe or not, I'll assume, for the sake of simplicity.
//         Could be replaced by a goroutine transaction system
var db *bolt.DB

var commands = botcommands.BotCommands{}

func main() {
	fmt.Printf("DoDdy %s starting\n", version)

	bot, err := discordgo.New("Bot " + testToken)
	if err != nil {
		panic(err.Error())
	}

	commands.Init(bot)
	bot.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}
		if len(m.Content) == 0 {
			return
		}
		commands.Parse(m)
	})

	db, err = bolt.Open("doddy.db", 0666, nil)
	if err != nil {
		panic("could not open boltdb: " + err.Error())
	}
	defer db.Close()

	//TODO: Reimplement prefixes
	/*if db.View(
		func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("Nodes"))
			c := b.Cursor()
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				prefixes[string(k)] = string(b.Bucket(k).Get([]byte("Prefix")))
			}
			return nil
		}) != nil {
		panic("could not read prefixes from boltdb: " + err.Error())
	}*/

	if bot.Open() != nil {
		panic("could not open bot: " + err.Error())
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	fmt.Println("DoDdy ready.\nPress Ctrl+C to exit.")
	<-sc
	bot.Close()
}
