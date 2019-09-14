package smug

import (
    "crypto/tls"
    "fmt"
    "time"

    libirc "github.com/thoj/go-ircevent"
)

type IrcBroker struct {
    conn *libirc.Connection
    channel string
    nick string
    prefix string
    server string
}

func (ib *IrcBroker) Name() string {
    return fmt.Sprintf("irc-%s-%s-as-%s", ib.server, ib.channel, ib.nick)
}

// args [server, channel, nick]
func (ib *IrcBroker) Setup(args ...string) {
    ib.server = args[0]
    ib.channel = args[1]
    ib.nick = args[2]
    ib.conn = libirc.IRC(ib.nick, "smug")
    ib.conn.VerboseCallbackHandler = true
    ib.conn.UseTLS = true  // XXX should be a param
    if ib.conn.UseTLS {
        ib.conn.TLSConfig = &tls.Config{InsecureSkipVerify: true} // XXX
    }
    ib.conn.AddCallback(
        "001",
        func(e *libirc.Event) {
            ib.conn.Join(ib.channel)
            ib.Put("hi. sup?")
        } )
    // ib.conn.AddCallback("366", func(e *irc.Event) { }) // ignore end of names
    err := ib.conn.Connect(ib.server)
    if err != nil {
        fmt.Printf("ERR %s", err)
        ib.conn = nil // error'd here, set this connection to nil XXX
    }
}


func (ib *IrcBroker) Put(msg string) {
    ib.conn.Privmsg(ib.channel, msg)
}


func (ib *IrcBroker) Publish(ev *Event) {
    ib.Put(fmt.Sprintf("|%s| %s", ev.Nick, ev.Text))
}


func (ib *IrcBroker) Run(dis Dispatcher) {
    // XXX this should ensure some sort of singleton to ensure Run should only
    // ever be called once...
    ib.conn.AddCallback("PRIVMSG", func (e *libirc.Event) {
        ev := &Event{
            Origin: ib,
            Nick: e.Nick,
            Text: e.Message(),
            ts: time.Now(),
        }
        dis.Broadcast(ev)
    })
}

