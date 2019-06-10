package main

import (
	"io"
	"log"
	"os"

	"barullo"

	"github.com/gordonklaus/portaudio"
)

var (
	channelNum      = 1
	bitDepthInBytes = 2
	bufferSize      = 64 * 10
)

const (
	sampleRate = 44100
)

func main() {
	//mPortaudio()
	mPortaudioTakeOnMe()
}

func test() {
	f, err := os.Open("/home/antonio/Downloads/Rick_Astley_-_Never_Gonna_Give_You_Up.mid")
	if err != nil {
		panic(err)
	}
	barullo.GetEventsFromMidi(4, sampleRate, f)
}

func mPortaudio() {
	portaudio.Initialize()
	defer portaudio.Terminate()

	buf := make([]float64, bufferSize)
	out := make([]float32, bufferSize)

	stream, err := portaudio.OpenDefaultStream(0, 1, sampleRate, bufferSize, &out)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		log.Fatal(err)
	}
	defer stream.Stop()

	f, err := os.Open("./assets/never_gonna_give_you_up.json")
	if err != nil {
		log.Fatal(err)
	}

	e1 := barullo.GetEventsFromMidi(5, sampleRate, f)
	f.Seek(0, os.SEEK_SET)
	e2 := barullo.GetEventsFromMidi(7, sampleRate, f)
	f.Seek(0, os.SEEK_SET)
	e3 := barullo.GetEventsFromMidi(10, sampleRate, f)

	eLength1 := e1[len(e1)-1].Offset + 6000
	eLength2 := e2[len(e2)-1].Offset + 6000
	eLength3 := e3[len(e3)-1].Offset + 6000

	eLength := eLength1
	if eLength2 > eLength1 {
		eLength = eLength2
	}
	if eLength2 > eLength {
		eLength = eLength3
	}

	seq1 := barullo.NewSequence(eLength, e1)
	seq2 := barullo.NewSequence(eLength, e2)
	seq3 := barullo.NewSequence(eLength, e3)

	sig1 := barullo.NewPulse(0.25, sampleRate, seq1)
	env1 := barullo.NewEnvelope(2000/4, 2000/4, 0.8, 10000/4, sig1, seq1)
	lp1 := barullo.NewLPFilter(env1, 300.3, 1.1)

	sig2 := barullo.NewTriangle(0.8, sampleRate, seq2)
	env2 := barullo.NewEnvelope(100, 0, 1.0, 500, sig2, seq2)
	lp2 := barullo.NewLPFilter(env2, 8000.3, 6.2)

	sig3 := barullo.NewPulse(0.1, sampleRate, seq3)
	env3 := barullo.NewEnvelope(2000/4, 2000/4, 0.8, 10000/4, sig3, seq3)
	lp3 := barullo.NewLPFilter(env3, 8000.3, 1.1)

	mixer := barullo.NewMixer([]barullo.Node{lp1, lp2, lp3}, []float64{0.8, 0.8, 0.8})

	var sampleOffset int64
	for {
		mixer.Get(int(sampleOffset), buf)

		sampleOffset += int64(bufferSize)

		f64ToF32Copy(out, buf)

		if err := stream.Write(); err != nil {
			log.Printf("error writing to stream : %v\n", err)
		}
	}
}

