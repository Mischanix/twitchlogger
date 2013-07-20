package main

import (
  "github.com/Mischanix/twitchlogger/justinfan"
)

var irc struct {
  client *justinfan.Client
}

func ircConsumer() {
  irc.client = justinfan.Connect()
  messages := irc.client.Messages()
  commands := irc.client.Commands()
  for {
    select {
    case msg := <-messages:
      db.msgBuffer.Add(msg)
    case cmd := <-commands:
      db.cmdBuffer.Add(cmd)
    }
  }
}

func updateChannels(channels []string) {
  irc.client.SetChannels(channels)
}
