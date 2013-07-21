package main

import (
  "github.com/Mischanix/twitchlogger/justinfan"
)

var irc struct {
  client      *justinfan.Client
  lastChannel map[string]string
}

func ircConsumer() {
  irc.client = justinfan.Connect()
  irc.lastChannel = make(map[string]string)
  messages := irc.client.Messages()
  commands := irc.client.Commands()
  for {
    select {
    case msg := <-messages:
      irc.lastChannel[msg.User] = msg.Channel
      db.msgBuffer.Add(msg)
    case cmd := <-commands:
      go processCommand(cmd)
    }
  }
}

func updateChannels(channels []string) {
  irc.client.SetChannels(channels)
}
