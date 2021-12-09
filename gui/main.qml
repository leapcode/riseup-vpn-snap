import QtQuick 2.0
import QtQuick.Controls 2.4
import QtQuick.Dialogs 1.2
import QtQuick.Controls.Material 2.1
import QtQuick.Layouts 1.14

import "./components"

ApplicationWindow {

    id: root
    visible: true

    property int appHeight: 460
    property int appWidth: 280
    property alias customTheme: themeLoader.item
    property bool drawerOn: false

    width: appWidth
    minimumWidth: appWidth
    maximumWidth: appWidth

    height: appHeight
    minimumHeight: appHeight
    maximumHeight: appHeight

    title: ctx ? ctx.appName : ""
    Material.accent: Material.Green

    property var ctx
    property var error: ""

    // TODO can move properties to some state sub-item to unclutter
    property bool isDonationService: false
    property bool showDonationReminder: false
    property var locationsModel: []
    // TODO get from persistance
    property var selectedGateway: "auto"

    // FIXME get svg icons
    property var icons: {
        "off": "qrc:/assets/icon/png/white/vpn_off.png",
        "on": "qrc:/assets/icon/png/white/vpn_on.png",
        "wait": "qrc:/assets/icon/png/white/vpn_wait_0.png",
        "blocked": "qrc:/assets/icon/png/white/vpn_blocked.png"
    }

    signal openDonateDialog()

    FontLoader {
        id: lightFont
        source: "qrc:/poppins-regular.ttf"
    }

    FontLoader {
        id: boldFont
        source: "qrc:/poppins-bold.ttf"
    }

    FontLoader {
        id: boldFontMonserrat
        source: "qrc:/monserrat-bold.ttf"
    }

    FontLoader {
        id: robotoFont
        source: "qrc:/roboto.ttf"
    }

    FontLoader {
        id: robotoBoldFont
        source: "qrc:/roboto-bold.ttf"
    }

    Loader {
        id: loader
        asynchronous: true
        anchors.fill: parent
    }

    Loader {
        id: themeLoader
        source: loadTheme()
    }


    Systray {
        id: systray
    }

    Connections {
        target: jsonModel
        function onDataChanged() {
            let j = jsonModel.getJson()
            if (qmlDebug) {
                console.debug(j)
            }
            ctx = JSON.parse(j)
            if (ctx != undefined) {
                locationsModel = getSortedLocations()
            }
            if (ctx.errors) {
                console.debug("errors, setting root.error")
                root.error = ctx.errors
            } else {
                root.error = ""
            }
            if (ctx.donateURL) {
                isDonationService = true
            }
            if (ctx.donateDialog == 'true') {
                showDonationReminder = true
            }
            if (isAutoLocation()) {
                root.selectedGateway = "auto"
            }
        }
    }

    function getSortedLocations() {
        let obj = ctx.locations
        var arr = []
        for (var prop in obj) {
            if (obj.hasOwnProperty(prop)) {
                arr.push({
                             "key": prop,
                             "value": obj[prop]
                         })
            }
        }
        arr.sort(function (a, b) {
            return a.value - b.value
        }).reverse()
        return Array.from(arr, (k,_) => k.key);
    }

    function isAutoLocation() {
        // FIXME there's something weird going on with newyork location...
        // it gets marked as auto, which from europe is a bug.
        let best = ctx.locationLabels[ctx.bestLocation]
        if (best == undefined) {
            return false
        }
        return (best[0] == ctx.currentLocation)
    }

    function bringToFront() {
        // FIXME does not work properly, at least on linux 
        if (visibility == 3) {
            showNormal()
        } else {
            show() 
        }
        raise()
        requestActivate()
    }

    function loadTheme() {
        let arr = flavor.split("/")
        var providerFlavor = arr[arr.length-1]
        console.debug("flavor: " + providerFlavor)
        if (providerFlavor == "riseup-vpn") {
            return "themes/Riseup.qml"
        } else if (providerFlavor== "calyx-vpn") {
            return "themes/Calyx.qml"
        } else {
            // we should do a Default theme, with a fallback
            // mechanism
            return "Riseup.qml"
        }
    }

    onSceneGraphError: function (error, msg) {
        console.debug("ERROR while initializing scene")
        console.debug(msg)
    }

    Component.onCompleted: {
        loader.source = "components/Splash.qml"
    }
}
