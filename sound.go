package main

import (
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

var (
	sampleRate = beep.SampleRate(44100)
	beepSound  beep.Streamer
)

func InitSound() {
	speaker.Init(sampleRate, sampleRate.N(time.Second/10))
	beepSound = squareWave(440)
}

func PlayBeep() {
	speaker.Play(beepSound)
}

func StopBeep() {
	speaker.Clear()
}

func squareWave(freq float64) beep.Streamer {
	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		for i := range samples {
			t := float64(i) / float64(sampleRate)
			if int(t*freq*2)%2 == 0 {
				samples[i][0] = 0.5
				samples[i][1] = 0.5
			} else {
				samples[i][0] = -0.5
				samples[i][1] = -0.5
			}
		}
		return len(samples), true
	})
}
