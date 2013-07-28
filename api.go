package main

import (
  "encoding/json"
  "github.com/Mischanix/applog"
  "net/http"
  "net/url"
  "time"
)

var (
  apiBase        = "https://api.twitch.tv/kraken/"
  updateInterval time.Duration
)

func apiConsumer() {
  updateInterval = time.Duration(config.UpdateInterval) * time.Millisecond

  go updateStreams()
  go ircConsumer()
  go statusConsumer()
  for {
    <-time.After(updateInterval)
    go updateStreams()
  }
}

func apiGet(path string, opts url.Values, result interface{}) error {
  req, err := http.NewRequest("GET", apiBase+path+"?"+opts.Encode(), nil)
  if err != nil {
    applog.Error("api.get: request error: %v", err)
    return err
  }
  req.Header.Add("Accept", "application/vnd.twitchtv.v3+json")
  resp, err := http.DefaultClient.Do(req)
  if err != nil {
    applog.Error("api.get: response error: %v", err)
    return err
  }
  err = json.NewDecoder(resp.Body).Decode(result)
  if err != nil {
    applog.Error("api.get: json decode error: %v", err)
    return err
  }
  err = resp.Body.Close()
  if err != nil {
    applog.Error("api.get: body.close error: %v", err)
    return err
  }
  return nil
}

func statusConsumer() {
  // Stagger statuses so we don't report empty userlists
  <-time.After(updateInterval / 2)
  go updateStatuses()
  for {
    <-time.After(updateInterval)
    go updateStatuses()
  }
}
