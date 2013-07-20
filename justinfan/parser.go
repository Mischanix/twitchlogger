package justinfan

import (
  "bufio"
  "errors"
  "io"
  "time"
)

type ircMessage struct {
  method, source, dest, theRest string
  received                      time.Time
}

type ircParser struct {
  scanner *bufio.Scanner
}

func newParser(r io.Reader) *ircParser {
  return &ircParser{bufio.NewScanner(r)}
}

func (p *ircParser) parseMessage() (*ircMessage, error) {
  msg := &ircMessage{}
  msg.received = time.Now()
  ok := p.scanner.Scan()
  if !ok {
    return nil, p.scanner.Err()
  }
  line := p.scanner.Bytes()
  var s, n int
  for ; n < 3; n++ {
    advance, word, err := bufio.ScanWords(line[s:], false)
    if err != nil {
      return nil, err
    }
    s += advance
    switch n {
    case 0:
      msg.source = string(word)
    case 1:
      msg.method = string(word)
    case 2:
      msg.dest = string(word)
    }
  }
  if s < len(line) {
    msg.theRest = string(line[s:])
  }
  return msg, nil
}

func parseCommand(msg *ircMessage) (*Command, error) {
  if msg.theRest == "" || msg.theRest[0] != ':' {
    return nil, errors.New("jtv command too short")
  }
  text := []byte(msg.theRest)[1:]
  var s, n int
  cmd := &Command{}
  cmd.Received = time.Now()
  for ; n < 2; n++ {
    advance, word, err := bufio.ScanWords(text[s:], false)
    if err != nil {
      return nil, err
    }
    s += advance
    if advance == 0 {
      word = text[s:]
      s = len(text)
    }
    switch n {
    case 0:
      cmd.Command = string(word)
    case 1:
      cmd.User = string(word)
    }
    if s == len(text) { // No User
      break
    }
  }
  if s < len(text) {
    cmd.Arg = string(text[s:])
  }
  return cmd, nil
}
