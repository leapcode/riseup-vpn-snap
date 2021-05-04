package bonafide

import (
	"errors"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"time"
)

const (
	maxGateways = 3
)

// A Gateway is a representation of gateways that is independent of the api version.
// If a given physical location offers different transports, they will appear as separate gateways.
type Gateway struct {
	Host         string
	IPAddress    string
	Location     string
	LocationName string
	CountryCode  string
	Ports        []string
	Protocols    []string
	Options      map[string]string
	Transport    string
}

// Load reflects the fullness metric that menshen returns, if available.
type Load struct {
	Host     string
	Fullness string
}

type gatewayDistance struct {
	gateway  Gateway
	distance int
}

/* a map between locations and hostnames, to be able to select by city */
type cityMap struct {
	gws map[string][]string
}

func (g *cityMap) Get(key string) []string {
	if val, ok := g.gws[key]; ok {
		return val
	}
	return make([]string, 0)
}

type gatewayPool struct {
	available  []Gateway
	userChoice []byte
	byCity     cityMap

	/* recommended is an array of hostnames, fetched from the old geoip service.
	 *  this should be deprecated in favor of recommendedWithLoad when new menshen is deployed */
	recommended         []string
	recommendedWithLoad []Load

	/* TODO locations are just used to get the timezone for each gateway. I
	* think it's easier to just merge that info into the version-agnostic
	* Gateway, that is passed from the eipService, and do not worry with
	* the location here */
	locations map[string]Location
}

func (p *gatewayPool) populateCityList() {
	for _, gw := range p.available {
		loc := gw.Location
		gws := p.cityMap.Get(loc)
		p.cityMap[loc] = append(gws, gw.Host)
	}
}

func (p *gatewayPool) getCities() []string {
	c := make([]string, 0)
	for city, _ := range p.byCity {
		c = append(c, city)
	}
	return c
}

func (p *gatewayPool) isValidCity(city string) bool {
	cities := p.getCities()
	valid := stringInSlice(city, cities)
	return valid
}

/* this method should only be used if we have no usable menshen list */
func (p *gatewayPool) getRandomGatewayByCity(city string) (Gateway, error) {
	if !p.isValidCity(city) {
		return Gateway{}, errors.New("bonafide: BUG not a valid city: " + city)
	}
	gws := p.byCity.Get(city)
	if len(gws) == 0 {
		return Gateway{}, errors.New("bonafide: BUG no gw for city " + city)
	}
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	host := gws[r.Intn(len(gws))]
	for _, gw := range p.available {
		if gw.Host == host {
			return gw, nil
		}
	}
	return Gateway{}, errors.New("bonafide: BUG should not reach here")
}

/* genLabels generates unique, human-readable labels for a gateway. It gives a serial
   number to each gateway in the same location (paris-1, paris-2,...). The
   current implementation will give a different label to each transport.
   An alternative (to discuss) would be to give the same label to the same hostname.
func (p *gatewayPool) genLabels() {
	acc := make(map[string]int)
	for i, gw := range p.available {
		if _, count := acc[gw.Location]; !count {
			acc[gw.Location] = 1
		} else {
			acc[gw.Location] += 1
		}
		gw.Label = gw.Location + "-" + strconv.Itoa(acc[gw.Location])
		p.available[i] = gw
	}
	for i, gw := range p.available {
		if acc[gw.Location] == 1 {
			gw.Label = gw.Location
			p.available[i] = gw
		}
	}
}
*/

/*
func (p *gatewayPool) getLabels() []string {
	labels := make([]string, 0)
	for _, gw := range p.available {
		labels = append(labels, gw.Label)
	}
	return labels
}

func (p *gatewayPool) isValidLabel(label string) bool {
	labels := p.getLabels()
	valid := stringInSlice(label, labels)
	return valid
}

func (p *gatewayPool) getGatewayByLabel(label string) (Gateway, error) {
	for _, gw := range p.available {
		if gw.Label == label {
			return gw, nil
		}
	}
	return Gateway{}, errors.New("bonafide: not a valid label")
}
*/

