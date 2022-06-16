package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi"
	channel "gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/reader"
	driver "gitlab.com/gomidi/rtmididrv"
	"gopkg.in/yaml.v3"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

type MidiIO struct {
	Ins  []midi.In
	Outs []midi.Out
}

type key struct {
	Name     string   `yaml:"name"`
	Aliases  []string `yaml:"aliases,flow"`
	Type     string   `yaml:"type"`
	Task     string   `yaml:"task"`
	MaxValue uint8    `yaml:"max_value"`
	KeyUp    string   `yaml:"key_up"`
	KeyDown  string   `yaml:"key_down"`
    Info     []int    `yaml:"info"`
}

type config struct {
	Port string `yaml:"port"`
	Keys []key  `yaml:"keys"`
}


var prev_values map[string]uint8

func (c config) getKey(name string) (key, error) {
	for _, key := range c.Keys {
		if key.Name == name {
			return key, nil
		}
		for _, alias := range key.Aliases {
			if alias == name {
				return key, nil
			}
		}
	}
	return key{}, fmt.Errorf("No configuration found.")
}

func initMidi() MidiIO {
	drv, err := driver.New()
	must(err)

	// make sure to close all open ports at the end
	defer drv.Close()

	ins, err := drv.Ins()
	must(err)

	outs, err := drv.Outs()
	must(err)

	midiIO := MidiIO{Ins: ins, Outs: outs}

	return midiIO
}
// YAML handler
func getConf() *config {
    var conf config
    env := os.Getenv("MIDI_MACRO_PATH")
    if (!(len(env) <0)){
        env = "/home/v0/midi-macro-master/config/midi_macros.yml"
    }
	yamlFile, err := ioutil.ReadFile(env)
	must(err)
	err = yaml.Unmarshal(yamlFile, &conf)
    must(err)

	return &conf
}

func main() {
	// List command
	var rootCmd = &cobra.Command{
		Use:   "midimacro",
		Short: "Tool to map macros to your MIDI controller",
		// Long:  ``,
	}

	// List command
	var midiMacroListCmd = &cobra.Command{
		Use:   "list",
		Short: "List of connected devices",
		Long:  `Prints list of connected MIDI devices. Configure this in your YAML config file`,
		Run: func(cmd *cobra.Command, args []string) {
			// Discover which port to use.
			midiIO := initMidi()
			printInPorts(midiIO.Ins)
		},
	}

	// Run command
	var midiMacroRunCmd = &cobra.Command{
		Use:   "run",
		Short: "Run MIDI event listener",
		Long:  `Starts MIDI event listener and macro handler`,
		Run: func(cmd *cobra.Command, args []string) {
			midiIO := initMidi()
			in, out := getIn(midiIO.Ins), midiIO.Outs[0]

			must(in.Open())
			must(out.Open())
			prev_values = make(map[string]uint8)

			rd := reader.New(
				reader.NoLogger(),

				// Fetch every message
				reader.Each(func(pos *reader.Position, msg midi.Message) {
					switch midi_message := msg.(type) {
					case channel.NoteOn:
						handle(midi_message)
					case channel.ControlChange:
						handle(midi_message)
					}
				}),
			)

			exit := make(chan string)

			// listen for MIDI
			go rd.ListenTo(getIn(midiIO.Ins))
			fmt.Println("MIDI event listener started!")

			for {
				select {
				case <-exit:
					os.Exit(0)
				}
			}
		},
	}

	rootCmd.AddCommand(midiMacroRunCmd)
	rootCmd.AddCommand(midiMacroListCmd)
	rootCmd.Execute()
}

func handle(midi_message midi.Message){
	conf := getConf()
    //fmt.Println(prev_value,prev_values[key.Name])
    var midi_key string
    var value uint8
    switch m := midi_message.(type) {
    case channel.NoteOn:
        midi_key = fmt.Sprint(m.Key())
    case channel.ControlChange:
        midi_key = fmt.Sprint(m.Controller())
        value = m.Value()
        // CC knob btn sends 127 then 0, we ignore 0
        if value == 0{ return }
    }
    key, err := conf.getKey(midi_key)

	if err != nil {
		fmt.Printf("Error getting key %s: %v\n", midi_key, err)
    }
    // write previous values
	prev_value, has_prev := prev_values[key.Name]
	if !has_prev {
		prev_value = value
	}
	prev_values[key.Name] = value

    // handle tasks
	if key.Task == "mouseMove" {
		mouseMove(value,key.Info)
	} else if key.Task == "mouseClickToggle" {
        if prev_value == 127{
            prev_values[key.Name] = 0
        } else {
            prev_values[key.Name] = 127
        }
        mouseClickToggle(prev_value)
    } else if key.Task == "mouseClick"{
        mouseClick()
	} else if key.Task == "mouseScroll"{
        mouseScroll(value,key.Info)
	} else {
        fmt.Println("running",key, midi_message)
        command := getCommand(key.Task)
        name, args := command[0], command[1:]

        cmd := exec.Command(name, args...)
        cmd.Run()
    }
}

func getIn(ports []midi.In) (def midi.In) {
	conf := getConf()
	portIndex, err := strconv.Atoi(conf.Port)

	if err == nil {
		return ports[portIndex]
	}

	for _, in := range ports {
		if strings.Contains(in.String(), conf.Port) {
			return in
		}
	}

	fmt.Printf("Failed to find port: %s\n", conf.Port)
	os.Exit(1)

	return
}

// Get task command
func getCommand(task string) []string {
	split := strings.Split(task, ",")
	return split
}

func printPort(port midi.Port) {
	fmt.Printf("[%v] %s\n", port.Number(), port.String())
}

func printInPorts(ports []midi.In) {
	fmt.Printf("MIDI IN Ports\n")
	for _, port := range ports {
		printPort(port)
	}
}
