package ketama

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
)

//Continuum models a sparse space that a given host occupies one or more points along
type Continuum struct {
	filename string

	mu        sync.RWMutex
	pointsMap map[uint32]*host
	points    []uint32
	hosts     map[string]*host
}

type host struct {
	name string
}

//BySize implements sort.Interface for unit32
type BySize []uint32

func (a BySize) Len() int           { return len(a) }
func (a BySize) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySize) Less(i, j int) bool { return a[i] < a[j] }

//GetHost looks up the host that is next nearest on the Continuum to where key hashes to and returns the name
func (c *Continuum) GetHost(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	point := hash(key)

	//nearest := c.findNearestPoint(point)
	nearest := c.findNearestPointBisect(point)
	h := c.pointsMap[nearest]
	return h.name
}

//Reload implements locking semantics to reload from a file
func (c *Continuum) Reload() bool {
	if c.filename == "" {
		return false
	}

	bytes, err := ioutil.ReadFile(c.filename)
	if err != nil {
		log.Fatal(err)
	}
	contents := string(bytes)
	contents = strings.Trim(contents, "\n ")
	hosts := strings.Split(contents, "\n")
	if len(hosts) < 1 {
		log.Fatal("watched file doesn't contain at least one newline")
	}
	parts := strings.Split(hosts[0], ",")

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(parts) == 1 {
		c.setHosts(hosts)
	} else if len(parts) == 2 {
		weightedHosts, err := parseHostWeights(hosts)
		if err != nil {
			log.Fatal(err)
		}
		c.setHostsWithWeights(weightedHosts)
	} else {
		log.Fatal("bad format for watched file, neither one nor two TSV fields on the first line")
	}
	return true
}

func (c *Continuum) findNearestPoint(point uint32) uint32 {
	//this is hideous linear walk through the array to find the first biggest point
	var firstBiggest uint32
	for _, p := range c.points {
		if p > point {
			firstBiggest = p
			break
		}
	}
	//check for point that is outside Continuum
	//we tried every point and found nothing bigger, therefore wrap
	if firstBiggest == 0 {
		firstBiggest = c.points[0]
	}
	return firstBiggest
}

func (c *Continuum) findNearestPointBisect(point uint32) uint32 {
	nearest := c.findNearestOffsetBisect(point)
	return c.points[nearest]
}

func (c *Continuum) findNearestOffsetBisect(point uint32) int {
	nearest := sort.Search(len(c.points), func(i int) bool { return c.points[i] > point })
	if nearest == len(c.points) {
		nearest = 0
	}
	return nearest
}

//GetHosts returns n hosts that are positioned adjecent on the continuum to provide the fundamental building block for replicas. Will return an error if n is larger than the set of known hosts.
func (c *Continuum) GetHosts(key string, n uint) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if n > uint(len(c.hosts)) {
		err := fmt.Errorf("Can't find %v hosts as requested, only know of %v hosts", n, len(c.hosts))
		return nil, err
	}
	point := hash(key)

	offset := c.findNearestOffsetBisect(point)

	var hosts []string
	var dupFound bool

	for len(hosts) != int(n) {
		dupFound = false
		point := c.points[offset]
		offset++
		if offset == len(c.points) {
			offset = 0
		}
		host := c.pointsMap[point]
		for _, h := range hosts {
			if h == host.name {
				dupFound = true
			}
		}
		if dupFound {
			continue
		}
		hosts = append(hosts, host.name)
	}
	return hosts, nil
}

//Make instantiates a Continuum with a set of hosts of equal weights
func Make(hosts []string) *Continuum {
	c := &Continuum{}
	c.setHosts(hosts)
	return c
}

//MakeWithWeights instantiates a Continuum with a set of hosts, each with an explicit weighting
func MakeWithWeights(hosts map[string]uint) *Continuum {
	c := &Continuum{}
	c.setHostsWithWeights(hosts)
	return c
}

//MakeWithFile instantiates a Continuun from a file that either lists a series of hosts or host, weight pairs. It is suitable to be used with a Watcher
func MakeWithFile(filename string) *Continuum {
	c := &Continuum{}
	c.filename = filename
	ok := c.Reload()
	if !ok {
		return nil
	}

	return c
}

//setHosts is a convience function when you have hosts with equal weight
func (c *Continuum) setHosts(hosts []string) {
	var weights = make(map[string]uint)
	for _, host := range hosts {
		weights[host] = 1
	}
	c.setHostsWithWeights(weights)
}

//setHostsWithWeights mirrors the java implementation of ketama and uses each host's relative weight to determine the number of points it occupies across the continuum
// https://github.com/RJ/ketama/blob/18cf9a7717dad0d8106a5205900a17617043fe2c/java_ketama/SockIOPool.java#L587-L607
func (c *Continuum) setHostsWithWeights(hostnames map[string]uint) {
	c.pointsMap = make(map[uint32]*host)
	c.hosts = make(map[string]*host)
	var totalWeight uint

	for _, weight := range hostnames {
		totalWeight += weight
	}

	for hostname, weight := range hostnames {
		h := &host{
			name: hostname,
		}
		c.hosts[hostname] = h
		factor := int(math.Floor((40 * float64(len(hostnames)) * float64(weight)) / float64(totalWeight)))
		//fmt.Println(factor)
		for i := 0; i < factor; i++ {
			key := h.name + "-" + strconv.Itoa(i)
			sum := md5.Sum([]byte(key))
			for j := 0; j < 4; j++ {
				point := uint32(sum[3+j*4])<<24 |
					uint32(sum[2+j*4])<<16 |
					uint32(sum[1+j*4])<<8 |
					uint32(sum[j*4])
				c.pointsMap[point] = h
				//fmt.Println("added point", point, "for host", h.name, "using key", key)
			}
		}
	}
	//use the points in the map to construct a sorted array to search later in GetHost
	for point := range c.pointsMap {
		c.points = append(c.points, point)
		sort.Sort(BySize(c.points))
	}
}

func hash(key string) uint32 {
	sum := md5.Sum([]byte(key))

	return uint32(sum[3])<<24 |
		uint32(sum[2])<<16 |
		uint32(sum[1])<<8 |
		uint32(sum[0])
}

func parseHostWeights(hosts []string) (map[string]uint, error) {
	hostWeights := make(map[string]uint)
	for _, line := range hosts {
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			return nil, errors.New("bad host file input")
		}
		i, err := strconv.Atoi(strings.Trim(parts[1], " "))
		if err != nil {
			log.Fatal(err)
		}
		hostWeights[parts[0]] = uint(i)
	}
	return hostWeights, nil
}
