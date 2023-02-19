package player

import (
	"context"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/voice"
	"github.com/sirupsen/logrus"
	"sync"
)

type Common struct {
	sync.Mutex
	heap map[string]*Player
	log  *logrus.Logger
}

func NewCommon(log *logrus.Logger) *Common {
	log.SetReportCaller(true)

	return &Common{heap: make(map[string]*Player), log: log}
}

func (c *Common) Add(ctx context.Context, v *voice.Session, guid discord.GuildID, cid discord.ChannelID, file string) {
	c.Lock()
	c.log.Info("adding player to guild ", guid, " with channel ", cid)
	activePlayer := c.Get(guid)
	if activePlayer != nil {
		c.log.Warning("found active player on heap")
		if activePlayer.cid != cid {
			c.log.Warning("found active player on heap with different channel id. stopping it and creating new one")
			activePlayer.Stop(true)
		} else {
			c.log.Warning("reusing voice session")
			activePlayer.Stop(false)
			v = activePlayer.Session
		}
	}

	c.log.Info("creating new player ", guid)
	player := NewPlayer(v, cid, file)
	c.add(guid, player)

	c.log.Info("playing ", guid, " ", cid, " ", file)
	go player.Play(ctx, cid)

	c.Unlock()
}

func (c *Common) Get(guid discord.GuildID) *Player {
	return c.heap[guid.String()]
}

func (c *Common) add(guid discord.GuildID, p *Player) {
	c.heap[guid.String()] = p
}

func (c *Common) Remove(guid discord.GuildID) {
	c.Lock()
	defer c.Unlock()

	player := c.Get(guid)
	if player != nil {
		c.log.Info("removing player from guild ", guid)

		player.Stop(true)
		delete(c.heap, guid.String())
	}
}
