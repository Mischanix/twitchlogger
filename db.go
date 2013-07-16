package main

import (
  "github.com/Mischanix/applog"
  "github.com/Mischanix/batcher"
  "github.com/Mischanix/twitchlogger/justinfan"
  "github.com/Mischanix/wait"
  "labix.org/v2/mgo"
)

var db struct {
  ready        *wait.Flag
  session      *mgo.Session
  database     *mgo.Database
  msgColl      *mgo.Collection
  statusColl   *mgo.Collection
  messages     *dbDocs
  statuses     *dbDocs
  msgBuffer    *batcher.Buffer
  statusBuffer *batcher.Buffer
}

func dbClient() {
  var err error
  db.ready = wait.NewFlag(false)
  if db.session, err = mgo.Dial(config.DbUrl); err != nil {
    applog.Error("mgo.Dial failure: %v", err)
  }
  db.database = db.session.DB(config.DbName)
  db.msgColl = db.database.C(config.MsgCollection)
  db.statusColl = db.database.C(config.StatusCollection)
  db.messages = &dbDocs{nil}
  db.statuses = &dbDocs{nil}
  db.msgBuffer = batcher.New(db.messages, batcher.ElementCountThreshold)
  db.statusBuffer = batcher.New(db.statuses, batcher.Manual)
  db.msgBuffer.SetThreshold(128)

  // Flush docs before exit
  stopped.Add(1)
  stopping.Once(true, func() {
    db.msgBuffer.Flush()
    db.statusBuffer.Flush()
    stopped.Done()
  })
  db.ready.Set(true)
}

type statusDoc struct {
  channel string
  status  string
  users   []string
  viewers int
}

type dbDocs struct {
  docs []interface{}
}

func (d *dbDocs) Add(e interface{}) {
  d.docs = append(d.docs, e)
}

func (d *dbDocs) Flush() {
  if d.docs == nil || len(d.docs) == 0 {
    return
  }
  var coll *mgo.Collection
  if _, ok := d.docs[0].(*justinfan.Message); ok {
    coll = db.msgColl
  } else if _, ok := d.docs[0].(*statusDoc); ok {
    coll = db.statusColl
  } else {
    panic("db: Flush called on docs of unknown type")
  }
  coll.Insert(d.docs...)
  applog.Debug("Wrote %d docs", d.Count())
  d.docs = nil
}

func (d *dbDocs) Count() int {
  return len(d.docs)
}
