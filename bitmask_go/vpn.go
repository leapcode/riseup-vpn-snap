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

package bitmask

import (
	"path"
)

const (
	openvpnManagementAddr = "127.0.0.1"
	openvpnManagementPort = "6061"
)

// StartVPN for provider
func (b *Bitmask) StartVPN(provider string) error {
	gateways, err := b.bonafide.getGateways()
	if err != nil {
		return err
	}
	err = b.launch.firewallStart(gateways)
	if err != nil {
		return err
	}

	arg := []string{"--nobind", "--verb", "1"}
	bonafideArgs, err := b.bonafide.getOpenvpnArgs()
	if err != nil {
		return err
	}
	arg = append(arg, bonafideArgs...)
	for _, gw := range gateways {
		arg = append(arg, "--remote", gw.IPAddress, "443", "tcp4")
	}
	certPemPath := b.getCertPemPath()
	arg = append(arg, "--client", "--tls-client", "--remote-cert-tls", "server", "--management-client", "--management", openvpnManagementAddr+" "+openvpnManagementPort, "--ca", b.getCaCertPath(), "--cert", certPemPath, "--key", certPemPath)
	return b.launch.openvpnStart(arg...)
}

// StopVPN or cancel
func (b *Bitmask) StopVPN() error {
	err := b.launch.firewallStop()
	if err != nil {
		return err
	}
	return b.launch.openvpnStop()
}

// GetStatus returns the VPN status
func (b *Bitmask) GetStatus() (string, error) {
	status, err := b.getOpenvpnState()
	if err != nil {
		status = Off
	}
	return status, nil
}

// InstallHelpers into the system
func (b *Bitmask) InstallHelpers() error {
	// TODO
	return nil
}

// VPNCheck returns if the helpers are installed and up to date and if polkit is running
func (b *Bitmask) VPNCheck() (helpers bool, priviledge bool, err error) {
	// TODO
	return true, true, nil
}

// ListGateways return the names of the gateways
func (b *Bitmask) ListGateways(provider string) ([]string, error) {
	gateways, err := b.bonafide.getGateways()
	if err != nil {
		return nil, err
	}
	gatewayNames := make([]string, len(gateways))
	for i, gw := range gateways {
		gatewayNames[i] = gw.Location
	}
	return gatewayNames, nil
}

// UseGateway selects name as the default gateway
func (b *Bitmask) UseGateway(name string) error {
	b.bonafide.setDefaultGateway(name)
	return nil
}

func (b *Bitmask) getCertPemPath() string {
	return path.Join(b.tempdir, "openvpn.pem")
}

func (b *Bitmask) getCaCertPath() string {
	return path.Join(b.tempdir, "cacert.pem")
}
