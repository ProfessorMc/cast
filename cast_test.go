package cast

import (
	"testing"
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
