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

package bonafide

import (
	"fmt"
	"io/ioutil"
)

type AnonymousAuthentication struct {
	bonafide *Bonafide
}

func (a *AnonymousAuthentication) GetPemCertificate() ([]byte, error) {
	resp, err := a.bonafide.client.Post(certAPI, "", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		resp, err = a.bonafide.client.Post(certAPI3, "", nil)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Get vpn cert has failed with status: %s", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}
