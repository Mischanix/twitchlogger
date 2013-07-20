package justinfan

import (
  "fmt"
  "github.com/Mischanix/applog"
  "github.com/Mischanix/wait"
  "os"
  "testing"
  "time"
)

var stopped = wait.NewFlag(false)

func init() {
  applog.SetOutput(os.Stdout)
  applog.Level = applog.DebugLevel
}

func TestConnect(t *testing.T) {
  c := Connect()
  c.SetChannels([]string{"riotgames"})
  go func() {
    stopChan := stopped.ChanFor(true)
    msgChan := c.Messages()
    cmdChan := c.Commands()
    for {
      select {
      case msg := <-msgChan:
        fmt.Println(msg)
      case cmd := <-cmdChan:
        fmt.Println(cmd)
      case <-stopChan:
        return
      }
    }
  }()
  <-time.After(10 * time.Second)
  c.conn.Close()
  <-time.After(10 * time.Second)
  c.Disconnect()
  // Ensure silence
  <-time.After(10 * time.Second)
  stopped.Set(true)
}
