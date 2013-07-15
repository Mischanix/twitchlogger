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

func ExampleConnect(t *testing.T) {
  c := Connect()
  c.SetChannels([]string{"sodapoppin"})
  go func() {
    stopChan := stopped.ChanFor(true)
    msgChan := c.Messages()
    for {
      select {
      case m := <-msgChan:
        fmt.Println(m)
      case <-stopChan:
        return
      }
    }
  }()
  <-time.After(5 * time.Second)
  c.SetChannels([]string{"reckful"})
  <-time.After(5 * time.Second)
  c.Disconnect()
  stopped.Set(true)
}
