package main

import (
	"fmt"
	"github.com/rubyist/circuitbreaker"
)

type Message struct {
	rc     chan Message
	target string
	msg    string
}

func (m *Message) Reply(r string) bool {
	m.rc <- Message{
		rc:     m.rc,
		target: m.target,
		msg:    r}
	return true
}

type AirTrafficControl struct {
	panel *circuit.Panel
	rc    chan Message
}

func NewAirTrafficControl() *AirTrafficControl {
	atc := &AirTrafficControl{
		panel: circuit.NewPanel(),
		rc:    make(chan Message)}
	go func() {
		defer close(atc.rc) // will prob trip of sender, add flag ?
		for {
			msg := <-atc.rc
			msg.Reply("WAIT")
			msg.Reply("CLEARED")
		}
	}()
	return atc
}

func (atc *AirTrafficControl) JoinQueue(h string) chan Message {
	rc := make(chan Message)
	atc.rc <- Message{
		rc:     rc,
		target: h,
		msg:    "JOIN"}
	return rc
}

func main() {
	atc := NewAirTrafficControl()
	rc := atc.JoinQueue("testhost")
	defer close(rc)
	abort := false
	waiting := true
	for abort == false && waiting == true {
		msg := <-rc
		switch msg.msg {
		case "WAIT":
			fmt.Println("waiting")
			waiting = true
		case "ABORT":
			fmt.Println("aborting")
			abort = true
		case "CLEARED":
			fmt.Println("cleared")
			waiting = false
		}
	}
	if !abort {
		fmt.Println("we're a go!")
	} else {
		fmt.Println("we've aborted!")
	}
}
