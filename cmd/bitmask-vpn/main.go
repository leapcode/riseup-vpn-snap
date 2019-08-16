// Copyright (C) 2018 LEAP
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

//go:generate python3 ../../branding/scripts/vendorize.py ../../branding/templates/bitmaskvpn/config.go ../../branding/config/vendor.conf ../../pkg/config/config.go

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/systray"
	"github.com/jmshal/go-locale"
	"golang.org/x/text/message"
)

const (
	logFile = "systray.log"
)

var version string

func main() {
	// on OSX sometimes the systray doesn't work (bitmask-systray#52)
	// locking the main thread into an OS thread fixes the problem
	runtime.LockOSThread()

	logger, err := config.ConfigureLogger(path.Join(config.Path, logFile))
	if err != nil {
		log.Println("Can't configure logger: ", err)
	} else {
		defer logger.Close()
	}

	conf := systray.ParseConfig()

	selectGateway := flag.Bool("select-gateway", false, "Enable gateway selection")
	obfs4 := flag.Bool("obfs4", false, "Use obfs4 to obfuscate the traffic is available in the provider")
	disableAutostart := flag.Bool("disable-autostart", false, "Disable the autostart for the next run")
	startVPN := flag.String("start-vpn", "", "Start the vpn in turned 'on' or 'off'")
	versionFlag := flag.Bool("version", false, "Version of the bitmask-systray")
	flag.Parse()
	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}
	if *startVPN != "" {
		if *startVPN != "on" && *startVPN != "off" {
			fmt.Println("-start-vpn should be 'on' or 'off'")
			os.Exit(1)
		}
		conf.StartVPN = *startVPN == "on"
	}
	if *selectGateway {
		conf.SelectGateway = *selectGateway
	}
	if *obfs4 {
		conf.Obfs4 = *obfs4
	}
	if *disableAutostart {
		conf.DisableAustostart = *disableAutostart
	}

	conf.Version = version
	conf.Printer = initPrinter()
	systray.Run(conf)
}

func initPrinter() *message.Printer {
	locale, err := go_locale.DetectLocale()
	if err != nil {
		log.Println("Error detecting the system locale: ", err)
	}

	return message.NewPrinter(message.MatchLanguage(locale, "en"))
}
