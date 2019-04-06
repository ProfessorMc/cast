package cast

import (
	"sync"
	"time"
)

// Relay implements an interface for a fan out pattern
// the goal is to have one input source pass data to multiple output recievers
// with each reciever getting the same data
type Relay interface {
	Start()
	Close()
	AddReceiver() chan []byte
}

type cast struct {
	input   <-chan []byte
	outputs []chan<- []byte
	end     chan struct{}
	add     chan chan []byte
	running bool
	mu      sync.Mutex
}

// New returns a broacast interface from an input channel
func New(input <-chan []byte) Relay {
	return &cast{
		input:   input,
		outputs: make([]chan<- []byte, 0),
		end:     make(chan struct{}),
		add:     make(chan chan []byte),
	}
}

func (c *cast) Start() {
	c.running = true
	go func() {
		for {
			select {
			// handle incoming data
			case data := <-c.input:
				//log.Print("message received")
				go c.broadcastMessage(data)
			case ch := <-c.add:
				c.outputs = append(c.outputs, ch)
			// handle end signal
			case <-c.end:
				for _, ch := range c.outputs {
					close(ch)
				}
				c.running = false
				return
			}

		}
	}()

}

func (c *cast) broadcastMessage(data []byte) {
	for _, ch := range c.outputs {
		go sendToReceiver(ch, data)
	}
}

func sendToReceiver(ch chan<- []byte, data []byte) {
	messageSent := make(chan bool, 0)
	go func(done chan bool) {
		ch <- data
		close(messageSent)
	}(messageSent)
	select {
	case <-messageSent:
		//log.Print("message sent to receiver")
		break
	case <-time.After(5 * time.Millisecond):
		//log.Print("message timed out")
		break
	}
}

func (c *cast) AddReceiver() chan []byte {
	ch := make(chan []byte, 1)
	if c.running {
		c.add <- ch
	} else {
		c.outputs = append(c.outputs, ch)
	}
	return ch
}

func (c *cast) Close() {
	c.end <- struct{}{}
}
