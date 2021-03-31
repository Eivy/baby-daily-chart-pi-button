package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	rpio "github.com/stianeikeland/go-rpio"
)

var URL = "https://us-central1-babydailychart.cloudfunctions.net/NewHappen"
var ID = os.Getenv("BABY_USER_ID")
var path = "baby.env"

func main() {
	log.Println("Start watching pins")
	readEnv()
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
	go watch(ch)
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":80", nil))
}

func readEnv() {
	_, err := os.Stat(path)
	if err != nil {
		return
	}
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		v := strings.Split(s.Text(), "=")
		switch v[0] {
		case "URL":
			URL = v[1]
		case "ID":
			ID = v[1]
		}
	}
}

func writeEnv() (err error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintln(f, "URL="+URL)
	fmt.Fprintln(f, "ID="+ID)
	return
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		showPage(w, r)
		return
	}
	if r.Method == http.MethodPost {
		changeVariables(w, r)
		return
	}
}

func showPage(w http.ResponseWriter, r *http.Request) {
	page := `<html>
	<body>
	  <form method="POST">
	  <label for="url">FunctionsURL:</label><input id="url" name="url" type="text" value="` + URL + `">
	  <label for="id">UserID:</label><input id="id" name="id" type="text" value="` + ID + `">
	  <button>Send</button>
	  </form>
	</body>
	</html>`
	_, err := w.Write([]byte(page))
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func changeVariables(w http.ResponseWriter, r *http.Request) {
	URL = r.FormValue("url")
	ID = r.FormValue("id")
	err := writeEnv()
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("Change values\nURL: %s\nID: %s", URL, ID)
	showPage(w, r)
}

func watch(ch chan string) {
	out := rpio.Pin(14)
	out.Mode(rpio.Output)
	for {
		select {
		case button := <-ch:
			log.Println("Send ", button)
			err := send(button)
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
	req.Header.Set("BABY_USER_ID", ID)
	req.Header.Set("BABY_BUTTON_NUM", button)
	_, err = http.DefaultClient.Do(req)
	return
}
