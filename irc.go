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
  for {
    db.msgBuffer.Add(<-messages)
  }
}

func updateChannels(channels []string) {
  irc.client.SetChannels(channels)
}
