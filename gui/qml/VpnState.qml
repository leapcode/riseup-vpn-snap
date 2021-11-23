import QtQuick 2.0
import QtQuick.Controls 2.12

import "logic.js" as Logic

StateGroup {

    state: ctx ? ctx.status : ""

    states: [
        State {
            name: "initializing"
        },
        State {
            name: "off"
            StateChangeScript {
                script: Logic.setStatus("off")
            }
            PropertyChanges {
                target: systray
                tooltip: Logic.toHuman("off")
                icon.source: icons["off"]
            }
            PropertyChanges {
                target: vpntoggle
                checked: false
                checkable: true
            }
            PropertyChanges {
                target: statusItem
                text: Logic.toHuman("off")
            }
            PropertyChanges {
                target: autoSelectionItem
                text: qsTr("Recommended")
            }
            PropertyChanges {
                target: mainStatus
                text: Logic.toHuman("off")
            }
            PropertyChanges {
                target: mainCurrentGateway
                text: ""
            }
        },
        State {
            name: "on"
            StateChangeScript {
                script: {
                    Logic.setNeedsReconnect(false)
                    brReconnect = false

                    if (needsDonate && !shownDonate) {
                        donate.visible = true
                        shownDonate = true
                        backend.donateSeen()
                    }
                }
            }
            PropertyChanges {
                target: systray
                tooltip: Logic.toHuman("on")
                icon.source: icons["on"]
            }
            PropertyChanges {
                target: vpntoggle
                checked: true
                checkable: true
            }
            PropertyChanges {
                target: statusItem
                text: Logic.toHuman("on")
            }
            PropertyChanges {
                target: autoSelectionItem
                text: {
                    if (autoSelectionButton.checked) {
                        //: %1 -> location to which the client is connected to
                        qsTr("Recommended (%1)").arg(locationStr())
                    } else {
                        qsTr("Recommended")
                    }
                }
            }
            PropertyChanges {
                target: mainStatus
                text: Logic.toHuman("on")
            }
            PropertyChanges {
                target: mainCurrentGateway
                //: %1 -> location to which the client is connected to
                text: qsTr("Connected to %1").arg(locationStr())
            }
        },
        State {
            name: "starting"
            PropertyChanges {
                target: systray
                tooltip: Logic.toHuman("connecting")
                icon.source: icons["wait"]
            }
            PropertyChanges {
                target: statusItem
                text: Logic.toHuman("connecting")
            }
            PropertyChanges {
                target: autoSelectionItem
                text: {
                    if (autoSelectionButton.checked) {
                        //: %1 -> location to which the client is connected to
                        qsTr("Recommended (%1)").arg(locationStr())
                    } else {
                        //: Clickable item to switch to the recommended location
                        qsTr("Recommended")
                    }
                }
            }
            PropertyChanges {
                target: mainStatus
                text: qsTr("Connecting…")
            }
            PropertyChanges {
                target: mainCurrentGateway
                text: ""
            }
            PropertyChanges {
                target: vpntoggle
                checked: true
                checkable: false
            }
        },
        State {
            name: "stopping"
            PropertyChanges {
                target: systray
                tooltip: Logic.toHuman("stopping")
                icon.source: icons["wait"]
            }
            PropertyChanges {
                target: statusItem
                text: Logic.toHuman("stopping")
            }
            PropertyChanges {
                target: autoSelectionItem
                //: Clickable item to switch to the recommended location
                text: qsTr("Recommended")
            }
            PropertyChanges {
                target: mainStatus
                text: Logic.toHuman("stopping")
            }
            PropertyChanges {
                target: mainCurrentGateway
                text: ""
            }
            PropertyChanges {
                target: vpntoggle
                checked: false
                checkable: false
            }
        },
        State {
            name: "failed"
            PropertyChanges {
                target: systray
                tooltip: Logic.toHuman("failed")
                icon.source: icons["wait"]
            }
            PropertyChanges {
                target: statusItem
                text: Logic.toHuman("failed")
            }
            PropertyChanges {
                target: autoSelectionItem
                //: Clickable item to swicth to the recommended location
                text: qsTr("Recommended")
            }
            PropertyChanges {
                target: mainStatus
                text: Logic.toHuman("failed")
            }
            PropertyChanges {
                target: mainCurrentGateway
                text: ""
            }
            PropertyChanges {
                target: vpntoggle
                checked: false
                checkable: true
            }
        }
    ]
}
