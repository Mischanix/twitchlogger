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
  go func() {
    for {
      msg := <-messages
      irc.lastChannel[msg.User] = msg.Channel
      db.msgBuffer.Add(msg)
    }
  }()
  go func() {
    for {
      processCommand(<-commands)
    }
  }()
}

func updateChannels(channels []string) {
  irc.client.SetChannels(channels)
}
