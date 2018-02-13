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
	"errors"
)

// StartVPN for provider
func (b *Bitmask) StartVPN(provider string) error {
	_, err := b.send("vpn", "start", provider)
	return err
}

// StopVPN or cancel
func (b *Bitmask) StopVPN() error {
	_, err := b.send("vpn", "stop")
	return err
}

// GetStatus returns the VPN status
func (b *Bitmask) GetStatus() (string, error) {
	res, err := b.send("vpn", "status")
	if err != nil {
		return "", err
	}
	return res["status"].(string), nil
}

// ListGateways return the names of the gateways
func (b *Bitmask) ListGateways(provider string) ([]string, error) {
	res, err := b.send("vpn", "list")
	if err != nil {
		return nil, err
	}

	names := []string{}
	locations, ok := res[provider].([]interface{})
	if !ok {
		return nil, errors.New("Can't read the locations for provider " + provider)
	}
	for i := range locations {
		loc := locations[i].(map[string]interface{})
		names = append(names, loc["name"].(string))
	}
	return names, nil
}

// UseGateway selects name as the default gateway
func (b *Bitmask) UseGateway(name string) error {
	_, err := b.send("vpn", "locations", name)
	return err
}
