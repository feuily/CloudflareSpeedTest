package task

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/XIU2/CloudflareSpeedTest/utils"
)

const (
	tcpConnectTimeout = time.Second * 1
	maxRoutine        = 1000
	httpingRoutineCap = 128
	defaultRoutines   = 200
	defaultPort       = 443
	defaultPingTimes  = 4
)

var (
	Routines      = defaultRoutines
	TCPPort   int = defaultPort
	PingTimes int = defaultPingTimes
)

type Ping struct {
	wg      *sync.WaitGroup
	m       *sync.Mutex
	ips     []*net.IPAddr
	csv     utils.PingDelaySet
	control chan bool
	bar     *utils.Bar
}

func checkPingDefault() {
	if Routines <= 0 {
		Routines = defaultRoutines
	} else if Routines > maxRoutine {
		Routines = maxRoutine
	}
	if Httping && Routines > httpingRoutineCap {
		utils.Yellow.Printf("[提示] HTTPing 模式下并发过高可能触发服务商限速，已将并发从 %d 自动降到 %d。\n", Routines, httpingRoutineCap)
		Routines = httpingRoutineCap
	}
	if TCPPort <= 0 || TCPPort >= 65535 {
		TCPPort = defaultPort
	}
	if PingTimes <= 0 {
		PingTimes = defaultPingTimes
	}
	configureRequestLimiter()
}

func NewPing() *Ping {
	checkPingDefault()
	ips := loadIPRanges()
	return &Ping{
		wg:      &sync.WaitGroup{},
		m:       &sync.Mutex{},
		ips:     ips,
		csv:     make(utils.PingDelaySet, 0),
		control: make(chan bool, Routines),
		bar:     utils.NewBar(len(ips), "可用:", ""),
	}
}

func (p *Ping) Run() utils.PingDelaySet {
	if len(p.ips) == 0 {
		return p.csv
	}
	launchPause := time.Duration(0)
	if Httping && Routines > 0 {
		launchPause = time.Millisecond * time.Duration(Routines/50)
		if launchPause < time.Millisecond {
			launchPause = time.Millisecond
		}
		if launchPause > 5*time.Millisecond {
			launchPause = 5 * time.Millisecond
		}
	}
	if Httping {
		utils.Cyan.Printf("开始延迟测速（模式：HTTP, 端口：%d, 范围：%v ~ %v ms, 丢包：%.2f)\n", TCPPort, utils.InputMinDelay.Milliseconds(), utils.InputMaxDelay.Milliseconds(), utils.InputMaxLossRate)
	} else {
		utils.Cyan.Printf("开始延迟测速（模式：TCP, 端口：%d, 范围：%v ~ %v ms, 丢包：%.2f)\n", TCPPort, utils.InputMinDelay.Milliseconds(), utils.InputMaxDelay.Milliseconds(), utils.InputMaxLossRate)
	}
	for _, ip := range p.ips {
		p.wg.Add(1)
		p.control <- false
		go p.start(ip)
		if launchPause > 0 {
			time.Sleep(launchPause)
		}
	}
	p.wg.Wait()
	p.bar.Done()
	sort.Sort(p.csv)
	return p.csv
}

func (p *Ping) start(ip *net.IPAddr) {
	defer p.wg.Done()
	p.tcpingHandler(ip)
	<-p.control
}

// bool connectionSucceed float32 time
func (p *Ping) tcping(ip *net.IPAddr) (bool, time.Duration) {
	waitForRequestSlot()
	startTime := time.Now()
	var fullAddress string
	if isIPv4(ip.String()) {
		fullAddress = fmt.Sprintf("%s:%d", ip.String(), TCPPort)
	} else {
		fullAddress = fmt.Sprintf("[%s]:%d", ip.String(), TCPPort)
	}
	conn, err := net.DialTimeout("tcp", fullAddress, tcpConnectTimeout)
	if err != nil {
		return false, 0
	}
	defer conn.Close()
	duration := time.Since(startTime)
	return true, duration
}

// pingReceived pingTotalTime
func (p *Ping) checkConnection(ip *net.IPAddr) (recv int, totalDelay time.Duration, colo string) {
	if Httping {
		recv, totalDelay, colo = p.httping(ip)
		return
	}
	colo = "" // TCPing 不获取 colo
	failed := 0
	for i := 0; i < PingTimes; i++ {
		if ok, delay := p.tcping(ip); ok {
			recv++
			totalDelay += delay
			continue
		}
		failed++
		if shouldAbortByLossRate(failed) {
			break
		}
	}
	return
}

func (p *Ping) appendIPData(data *utils.PingData) int {
	p.m.Lock()
	defer p.m.Unlock()
	p.csv = append(p.csv, utils.CloudflareIPData{
		PingData: data,
	})
	return len(p.csv)
}

func (p *Ping) availableCount() int {
	p.m.Lock()
	defer p.m.Unlock()
	return len(p.csv)
}

// handle tcping
func (p *Ping) tcpingHandler(ip *net.IPAddr) {
	recv, totalDlay, colo := p.checkConnection(ip)
	if recv == 0 {
		p.bar.Grow(1, strconv.Itoa(p.availableCount()))
		return
	}
	data := &utils.PingData{
		IP:       ip,
		Sended:   PingTimes,
		Received: recv,
		Delay:    totalDlay / time.Duration(recv),
		Colo:     colo,
	}
	p.bar.Grow(1, strconv.Itoa(p.appendIPData(data)))
}
