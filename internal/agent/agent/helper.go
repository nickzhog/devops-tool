package agent

import (
	"bytes"
	"context"
	"io"
	"math/rand"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/nickzhog/devops-tool/pkg/encryption"
	"github.com/shirou/gopsutil/mem"
)

func (a *agent) sendRequest(ctx context.Context, url string, postData []byte) ([]byte, error) {
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}

	if a.publicKey != nil && len(postData) > 0 {
		newPostData, err := encryption.EncryptData(postData, a.publicKey)
		if err != nil {
			a.logger.Error(err)
			return nil, err
		}

		postData = newPostData
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(postData))
	if err != nil {
		return nil, err
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	ip := strings.Split(addrs[0].String(), "/")
	request.Header.Add("X-Real-IP", ip[0])

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	answer, err := io.ReadAll(res.Body)

	return answer, err
}

func setGaugeMetrics(m map[string]float64) {
	var memstat runtime.MemStats
	runtime.ReadMemStats(&memstat)

	m["Alloc"] = float64(memstat.Alloc)
	m["BuckHashSys"] = float64(memstat.BuckHashSys)
	m["Frees"] = float64(memstat.Frees)
	m["GCCPUFraction"] = float64(memstat.GCCPUFraction)
	m["GCSys"] = float64(memstat.GCSys)
	m["HeapAlloc"] = float64(memstat.HeapAlloc)
	m["HeapIdle"] = float64(memstat.HeapIdle)
	m["HeapInuse"] = float64(memstat.HeapInuse)
	m["HeapObjects"] = float64(memstat.HeapObjects)
	m["HeapReleased"] = float64(memstat.HeapReleased)
	m["HeapSys"] = float64(memstat.HeapSys)
	m["LastGC"] = float64(memstat.LastGC)
	m["Lookups"] = float64(memstat.Lookups)
	m["MCacheInuse"] = float64(memstat.MCacheInuse)
	m["MCacheSys"] = float64(memstat.MCacheSys)
	m["MSpanInuse"] = float64(memstat.MSpanInuse)
	m["MSpanSys"] = float64(memstat.MSpanSys)
	m["Mallocs"] = float64(memstat.Mallocs)
	m["NextGC"] = float64(memstat.NextGC)
	m["NumForcedGC"] = float64(memstat.NumForcedGC)
	m["NumGC"] = float64(memstat.NumGC)
	m["OtherSys"] = float64(memstat.OtherSys)
	m["PauseTotalNs"] = float64(memstat.PauseTotalNs)
	m["StackInuse"] = float64(memstat.StackInuse)
	m["StackSys"] = float64(memstat.StackSys)
	m["Sys"] = float64(memstat.Sys)
	m["TotalAlloc"] = float64(memstat.TotalAlloc)

	m["RandomValue"] = float64(rand.Int63n(1000))

	//gopsutil
	mem, _ := mem.VirtualMemory()

	m["CPUutilization1"] = float64(mem.UsedPercent)
	m["TotalMemory"] = float64(mem.Total)
	m["FreeMemory"] = float64(mem.Free)

}
