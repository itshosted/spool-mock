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
	var code, msgtop string
	head := true
	body := true

	if msgid[0] == '<' {
		msgtop = "0 " + msgid
	} else {
		msgtop = msgid + " " + "<aac@bb.cc>"
	}

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

	if msgid == "<aaa@bb.cc>" || msgid == "123" {
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

	conn.Send(code + " " + msgtop)

	if msgid == "<aab@bb.cc>" || msgid == "124" {
		// fake a broken
		conn.Send(raw[0:50])
		conn.Close()
	} else {
		conn.Send(raw)
	}
	conn.Send("\r\n.") // additional \r\n auto-added
	if msgid == "<close@bb.cc>" || msgid == "500" {
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

func Xover(conn *client.Conn, tok []string) {
	// xover 7824800-7824865
	if len(tok) != 2 {
		conn.Send("501 Invalid syntax.")
		return
	}
	if tok[1] == "7824800-7824826" {
		conn.Send("224 Overview follows.")
		raw := `7824800	ABC 123 | Me	Name <qmumrmAB8Q8CxnO8j-smpGa1vPJ-sTcVrr6oHIfGkfcd7vF6o92vjRbUWz0fREIBxd.megzPlvpuGIkfVnOLYp6Uu78uzd5l28c5tl-shqPgHUtjWRRpiOnix4XZDrXYq0lI@17a00d75b03c10d23d85z03.3400118167.10.1443520028.1.NL.Gx0rio4h-sMFXi6sHL3CgB4t-sRFz0-sCaaQ2-slCQ33Xny4Ervrh87mtiE7kVh9TmwF>	Tue, 29 Sep 2015 09:47:08 GMT	<WKueCVYSDKcF14KVg8Qse@spot.net>		3017	11	Xref: artnum free.pt:7824800
7824801	A b's Z Abc - Person | hotmama	hotmama <rbsHeWOltXSPohqOb5cK1bo7e0aKcDewG5MLp3LORmP2CCPuszZUvMNSrKoQYlkuXbqCfjlwvDobYaSfBhz-pdGI6IARjweEeXSJmjIpXQKkQ-pJF1hq-pbl6x85gQPuVp0MTSjS91emBCTCsgsFHXJAUA-p3gR-pgne42u94Dy421eM=.oFSFmeMqBIP-s6I6JnNAXmhepNSDExw662pJG-pfiaubqB2ED5ZPkGafJXljSGmC2xOE8KrmkJQJU7ZLR8wp9f2TWnoCI63fJ8aQw2G-sXVnmsvlS	29 Sep 2015 09:50:42 GMT	<ck5tbXJUWjY3hAHLlvHATfU4308@spot.net>		3154	9	Xref: artnum free.pt:7824801
7824802	Text - Name LName - Derp derp | derp	Derp45 <vOD8F13AlBel-sUwaD1PMGBqs10-pSgwRn4e2-sqa3nTto9M1Go-sfBv4DPzy9ByTbBZ.fSP3JtI-sWEEC5zWKcU9d28IQdNhx08SkdN-sV0YqNhL4QrN2P1iFVjieJSO9-s44il@17a01d23d75b03c10z03.336868165.10.1443520400.1.NL.fMgLmXnHp5oQRTDLq2nLSZnaiGwjR1rxPMpOUPePBEjo6mzYjb-s0-smw32MEUCQAR>	Tue, 29 Sep 2015 11:53:47 +0200	<KPuPKc4hViAj18KVgU92q@spot.net>		1801	2	Xref: artnum free.pt:7824802
7824803	[Subj] My Sister's Text Text - Text Text | hotmama	hotmama <rbsHeWOltXSPohqOb5cK1bo7e0aKcDewG5MLp3LORmP2CCPuszZUvMNSrKoQYlkuXbqCfjlwvDobYaSfBhz-pdGI6IARjweEeXSJmjIpXQKkQ-pJF1hq-pbl6x85gQPuVp0MTSjS91emBCTCsgsFHXJAUA-p3gR-pgne42u94Dy421eM=.RL6EsiNtP5kxupOlUPV9eVf40G5Qe5QazLoj35wO6YRXf8jsqpRejBGjeiE5Y5d1hEDvtlG-shsIW8ukrnTKdYctQ9FWaUh2JV8CjLstLAQOTjt0T	29 Sep 2015 09:49:28 GMT	<SG84aTE0ZWdjFpC3SuuDoPw1NLR@spot.net>		3086	9	Xref: artnum free.pt:7824803
7824804	Text - AB ABCDEFGG | hotmama	hotmama <rbsHeWOltXSPohqOb5cK1bo7e0aKcDewG5MLp3LORmP2CCPuszZUvMNSrKoQYlkuXbqCfjlwvDobYaSfBhz-pdGI6IARjweEeXSJmjIpXQKkQ-pJF1hq-pbl6x85gQPuVp0MTSjS91emBCTCsgsFHXJAUA-p3gR-pgne42u94Dy421eM=.jVMWwPY9TkbBFu8DOQeBmSrJU5F3RMrEe5qxzxhOqBEg9FUIQMVWaXBMndMAIMrSHXwLX7Vp0cL9yuoD7l1kJPsGM1l-sIoH4ajIixcCMOaVCsnwa	29 Sep 2015 09:47:56 GMT	<cFNFWGoyOXh1EagYvJjQDCNbp8Q@spot.net>		3003	7	Xref: artnum free.pt:7824804
7824806	Real Text Stories - Person One & Person Two | hotmama	hotmama <rbsHeWOltXSPohqOb5cK1bo7e0aKcDewG5MLp3LORmP2CCPuszZUvMNSrKoQYlkuXbqCfjlwvDobYaSfBhz-pdGI6IARjweEeXSJmjIpXQKkQ-pJF1hq-pbl6x85gQPuVp0MTSjS91emBCTCsgsFHXJAUA-p3gR-pgne42u94Dy421eM=.RPErjYrF3WMgc3iQ1nFsAkuLbTutxtAKAydRSEOl6YJxc2DVxUeX3lDdhS4oPzcwI-pZzz7VERwqOAGm2zXGkuprWjqVBnBbYWvcx8vt8X5DePkX-	29 Sep 2015 09:52:56 GMT	<RHJYc0NyS2dFO4EOGokhS3ChEZY@spot.net>		3264	9	Xref: artnum free.pt:7824806
7824807	[Subj] Hard Text - Person Derp | hotmama	hotmama <rbsHeWOltXSPohqOb5cK1bo7e0aKcDewG5MLp3LORmP2CCPuszZUvMNSrKoQYlkuXbqCfjlwvDobYaSfBhz-pdGI6IARjweEeXSJmjIpXQKkQ-pJF1hq-pbl6x85gQPuVp0MTSjS91emBCTCsgsFHXJAUA-p3gR-pgne42u94Dy421eM=.a7wnSMEyBU-pXA1Esw45rd0stA4rd2FOD3w-sXb5O0NbMhrgoaaMpXbYjSTAAHBH0JXVlABT7KpgtmDr5c5NXyyMZRakDm6CAxW1eu-pLOEYkyZcE	29 Sep 2015 09:54:26 GMT	<REtUdXZ1OFlaKfgH9yu1BtA8UDG@spot.net>		3263	9	Xref: artnum free.pt:7824807
7824808	Super cool derp S01E02 720p TestLip H264 | Bassie10	Bassie10 <3hKtBrWCqFv055OmdF25pEaWgNjp0yG7EwB-sm1ivvhGXJ8I0zP7AHMgTkrrv7lg3.IlaGikrm4FZ4eh0kmtXaAczBd-pmBGM97fvDu9ht2NUVCoFLhHanOvrSJFw4HbE02@17a09b04d11c00c10d06z01.726336313.10.1443520698.1.NL.J2l0mXvqRhgIw1rbTNistDVg4tNdTagZrw3SC0VHb9Rp-sAdxAxO2Epi4wnDcyLZk>	29 Sep 2015 09:58:16 GMT	<fDh9DBtij5oumAKVgnxVl@spot.net>		6311	3	Xref: artnum free.pt:7824808
7824809	[WOUW] MOUW Derp Herp. 13 | hotmama	hotmama <rbsHeWOltXSPohqOb5cK1bo7e0aKcDewG5MLp3LORmP2CCPuszZUvMNSrKoQYlkuXbqCfjlwvDobYaSfBhz-pdGI6IARjweEeXSJmjIpXQKkQ-pJF1hq-pbl6x85gQPuVp0MTSjS91emBCTCsgsFHXJAUA-p3gR-pgne42u94Dy421eM=.bMMK3byHT92HCeARYcOn9p23oyfi7JoPjmTvLqyAqkPZAsCdctgb3iyZQJ5xgXmXP2A0l6XN5cSQDVNFFqHe08Fpvw-s4AC-pi9b62G9CiAJGrSK3	29 Sep 2015 09:59:19 GMT	<R2o3WE5ROWVqAX3DOiNyfq46GkC@spot.net>		3163	9	Xref: artnum free.pt:7824809
7824810	Random subject here [Spoil HD] | hoil	La4444 <vOD8F13AlBel-sUwaD1PMGBqs10-pSgwRn4e2-sqa3nTto9M1Go-sfBv4DPzy9ByTbBZ.pq1WzTNmxhv56FLTPkQfFb5pjzaOKazJoW2MAwEJeanimtosyyC-pvjumQkSzmSHE@17a01d23d75b03c10z03.427711992.10.1443521055.1.NL.BSz6C4oIqFXIXpgFBkgS-s5ND5SG9bsC8TAVZGZxVQUpGeVlx0Ol-sc8flUoAb-p6GD>Tue, 29 Sep 2015 12:04:42 +0200	<aRTJQ1S78V4HGIKVgaZ5y@spot.net>		1971	2	Xref: artnum free.pt:7824810
7824811	[Subj] It's A abc thing #200 | hotmama	hotmama <rbsHeWOltXSPohqOb5cK1bo7e0aKcDewG5MLp3LORmP2CCPuszZUvMNSrKoQYlkuXbqCfjlwvDobYaSfBhz-pdGI6IARjweEeXSJmjIpXQKkQ-pJF1hq-pbl6x85gQPuVp0MTSjS91emBCTCsgsFHXJAUA-p3gR-pgne42u94Dy421eM=.eKMfrw-pjhYj9GI4uFp6DlGlqX4U-s3X2iNA3Y1Ew8qSOdgkvPwMRkacSsgu3MYem35A2ef3EEKEGd8DPah1lrU5p4wOahShnsjUpxmDx9US66qJZ	29 Sep 2015 10:02:54 GMT	<SUxMM0NqOWdnVsITT3eCBtTVqyR@spot.net>		4117	14	Xref: artnum free.pt:7824811
7824812	[Hurts] The real text returns | hotmama	hotmama <rbsHeWOltXSPohqOb5cK1bo7e0aKcDewG5MLp3LORmP2CCPuszZUvMNSrKoQYlkuXbqCfjlwvDobYaSfBhz-pdGI6IARjweEeXSJmjIpXQKkQ-pJF1hq-pbl6x85gQPuVp0MTSjS91emBCTCsgsFHXJAUA-p3gR-pgne42u94Dy421eM=.Wi4g51cLUeh2PLIA9Fn-pt0KR4R-sr7ODCqXP-pvRrs-p9I5fnXuJ1WouHbYsaMty37lSJvcfIQq83nf89q6KXf45gFhqbR69-sExD-sua4DcRs8v	29 Sep 2015 10:01:05 GMT	<SEZORlRCTTJoPKz5T7mFvEe455l@spot.net>		4967	13	Xref: artnum free.pt:7824812
7824813	Pain Gain (Herp Derp Productions) | hotmama	hotmama <rbsHeWOltXSPohqOb5cK1bo7e0aKcDewG5MLp3LORmP2CCPuszZUvMNSrKoQYlkuXbqCfjlwvDobYaSfBhz-pdGI6IARjweEeXSJmjIpXQKkQ-pJF1hq-pbl6x85gQPuVp0MTSjS91emBCTCsgsFHXJAUA-p3gR-pgne42u94Dy421eM=.rZKeJCxvWgTBBAXtV4HdpjsO-p3i2jVjQPASuZMZvrezlPKfsAz4cxrbTIk8iGi3lO-p5TBVGLv2-pChAkT8HudmCEDMWkB4be4R9da5JXMJRsdwX	29 Sep 2015 10:06:40 GMT	<enFqSGtyb1U5L0Lpu3ebdSYmGXB@spot.net>		4343	14	Xref: artnum free.pt:7824813
7824814	More random text | hotmama	hotmama <rbsHeWOltXSPohqOb5cK1bo7e0aKcDewG5MLp3LORmP2CCPuszZUvMNSrKoQYlkuXbqCfjlwvDobYaSfBhz-pdGI6IARjweEeXSJmjIpXQKkQ-pJF1hq-pbl6x85gQPuVp0MTSjS91emBCTCsgsFHXJAUA-p3gR-pgne42u94Dy421eM=.oyXA0Dmlc79hkshPH7SavGLLE74nuIcVJnge1OfqKeqtDUhkfpRGNFed-st6urYNglH3887E60A6QZ3-shSASHmhfU4BKw-s4Bzjg9qWoZ4uqj0M-	29 Sep 2015 10:07:27 GMT	<N0tPaU9NUmphQptrnd2xxxciWdG@spot.net>		3919	13	Xref: artnum free.pt:7824814
7824815	Text random (alive) | hotmama	hotmama <rbsHeWOltXSPohqOb5cK1bo7e0aKcDewG5MLp3LORmP2CCPuszZUvMNSrKoQYlkuXbqCfjlwvDobYaSfBhz-pdGI6IARjweEeXSJmjIpXQKkQ-pJF1hq-pbl6x85gQPuVp0MTSjS91emBCTCsgsFHXJAUA-p3gR-pgne42u94Dy421eM=.Tmg6LsfOzQV9ABQ-sMGpJjzv228lgr-pg-sdWPkK9jmNamZVOLLTOfRqFqSXrX1WmL6lMHNVsrwrLkuI4Sg74I-pt8NH-pPHP0tQajCfoLdWGt3BN	29 Sep 2015 10:07:53 GMT	<cXV1Y3NSNk5HL112F1VFeLCTS54@spot.net>		4040	14	Xref: artnum free.pt:7824815
7824816	Blabla kleur text 1.2.3.4 Nederlands	Citp <qdbYG42osB9nKHQjfG7kB7tiAUR12G32Xde0YoX5HrvoK28GOkG9vKRqCQ-pyYgmB.D2MPIIVyaQqNV8dVNic28tN2DT2fek3k9uf-pPINziMqY-s1E76dx-pLKbzE-pCOTdhn@47a00b09.8016910.10.1443521548.1.NL.RP1U2inwB5npih-sf5idprFU4Fn8ujZFbIsP1DTFpoDqO1tnHHDakdOI09-s0aAP7B>	Tue, 29 Sep 2015 10:12:28 GMT	<ak7L2BDy8rUCmQKVgyEjX@spot.net>		5597	28	Xref: artnum free.pt:7824816
7824817	Rokende text (2015) | teens	KarinaAva <vOD8F13AlBel-sUwaD1PMGBqs10-pSgwRn4e2-sqa3nTto9M1Go-sfBv4DPzy9ByTbBZ.lEtBK0TzVR3ej8rlW8E7Myl5uQ3EMO4XQBkKDyK6aGBDkuZRKM8IvXi-sOoqg8X8d@17a01d23d75b03c10z03.815909751.10.1443521709.1.NL.XXSODRVY8EqYBXLcXK7cWL0PycsjZ23n9Xr13LUFWcfd0Wwd9jkEWBOl-pXssDsqp>	Tue, 29 Sep 2015 12:15:36 +0200	<do0xhC09YQAqmQKVg3ran@spot.net>		1656	2	Xref: artnum free.pt:7824817
7824818	RandText RandText-Januari 1980	Spiegel <wYKtGrS-phsIYo5r0vfJl0OgrhCnX83N6o5sil82HlZVlVLXJ7uhQ5nFTwo4NHX4X.J6xf6AWN-s-swKfxMNSce-s6CIXYZmsS-shyRs6Ht2ChkFDN2RlqlZquEI66fa94dtoT@17a05c04d44z02.35876240.10.1443525368.1.NL.oLoRAXNqhI0DsEo0TWWs8jbDRWASZ-p09GtVYiM5YCoby-psttmXeEq5gMqd75It17>	Tue, 29 Sep 2015 10:16:17 GMT	<xt3UYlb6TpE8HIKVg7WcE@spot.net>		1921	7	Xref: artnum free.pt:7824818
7824819	RandText RandText RandText-RandText 2015	Spiegel <wYKtGrS-phsIYo5r0vfJl0OgrhCnX83N6o5sil82HlZVlVLXJ7uhQ5nFTwo4NHX4X.PgdH9DOxZUuGfh29UXIkE5sdlcRsG8FeSuTZl4rRVorvDprwQDm6mYs0coXKBZYw@17a05c04d44z02.34697887.10.1443525229.1.NL.no69vuJNnGConbBGDeCkWavo5jjkBF55-sEZ9ohda6nRHfPgLVFQXti9R-pzxfGFiw>	Tue, 29 Sep 2015 10:13:58 GMT	<66eO047ED38anIKVgGUkE@spot.net>		2046	7	Xref: artnum free.pt:7824819
7824820	RandText - RandText RandText RandText RandText	Citp <qdbYG42osB9nKHQjfG7kB7tiAUR12G32Xde0YoX5HrvoK28GOkG9vKRqCQ-pyYgmB.GBkDCiq-pwGVwFiVH56oP6NGIpRHXY9KX6w-pCPIdUCy7B3e4p1-pj-sg9pt7MPRs9c3@47a00b26.449715654.10.1443521892.1.NL.LJbXIYUviRJ8HKwvjJjb2d2aCIisdot9rWrOB0KhQY1E7TMWsxH-sOQwFDObAH-s8T>	Tue, 29 Sep 2015 10:18:11 GMT	<MPpNkB5SGC0YWUKVgPZfP@spot.net>		3417	7	Xref: artnum free.pt:7824820
7824821	RandText RandText RandText 11 Derp 99	Jaaprond <udOJLzYJY4T2EPdzAgV9bfao7PNXUU3pZzEVTPylFNgNPcLTe2u666ZfwxaNwlJ-s.j3xs1MoKVEJLpmBbFM9GQvBl1076Sn1xltHkfAyD7B38F2un7WOXq4311qtCJFRC@17a00b04d11b03c02c10d05d50z01.997732875.10.1443522033.1.NL.mTyg6aWIFkVPxJtSDvk2-sgyOB-pfz5Z2zmlBgszszqx0Gl4CX4ZVBoJdKtmi8W3LW>	Tue, 29 Sep 2015 10:20:33 GMT	<wri1oSMXsuo7WUKVgEvgB@spot.net>		2484	3	Xref: artnum free.pt:7824821
7824822	RandText RandText RandText RandText II | sanook	sanook <sBhTrAj4CZLAhpiecI3B7jtR5708ko-pwWuVKa9-srPHxepNMPG0chq0rRNMV5yC3r.OYEWFMCrocLGqIgmFpleF0hJ4xLJr1vhO3vbiv1sOYjUMMHUIGxE8VZEn-sqdHlTF@27a00b00c07d03d33z00.199394624.10.1443522149.1.NL.m7-sWQhOqQQQrW5XHa15JwEmrQ2PHA4gMRhEjJUsMHr163b1Yd-s3yEXJIHUMNSizu>	Tue, 29 Sep 2015 10:22:29 GMT	<9kouoCxpUCUX2YKVgBO4H@spot.net>		2423	7	Xref: artnum free.pt:7824822
7824823	RandText RandText RandText RandText 2019	Spiegel <wYKtGrS-phsIYo5r0vfJl0OgrhCnX83N6o5sil82HlZVlVLXJ7uhQ5nFTwo4NHX4X.lgzD4Xlw-pkzv0jG39ALTmMQwcIhLjLl2qeAgpUhKUT2-pvhmCtFnxWMbKXWz5NpS5@17a05c04d44z02.24441869.10.1443525673.1.NL.Jl-pc8xO01Oge1lMUe0I1PJWXqkZrWgl5i6-svSJ-pnY6Px-p5Nnwm6Sa8C7KSoHbfms>	Tue, 29 Sep 2015 10:21:28 GMT	<wrKlxcpp3YsInQKVgd46Q@spot.net>		2279	10	Xref: artnum free.pt:7824823
7824824	RandText RandText RandText RandText | fetish	FuMyAs <vOD8F13AlBel-sUwaD1PMGBqs10-pSgwRn4e2-sqa3nTto9M1Go-sfBv4DPzy9ByTbBZ.inmhyJJbL0NsXLY0s1U64id66c84CR-p8SWO8GLXSlRCj6c9sZloUsuSYrH-p9WRAL@17a01d23d75b03c10z03.132048357.10.1443522363.1.NL.fRYxBB2eChEvd3HVLj26h1iNTPBT9f4ccSfoAyaBP-s3CgZLpgIzatPdRPbYS4ED-p>	Tue, 29 Sep 2015 12:26:30 +0200	<tstfvWUQMuEOWcKVgJgU6@spot.net>		2163	2	Xref: artnum free.pt:7824824
7824826	RandText RandText RandText RandText RandText 2015	Spiegel <wYKtGrS-phsIYo5r0vfJl0OgrhCnX83N6o5sil82HlZVlVLXJ7uhQ5nFTwo4NHX4X.FALaRkOgDeYvxiZifFoeuRAmBumNPsO6RN1c1g6mHclrrWtpQvCdzkWVlE8NysCv@17a05c04d44z02.22635304.10.1443526055.1.NL.W5-sZ72AX0ZTetdgqa4a1vQZHc32sfCCoRLGbgdz1v4J7CmJNY3TWKNGGWJdtwLXx>	Tue, 29 Sep 2015 10:27:44 GMT	<BiFknR3m2rUn3UKVg2cCo@spot.net>		2358	11	Xref: artnum free.pt:7824826
.`;
		raw = strings.Replace(raw, "\n", "\r\n", -1)
		conn.Send(raw);
		return
	}

	conn.Send("501 No test")
}

func Xhdr(conn *client.Conn, tok []string) {
	// xhdr Date 7824860-7824865
	if len(tok) != 3 {
		conn.Send("501 Invalid syntax.")
		return
	}

	if tok[1] == "derp" {
		conn.Send("503 Header type unsupported.")
		return
	}

	if tok[1] == "Date" && tok[2] == "<aaa@spot.red>" {
		conn.Send("501 Syntax error Unparsable input: aaa@spot.red")
		return
	}

	if tok[1] == "Date" && tok[2] == "7824860-7824865" {
		conn.Send("221 Date headers follow.")
		raw := `7824860 Tue, 29 Sep 2015 13:39:23 +0200
7824861 Tue, 29 Sep 2015 13:40:33 +0200
7824862 Tue, 29 Sep 2015 13:42:44 +0200
7824863 Tue, 29 Sep 2015 11:43:36 GMT
7824864 Tue, 29 Sep 2015 13:53:39 +0200
7824865 Tue, 29 Sep 2015 11:49:51 GMT
.`;
		raw = strings.Replace(raw, "\n", "\r\n", -1)
		conn.Send(raw);
		return
	}

	conn.Send("501 No test")
}

func Stat(conn *client.Conn, tok []string) {
	// 223 0 <U5leXguPzRAUFUKVggE7M@spot.net> status
	if len(tok) != 2 {
		conn.Send("501 Invalid syntax.")
		return
	}
	if tok[1] == "<close@bb.cc>" {
		conn.Send("223 0 <close@bb.cc> status")
		return
	} else if tok[1] == "500" {
		conn.Send("223 500 <close@bb.cc> status")
		return
	}
	conn.Send("501 No test")
}

func List(conn *client.Conn, tok []string) {
	if len(tok) == 1 {
		conn.Send("215 active file follows.")
		raw := `alt.pri 1 1 Y
alt.media.dvd.hack.samsung 1 1 y
macromedia.director.3d 1 1 Y
alt.tasteless.bottomfeeders 5 1 Y
.`;
		raw = strings.Replace(raw, "\n", "\r\n", -1)
		conn.Send(raw);
		return
	}

	conn.Send("501 No test")
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
		} else if cmd == "STAT" {
			Stat(conn, tok)
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
			Group(conn, tok)
		} else if cmd == "NOOP" {
			conn.Send("500 Unsupported.")
		} else if cmd == "POST" {
			PostArticle(conn)
		} else if cmd == "XOVER" {
			Xover(conn, tok)
		} else if cmd == "XHDR" {
			Xhdr(conn, tok)
		} else if cmd == "LIST" {
			List(conn, tok)
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