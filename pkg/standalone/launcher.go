// +build !linux
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

package standalone

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"0xacab.org/leap/bitmask-vpn/pkg/standalone/bonafide"
)

const (
	helperAddr = "http://localhost:7171"
)

type launcher struct {
}

func newLauncher() (*launcher, error) {
	return &launcher{}, nil
}

func (l *launcher) close() error {
	return nil
}

func (l *launcher) check() (helpers bool, priviledge bool, err error) {
	return true, true, nil
}

func (l *launcher) openvpnStart(flags ...string) error {
	byteFlags, err := json.Marshal(flags)
	if err != nil {
		return err
	}
	return l.send("/openvpn/start", byteFlags)
}

func (l *launcher) openvpnStop() error {
	return l.send("/openvpn/stop", nil)
}

func (l *launcher) firewallStart(gateways []bonafide.Gateway) error {
	ipList := make([]string, len(gateways))
	for i, gw := range gateways {
		ipList[i] = gw.IPAddress
	}
	byteIPs, err := json.Marshal(ipList)
	if err != nil {
		return err
	}
	return l.send("/firewall/start", byteIPs)
}

func (l *launcher) firewallStop() error {
	return l.send("/firewall/stop", nil)
}

func (l *launcher) firewallIsUp() bool {
	var isup bool = false
	res, err := http.Post(helperAddr+"/firewall/isup", "", nil)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Printf("Got an error status code for firewall/isup: %v\n", res.StatusCode)
		isup = false
	} else {
		upStr, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Errorf("Error getting body for firewall/isup: %q", err)
			return false
		}
		isup, err = strconv.ParseBool(string(upStr))
		if err != nil {
			fmt.Errorf("Error parsing body for firewall/isup: %q", err)
			return false
		}
	}
	return isup
}

func (l *launcher) send(path string, body []byte) error {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	res, err := http.Post(helperAddr+path, "", reader)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	resErr, err := ioutil.ReadAll(res.Body)
	if len(resErr) > 0 {
		return fmt.Errorf("Helper returned an error: %q", resErr)
	}
	return err
}
