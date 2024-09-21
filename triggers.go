package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/hypebeast/go-osc/osc"
)

type TriggerPoint struct {
	Path string    `yaml:"path"`
	Time time.Time `yaml:"time"`
	Done bool      `yaml:"-"`
}

func TriggerFromString(path, t string) TriggerPoint {
	sparts := strings.Split(t, ":")
	parts := make([]int, 3)
	var err error
	for i := range sparts {
		parts[i], err = strconv.Atoi(sparts[i])
		if err != nil {
			parts[i] = 0
		}
	}
	tp := TriggerPoint{
		Path: path,
		Time: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), parts[0], parts[1], parts[2], 0, time.Now().Location()),
		Done: false,
	}
	return tp
}
func TriggerFromTime(path string, t time.Time) TriggerPoint {
	tp := TriggerPoint{
		Path: path,
		Time: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Now().Location()),
		Done: false,
	}
	return tp
}

type Triggers struct {
	points []TriggerPoint
	ticker *time.Ticker
	lock   sync.RWMutex
	done   chan bool
	Host   string
	Port   int
}

func (t *Triggers) AddPoint(tp TriggerPoint) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.points = append(t.points, tp)
	sort.SliceStable(t.points, func(i, j int) bool {
		return t.points[i].Time.Before(t.points[j].Time)
	})
}

func (t *Triggers) RemovePoint(index int) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if len(t.points) > index+1 {
		t.points = append(t.points[:index], t.points[index+1:]...)
	} else {
		t.points = t.points[:index]
	}
}

func (t *Triggers) GetTriggers() []TriggerPoint {
	t.lock.RLock()
	defer t.lock.RUnlock()
	buf := make([]TriggerPoint, len(t.points))
	_ = copy(buf, t.points)
	return buf
}

func (t *Triggers) Stop() {
	t.ticker.Stop()
	t.done <- true
}

func (t *Triggers) Start() {
	go func() {
		for {
			select {
			case <-t.done:
				return
			case <-t.ticker.C:
				t.lock.Lock()
				for i, point := range t.points {
					if t.Host == "" || t.Port == 0 {
						continue
					}
					if !point.Done && point.Time.Before(time.Now()) {
						LOG(fmt.Sprintf("sending point %s at %s", point.Path, point.Time.String()))
						point.Done = true
						client := osc.NewClient(t.Host, t.Port)
						msg := osc.NewMessage(point.Path)
						err := client.Send(msg)
						if err != nil {
							LOG(err.Error())
						}
						t.points[i] = point
					}
				}
				t.lock.Unlock()
			}
		}
	}()
}

func (t *Triggers) ToRows() []table.Row {
	t.lock.RLock()
	defer t.lock.RUnlock()
	trs := make([]table.Row, len(t.points))
	for i := range trs {
		dne := ""
		if t.points[i].Done {
			dne = "x"
		}
		trs[i] = []string{t.points[i].Time.Format("15:04:05"), t.points[i].Path, dne}
	}
	return trs
}

func NewTriggers(host string, port int) *Triggers {
	trg := Triggers{
		points: make([]TriggerPoint, 0),
		ticker: time.NewTicker(1 * time.Second),
		done:   make(chan bool),
		Host:   host,
		Port:   port,
	}
	trg.Start()
	return &trg
}
