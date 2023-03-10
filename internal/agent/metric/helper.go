package metric

import (
	"bytes"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/mem"
)

func sendRequest(url string, postData []byte) ([]byte, error) {
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(postData))
	if err != nil {
		return nil, err
	}

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
