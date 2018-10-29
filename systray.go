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

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"0xacab.org/leap/bitmask-systray/bitmask"
	"0xacab.org/leap/bitmask-systray/icon"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
)

type bmTray struct {
	bm            bitmask.Bitmask
	conf          *systrayConfig
	notify        *notificator
	waitCh        chan bool
	mStatus       *systray.MenuItem
	mTurnOn       *systray.MenuItem
	mTurnOff      *systray.MenuItem
	mDonate       *systray.MenuItem
	mCancel       *systray.MenuItem
	activeGateway *gatewayTray
	autostart     autostart
}

type gatewayTray struct {
	menuItem *systray.MenuItem
	name     string
}

func run(bm bitmask.Bitmask, conf *systrayConfig, notify *notificator, as autostart) {
	// XXX this removes the snap error message, but produces an invisible icon.
	// https://0xacab.org/leap/riseup_vpn/issues/44
	// os.Setenv("TMPDIR", "/var/tmp")
	bt := bmTray{bm: bm, conf: conf, notify: notify, autostart: as}
	systray.Run(bt.onReady, bt.onExit)
}

func (bt bmTray) onExit() {
	status, _ := bt.bm.GetStatus()
	if status != "off" {
		bt.bm.StopVPN()
	}
	log.Println("Closing systray")
}

func (bt *bmTray) onReady() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	systray.SetIcon(icon.Off)

	bt.mStatus = systray.AddMenuItem(printer.Sprintf("Checking status..."), "")
	bt.mStatus.Disable()
	bt.mTurnOn = systray.AddMenuItem(printer.Sprintf("Turn on"), "")
	bt.mTurnOn.Hide()
	bt.mTurnOff = systray.AddMenuItem(printer.Sprintf("Turn off"), "")
	bt.mTurnOff.Hide()
	bt.mCancel = systray.AddMenuItem(printer.Sprintf("Cancel"), printer.Sprintf("Cancel connection to %s", applicationName))
	bt.mCancel.Hide()
	systray.AddSeparator()

	if bt.conf.SelectGateway {
		bt.addGateways()
	}

	mHelp := systray.AddMenuItem(printer.Sprintf("Help..."), "")
	bt.mDonate = systray.AddMenuItem(printer.Sprintf("Donate..."), "")
	mAbout := systray.AddMenuItem(printer.Sprintf("About..."), "")
	systray.AddSeparator()

	mQuit := systray.AddMenuItem(printer.Sprintf("Quit"), "")

	go func() {
		ch := bt.bm.GetStatusCh()
		if status, err := bt.bm.GetStatus(); err != nil {
			log.Printf("Error getting status: %v", err)
		} else {
			bt.changeStatus(status)
		}

		for {
			select {
			case status := <-ch:
				log.Println("status: " + status)
				bt.changeStatus(status)

			case <-bt.mTurnOn.ClickedCh:
				log.Println("on")
				bt.changeStatus("starting")
				bt.bm.StartVPN(provider)
				bt.conf.setUserStoppedVPN(false)
			case <-bt.mTurnOff.ClickedCh:
				log.Println("off")
				bt.changeStatus("stopping")
				bt.bm.StopVPN()
				bt.conf.setUserStoppedVPN(true)
			case <-bt.mCancel.ClickedCh:
				log.Println("cancel")
				bt.changeStatus("stopping")
				bt.bm.StopVPN()
				bt.conf.setUserStoppedVPN(true)

			case <-mHelp.ClickedCh:
				open.Run("https://riseup.net/vpn/support")
			case <-bt.mDonate.ClickedCh:
				bt.conf.setDonated()
				open.Run("https://riseup.net/vpn/donate")
			case <-mAbout.ClickedCh:
				bitmaskVersion, err := bt.bm.Version()
				versionStr := version
				if err != nil {
					log.Printf("Error getting version: %v", err)
				} else if bitmaskVersion != "" {
					versionStr = fmt.Sprintf("%s (bitmaskd %s)", version, bitmaskVersion)
				}
				bt.notify.about(versionStr)

			case <-mQuit.ClickedCh:
				err := bt.autostart.Disable()
				if err != nil {
					log.Printf("Error disabling autostart: %v", err)
				}
				systray.Quit()
			case <-signalCh:
				systray.Quit()

			case <-time.After(5 * time.Second):
				if status, err := bt.bm.GetStatus(); err != nil {
					log.Printf("Error getting status: %v", err)
				} else {
					bt.changeStatus(status)
				}
			}
		}
	}()
}

