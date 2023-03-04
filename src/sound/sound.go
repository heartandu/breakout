package sound

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
)

const (
	samplingRate  = 44100
	numOfChannels = 2
	audioBitDepth = 2

	bgMusicFileName     = "resources/sounds/breakout.mp3"
	nsbBleepFileName    = "resources/sounds/bleep.mp3"
	sbBleepFileName     = "resources/sounds/solid.wav"
	powerUpFileName     = "resources/sounds/powerup.wav"
	paddleBleepFileName = "resources/sounds/bleep.wav"
)

type Player struct {
	context *oto.Context

	bgMusic     oto.Player
	bgMusicStop chan struct{}
	nsbBleep    oto.Player
	sbBleep     oto.Player
	powerUp     oto.Player
	paddleBleep oto.Player
}

func NewPlayer() (*Player, error) {
	context, readyChan, err := oto.NewContext(samplingRate, numOfChannels, audioBitDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to create context: %w", err)
	}

	<-readyChan

	p := &Player{context: context}

	if err = p.initSounds(); err != nil {
		return nil, fmt.Errorf("failed to init sounds: %w", err)
	}

	return p, nil
}

func (p *Player) initSounds() error {
	var err error

	p.bgMusic, err = p.initSoundPlayer(bgMusicFileName)
	if err != nil {
		return fmt.Errorf("failed to init bg sound player: %w", err)
	}

	p.bgMusicStop = make(chan struct{})

	p.nsbBleep, err = p.initSoundPlayer(nsbBleepFileName)
	if err != nil {
		return fmt.Errorf("failed to init non solid block bleep player: %w", err)
	}

	p.sbBleep, err = p.initSoundPlayer(sbBleepFileName)
	if err != nil {
		return fmt.Errorf("failed to init solid block bleep player: %w", err)
	}

	p.powerUp, err = p.initSoundPlayer(powerUpFileName)
	if err != nil {
		return fmt.Errorf("failed to init power up player: %w", err)
	}

	p.paddleBleep, err = p.initSoundPlayer(paddleBleepFileName)
	if err != nil {
		return fmt.Errorf("failed to init paddle bleep player: %w", err)
	}

	return nil
}

func (p *Player) initSoundPlayer(fileName string) (oto.Player, error) {
	file, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read sound file '%s': %w", fileName, err)
	}

	var decoder io.Reader

	fileReader := bytes.NewReader(file)
	extensionIndex := strings.LastIndex(fileName, ".")

	switch fileName[extensionIndex+1:] {
	case "mp3":
		decoder, err = mp3.NewDecoder(fileReader)
	case "wav":
		decoder, err = wav.DecodeWithoutResampling(fileReader)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create decoder: %w", err)
	}

	return p.context.NewPlayer(decoder), nil
}

func (p *Player) PlayBgMusic() {
	playLoop(p.bgMusic, p.bgMusicStop)
}

func (p *Player) PlayNonSolidBlockBleep() {
	play(p.nsbBleep)
}

func (p *Player) PlaySolidBlockBleep() {
	play(p.sbBleep)
}

func (p *Player) PlayPowerUp() {
	play(p.powerUp)
}

func (p *Player) PlayPaddleBleep() {
	play(p.paddleBleep)
}

func (p *Player) Cleanup() error {
	close(p.bgMusicStop)

	if err := p.context.Suspend(); err != nil {
		return fmt.Errorf("failed to suspend context: %w", err)
	}

	return nil
}

func playLoop(p oto.Player, stopChan chan struct{}) {
	go func() {
	Loop:
		for {
			select {
			case <-stopChan:
				p.Pause()
				break Loop
			default:
				if !p.IsPlaying() {
					play(p)
				}
			}
		}
	}()
}

func play(p oto.Player) {
	_, err := p.(io.Seeker).Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}
	p.Play()
}
