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
  c.SetChannels([]string{"wcs_gsl"})
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
  fmt.Println(len(c.Users("wcs_gsl")))
  c.SetChannels([]string{"wcs_osl"})
  <-time.After(5 * time.Second)
  fmt.Println(len(c.Users("wcs_osl")))
  c.Disconnect()
  stopped.Set(true)
}
