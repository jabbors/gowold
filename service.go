package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/sabhiram/go-wol/wol"
)

const (
	indexPage = `
<html>
<head>
<title>gowold</title>
<link rel="stylesheet" href="http://yui.yahooapis.com/pure/0.6.0/pure-min.css">
<style type="text/css">
#container {width:100%;height:100%;position:absolute;vertical-align:middle;text-align:center;}
#centered {margin-left:auto;margin-right:auto;margin-top:20%;margin-bottom:auto;display:block;}
.button-green, .button-red {color: white; border-radius: 4px; text-shadow: 0 1px 1px rgba(0, 0, 0, 0.2);}
.button-green {background: rgb(28, 184, 65);}
.button-red {background: rgb(202, 60, 60);}
</style>
</head>
<body>
<div id="container">
<div id="centered">
<form method="post">
<input class="button-BUTTONCOLORPLACEHOLDER pure-button" type="submit" name="action" value="ACTIONPLACEHOLDER" />
</form>
</div>
</div>
</body>
</html>
`
	actionStart      = "start"
	actionStop       = "stop"
	buttonColorGreen = "green"
	buttonColorRed   = "red"
)

// wolService broadcasts magic packets to target device identfied by a MAC address
type wolService struct {
	activated      bool
	broadcasting   bool
	startBroadcast time.Time
	lastBroadcast  time.Time
	macAddr        string
	broadcastIP    string
	broadcastPort  string
	resendInterval time.Duration
	udpAddr        *net.UDPAddr
	magicPacket    []byte
	err            error
}

func NewWolService(macAddr, broadcastIP, broadcastPort string) (*wolService, error) {
	ws := wolService{
		activated:      false,
		broadcasting:   false,
		macAddr:        macAddr,
		broadcastIP:    broadcastIP,
		broadcastPort:  broadcastPort,
		resendInterval: time.Minute,
	}

	var err error
	bcastAddr := fmt.Sprintf("%s:%s", ws.broadcastIP, ws.broadcastPort)
	ws.udpAddr, err = net.ResolveUDPAddr("udp", bcastAddr)
	if err != nil {
		return nil, err
	}

	// Create a magic packet
	mp, err := wol.New(ws.macAddr)
	if err != nil {
		return nil, err
	}

	// Transform into stream of bytes to send.
	ws.magicPacket, err = mp.Marshal()
	if err != nil {
		return nil, err
	}

	ws.start()
	return &ws, nil
}

func (ws *wolService) start() {
	go ws.run()
}

func (ws *wolService) run() {
	conn, err := net.DialUDP("udp", nil, ws.udpAddr)
	if err != nil {
		ws.err = err
		return
	}
	defer conn.Close()

	for {
		time.Sleep(time.Second * 5)
		if !ws.activated {
			continue
		}
		if time.Now().Before(ws.lastBroadcast.Add(ws.resendInterval)) {
			continue
		}

		log.Printf("attempting to send a magic packet to device with MAC '%s'", ws.macAddr)

		n, err := conn.Write(ws.magicPacket)
		if err == nil && n != 102 {
			ws.err = fmt.Errorf("error: magic packet sent was %d bytes (expected 102 bytes sent)", n)
			log.Println(err)
			continue
		}
		if err != nil {
			ws.err = err
			log.Printf("error: %s", ws.err)
			continue
		}

		log.Println("magic packet was sent successfully")
		ws.broadcasting = true
		ws.lastBroadcast = time.Now()
	}
}

func (ws *wolService) IndexHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Measure time spent executing
	start := time.Now()

	a := r.FormValue("action")
	if a == actionStop {
		ws.activated = false
		ws.broadcasting = false
	}
	if a == actionStart {
		ws.activated = true
		ws.startBroadcast = time.Now()
	}

	action := actionStart
	buttonColor := buttonColorGreen
	if ws.activated {
		action = actionStop
		buttonColor = buttonColorRed
	}

	// Logs [source IP] [request method] [request URL] [HTTP status] [time spent serving request]
	log.Printf("%v\t \"%v - %v\"\t%v\t%v", sourceIP(r), r.Method, r.RequestURI, http.StatusOK, time.Since(start))
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, strings.Replace(strings.Replace(indexPage, "ACTIONPLACEHOLDER", action, -1), "BUTTONCOLORPLACEHOLDER", buttonColor, -1))
}

func (ws *wolService) StatusHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Measure time spent executing
	start := time.Now()

	if !ws.broadcasting {
		if ws.err != nil {
			body := fmt.Sprintf("{\"error\":\"%s\"}", ws.err)
			// Logs [source IP] [request method] [request URL] [HTTP status] [time spent serving request]
			log.Printf("%v\t \"%v - %v\"\t%v\t%v", sourceIP(r), r.Method, r.RequestURI, http.StatusInternalServerError, time.Since(start))
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, body)
			return
		}

		// Logs [source IP] [request method] [request URL] [HTTP status] [time spent serving request]
		log.Printf("%v\t \"%v - %v\"\t%v\t%v", sourceIP(r), r.Method, r.RequestURI, http.StatusNoContent, time.Since(start))
		w.WriteHeader(http.StatusNoContent)
		return
	}

	body := fmt.Sprintf("{\"started_at\":\"%d\", \"last_broadcast\":\"%d\"}", ws.startBroadcast.Unix(), ws.lastBroadcast.Unix())
	// Logs [source IP] [request method] [request URL] [HTTP status] [time spent serving request]
	log.Printf("%v\t \"%v - %v\"\t%v\t%v", sourceIP(r), r.Method, r.RequestURI, http.StatusOK, time.Since(start))

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, body)
}

func sourceIP(r *http.Request) string {
	var ip string
	header := r.Header.Get("X-Forwarded-For")
	if header != "" {
		ips := strings.Split(header, ",")
		ip = strings.TrimSpace(ips[0])
	} else {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	return ip
}
