package main

import (
  "github.com/Mischanix/applog"
  "net/url"
  "strconv"
  "time"
)

type streamsResponse struct {
  Streams []struct {
    Channel struct {
      Status string
      Name   string
    }
    Viewers int
  }
}

type streamStatus struct {
  Status  string
  Viewers int
}

var streams struct {
  status map[string]streamStatus
  list   []string
}

func updateStreams() {
  streams.list = nil
  if streams.status == nil {
    streams.status = make(map[string]streamStatus)
  }
  // anti-dos limits
  limit := 20
  count := 0
  var offset int
  done := false
  for !done {
    var curr streamsResponse
    if count > limit {
      break
    }
    opts := url.Values{}
    opts.Add("limit", "100")
    opts.Add("offset", strconv.Itoa(offset))
    err := apiGet("streams", opts, &curr)
    count++
    if err != nil {
      applog.Error("updateStreams.err: %v", err)
      <-time.After(5 * time.Second)
      continue
    }

    for _, stream := range curr.Streams {
      if stream.Viewers < config.ViewerThreshold {
        done = true
        break
      }
      streamName := stream.Channel.Name
      streams.status[streamName] = streamStatus{
        stream.Channel.Status,
        stream.Viewers,
      }
      streams.list = append(streams.list, streamName)
    }
    <-time.After(1 * time.Second)
    offset += 100
  }

  for name, _ := range streams.status {
    found := false
    for _, n := range streams.list {
      if name == n {
        found = true
        break
      }
    }
    if !found {
      delete(streams.status, name)
    }
  }

  applog.Info(
    "streams.updateStreams: Got %d streams / %d results from kraken",
    len(streams.status), len(streams.list),
  )
  updateChannels(streams.list)
}

func updateStatuses() {
  for name, status := range streams.status {
    db.statusBuffer.Add(&statusDoc{
      name,
      time.Now(),
      status.Status,
      status.Viewers,
    })
  }
  db.statusBuffer.Flush()
}
