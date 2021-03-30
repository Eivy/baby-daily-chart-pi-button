package main

import (
	"log"
	"net/http"
	"os"
	"time"

	rpio "github.com/stianeikeland/go-rpio"
)

const URL = "https://us-central1-babydailychart.cloudfunctions.net/NewHappen"

func main() {
	log.Println("Start watching pins")
	err := rpio.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer rpio.Close()
	ch := make(chan string)
	for k, v := range map[string]int{
		"1": 4,
		"2": 17,
		"3": 27,
	} {
		go readButton(ch, k, v)
	}
	out := rpio.Pin(14)
	out.Mode(rpio.Output)
	for {
		select {
		case button := <-ch:
			log.Println("Send ", button)
			err = send(button)
			if err != nil {
				log.Println("[ERROR]", button, err)
				blink(out)
			}
		}
	}
}

func blink(pin rpio.Pin) {
	for range []int{1, 2} {
		pin.High()
		time.Sleep(time.Millisecond * 200)
		pin.Low()
		time.Sleep(time.Millisecond * 200)
	}
}

func readButton(ch chan string, button string, pin int) (err error) {
	in := rpio.Pin(pin)
	in.Mode(rpio.Input)
	in.PullDown()
	var current rpio.State
	before := in.Read()
	for {
		current = in.Read()
		if current == rpio.High && current != before {
			ch <- button
		}
		before = current
		time.Sleep(time.Millisecond * 200)
	}
}

func send(button string) (err error) {
	req, err := http.NewRequest("POST", URL, nil)
	if err != nil {
		return
	}
	req.Header.Set("BABY_USER_ID", os.Getenv("BABY_USER_ID"))
	req.Header.Set("BABY_BUTTON_NUM", button)
	_, err = http.DefaultClient.Do(req)
	return
}
