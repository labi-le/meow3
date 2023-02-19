package internal

import (
	"context"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/diamondburned/arikawa/v3/voice"
	"github.com/labi-le/meow3/pkg/player"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

type resource struct {
	bot    *state.State
	common *player.Common
	log    *logrus.Logger

	musicDir string

	self *discord.User
}

func RegisterHandlers(bot *state.State, logger *logrus.Logger, common *player.Common, musicDir string) {
	me, err := bot.Me()
	if err != nil {
		panic(errors.Wrap(err, "failed to get self id"))
	}

	res := &resource{
		bot:      bot,
		common:   common,
		log:      logger,
		musicDir: musicDir,
		self:     me,
	}

	res.registerAutocomplete()

	bot.AddHandler(res.autocomplete)
	bot.AddHandler(res.Help)
	bot.AddHandler(res.Meow)
	bot.AddHandler(res.StopMeow)
	bot.AddHandler(res.AutoExitVoice)
	bot.AddHandler(res.Playlist)
}

func (r *resource) Help(c *gateway.MessageCreateEvent) {
	if c.Author.Bot {
		return
	}

	if c.Content != "/help" {
		return
	}

	msg := `**Meow3** - a heterosexual discord bot for playing music

**/meow** play random track in voice channel
**/stop** stop playing music
**/playlist** show playlist
**/help** show this Help`

	r.bot.SendMessage(c.ChannelID, msg)
}

func (r *resource) Meow(c *gateway.MessageCreateEvent) {
	if c.Author.Bot {
		return
	}

	if c.Content != "/Meow" {
		return
	}

	r.play(c.GuildID, c.Author.ID)

	r.bot.SendMessage(c.ChannelID, "you are not in voice channel")

}

func (r *resource) play(guild discord.GuildID, uid discord.UserID) {
	states, err := r.bot.VoiceStates(guild)
	if err != nil {
		r.log.Error(errors.Wrap(err, "failed to get voice states"))
	}

	for _, s := range states {
		if s.UserID == uid {

			track := player.SelectRandomTrack(os.DirFS(r.musicDir))
			r.log.Print("add track " + track)
			v, err := voice.NewSession(r.bot)
			if err != nil {
				r.log.Error(errors.Wrap(err, "failed to get voice state"))
				return
			}
			r.common.Add(
				context.Background(),
				v,
				s.GuildID,
				s.ChannelID,
				track,
			)

			return

		}
	}
}

func (r *resource) AutoExitVoice(c *gateway.VoiceStateUpdateEvent) {
	if c.UserID == r.self.ID {
		return
	}

	// if empty, leave voice channel
	states, err := r.bot.VoiceStates(c.GuildID)
	if err != nil {
		return
	}

	type UserInVoice map[discord.ChannelID][]discord.UserID

	channels := make(UserInVoice)
	for _, voiceState := range states {
		channels[voiceState.ChannelID] = append(channels[voiceState.ChannelID], voiceState.UserID)
	}

	var chanWithBot discord.ChannelID
	for channelID, users := range channels {
		for _, user := range users {
			if user == r.self.ID {
				chanWithBot = channelID
				break
			}
		}
	}

	if chanWithBot == 0 {
		return
	}

	if len(channels[chanWithBot]) == 1 {
		r.common.Remove(c.GuildID)
	}
}

func (r *resource) StopMeow(c *gateway.MessageCreateEvent) {
	if c.Author.Bot {
		return
	}

	if c.Content != "/stop" {
		return
	}

	r.common.Remove(c.GuildID)
}

func (r *resource) Playlist(c *gateway.MessageCreateEvent) {
	if c.Author.Bot {
		return
	}

	if c.Content != "/playlist" {
		return
	}

	playlist := player.GetAllTracks(os.DirFS(r.musicDir))

	r.bot.SendMessage(c.ChannelID, playlist.String())
}

func (r *resource) registerAutocomplete() {
	app, err := r.bot.CurrentApplication()
	if err != nil {
		r.log.Error("Failed to get application ID:", err)
	}

	newCommands := []api.CreateCommandData{
		{
			Name:        "meow",
			Description: "Play random track in voice channel",
		},

		{
			Name:        "stop",
			Description: "Stop playing music",
		},
	}

	//if _, err := r.bot.BulkOverwriteGuildCommands(app.ID, discord.GuildID(1060870426617708634), newCommands); err != nil {
	//	r.log.Error("failed to create guild command:", err)
	//}
	//
	if _, err := r.bot.BulkOverwriteCommands(app.ID, newCommands); err != nil {
		r.log.Error("failed to create guild command:", err)
	}
}

func (r *resource) autocomplete(e *gateway.InteractionCreateEvent) {
	var resp api.InteractionResponse
	switch d := e.Data.(type) {
	case *discord.CommandInteraction:
		switch d.Name {
		case "meow":
			resp = api.InteractionResponse{
				Type: api.MessageInteractionWithSource,
				Data: &api.InteractionResponseData{
					Content: option.NewNullableString("Playing"),
				},
			}

			r.play(e.GuildID, e.Member.User.ID)
			break
		case "stop":
			resp = api.InteractionResponse{
				Type: api.MessageInteractionWithSource,
				Data: &api.InteractionResponseData{
					Content: option.NewNullableString("Stop"),
				},
			}

			r.common.Remove(e.GuildID)

			break
		}
		//case *discord.AutocompleteInteraction:
		//default:
		//	spew.Dump(d)
		//	return
	}

	if err := r.bot.RespondInteraction(e.ID, e.Token, resp); err != nil {
		log.Println("failed to send interaction callback:", err)
	}
}
