import QtQuick 2.15
import QtQuick.Controls 2.4
import QtQuick.Dialogs 1.2
import QtQuick.Controls.Material 2.1

import "../themes/themes.js" as Theme

ToolBar {
    visible: stackView.depth > 1
    Material.foreground: Material.Black
    Material.background: customTheme.bgColor
    Material.elevation: 0

    contentHeight: settingsButton.implicitHeight

    ToolButton {
        id: settingsButton
        anchors {
            left: parent.left
            // margin needed at least for the Locations panel
            leftMargin: 5
        }
        font.pixelSize: Qt.application.font.pixelSize * 1.6
        icon.source: "../resources/arrow-left.svg"
        HoverHandler {
            cursorShape: Qt.PointingHandCursor
        }
        onClicked: {
            if (stackView.depth > 1) {
                stackView.pop()
            } else {
                settingsDrawer.toggle()
            }
        }
    }

    Label {
        text: stackView.currentItem.title
        font.bold: true
        anchors.centerIn: parent
    }
}
