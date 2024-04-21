import QtQuick 2.15
import QtQuick.Controls 2.2
import QtQuick.Layouts 1.14
import QtQuick.Controls.Material 2.1
import QtGraphicalEffects 1.0
import Qt.labs.settings 1.0

import "../themes/themes.js" as Theme

ThemedPage {
    title: qsTr("Preferences")

    Settings {
        id: settings
        property string locale: locale
    }

    ScrollView {
        clip: true
        anchors {
            fill: parent
        }

        Rectangle {
            implicitHeight: getBoxHeight() + 20
            implicitWidth: root.appWidth
            width: root.appWidth * 0.80
            // FIXME - just the needed height
            height: getBoxHeight()
            radius: 10
            color: "white"

            anchors {
                // fill: parent
                right: parent.right
                left: parent.left
                top: parent.top
                margins: 10
            }

            ColumnLayout {
                id: prefCol
                width: root.appWidth * 0.80

                Rectangle {
                    id: turnOffWarning
                    visible: false
                    height: 20
                    width: parent.width
                    color: "white"

                    Label {
                        color: "red"
                        text: qsTr("Turn off the VPN to make changes")
                        width: prefCol.width
                    }
                    Layout.topMargin: 10
                    Layout.leftMargin: 10
                    Layout.rightMargin: 10
                }

                RowLayout {
                    Layout.topMargin: 10
                    Layout.leftMargin: 10
                    Layout.rightMargin: 10

                    Label {
                        id: languageLabel
                        text: qsTr("Language")
                        font.bold: true
                    }

                    Icon {
                        id: languageIcon
                        sourceSize.height: 20
                        sourceSize.width: 20
                        source: "../resources/language.svg"
                    }
                }

                ComboBox {
                    id: languageComboBox
                    Layout.leftMargin: 10
                    Layout.rightMargin: 10
                    Layout.preferredWidth: 240
                    model: locales
                    textRole: 'name'
                    valueRole: 'locale'

                    Material.elevation: 0

                    Component.onCompleted: {
                        currentIndex = indexOfValue(settings.locale)
                    }
                    onActivated: {
                        backend.setLocale(currentValue)
                    }
                }

                Label {
                    id: circumLabel
                    text: qsTr("Censorship circumvention")
                    font.bold: true
                    Layout.topMargin: 10
                    Layout.leftMargin: 10
                    Layout.rightMargin: 10
                }

                Label {
                    text: qsTr("These techniques can bypass censorship, but are slower. Use them only when needed.")
                    color: Material.foreground
                    visible: true
                    wrapMode: Text.Wrap
                    font.pixelSize: Theme.fontSize - 3
                    Layout.leftMargin: 10
                    Layout.rightMargin: 10
                    Layout.preferredWidth: 240
                }

                MaterialCheckBox {
                    id: useBridgesCheckBox
                    enabled: areBridgesAvailable()
                    checked: false
                    text: qsTr("Use obfs4 bridges")
                    // TODO refactor - this sets wrapMode on checkbox
                    contentItem: Label {
                        text: useBridgesCheckBox.text
                        font: useBridgesCheckBox.font
                        horizontalAlignment: Text.AlignLeft
                        verticalAlignment: Text.AlignVCenter
                        leftPadding: useBridgesCheckBox.indicator.width + useBridgesCheckBox.spacing
                        wrapMode: Label.Wrap
                    }
                    Layout.leftMargin: 10
                    Layout.rightMargin: 10
                    HoverHandler {
                        cursorShape: Qt.PointingHandCursor
                    }
                    onClicked: {
                        // TODO there's a corner case that needs to be dealt with in the backend,
                        // if an user has a manual location selected and switches to bridges:
                        // we need to fallback to "auto" selection if such location does not
                        // offer bridges
                        useBridges(checked)
                        useUDP.enabled = !checked
                    }
                }

                Label {
                    text: qsTr("Traffic is obfuscated to bypass blocks")
                    color: useBridgesCheckBox.enabled ? Material.foreground : Material.hintTextColor
                    visible: true
                    wrapMode: Text.Wrap
                    font.pixelSize: Theme.fontSize - 5
                    Layout.leftMargin: 36
                    Layout.rightMargin: 15
                    Layout.bottomMargin: 5
                    Layout.topMargin: -5
                    Layout.preferredWidth: 220
                }

                MaterialCheckBox {
                    id: useSnowflake
                    text: qsTr("Use Snowflake")
                    enabled: false
                    checked: false
                    HoverHandler {
                        cursorShape: Qt.PointingHandCursor
                    }
                    Layout.leftMargin: 10
                    Layout.rightMargin: 10
                    Layout.preferredWidth: 240
                    onClicked: {
                        doUseSnowflake(checked)
                    }
                }

                Label {
                    text: qsTr("Snowflake needs Tor installed in your system")
                    color: useSnowflake.enabled ? Material.foreground : Material.hintTextColor
                    visible: true
                    wrapMode: Text.Wrap
                    font.pixelSize: Theme.fontSize - 5
                    Layout.leftMargin: 36
                    Layout.rightMargin: 15
                    Layout.bottomMargin: 5
                    Layout.topMargin: -5
                    Layout.preferredWidth: 220
                }

                Label {
                    text: qsTr("Transport")
                    font.bold: true
                    Layout.leftMargin: 10
                    Layout.rightMargin: 10
                    Layout.topMargin: 8
                }

                Label {
                    text: qsTr("UDP can make the VPN faster. It might be blocked on some networks.")
                    width: parent.width
                    color: Material.foreground
                    visible: true
                    wrapMode: Text.Wrap
                    font.pixelSize: Theme.fontSize - 3
                    Layout.leftMargin: 10
                    Layout.rightMargin: 10
                    Layout.preferredWidth: 240
                }

                MaterialCheckBox {
                    id: useUDP
                    text: qsTr("Use UDP if available")
                    enabled: false
                    checked: false
                    Layout.leftMargin: 10
                    Layout.rightMargin: 10
                    HoverHandler {
                        cursorShape: Qt.PointingHandCursor
                    }
                    onClicked: {
                        doUseUDP(checked)
                        useBridgesCheckBox.enabled = areBridgesAvailable()
                    }
                }
            }
        }
    }

    StateGroup {
        state: ctx ? ctx.status : "off"
        states: [
            State {
                name: "on"
                PropertyChanges {
                    target: turnOffWarning
                    visible: true
                }
                PropertyChanges {
                    target: useBridgesCheckBox
                    enabled: false
                }
                PropertyChanges {
                    target: useUDP
                    enabled: false
                }
            },
            State {
                name: "starting"
                PropertyChanges {
                    target: turnOffWarning
                    visible: true
                }
                PropertyChanges {
                    target: useBridgesCheckBox
                    enabled: false
                }
                PropertyChanges {
                    target: useUDP
                    enabled: false
                }
            },
            State {
                name: "off"
                PropertyChanges {
                    target: turnOffWarning
                    visible: false
                }
                PropertyChanges {
                    target: useBridgesCheckBox
                    enabled: true
                }
                PropertyChanges {
                    target: useUDP
                    enabled: true
                }
            }
        ]
    }

    function areBridgesAvailable() {
        // FIXME check if provider offers it
        let providerSupport = true
        return providerSupport && !useUDP.checked
    }

    function useBridges(value) {
        if (value == true) {
            backend.setTransport("obfs4")
        } else {
            backend.setTransport("openvpn")
        }
    }

    function doUseUDP(value) {
        backend.setUDP(value)
    }

    function doUseSnowflake(value) {
        backend.setSnowflake(value)
    }

    function getBoxHeight() {
        return prefCol.height + 15
    }

    Component.onCompleted: {
        if (ctx && ctx.offersUdp == "false") {
            useUDP.enabled = false
        }
        if (ctx && ctx.offersUdp && ctx.udp == "true") {
            useUDP.checked = true
        }
        if (ctx && ctx.transport == "obfs4") {
            useBridgesCheckBox.checked = true
            useUDP.enabled = false
        }
        // disabled for now, will be on next release
        /*
        if (ctx && ctx.hasTor == "true") {
            useSnowflake.enabled = true
        }
        if (ctx && ctx.hasTor && ctx.snowflake == "true") {
            useSnowflake.checked = true
        }
        */
    }
}
