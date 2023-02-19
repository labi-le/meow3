package player

import (
	"context"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/voice"
	"github.com/diamondburned/arikawa/v3/voice/udp"
	"github.com/diamondburned/oggreader"
	"github.com/pkg/errors"
	"io"
	"os/exec"
	"strconv"
	"time"
)

type Player struct {
	Session *voice.Session
	Cmd     *exec.Cmd
	Pipe    io.ReadCloser
	cid     discord.ChannelID
	stop    chan struct{}
}

// Optional constants to tweak the Opus stream.
const (
	frameDuration = 60 // ms
	timeIncrement = 2880
)

func NewPlayer(v *voice.Session, cid discord.ChannelID, file string) *Player {
	v.SetUDPDialer(udp.DialFuncWithFrequency(
		frameDuration*time.Millisecond, // correspond to -frame_duration
		timeIncrement,
	))

	player := createFFMpeg(file)
	player.Session = v
	player.stop = make(chan struct{})
	player.cid = cid

	go func() {
		select {
		case <-player.stop:
			player.cleanup()
			return
		}
	}()

	return player
}

func createFFMpeg(file string) *Player {
	ffmpeg := exec.Command(
		"ffmpeg", "-hide_banner", "-loglevel", "error",
		// Streaming is slow, so a single thread is all we need.
		"-threads", "1",
		// Input file.
		"-i", file,
		// Output format; leave as "libopus".
		"-c:a", "libopus",
		// Bitrate in kilobits. This doesn't matter, but I recommend 96k as the
		// sweet spot.
		"-b:a", "128k",
		// Frame duration should be the same as what's given into
		// udp.DialFuncWithFrequency.
		"-frame_duration", strconv.Itoa(frameDuration),
		// Disable variable bitrate to keep packet sizes consistent. This is
		// optional.
		"-vbr", "off",
		// Output format, which is opus, so we need to unwrap the opus file.
		"-f", "opus",
		"-",
	)

	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		panic(errors.Wrap(err, "failed to get stdout pipe"))
	}

	if err := ffmpeg.Start(); err != nil {
		panic(errors.Wrap(err, "failed to Play ffmpeg"))
	}

	return &Player{
		Cmd:  ffmpeg,
		Pipe: stdout,
	}
}

func (p *Player) Play(ctx context.Context, id discord.ChannelID) error {
	defer p.cleanup()

	// Join the voice channel.
	if err := p.Session.JoinChannelAndSpeak(ctx, id, false, false); err != nil {
		return errors.Wrap(err, "failed to join channel")
	}

	// Start decoding FFmpeg's OGG-container output and extract the raw Opus
	// frames into the stream.
	if err := oggreader.DecodeBuffered(p.Session, p.Pipe); err != nil {
		if !errors.Is(err, udp.ErrManagerClosed) {
			return errors.Wrap(err, "failed to decode ogg")
		}

	}

	// Wait until FFmpeg finishes writing entirely and leave.
	_ = p.Cmd.Wait()

	return nil
}

func (p *Player) Stop(leave bool) {
	p.stop <- struct{}{}

	if leave {
		_ = p.Session.Leave(context.TODO())
	}
}

func (p *Player) cleanup() {
	_ = p.Cmd.Process.Kill()
	_ = p.Pipe.Close()
}