var channels = []struct {
	events func(io.Reader) []barullo.Event
	source func([]barullo.Event, int) *barullo.LPFilter
	volume float64
}{
	{
		events: func(f io.Reader) []barullo.Event {
			return barullo.GetEventsFromMidi(0, sampleRate, f)
		},
		source: func(evts []barullo.Event, length int) *barullo.LPFilter {
			seq := barullo.NewSequence(length, evts)
			sig := barullo.NewPulse(0.25, sampleRate, seq)
			env := barullo.NewEnvelope(2000/4, 2000/4, 0.8, 10000/4, sig, seq)

			return barullo.NewLPFilter(env, 300.3, 1.1)
		},
		volume: 0.4,
	},
	{
		events: func(f io.Reader) []barullo.Event {
			return barullo.GetEventsFromMidi(2, sampleRate, f)
		},
		source: func(evts []barullo.Event, length int) *barullo.LPFilter {
			seq := barullo.NewSequence(length, evts)
			sig := barullo.NewPulse(0.2, sampleRate, seq)
			env := barullo.NewEnvelope(2000/4, 2000/4, 0.8, 10000/4, sig, seq)
			return barullo.NewLPFilter(env, 5000.3, 6.2)
		},
		volume: 0.05,
	},
	{
		events: func(f io.Reader) []barullo.Event {
			return barullo.GetEventsFromMidi(3, sampleRate, f)
		},
		source: func(evts []barullo.Event, length int) *barullo.LPFilter {
			seq := barullo.NewSequence(length, evts)
			sig := barullo.NewTriangle(0.8, sampleRate, seq)
			env := barullo.NewEnvelope(100, 0, 1.0, 500, sig, seq)

			return barullo.NewLPFilter(env, 8000.3, 6.2)

		},
		volume: 0.1,
	},
	{
		events: func(f io.Reader) []barullo.Event {
			return barullo.GetEventsFromMidi(4, sampleRate, f)
		},
		source: func(evts []barullo.Event, length int) *barullo.LPFilter {
			seq := barullo.NewSequence(length, evts)
			sig := barullo.NewSignal(barullo.Sin, sampleRate, seq)
			env := barullo.NewEnvelope(1000/4, 1000/4, 0.8, 10000/4, sig, seq)
			return barullo.NewLPFilter(env, 5000.3, 6.2)
		},
		volume: 0.1,
	},
	{
		events: func(f io.Reader) []barullo.Event {
			return barullo.GetEventsFromMidi(5, sampleRate, f)
		},
		source: func(evts []barullo.Event, length int) *barullo.LPFilter {
			seq := barullo.NewSequence(length, evts)
			sig := barullo.NewPulse(0.25, sampleRate, seq)
			env := barullo.NewEnvelope(2000/4, 2000/4, 0.8, 10000/4, sig, seq)

			return barullo.NewLPFilter(env, 600.3, 1.1)
		},
		volume: 0.2,
	},
	{
		events: func(f io.Reader) []barullo.Event {
			return barullo.GetEventsFromMidi(7, sampleRate, f)
		},
		source: func(evts []barullo.Event, length int) *barullo.LPFilter {
			seq := barullo.NewSequence(length, evts)
			sig := barullo.NewSignal(barullo.Noise, sampleRate, seq)
			env := barullo.NewEnvelope(2000/4, 2000/4, 0.8, 10000/4, sig, seq)
			return barullo.NewLPFilter(env, 5000.3, 6.2)
		},
		volume: 0.2,
	},
}

func mPortaudioTakeOnMe() {
	portaudio.Initialize()
	defer portaudio.Terminate()

	buf := make([]float64, bufferSize)
	out := make([]float32, bufferSize)

	stream, err := portaudio.OpenDefaultStream(0, 1, sampleRate, bufferSize, &out)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		log.Fatal(err)
	}
	defer stream.Stop()

	f, err := os.Open("./assets/Take On Me 8 (Karaoke).mid")
	if err != nil {
		log.Fatal(err)
	}

	var allEvents [][]barullo.Event
	var maxlen int
	for _, ch := range channels {
		events := ch.events(f)
		allEvents = append(allEvents, events)
		f.Seek(0, os.SEEK_SET)
		eLength := events[len(events)-1].Offset + 6000
		if eLength > maxlen {
			maxlen = eLength
		}
	}

	var nodes []barullo.Node
	var volumes []float64
	for i, ch := range channels {
		nodes = append(nodes, ch.source(allEvents[i], maxlen))
		volumes = append(volumes, ch.volume)
	}

	mixer := barullo.NewMixer(nodes, volumes)

	var sampleOffset int64
	sampleOffset = 44100 * 8
	for {
		mixer.Get(int(sampleOffset), buf)

		sampleOffset += int64(bufferSize)

		f64ToF32Copy(out, buf)

		if err := stream.Write(); err != nil {
			log.Printf("error writing to stream : %v\n", err)
		}
	}
}

func f64ToF32Copy(dst []float32, src []float64) {
	for i := range src {
		dst[i] = float32(src[i])
	}
}
