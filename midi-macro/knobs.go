package main

import (
	"fmt"
	"os/exec"
//	"time"
)

func run(args ...string){
    cmd := exec.Command("/usr/bin/xdotool",args...)
    cmd.Run()
}

// Volume controller
func updateVolume(maxValue uint8, value uint8) {
	volume := float32(value) / (float32(maxValue) / 10) // 10 is the max volume

	cmd := exec.Command("/usr/bin/osascript", "-e", "set volume "+fmt.Sprint(volume))
	cmd.Run()
}

// Brightness controller
func updateBrightness(maxValue uint8, value uint8) {
	brightness := float32(value) / (float32(maxValue) / 1) // 1 is the max brightness
	run("/usr/local/bin/brightness", fmt.Sprint(brightness))
}

func keyPress(maxValue uint8, value uint8, prev_value uint8, key key) {
    var action string
    if value > prev_value {
        action = key.KeyUp
    } else if value < prev_value {
        action = key.KeyDown
    }

    if action != "" {
        run("key", action)
    }
}


func mouseMove(value uint8, info []int) {
    action :=-1
    if (value == 127){ action = 1 }
    x := info[0]
    y := info[1]
    xx := fmt.Sprint(action*x*20)
    yy := fmt.Sprint(action*y*20)
    //fmt.Println(info,x,y,xx,yy)
    run("mousemove_relative","--",xx,yy,"--sync")
}

func mouseClickToggle(prev_value uint8) {
    fmt.Println(prev_value)
    arg:= ""
    if (prev_value ==0){
        arg = "mousedown"
    } else {
        arg = "mouseup"
    }

    run(arg,"1","--sync")
}

func mouseClick() {
    run("click","1")
}

func mouseScroll(value uint8,info []int){
    // 0 vertical 1 horizontal
    action:=0
    if info[0] == 0{
        action =4
        if (value == 127){ action = 5 }
    } else if info[0]==1{
        action =7
        if (value == 127){ action = 6 }
    }
    arg := fmt.Sprint(action)
    run("click",arg,"--sync")
}
