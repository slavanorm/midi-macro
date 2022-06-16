midi-macro
==========

Use your MIDI controller (pads, knobs, sliders, keys etc.) to trigger macros.

To build:
`sudo apt-get install libasound2-dev`
`go build -o midimacro midi-macro/*.go`

Then:
`./midimacro list`

Pick your device name from list. 
Set its name to the config file, and point an environment variable to it: 
`export MIDI_MACRO_PATH=/path/to/midi_macros.yml`

And finally
`./midimacro run`

differences from master:

main.go:
naive refactor and changes to allow new functions in knobs.go file.

configs:
i set it up to employ my useless OP-1 to use it instead of mouse with help of xdotools.

it sounds strange, but helps a lot against a ton of mouseclicking in VCV rack.
