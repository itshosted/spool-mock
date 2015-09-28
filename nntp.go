package main

import (
	"fmt"
	"spool-mock/config"
	"net"
	"spool-mock/client"
	"strings"
	"io"
	"net/textproto"
	"bufio"
	"bytes"
	"spool-mock/dotreader"
)

func Quit(conn *client.Conn, tok []string) {
	conn.Send("205 Bye.")
}

func Unsupported(conn *client.Conn, tok []string) {
	fmt.Println(fmt.Sprintf("WARN: C(%s): Unsupported cmd %s", conn.RemoteAddr(), tok[0]))
	conn.Send("500 Unsupported.")
}

func read(conn *client.Conn, msgid string, msgtype string) {
	var code string
	head := true
	body := true

	if msgtype == "ARTICLE" {
		code = "220"
	} else if msgtype == "HEAD" {
		code = "221"
		body = false
	} else if msgtype == "BODY" {
		code = "222"
		head = false
	} else {
		panic("Should not get here")
	}

	if msgid == "<aaa@bb.cc>" {
		conn.Send("500 msgid means fivehundred err")
		return
	}

	var raw string
	if head {
	raw += `Path: asg009!abp002.ams.xsnews.nl!abuse.newsxs.nl!not-for-mail
From: Zinitzio <x8F4zpNLByt8Vhh1hyFBTcarWqKeqTszySrxYJUNrGyj64VA761YahKczcyROsOv.N5UyksLragucHTY7hXbIf3OraQSwtjjJX6PcYubvlsh6oPDUGuY1j0b4Z7i6xnio@47a00b01.16110764.10.1443172883.1.NL.v8r0DMvyrMxvrV9wjB9RklWe-p-p1ZChfS4lxGsMNtRWMbyLXZonEJ6Lp3usHDsLnG>
Subject: Mkv Tool Nix 8.4.0 NL | Zgp
Newsgroups: free.pt
Message-ID: <pTgQyybcKwYEhIFVg2wH7@spot.net>
X-Newsreader: Spotnet 2.0.0.114
X-XML: <Spotnet><Posting><Key>7</Key><Created>1443172883</Created><Poster>Zinitzio</Poster><Tag>Zgp</Tag><Title>Mkv Tool Nix 8.4.0 NL</Title><Description>Iedere Mkv (x264) film heeft meerdere sporen. Met dit programma kun je sporen verwijderen of toevoegen. Heb je een film zonder ondertitel dan kun je die makkelijk toevoegen.[br][br]In deze spot zitten de volgende onderdelen:[br][br]Mkv Tool Nix 8.4.0</Description><Image Width='350' Height='350'><Segment>Ldqj0ABsZDMEhIFVgyrLc@spot.net</Segment></Image><Size>16110764</Size><Category>04<Sub>04a00</Sub><Sub>04b01</Sub></Category><NZB><Segment>sm0Ls136Ir4EhIFVgj4Dg@spot.net</Segment></NZB></Posting></Spotnet>
X-XML-Signature: mMXtDVvEzuAz5soJzKcpsd042VQY2M306o418-pOYtLIxv7DN5lDzAO3rB3EakfZT
X-User-Key: <RSAKeyValue><Modulus>x8F4zpNLByt8Vhh1hyFBTcarWqKeqTszySrxYJUNrGyj64VA761YahKczcyROsOv</Modulus><Exponent>AQAB</Exponent></RSAKeyValue>
X-User-Signature: N5UyksLragucHTY7hXbIf3OraQSwtjjJX6PcYubvlsh6oPDUGuY1j0b4Z7i6xnio
Content-Type: text/plain; charset=ISO-8859-1
Content-Transfer-Encoding: 8bit
X-Complaints-To: abuse@newsxs.nl
Organization: Newsxs
Date: Fri, 25 Sep 2015 11:21:23 +0200
Lines: 5
NNTP-Posting-Date: Fri, 25 Sep 2015 11:21:23 +0200`
}
if head && body {
	raw += "\n\n"
}
if body {
raw += `Iedere Mkv (x264) film heeft meerdere sporen. Met dit programma kun je sporen verwijderen of toevoegen. Heb je een film zonder ondertitel dan kun je die makkelijk toevoegen.

In deze spot zitten de volgende onderdelen:

Mkv Tool Nix 8.4.0`
}

	raw = strings.Replace(raw, "\n", "\r\n", -1)

	conn.Send(code + " " + msgid)

	if msgid == "<aab@bb.cc>" {
		// fake a broken
		conn.Send(raw[0:50])
		conn.Close()
	} else {
		conn.Send(raw)
	}
	conn.Send("\r\n.") // additional \r\n auto-added
	if msgid == "<close@bb.cc>" {
		conn.Close()
	}
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

	if b.String() != "\r\nBody.\r\nBody1\r\nBody2 ohyeay?\r\n.\r\n" {
		conn.Send("500 Body does not match hardcoded compare value.")
		return
	}
	conn.Send("240 Posted.")
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

func Group(conn *client.Conn, tok []string) {
	if len(tok) != 2 {
		conn.Send("501 Invalid syntax.")
		return
	}
	if tok[1] == "nosuch.group" {
		conn.Send("411 No such group.")
		return
	} else if tok[1] == "standard.group" {
		conn.Send("211 300007627 8974530000 9274537627 standard.group")
		return
	}

	conn.Send("501 No test for given groupname")
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
		} else if cmd == "GROUP" {
			// GROUP x
			Group(conn, tok)
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