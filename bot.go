package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	// Token Discord API Bot Token
	Token  string
	prefix = "!"
	// DefaultRole String containing the role name to put authetnicated users in
	DefaultRole = "Members"
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	// Create new discord session using token provided on command line via -t
	dg, err := discordgo.New("Bot " + Token)

	// Exit if we fail to connect for whatever reason
	if err != nil {
		fmt.Println("Error creating discord session: ", err)
		os.Exit(1)
	}

	// Register handlers for discord events
	dg.AddHandler(messageCreate)

	// Connect to discord
	err = dg.Open()
	if err != nil {
		fmt.Println("Error connecting to discord websocket: ", err)
		os.Exit(1)
	}

	fmt.Println("Bot connected to Discord! Press Ctrl-C to shutdown")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ignore messages we send
	if m.Author.ID == s.State.User.ID {
		return
	}

	// if a message length is greater than 1 character and the prefix matches our set prefix,
	// split the command from the prefix and store the rest as arguments to be passed on to
	// the appropiate handler function
	if len(m.Content) > 1 && strings.HasPrefix(m.Content, prefix) {
		messageArray := strings.Split(m.Content, " ")
		command := strings.ToLower(strings.TrimPrefix(messageArray[0], prefix))
		args := messageArray[1:]

		switch command {
		case "iam":
			authHandler(s, m, args)
		case "whois":
			whoisHandler(s, m, args)
		}
	}
}

func authHandler(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Bomb out if there's less than 3 arguments
	if len(args) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Missing argument. Please use !iam server firstname lastname")
		return
	}

	// Build character name
	characterFirstName := strings.Title(args[1])
	characterLastName := strings.Title(args[2])
	characterName := characterFirstName + " " + characterLastName

	// Attempt to change nickname
	err := s.GuildMemberNickname(m.GuildID, m.Author.ID, characterName)
	if err != nil {
		fmt.Println("Failed to change nickname for ", m.Author.Username)
		fmt.Println(err)
	}

	// Add default role to member
	dguild, err := s.Guild(m.GuildID)
	if err != nil {
		fmt.Println(err)
		return
	}

	role := getGuildRoleByName(DefaultRole, dguild)
	err = s.GuildMemberRoleAdd(m.GuildID, m.Author.ID, role)
	if err != nil {
		msg := "Failed to add role " + DefaultRole + " to " + m.Author.Username
		fmt.Println(msg)
		fmt.Println(err)
		s.ChannelMessageSend(m.ChannelID, msg)
		return
	}

	// Success! Let the user know
	s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+"> authenticated as **"+characterName+"**")

	// TODO: Serialize this data to file as JSON so it can be retreived later
}

func getGuildRoleByName(name string, guild *discordgo.Guild) string {
	for _, role := range guild.Roles {
		if role.Name == name {
			return role.ID
		}
	}
	return ""
}

func whoisHandler(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
}
