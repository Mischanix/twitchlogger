// Cute read-only TwitchTV IRC client
package justinfan

import (
  "github.com/Mischanix/applog"
  "github.com/Mischanix/wait"
  "net"
  "time"
)

const ircDial = "irc.twitch.tv:6667"
const ircPass = "~"
const ircUser = "justinfan735733"

type Message struct {
  User     string
  Channel  string
  Received time.Time
  Message  string
}

type Client struct {
  conn      net.Conn
  channels  map[string]bool
  joinQueue []string
  partQueue []string
  connected *wait.Flag
  valid     *wait.Flag
  messages  chan *Message
}

// Connect creates a new client and begins and authenticates the IRC session.
func Connect() *Client {
  c := &Client{
    nil,
    make(map[string]bool),
    nil, nil,
    wait.NewFlag(false),
    wait.NewFlag(true),
    make(chan *Message, 64),
  }
  conn, err := net.Dial("tcp", ircDial)
  if err != nil {
    applog.Error("justinfan.Connect failed: %v", err)
  }
  c.conn = conn
  go c.readHandler()
  go c.channelManager()
  return c
}

// Disconnect closes the client.
func (c *Client) Disconnect() {
  c.connected.Set(false)
  c.conn.Close()
  c.valid.Set(false)
}

// SetChannels updates the client to monitor the channels in channelNames.
func (c *Client) SetChannels(channelNames []string) {
  m := make(map[string]bool, len(channelNames))
  for _, name := range channelNames {
    if _, ok := c.channels[name]; !ok {
      c.joinQueue = append(c.joinQueue, name)
    }
    m[name] = true
  }
  for name, _ := range c.channels {
    if _, ok := m[name]; !ok {
      c.partQueue = append(c.partQueue, name)
    }
  }
  c.valid.Set(false)
}

func (c *Client) Messages() <-chan *Message {
  return (<-chan *Message)(c.messages)
}

func (c *Client) readHandler() {
  buffer := make([]byte, 1024)
  var unprocessed []byte
  for c.conn != nil {
    buf := buffer
    l, err := c.conn.Read(buf)
    if unprocessed != nil {
      buf = append(unprocessed, buf[:l]...)
      l += len(unprocessed)
    }
    if err != nil {
      applog.Error("justinfan.read error: %v", err)
      return
    }

    var (
      // t = token start, n = stage, b = beginning
      t, n, b                       int
      method, source, dest, theRest string
    )
    for s := 0; s < l; s++ {
      switch buf[s] {
      case ' ':
        switch n {
        case 0:
          source = string(buf[t:s])
        case 1:
          method = string(buf[t:s])
        case 2:
          dest = string(buf[t:s])
        }
        if n < 3 {
          t = s + 1
          n++
        }
      case '\r', '\n':
        // This treats \n\r and \r as valid as well.  Regardless, tmi only uses
        // \r\n.
        if t != s {
          switch n {
          case 2: // If there is no message
            dest = string(buf[t:s])
          case 3:
            theRest = string(buf[t:s])
          }
          c.handleMessage(method, source, dest, theRest)
        }
        n = 0
        b = s + 1
        t = b
      }
    }
    if b <= l {
      unprocessed = make([]byte, l-b)
      copy(unprocessed, buf[b:l])
    } else {
      unprocessed = nil
    }
  }
}

func (c *Client) handleMessage(method, source, dest, theRest string) {
  switch method {
  case "001":
    applog.Info("justinfan: connection successful")
    c.connected.Set(true)
  case "PRIVMSG":
    user := clientToUsername(source)
    if user == "jtv" {
      return
    }
    if theRest == "" || dest[0] != '#' || theRest[0] != ':' {
      return
    }
    msg := &Message{
      user,
      dest[1:],
      time.Now(),
      theRest[1:],
    }
    c.messages <- msg
  }
}

func clientToUsername(client string) string {
  var start int
  for s := 0; s < len(client); s++ {
    switch client[s] {
    case ':':
      start = s + 1
    case '!':
      return client[start:s]
    }
  }
  return client[start:]
}

func (c *Client) channelManager() {
  c.writeLine("PASS " + ircPass)
  c.writeLine("NICK " + ircUser)
  c.connected.WaitFor(true)
  disconnection := c.connected.ChanFor(false)
  for {
    select {
    case <-disconnection:
      return
    case <-c.valid.ChanFor(false):
    }

    for _, name := range c.partQueue {
      c.writeLine("PART #" + name)
      delete(c.channels, name)
    }
    for _, name := range c.joinQueue {
      c.writeLine("JOIN #" + name)
      c.channels[name] = true
    }

    c.partQueue = nil
    c.joinQueue = nil
    c.valid.Set(true)
  }
}

func (c *Client) writeLine(line string) {
  c.conn.Write([]byte(line + "\n"))
}
