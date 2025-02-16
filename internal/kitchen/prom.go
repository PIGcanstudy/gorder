// package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"log"
// 	"net/http"
// 	"time"

// 	"github.com/prometheus/client_golang/prometheus"
// 	"github.com/prometheus/client_golang/prometheus/collectors"
// 	"github.com/prometheus/client_golang/prometheus/promhttp"
// 	"golang.org/x/exp/rand"
// )

// const (
// 	testAddr = "localhost:9123"
// )

// type request struct {
// 	StatusCode string
// }

// func produceData() {
// 	codes := []string{"503", "404", "400", "200", "304", "500"}
// 	for {
// 		body, _ := json.Marshal(request{
// 			StatusCode: codes[rand.Intn(len(codes))],
// 		})
// 		requestBody := bytes.NewBuffer(body)
// 		http.Post("http://"+testAddr, "application/json", requestBody)
// 		log.Printf("send request=%s to %s", requestBody.String(), testAddr)
// 		time.Sleep(2 * time.Second)
// 	}
// }

// var httpStatusCodeCounter = prometheus.NewCounterVec(
// 	prometheus.CounterOpts{
// 		Name: "http_status_code_counter",
// 		Help: "Count http status code",
// 	},
// 	[]string{"status_code"},
// )

// func sendMetricsHandler(w http.ResponseWriter, r *http.Request) {
// 	var req request
// 	defer func() {
// 		httpStatusCodeCounter.WithLabelValues(req.StatusCode).Inc()
// 		log.Printf("add 1 to %s", req.StatusCode)
// 	}()
// 	_ = json.NewDecoder(r.Body).Decode(&req)
// 	log.Printf("receive req:%+v", req)
// 	_, _ = w.Write([]byte(req.StatusCode))
// }

// func main() {
// 	go produceData()
// 	reg := prometheus.NewRegistry()
// 	prometheus.WrapRegistererWith(prometheus.Labels{"serviceName": "demo-service"}, reg).MustRegister(
// 		collectors.NewGoCollector(),                                       // 采集go运行时信息
// 		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}), // 采集处理器信息
// 		httpStatusCodeCounter,                                             // 采集自定义指标
// 	)
// 	// localhost:9123/metrics 启动http服务器
// 	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
// 	http.HandleFunc("/", sendMetricsHandler)
// 	log.Fatal(http.ListenAndServe(testAddr, nil))
// }
