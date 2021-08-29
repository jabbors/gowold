package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/kelseyhightower/envconfig"
)

const (
	appName = "gowald"
)

var (
	version string
)

// appConfig represents the configuration.
type appConfig struct {
	Host             string `default:"0.0.0.0" desc:"host to bind to"`
	Port             int    `default:"8080" desc:"port to bind to"`
	BroadcastIP      string `default:"255.255.255.255" envconfig:"BROADCAST_IP" desc:"The IP on wich to broadcast packets to"`
	BroadcastUDPPort string `default:"9" envconfig:"BROADCAST_UDP_PORT" desc:"UDP port on which to broadcat packets to"`
	TargetMAC        string `default:"" required:"true" envconfig:"TARGET_MAC" desc:"MAC address of the device to recive the magic packet"`
}

// parse options from the environment. Return an error if parsing fails.
func (a *appConfig) parse() {
	defaultUsage := flag.Usage
	flag.Usage = func() {
		// Show default usage for the app (lists flags, etc).
		defaultUsage()
		fmt.Fprint(os.Stderr, "\n")

		err := envconfig.Usage("", a)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n\n", err.Error())
			os.Exit(1)
		}
	}

	var verFlag bool
	flag.BoolVar(&verFlag, "version", false, "print version and exit")
	flag.Parse()

	// Print version and exit if -version flag is passed.
	if verFlag {
		fmt.Printf("gowald: version=%s\n", version)
		os.Exit(0)
	}

	err := envconfig.Process("", a)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n\n", err.Error())
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	config := &appConfig{}
	config.parse()

	_, err := net.ParseMAC(config.TargetMAC)
	if err != nil {
		panic(err)
	}

	ws, err := NewWolService(config.TargetMAC, config.BroadcastIP, config.BroadcastUDPPort)
	if err != nil {
		panic(err)
	}

	router := httprouter.New()
	router.RedirectTrailingSlash = true
	router.GET("/", ws.IndexHandler)
	router.POST("/", ws.IndexHandler)
	router.GET("/status", ws.StatusHandler)
	log.Printf("listening on %s:%d", config.Host, config.Port)
	log.Fatal(http.ListenAndServe(config.Host+":"+strconv.Itoa(config.Port), router))
}
