package atc

import (
	"fmt"
	"github.com/rubyist/circuitbreaker"
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
			fmt.Println("circuit report success")
			cb.Success()
		case "FAILURE":
			fmt.Println("circuit report failure")
			cb.Fail()
		}

	}
	return true
}

func NewAirTrafficControl() *AirTrafficControl {
	atc := &AirTrafficControl{
		panel: circuit.NewPanel(),
		rc:    make(chan Message)}
	go runAirTrafficControl(atc.rc) // start its go routine
	return atc
}

func (atc *AirTrafficControl) GetClearance(h string) bool {
	rc := make(chan Message)
	atc.rc <- Message{
		rc:     rc,
		target: h,
		msg:    "JOIN"}
	msg := <-rc
	return msg.msg == "CLEARED"
}

func (atc *AirTrafficControl) ReportFailure(h string) bool {
	atc.rc <- Message{
		rc:     nil,
		target: h,
		msg:    "FAILURE"}
	return true
}

func (atc *AirTrafficControl) ReportSuccess(h string) bool {
	atc.rc <- Message{
		rc:     nil,
		target: h,
		msg:    "SUCCESS"}
	return true
}
