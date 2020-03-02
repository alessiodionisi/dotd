package main

import (
	"flag"
	"fmt"
	"github.com/adnsio/dotd"
	"os"
)

var (
	addrFlag     string
	upstreamFlag string
	verboseFlag  bool
	versionFlag  bool
	fileLogFlag  string
	version      = "0.0.0"
	commit       = "commithash"
)

func init() {
	flag.StringVar(&addrFlag, "address", "[::]:53", "udp address")
	flag.StringVar(&upstreamFlag, "upstream", "https://1.1.1.1/dns-query", "upstream dns server")
	flag.BoolVar(&verboseFlag, "v", false, "print logs on terminal")
	flag.StringVar(&fileLogFlag, "flogs", "", "print logs to file")
	flag.BoolVar(&versionFlag, "version", false, "output version")
}

func main() {
	appString := fmt.Sprintf("dotd version %s %s", version, commit)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n\nUsage: dotd [options]\n", appString)
		flag.PrintDefaults()
	}

	flag.Parse()

	if versionFlag {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n", appString)
		os.Exit(2)
	}

	fmt.Printf("%s\n\n", appString)

	srv := dotd.New(&dotd.Config{
		Addr:     addrFlag,
		Upstream: upstreamFlag,
		Verbose:  verboseFlag,
		FileLog:  fileLogFlag,
	})

	srv.Listen()
}
