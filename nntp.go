package main

import (
	"fmt"
	"spool-mock/config"
	"net"
	"spool-mock/client"
	"strings"
	"spool-mock/db"
	"io"
	"net/textproto"
	"bufio"
	"bytes"
	"spool-mock/dotreader"
	//"spool-mock/headreader"
	//"spool-mock/bodyreader"
)

func Quit(conn *client.Conn, tok []string) {
	conn.Send("205 Bye.")
}

func Unsupported(conn *client.Conn, tok []string) {
	fmt.Println(fmt.Sprintf("WARN: C(%s): Unsupported cmd %s", conn.RemoteAddr(), tok[0]))
	conn.Send("500 Unsupported.")
}

func read(conn *client.Conn, msgid string, msgtype string) {
	read, usrErr, sysErr := db.Read(
		db.ReadInput{Msgid: msgid[1:len(msgid)-1], Type: msgtype},
	)
	if sysErr != nil {
		fmt.Println("WARN: " + sysErr.Error())
		conn.Send("500 Failed processing")
		return
	}
	if usrErr != nil {
		conn.Send("400 " + usrErr.Error())
		return
	}
	defer read.Close()

	var code string
	if msgtype == "ARTICLE" {
		code = "220"
	} else if msgtype == "HEAD" {
		code = "221"
	} else if msgtype == "BODY" {
		code = "222"
	} else {
		panic("Should not get here")
	}

	conn.Send(code + " " + msgid)
	if _, e := io.Copy(conn.GetWriter(), read); e != nil {
		fmt.Println("WARN: " + e.Error())
		conn.Send("500 Failed forwarding")
		return
	}
	conn.Send("\r\n.") // additional \r\n auto-added
}

func PostArticle(conn *client.Conn) {
	conn.Send("340 Start posting.")

	b := new(bytes.Buffer)
	br := bufio.NewReader(conn.GetReader())
	r := textproto.NewReader(br)

	fmt.Println("PostArticle head.")
	m, e := r.ReadMIMEHeader()
	if e != nil {
		conn.Send("440 Failed reading header")
		return
	}

	fmt.Println("PostArticle body.")
	if _, e := io.Copy(b, dotreader.New(br)); e != nil {
		conn.Send("440 Failed reading body")
		return
	}

	if val := m.Get("X-Accept"); val == "DENY" {
		conn.Send("440 Deny test.")
		return
	}

	// 	b.String()
	if b.String() != "Body.\r\nBody1\r\nBody2 ohyeay?\r\n" {
		conn.Send("500 Body broken?")
		return
	}
	conn.Send("240 Posted " + m.Get("X-MsgId"))

	/*buf := new(bytes.Buffer)
	br := dotreader.New(conn.GetReader())

	// TODO: Unsafe copy ALL
	fmt.Println("PostArticle read.")
	if _, e := io.Copy(buf, br); e != nil {
		conn.Send("500 Copy error.")
		return
	}
	fmt.Println("PostArticle parse.")
	headBody := strings.SplitN(buf.String(), "\r\n\r\n", 2)

	r := textproto.NewReader(bufio.NewReader(strings.NewReader(headBody[0])))
	m, e := r.ReadMIMEHeader()
	if e != nil {
		conn.Send("441 Failed reading header")
		return
	}
	if val := m.Get("X-Accept"); val == "DENY" {
		conn.Send("441 Deny test.")
		return
	}
	fmt.Println("PostArticle body.")

	if headBody[1] != "Body.\r\nBody1\r\nBody2 ohyeay?\r\n" {
		conn.Send("500 Body broken?")
		return
	}

	conn.Send("240 Posted " + m.Get("X-MsgId"))*/
}

func Article(conn *client.Conn, tok []string) {
	if len(tok) != 2 {
		conn.Send("501 Invalid syntax.")
		return
	}
	read(conn, tok[1], "ARTICLE")
}

func Head(conn *client.Conn, tok []string) {
	if len(tok) != 2 {
		conn.Send("501 Invalid syntax.")
		return
	}
	read(conn, tok[1], "HEAD")
}

func Body(conn *client.Conn, tok []string) {
	if len(tok) != 2 {
		conn.Send("501 Invalid syntax.")
		return
	}
	read(conn, tok[1], "BODY")
}

func req(conn *client.Conn) {
	conn.Send("200 StoreD")
	for {
		tok, e := conn.ReadLine()
		if e != nil {
			fmt.Println(fmt.Sprintf("WARN: C(%s): %s", conn.RemoteAddr(), e.Error()))
			break
		}

		cmd := strings.ToUpper(tok[0])
		if cmd == "QUIT" {
			Quit(conn, tok)
			break
		} else if cmd == "ARTICLE" {
			Article(conn, tok)
		} else if cmd == "HEAD" {
			Head(conn, tok)
		} else if cmd == "BODY" {
			Body(conn, tok)
		} else if cmd == "AUTHINFO" {
			sub := strings.ToUpper(tok[1])
			if sub == "USER" {
				conn.Send("381 Need more.")
			} else if sub == "PASS" {
				if tok[2] == "test" {
					conn.Send("281 Authentication accepted.")
				}
			}
		} else if cmd == "NOOP" {
			conn.Send("500 Unsupported.")
		} else if cmd == "POST" {
			PostArticle(conn)
		} else {
			Unsupported(conn, tok)
			break
		}
	}

	conn.Close()
	if config.Verbose {
		fmt.Println(fmt.Sprintf("C(%s) Closed", conn.RemoteAddr()))
	}
}

func nntpListen(listen string) error {
	sock, err := net.Listen("tcp", listen)
	if err != nil {
		return err
	}
	if config.Verbose {
		fmt.Println("nntpd listening on " + listen)
	}

	for {
		conn, err := sock.Accept()
		if err != nil {
			panic(err)
		}
		if config.Verbose {
			fmt.Println(fmt.Sprintf("C(%s) New", conn.RemoteAddr()))
		}

		go req(client.New(conn))
	}
}