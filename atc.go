package main

import (
	"fmt"
	"github.com/rubyist/circuitbreaker"
	"time"
)

type Message struct {
	rc     chan Message
	target string
	msg    string
}

func (m *Message) Reply(r string, rc chan Message) bool {
	m.rc <- Message{
		rc:     rc,
		target: m.target,
		msg:    r}
	return true
}

type AirTrafficControl struct {
	panel *circuit.Panel
	rc    chan Message
}

func runAirTrafficControl(rc chan Message) bool {
	defer close(rc) // will prob trip of sender, add flag ?
	panel := circuit.NewPanel()
	for {
		msg := <-rc
		cb, ok := panel.Get(msg.target)
		switch msg.msg {
		case "JOIN":
			// add new breaker for this target if first time seen
			if !ok {
				// panel.Add(msg.target, circuit.NewThresholdBreaker(1))
				panel.Add(msg.target, circuit.NewRateBreaker(0.10, 10))
			}
			if cb.Ready() {
				msg.Reply("CLEARED", rc)
			} else {
				msg.Reply("ABORT", rc)
			}
		case "SUCCESS":
			cb.Success()
		case "FAILURE":
			cb.Fail()
		}

	}
	return true
}

func NewAirTrafficControl() *AirTrafficControl {
	atc := &AirTrafficControl{
		panel: circuit.NewPanel(),
		rc:    make(chan Message)}
	go runAirTrafficControl(atc.rc)
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
	fmt.Println("--TESTING")

	atc := NewAirTrafficControl()
	for i := 0; i < 11113; i++ {
		time.Sleep(1000000)

		rc := atc.JoinQueue("testtarget")
		defer close(rc)

		abort := false
		waiting := true
		var msg Message
		for abort == false && waiting == true {
			msg = <-rc
			switch msg.msg {
			case "WAIT":
				waiting = true
			case "ABORT":
				abort = true
			case "CLEARED":
				waiting = false
			}
		}
		if !abort {
			fmt.Println("we're a go!", i)
			msg.Reply("FAILURE", rc)
		} else {
			fmt.Println("we've aborted!", i)
		}
		time.Sleep(5)
	}
	fmt.Println("--DONE TESTING")
}
