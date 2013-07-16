package main

import (
  "github.com/Mischanix/applog"
  "github.com/Mischanix/evconf"
  "github.com/Mischanix/wait"
  "os"
  "os/signal"
  "sync"
  "syscall"
)

var ready = wait.NewFlag(false)
var stopping = wait.NewFlag(false)
var stopped = &sync.WaitGroup{}

var logStdout = false

var config struct {
  DbUrl            string `json:"db_url"`
  DbName           string `json:"db_name"`
  MsgCollection    string `json:"msg_collection"`
  StatusCollection string `json:"status_collection"`
  UpdateInterval   int    `json:"update_interval"`  // in milliseconds
  ViewerThreshold  int    `json:"viewer_threshold"` // lower bound
}

func defaultConfig() {
  config.DbUrl = "localhost"
  config.DbName = "tl-dev"
  config.MsgCollection = "messages"
  config.StatusCollection = "statuses"
  config.UpdateInterval = 1000 * 60 * 5 // 5m
  config.ViewerThreshold = 100
}

func main() {
  // Termination setup
  go func() {
    c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGTERM)
    signal.Notify(c, os.Interrupt)
    <-c
    stopping.Set(true)
  }()
  go func() { // For Windows
    if _, err := os.Stdin.Read(nil); err == nil {
      stopping.Set(true)
    }
  }()

  // Log setup
  applog.Level = applog.InfoLevel
  if logStdout {
    applog.SetOutput(os.Stdout)
  } else {
    if logFile, err := os.OpenFile(
      "twitchlogger.log",
      os.O_WRONLY|os.O_CREATE|os.O_APPEND,
      os.ModeAppend|0666,
    ); err != nil {
      applog.SetOutput(os.Stdout)
      applog.Error("Unable to open log file: %v", err)
    } else {
      applog.SetOutput(logFile)
    }
  }
  applog.Info("starting...")

  // Config setup
  conf := evconf.New("twitchlogger.json", &config)
  conf.OnLoad(func() {
    ready.Set(true)
  })
  conf.StopWatching()
  defaultConfig()
  go func() {
    conf.Ready()
  }()

  ready.WaitFor(true)

  // Application
  go dbClient()
  go apiConsumer()
  stopping.WaitFor(true)
  applog.Info("exiting...")
  stopped.Wait()
}
