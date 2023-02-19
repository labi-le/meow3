package main

import (
	"context"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/labi-le/meow3/internal"
	"github.com/labi-le/meow3/pkg/player"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-envconfig"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

type Config struct {
	AccessToken string `env:"ACCESS_TOKEN,required"`
}

const MusicDir = "meow3-music"

//const MusicDir = "/home/labile/drive/meow3-music"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var conf Config
	if err := envconfig.Process(ctx, &conf); err != nil {
		panic(errors.Wrap(err, "failed to load config"))
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	logger.Info("start meow3")

	logger.Infof("music dir: %s", MusicDir)
	logger.Infof("config: %v", conf)

	bot := state.New("Bot " + conf.AccessToken)
	bot.AddIntents(
		gateway.IntentGuilds |
			gateway.IntentGuildPresences |
			gateway.IntentGuildVoiceStates |
			gateway.IntentGuildMessages |
			gateway.IntentGuildMembers,
	)

	common := player.NewCommon(logger)
	internal.RegisterHandlers(bot, logger, common, MusicDir)

	defer bot.Close()

	if err := bot.Open(ctx); err != nil {
		logger.Error(errors.Wrap(err, "failed to open bot"))
	}

	<-ctx.Done()
}