func (bt *bmTray) addGateways() {
	gatewayList, err := bt.bm.ListGateways(provider)
	if err != nil {
		log.Printf("Gateway initialization error: %v", err)
		return
	}

	mGateway := systray.AddMenuItem(printer.Sprintf("Route traffic through"), "")
	mGateway.Disable()
	for i, city := range gatewayList {
		menuItem := systray.AddMenuItem(city, printer.Sprintf("Use %s %v gateway", applicationName, city))
		gateway := gatewayTray{menuItem, city}

		if i == 0 {
			menuItem.Check()
			menuItem.SetTitle("*" + city)
			bt.activeGateway = &gateway
		} else {
			menuItem.Uncheck()
		}

		go func(gateway gatewayTray) {
			for {
				<-menuItem.ClickedCh
				gateway.menuItem.SetTitle("*" + gateway.name)
				gateway.menuItem.Check()

				bt.activeGateway.menuItem.Uncheck()
				bt.activeGateway.menuItem.SetTitle(bt.activeGateway.name)
				bt.activeGateway = &gateway

				bt.bm.UseGateway(gateway.name)
			}
		}(gateway)
	}

	systray.AddSeparator()
}

func (bt *bmTray) changeStatus(status string) {
	bt.mTurnOn.SetTitle(printer.Sprintf("Turn on"))
	if bt.waitCh != nil {
		bt.waitCh <- true
		bt.waitCh = nil
	}

	var statusStr string
	switch status {
	case "on":
		systray.SetIcon(icon.On)
		statusStr = printer.Sprintf("%s on", applicationName)
		bt.mTurnOn.Hide()
		bt.mTurnOff.Show()
		bt.mCancel.Hide()

	case "off":
		systray.SetIcon(icon.Off)
		statusStr = printer.Sprintf("%s off", applicationName)
		bt.mTurnOn.Show()
		bt.mTurnOff.Hide()
		bt.mCancel.Hide()

	case "starting":
		bt.waitCh = make(chan bool)
		go bt.waitIcon()
		statusStr = printer.Sprintf("Connecting to %s", applicationName)
		bt.mTurnOn.Hide()
		bt.mTurnOff.Hide()
		bt.mCancel.Show()

	case "stopping":
		bt.waitCh = make(chan bool)
		go bt.waitIcon()
		statusStr = printer.Sprintf("Stopping %s", applicationName)
		bt.mTurnOn.Hide()
		bt.mTurnOff.Hide()
		bt.mCancel.Hide()

	case "failed":
		systray.SetIcon(icon.Blocked)
		bt.mTurnOn.SetTitle(printer.Sprintf("Retry"))
		statusStr = printer.Sprintf("%s blocking internet", applicationName)
		bt.mTurnOn.Show()
		bt.mTurnOff.Show()
		bt.mCancel.Hide()
	}

	systray.SetTooltip(statusStr)
	bt.mStatus.SetTitle(statusStr)
}

func (bt *bmTray) waitIcon() {
	icons := [][]byte{icon.Wait0, icon.Wait1, icon.Wait2, icon.Wait3}
	for i := 0; true; i = (i + 1) % 4 {
		systray.SetIcon(icons[i])

		select {
		case <-bt.waitCh:
			return
		case <-time.After(time.Millisecond * 500):
			continue
		}
	}
}
