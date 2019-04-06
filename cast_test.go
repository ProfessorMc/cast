package cast

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestOutputs(t *testing.T) {

	// where our values come from
	producer := make(chan []byte)

	// output channels
	outputs := make([]chan []byte, 0)

	// init relay, add channels
	relay := New(producer)
	for i := 0; i < 10; i++ {
		outputs = append(outputs, relay.AddReceiver())
	}

	relay.Start()

	// value added to producer
	producer <- []byte("Hello World")

	for i, ch := range outputs {
		val := <-ch
		if string(val) != "Hello World" {
			t.Errorf("Channel %d not recieving value", i)
		}

	}
}

func TestAdding(t *testing.T) {
	// adding before or after calling Start should not change behavior

	producer := make(chan []byte)

	relay := New(producer)

	output1 := relay.AddReceiver()
	output2 := relay.AddReceiver()

	relay.Start()

	producer <- []byte("Testing")

	val1 := <-output1
	if string(val1) != "Testing" {
		t.Error("Output channel added before start not recieving value")
	}

	val2 := <-output2
	if string(val2) != "Testing" {
		t.Error("Output channel after Start not recieving value")
	}

	output3 := relay.AddReceiver()
	output4 := relay.AddReceiver()

	producer <- []byte("Testing2")
	val1 = <-output1
	if string(val1) != "Testing2" {
		t.Error("Output channel added before start not recieving value")
	}

	val2 = <-output2
	if string(val2) != "Testing2" {
		t.Error("Output channel after Start not recieving value")
	}

	val3 := <-output3
	if string(val3) != "Testing2" {
		t.Error("Output channel added before start not recieving value")
	}

	val4 := <-output4
	if string(val4) != "Testing2" {
		t.Error("Output channel after Start not recieving value")
	}

}

func TestClose(t *testing.T) {
	producer := make(chan []byte)

	relay := New(producer)

	output1 := relay.AddReceiver()
	output2 := relay.AddReceiver()
	output3 := relay.AddReceiver()
	output4 := relay.AddReceiver()

	relay.Start()
	relay.Close()
	// These will block of close doesn't work.
	<-output1
	<-output2
	<-output3
	<-output4

}

// Test produces messages until timeout, passes when 10 messages are received by either the fast or slow client.
// Producer write messages at 2 Hz
// Client 1 Received at 1 Hz
// Client 2 Receives at 4 Hz
// Note, there will be a lag in receiver and producer during startup so num sent should be higher than the 10 received to kill test.
// Client 1 should have about half the the messages of Client 2
func TestSlowReceiverCase(t *testing.T) {
	producer := make(chan []byte)

	// Test will prodcue at .5 second, will send 5 for test.
	relay := New(producer)

	// This one will not, sample at .75 second
	output1 := relay.AddReceiver()

	// This one will be able to keep up, sample at .25 second
	output2 := relay.AddReceiver()
	relay.Start()
	t.Log("starting test producer")
	numMessageSent := 0
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			//t.Log("Sending Message")
			producer <- []byte(fmt.Sprintf("Testing%d", numMessageSent))
			numMessageSent++
		}
	}()
	numMessageReceivedCh1 := 0
	numMessageReceivedCh2 := 0
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
	loop1:
		for {
			select {
			case <-output1:
				time.Sleep(1000 * time.Millisecond)
				numMessageReceivedCh1++
				if numMessageReceivedCh1 >= 10 || numMessageReceivedCh2 >= 10 {
					break loop1
				}
			case <-time.After(4 * time.Second):
				t.Fatalf("failed slow receiver")
				t.Logf("Received on 1: %d messages", numMessageReceivedCh1)
				t.Logf("Received on 2: %d messages", numMessageReceivedCh2)
				break loop1
			}
		}
		wg.Done()
	}()
	go func() {
	loop2:
		for {
			select {
			case <-output2:
				time.Sleep(250 * time.Millisecond)
				numMessageReceivedCh2++
				if numMessageReceivedCh1 >= 10 || numMessageReceivedCh2 >= 10 {
					break loop2
				}
			case <-time.After(4 * time.Second):
				t.Fatalf("failed slow receiver")
				t.Logf("Received on 1: %d messages", numMessageReceivedCh1)
				t.Logf("Received on 2: %d messages", numMessageReceivedCh2)
				break loop2
			}
		}
		wg.Done()
	}()
	wg.Wait()
	t.Logf("Sent: %d messages", numMessageSent)
	t.Logf("Received on 1: %d messages", numMessageReceivedCh1)
	t.Logf("Received on 2: %d messages", numMessageReceivedCh2)

}
