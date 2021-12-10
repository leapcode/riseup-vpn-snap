package main

/* a wrapper around bitmask that exposes status to a QtQml gui.
   Have a look at the pkg/backend module for further enlightment. */

import (
	"C"
	"unsafe"

	"0xacab.org/leap/bitmask-vpn/pkg/backend"
)

//export GetBitmaskVersion
func GetBitmaskVersion() *C.char {
	return (*C.char)(backend.GetVersion())
}

//export Login
func Login(username, password string) {
	backend.Login(username, password)
}

//export SwitchOn
func SwitchOn() {
	backend.SwitchOn()
}

//export SwitchOff
func SwitchOff() {
	backend.SwitchOff()
}

//export UseLocation
func UseLocation(label string) {
	// a bit of a hack to force the compiler to copy the string
	// so the original C++ string will not be used as it will be changed down the line
	location := string([]byte(label))
	backend.UseLocation(location)
}

//export UseAutomaticGateway
func UseAutomaticGateway() {
	backend.UseAutomaticGateway()
}

//export SetTransport
func SetTransport(transport string) {
	backend.SetTransport(string(transport))
}

//export GetTransport
func GetTransport() *C.char {
	return (*C.char)(backend.GetTransport())
}

//export SetUDP
func SetUDP(udp bool) {
	backend.SetUDP(udp)
}

//export SetSnowflake
func SetSnowflake(snowflake bool) {
	backend.SetSnowflake(snowflake)
}

//export Quit
func Quit() {
	backend.Quit()
}

//export DonateAccepted
func DonateAccepted() {
	backend.DonateAccepted()
}

//export DonateSeen
func DonateSeen() {
	backend.DonateSeen()
}

//export SubscribeToEvent
func SubscribeToEvent(event string, f unsafe.Pointer) {
	backend.SubscribeToEvent(event, f)
}

//export InitializeBitmaskContext
func InitializeBitmaskContext(provider string,
	jsonPtr unsafe.Pointer, jsonLen C.int,
	obfs4 bool, disableAutostart bool, startVPN string) {
	json := C.GoBytes(jsonPtr, jsonLen)
	opts := backend.InitOptsFromJSON(provider, string(json))
	opts.Obfs4 = obfs4
	opts.DisableAutostart = disableAutostart
	opts.StartVPN = startVPN
	go backend.InitializeBitmaskContext(opts)
}

//export InitializeTestBitmaskContext
func InitializeTestBitmaskContext(provider string,

	jsonPtr unsafe.Pointer, jsonLen C.int) {
	json := C.GoBytes(jsonPtr, jsonLen)
	opts := backend.InitOptsFromJSON(provider, string(json))
	opts.DisableAutostart = true
	opts.SkipLaunch = true
	opts.StartVPN = "no"
	backend.InitializeBitmaskContext(opts)
	backend.EnableMockBackend()
}

//export EnableWebAPI
func EnableWebAPI(port string) {
	backend.EnableWebAPI(port)
}

//export RefreshContext
func RefreshContext() *C.char {
	return (*C.char)(backend.RefreshContext())
}

//export ResetError
func ResetError(errname string) {
	backend.ResetError(errname)
}

//export ResetNotification
func ResetNotification(label string) {
	backend.ResetNotification(label)
}

//export InstallHelpers
func InstallHelpers() {
	backend.InstallHelpers()
}

func main() {}
