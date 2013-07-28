package main

import (
  "github.com/Mischanix/twitchlogger/justinfan"
  "github.com/Mischanix/wait"
  "time"
)

var irc struct {
  client      *justinfan.Client
  lastMessage map[string]*justinfan.Message
  userSignals map[string]*wait.Flag
}

func ircConsumer() {
  irc.client = justinfan.Connect()
  irc.lastMessage = make(map[string]*justinfan.Message)
  irc.userSignals = make(map[string]*wait.Flag)
  go func() {
    messages := irc.client.Messages()
    for {
      processMessage(<-messages)
    }
  }()
  go func() {
    commands := irc.client.Commands()
    for {
      go processCommand(<-commands)
    }
  }()
  go func() {
    for {
      <-time.After(5 * time.Minute)
      for user, msg := range irc.lastMessage {
        if time.Now().Sub(msg.Received) > 5*time.Minute {
          delete(irc.lastMessage, user)
        }
      }
    }
  }()
}

func updateChannels(channels []string) {
  irc.client.SetChannels(channels)
}

func processMessage(message *justinfan.Message) {
  db.msgBuffer.Add(message)

  irc.lastMessage[message.User] = message
  if signal, ok := irc.userSignals[message.User]; ok {
    signal.Set(true)
  } else {
    irc.userSignals[message.User] = wait.NewFlag(true)
  }
  // Hold the line open for a reasonable time period.  Since processCommand is
  // used as a goroutine while processMessage is used inline, messages are
  // often processed before the commands are processed, even though the
  // commands were parsed and reached ircConsumer first.  It's also reasonable
  // to believe that a human user will not speak in two different channels
  // within 100ms.
  go func() {
    <-time.After(100 * time.Millisecond)
    delete(irc.userSignals, message.User)
  }()
}

func processCommand(cmd *justinfan.Command) {
  if precedesMessage(cmd) {
    if signal, ok := irc.userSignals[cmd.User]; ok {
      signal.WaitFor(true)
    } else {
      signal = wait.NewFlag(false)
      irc.userSignals[cmd.User] = signal
      select {
      case <-signal.ChanFor(true):
      // Timeout needed for specialuser staff or admin, sub-only mode, channel
      // parts, and randomly.
      case <-time.After(10 * time.Second):
        signal.Set(true)
        delete(irc.userSignals, cmd.User)
      }
    }
  }

  // SPECIALUSER moderator is sent before and in addition to staff/admin.
  if cmd.Command == "SPECIALUSER" &&
    (cmd.Arg == "moderator" || cmd.Arg == "staff" || cmd.Arg == "admin") {
  } else {
    if msg, ok := irc.lastMessage[cmd.User]; ok {
      cmd.Channel = msg.Channel
    }
  }
  insertCommand(cmd)
}

func precedesMessage(cmd *justinfan.Command) bool {
  if cmd.Command == "SPECIALUSER" &&
    (cmd.Arg == "moderator" || cmd.Arg == "staff" || cmd.Arg == "admin") {
    return false
  }
  if cmd.Command == "CLEARCHAT" {
    return false
  }
  return true
}
