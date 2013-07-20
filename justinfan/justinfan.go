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

// Command is a command sent by Jtv such as USERCOLOR or CLEARCHAT
type Command struct {
  User     string
  Received time.Time
  Command  string
  Arg      string
}

type Client struct {
  conn      net.Conn
  parser    *ircParser
  channels  map[string]bool
  joinQueue []string
  partQueue []string
  connected *wait.Flag
  valid     *wait.Flag
  messages  chan *Message
  commands  chan *Command
}

// Connect creates a new client and begins and authenticates the IRC session.
func Connect() *Client {
  c := &Client{
    nil, nil,
    make(map[string]bool),
    nil, nil,
    wait.NewFlag(false),
    wait.NewFlag(true),
    make(chan *Message, 64),
    make(chan *Command, 64),
  }
  c.connectionManager()
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

func (c *Client) Commands() <-chan *Command {
  return (<-chan *Command)(c.commands)
}

func (c *Client) readHandler() {
  for c.conn != nil {
    ircMsg, err := c.parser.parseMessage()
    if err != nil {
      applog.Error("justinfan.read error: %v", err)
      c.connected.Set(false)
      return
    }
    c.handleMessage(ircMsg)
  }
}

func (c *Client) handleMessage(msg *ircMessage) {
  if msg == nil {
    return
  }
  switch msg.method {
  case "001":
    applog.Info("justinfan: connection successful")
    c.connected.Set(true)
  case "PRIVMSG":
    user := clientToUsername(msg.source)
    if user == "jtv" {
      cmd, err := parseCommand(msg)
      if err != nil {
        applog.Error("justinfan: error parsing jtv command: %v", err)
      }
      if cmd.Command == "HISTORYEND" {
        break
      }
      c.commands <- cmd
    } else {
      if msg.theRest == "" || msg.theRest[0] != ':' || msg.dest[0] != '#' {
        break
      }
      privmsg := &Message{
        user,
        msg.dest[1:],
        msg.received,
        msg.theRest[1:],
      }
      c.messages <- privmsg
    }
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

func (c *Client) connectionManager() {
  conn, err := net.Dial("tcp", ircDial)
  if err != nil {
    applog.Error("justinfan.Connect failed: %v", err)
  }
  c.conn = conn
  c.parser = newParser(c.conn)
}

func (c *Client) channelManager() {
  c.writeLine("TWITCHCLIENT 2")
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
