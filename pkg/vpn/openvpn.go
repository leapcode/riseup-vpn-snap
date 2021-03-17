// Copyright (C) 2018-2020 LEAP
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

package vpn

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"0xacab.org/leap/bitmask-vpn/pkg/vpn/bonafide"
	"0xacab.org/leap/shapeshifter"
)

const (
	openvpnManagementAddr = "127.0.0.1"
	openvpnManagementPort = "6061"
)

// StartVPN for provider
func (b *Bitmask) StartVPN(provider string) error {
	var proxy string
	if b.transport != "" {
		var err error
		proxy, err = b.startTransport()
		if err != nil {
			return err
		}
	}

	if !b.CanStartVPN() {
		return errors.New("BUG: cannot start vpn")
	}
	err := b.startOpenVPN(proxy)
	return err
}

func (b *Bitmask) CanStartVPN() bool {
	/* FIXME this is not enough. We should check, if provider needs
	* credentials, if we have a valid token, otherwise remove it and
	make sure that we're asking for the credentials input */
	return !b.bonafide.NeedsCredentials()
}

func (b *Bitmask) startTransport() (proxy string, err error) {
	proxy = "127.0.0.1:4430"
	if b.shapes != nil {
		return proxy, nil
	}

	gateways, err := b.bonafide.GetGateways(b.transport)
	if err != nil {
		return "", err
	}
	if len(gateways) == 0 {
		log.Printf("No gateway for transport %s in provider", b.transport)
		return "", nil
	}

	for _, gw := range gateways {
		if _, ok := gw.Options["cert"]; !ok {
			continue
		}
		b.shapes = &shapeshifter.ShapeShifter{
			Cert:      gw.Options["cert"],
			Target:    gw.IPAddress + ":" + gw.Ports[0],
			SocksAddr: proxy,
		}
		go b.listenShapeErr()
		if iatMode, ok := gw.Options["iat-mode"]; ok {
			b.shapes.IatMode, err = strconv.Atoi(iatMode)
			if err != nil {
				b.shapes.IatMode = 0
			}
		}
		err = b.shapes.Open()
		if err != nil {
			log.Printf("Can't connect to transport %s: %v", b.transport, err)
			continue
		}
		return proxy, nil
	}
	return "", fmt.Errorf("No working gateway for transport %s: %v", b.transport, err)
}

func (b *Bitmask) listenShapeErr() {
	ch := b.shapes.GetErrorChannel()
	for {
		err, more := <-ch
		if !more {
			return
		}
		log.Printf("Error from shappeshifter: %v", err)
	}
}

func (b *Bitmask) startOpenVPN(proxy string) error {
	certPemPath, err := b.getCert()
	if err != nil {
		return err
	}
	arg, err := b.bonafide.GetOpenvpnArgs()
	if err != nil {
		return err
	}

	if proxy == "" {
		gateways, err := b.bonafide.GetGateways("openvpn")
		if err != nil {
			return err
		}
		err = b.launch.firewallStart(gateways)
		if err != nil {
			return err
		}

		for _, gw := range gateways {
			for _, port := range gw.Ports {
				arg = append(arg, "--remote", gw.IPAddress, port, "tcp4")
			}
		}
	} else {
		gateways, err := b.bonafide.GetGateways(b.transport)
		if err != nil {
			return err
		}
		err = b.launch.firewallStart(gateways)
		if err != nil {
			return err
		}

		proxyArgs := strings.Split(proxy, ":")
		arg = append(arg, "--remote", proxyArgs[0], proxyArgs[1], "tcp4")
	}
	arg = append(arg,
		"--verb", "1",
		"--management-client",
		"--management", openvpnManagementAddr, openvpnManagementPort,
		"--ca", b.getCaCertPath(),
		"--cert", certPemPath,
		"--key", certPemPath)
	return b.launch.openvpnStart(arg...)
}

func (b *Bitmask) getCert() (certPath string, err error) {
	certPath = b.getCertPemPath()

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		log.Println("Fetching certificate to", certPath)
		cert, err := b.bonafide.GetPemCertificate()
		if err != nil {
			return "", err
		}
		err = ioutil.WriteFile(certPath, cert, 0600)
	}

	return certPath, err
}

// StopVPN or cancel
func (b *Bitmask) StopVPN() error {
	err := b.launch.firewallStop()
	if err != nil {
		return err
	}
	if b.shapes != nil {
		b.shapes.Close()
		b.shapes = nil
	}
	return b.launch.openvpnStop()
}

// ReloadFirewall restarts the firewall
func (b *Bitmask) ReloadFirewall() error {
	err := b.launch.firewallStop()
	if err != nil {
		return err
	}

	status, err := b.GetStatus()
	if err != nil {
		return err
	}

	if status != Off {
		gateways, err := b.bonafide.GetGateways("openvpn")
		if err != nil {
			return err
		}
		return b.launch.firewallStart(gateways)
	}
	return nil
}

// GetStatus returns the VPN status
func (b *Bitmask) GetStatus() (string, error) {
	status, err := b.getOpenvpnState()
	if err != nil {
		status = Off
	}
	if status == Off && b.launch.firewallIsUp() {
		return Failed, nil
	}
	return status, nil
}

func (b *Bitmask) InstallHelpers() error {
	// TODO use pickle module from here
	return nil
}

// VPNCheck returns if the helpers are installed and up to date and if polkit is running
func (b *Bitmask) VPNCheck() (helpers bool, privilege bool, err error) {
	return b.launch.check()
}

func (b *Bitmask) ListGatewaysByCity(transport string) (map[string]bonafide.Gateway, error) {
	gwForCities, err := b.bonafide.PickGatewayForCities(transport)
	return gwForCities, err
}

func (b *Bitmask) GetGatewayDetails(host string) (interface{}, error) {
	gw, err := b.bonafide.GetGatewayDetails(host)
	if err != nil {
		return bonafide.Gateway{}, err
	}
	return gw, nil
}

// UseGateway selects a gateway, by label, as the default gateway
func (b *Bitmask) UseGateway(label string) error {
	b.bonafide.SetManualGateway(label)
	return nil
}

// UseTransport selects an obfuscation transport to use
func (b *Bitmask) UseTransport(transport string) error {
	if transport != "obfs4" {
		return fmt.Errorf("Transport %s not implemented", transport)
	}
	b.transport = transport
	return nil
}

func (b *Bitmask) getCertPemPath() string {
	return path.Join(b.tempdir, "openvpn.pem")
}

func (b *Bitmask) getCaCertPath() string {
	return path.Join(b.tempdir, "cacert.pem")
}