func (p *gatewayPool) getGatewayByIP(ip string) (Gateway, error) {
	for _, gw := range p.available {
		if gw.IPAddress == ip {
			return gw, nil
		}
	}
	return Gateway{}, errors.New("bonafide: not a valid ip address")
}

func (p *gatewayPool) setAutomaticChoice() {
	p.userChoice = []byte("")
}

func (p *gatewayPool) setUserChoice(city []byte) error {
	if !p.isValidCity(string(city)) {
		return errors.New("bonafide: not a valid city for gateway choice")
	}
	p.userChoice = city
	return nil
}

func (p *gatewayPool) setRanking(hostnames []string) {
	hosts := make([]string, 0)
	for _, gw := range p.available {
		hosts = append(hosts, gw.Host)
	}

	for _, host := range hostnames {
		if !stringInSlice(host, hosts) {
			log.Println("ERROR: invalid host in recommended list of hostnames", host)
			return
		}
	}

	p.recommended = hostnames
}

func (p *gatewayPool) getBest(transport string, tz, max int) ([]Gateway, error) {
	gws := make([]Gateway, 0)
	if len(p.userChoice) != 0 {
		/* FIXME this is random because we still do not get menshen to return us load. after "new" menshen is deployed,
		   we can just get them by the order menshen reeturned */
		gw, err := p.getRandomGatewayByCity(string(p.userChoice))
		gws = append(gws, gw)
		return gws, err
	} else if len(p.recommended) != 0 {
		return p.getGatewaysFromMenshen(transport, max)
	} else {
		return p.getGatewaysByTimezone(transport, tz, max)
	}
}

func (p *gatewayPool) getAll(transport string, tz int) ([]Gateway, error) {
	if len(p.recommended) != 0 {
		return p.getGatewaysFromMenshen(transport, 999)
	} else {
		return p.getGatewaysByTimezone(transport, tz, 999)
	}
}

func (p *gatewayPool) getGatewaysFromMenshen(transport string, max int) ([]Gateway, error) {
	gws := make([]Gateway, 0)
	for _, host := range p.recommended {
		for _, gw := range p.available {
			if gw.Transport != transport {
				continue
			}
			if gw.Host == host {
				gws = append(gws, gw)
			}
			if len(gws) == max {
				goto end
			}
		}
	}
end:
	return gws, nil
}

func (p *gatewayPool) getGatewaysByTimezone(transport string, tzOffsetHours, max int) ([]Gateway, error) {
	gws := make([]Gateway, 0)
	gwVector := []gatewayDistance{}

	for _, gw := range p.available {
		if gw.Transport != transport {
			continue
		}
		distance := 13
		gwOffset, err := strconv.Atoi(p.locations[gw.Location].Timezone)
		if err != nil {
			log.Printf("Error sorting gateways: %v", err)
			return gws, err
		} else {
			distance = tzDistance(tzOffsetHours, gwOffset)
		}
		gwVector = append(gwVector, gatewayDistance{gw, distance})
	}
	rand.Seed(time.Now().UnixNano())
	cmp := func(i, j int) bool {
		if gwVector[i].distance == gwVector[j].distance {
			return rand.Intn(2) == 1
		}
		return gwVector[i].distance < gwVector[j].distance
	}
	sort.Slice(gwVector, cmp)

	for _, gw := range gwVector {
		gws = append(gws, gw.gateway)
		if len(gws) == max {
			break
		}
	}
	return gws, nil
}

func newGatewayPool(eip *eipService) *gatewayPool {
	p := gatewayPool{}
	p.available = eip.getGateways()
	p.locations = eip.Locations
	p.populateCityList()
	return &p
}

func tzDistance(offset1, offset2 int) int {
	abs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}
	distance := abs(offset1 - offset2)
	if distance > 12 {
		distance = 24 - distance
	}
	return distance
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
