/* All the exported functions should be added here */

package backend

import (
	"C"
	"encoding/json"
	"log"
	"strconv"
	"unsafe"

	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
	"0xacab.org/leap/bitmask-vpn/pkg/config/version"
	"0xacab.org/leap/bitmask-vpn/pkg/pickle"
	"0xacab.org/leap/bitmask-vpn/pkg/pid"
)

func Login(username, password string) {
	success, err := ctx.bm.DoLogin(username, password)
	if err != nil {
		if err.Error() == "TokenErrTimeout" {
			ctx.Errors = "bad_auth_timeout"
		} else if err.Error() == "TokenErrBadStatus 502" {
			ctx.Errors = "bad_auth_502"
		} else {
			log.Println("ERROR: bad login", err)
			ctx.Errors = "bad_auth"
		}
	} else if success {
		log.Printf("Logged in as %s", username)
		ctx.LoginOk = true
		ctx.LoginDialog = false
	} else {
		log.Printf("Failed to login as %s", username)
		ctx.LoginDialog = true
		ctx.Errors = "bad_auth"
	}
	go ctx.updateStatus()
}

func SwitchOn() {
	go setStatus(starting)
	go startVPN()
}

func SwitchOff() {
	go setStatus(stopping)
	go stopVPN()
}

func Quit() {
	if ctx.autostart != nil {
		ctx.autostart.Disable()
	}
	if ctx.Status != off {
		go setStatus(stopping)
		ctx.cfg.SetUserStoppedVPN(false)
	} else {
		ctx.cfg.SetUserStoppedVPN(true)
	}
	if ctx.bm != nil {
		ctx.bm.Close()
	}
	pid.ReleasePID()
}

func DonateAccepted() {
	donateAccepted()
}

func SubscribeToEvent(event string, f unsafe.Pointer) {
	subscribe(event, f)
}

type Providers struct {
	Default string                 `json:"default"`
	Data    []bitmask.ProviderOpts `json:"providers"`
}

type InitOpts struct {
	ProviderOptions  *bitmask.ProviderOpts
	SkipLaunch       bool
	Obfs4            bool
	DisableAutostart bool
	StartVPN         string
}

func InitOptsFromJSON(provider, providersJSON string) *InitOpts {
	providers := Providers{}
	err := json.Unmarshal([]byte(providersJSON), &providers)
	if err != nil {
		log.Println("ERROR while parsing json:", err)
	}
	if len(providers.Data) != 1 {
		panic("BUG: we do not support multi-provider yet")
	}
	providerOpts := &providers.Data[0]
	return &InitOpts{ProviderOptions: providerOpts}
}

func InitializeBitmaskContext(opts *InitOpts) {
	bitmask.ConfigureProvider(opts.ProviderOptions)

	initOnce.Do(func() { initializeContext(opts) })
	if ctx.bm != nil {
		ctx.LoginDialog = ctx.bm.NeedsCredentials()
		go ctx.updateStatus()
	}
	if ctx.AskForDonations {
		runDonationReminder()
	}
}

func RefreshContext() *C.char {
	c, _ := ctx.toJson()
	return C.CString(string(c))
}

func ResetError(errname string) {
	if ctx.Errors == errname {
		ctx.Errors = ""
	}
}

func ResetNotification(label string) {
	switch label {
	case "login_ok":
		ctx.LoginOk = false
		break
	default:
		break
	}
	go trigger(OnStatusChanged)
}

func InstallHelpers() {
	pickle.InstallHelpers()
}

func EnableMockBackend() {
	log.Println("[+] Mocking ui interaction on port 8080. \nTry 'curl localhost:8080/{on|off|failed}' to toggle status.")
	go enableMockBackend()
}

func EnableWebAPI(port string) {
	intPort, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal("Cannot parse port", port)
	}
	go enableWebAPI(intPort)
}

func GetVersion() *C.char {
	return C.CString(version.VERSION)
}
