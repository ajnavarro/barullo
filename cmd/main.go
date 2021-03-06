package main

import (
	"log"
	"math"

	"barullo"

	"github.com/gordonklaus/portaudio"
)

var (
	channelNum      = 1
	bitDepthInBytes = 2
	bufferSize      = 64
)

func main() {
	mPortaudio()
}

var freqs = []float64{
	16.35, // C
	18.35, // D
	20.60, // E
	21.83, // F
}

func mPortaudio() {
	portaudio.Initialize()
	defer portaudio.Terminate()

	sinBuf := make([]float64, bufferSize)
	envBuf := make([]float64, bufferSize)
	out := make([]float32, bufferSize)

	stream, err := portaudio.OpenDefaultStream(0, 1, 44100, bufferSize, &out)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		log.Fatal(err)
	}
	defer stream.Stop()

	env := barullo.NewEnvelope(2000, 2000, 0.8, 10000, sinBuf)

	var freq float64
	var sampleOffset int64
	for {
		freq = freqs[(sampleOffset/44100)%int64(len(freqs))]
		freq = freq * 16 // change octave

		oscSin(sampleOffset, freq, sinBuf)
		env.Get(int(sampleOffset), envBuf)
		sampleOffset += int64(bufferSize)

		f64ToF32Copy(out, envBuf)

		if err := stream.Write(); err != nil {
			log.Printf("error writing to stream : %v\n", err)
		}
	}
}

const twoPi = math.Pi * 2.0

func oscSin(o int64, freq float64, buf []float64) {
	var i int64
	for i = 0; i < int64(len(buf)); i++ {
		pos := float64(o + i)
		buf[i] = math.Sin((twoPi / 44100.0) * freq * pos)
	}
}

func f64ToF32Copy(dst []float32, src []float64) {
	for i := range src {
		dst[i] = float32(src[i])
	}
}
