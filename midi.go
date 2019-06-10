package barullo

import (
	"fmt"
	"io"
	"time"

	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/midi/smf"
	"gitlab.com/gomidi/midi/smf/smfreader"
)

const notesString = "C C#D D#E F F#G G#A A#B "

func keyToNoteAndOctave(key uint8) (note string, octave int) {
	octave = int(key/12 - 1)
	note = notesString[(key%12)*2 : (key%12)*2+2]
	return
}

func calcDuration(p *mid.Position, ticks smf.MetricTicks, bpm float64) (dur time.Duration) {
	ppq := ticks.Ticks4th()
	usPerTick := bpm / float64(ppq)
	nsPerTick := usPerTick * 1000
	return time.Duration(float64(p.AbsoluteTicks) * nsPerTick)
}

func GetEventsFromMidi(ch, sampleRate int, r io.Reader) []Event {
	var events []Event

	rd := mid.NewReader(mid.NoLogger())
	//rd := mid.NewReader()
	smfr := smfreader.New(r)

	if err := smfr.ReadHeader(); err != nil {
		panic(err)
	}

	header := smfr.Header()
	ticks := header.TimeFormat.(smf.MetricTicks)

	var defaultBpms float64 = 120

	tempoBPM := func(p mid.Position, bpm float64) {
		fmt.Println("TEMPOOOOOOOooo", bpm)
		defaultBpms = bpm
	}

	rd.Msg.Meta.TempoBPM = tempoBPM

	noteOn := func(p *mid.Position, channel, key, vel uint8) {
		if int(channel) == ch {
			dur := calcDuration(p, ticks, defaultBpms)
			note, octave := keyToNoteAndOctave(key)
			fmt.Println("NOTE ON", "CHANNEL", ch, "DURATION:", dur, "NOTE", note, "OCTAVE", octave)
			e := Event{
				Offset: int(dur/time.Second) * sampleRate,
				Note:   note,
				Octave: octave,
				Key:    NotePress,
			}

			events = append(events, e)
		}
	}

	noteOff := func(p *mid.Position, channel, key, vel uint8) {
		if int(channel) == ch {
			dur := calcDuration(p, ticks, defaultBpms)
			note, octave := keyToNoteAndOctave(key)
			fmt.Println("NOTE OFF", "CHANNEL", ch, "DURATION:", dur, "NOTE", note, "OCTAVE", octave)
			e := Event{
				Offset: int(dur/time.Second) * sampleRate,
				Note:   note,
				Octave: octave,
				Key:    NoteRelease,
			}

			events = append(events, e)
		}
	}

	rd.Msg.Channel.NoteOn = noteOn
	rd.Msg.Channel.NoteOff = noteOff

	if err := rd.ReadSMFFrom(smfr); err != nil {
		panic(err)
	}

	if len(events) == 0 {
		panic(fmt.Sprintln("NO EVENTS ADDED TO CHANNEL", ch))
	}
	return events
}
