// +build integration
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

package bonafide

import (
	"bytes"
	"os"
	"testing"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
)

type SIPCreds struct {
	userOk, passOk string
}

func getFromEnv(name, defaultVar string) string {
	val, ok := os.LookupEnv(name)
	if !ok {
		return defaultVar
	}
	return val
}

func getSIPCreds() SIPCreds {
	userOk := getFromEnv("SIP_USER_OK", "test_user_ok")
	passOk := getFromEnv("SIP_PASS_OK", "test_pass_ok")
	creds := SIPCreds{
		userOk: userOk,
		passOk: passOk,
	}
	return creds
}

func initTestConfigSIP() {
	config.APIURL = "https://localhost:8000/"
}

func _TestSIPIntegrationGetCert(t *testing.T) {
	// FIXME: we need to automatize the webapi start
	initTestConfigSIP()
	creds := getSIPCreds()

	b := New()
	b.auth = &sipAuthentication{b.client, b.getURL("auth")}
	ok, err := b.DoLogin(creds.userOk, creds.passOk)
	if err != nil {
		t.Fatal("DoLogin returned an error: ", err)
	}
	if !ok {
		t.Fatal("Invalid credentials")
	}

	cert, err := b.GetPemCertificate()
	if err != nil {
		t.Fatal("getCert returned an error: ", err)
	}

	if !bytes.Contains(cert, privateKeyHeader) {
		t.Errorf("No private key present: \n%q", cert)
	}

	if !bytes.Contains(cert, certHeader) {
		t.Errorf("No cert present: \n%q", cert)
	}

	/* TODO -- check we receive 401 for bad credentials */
	/* TODO -- check we receive 4xx for no credentials */
}
