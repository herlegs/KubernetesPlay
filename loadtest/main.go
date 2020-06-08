package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/herlegs/KafkaPlay/testratelimiter/rate"
)

const (
	defaultWorkerNum          = 10
	defaultStatsIntervalInSec = 5
)

var (
	qpsLimit      int64
	endpoint      string
	workerNum     int
	statsInterval time.Duration

	stats []*Result
	lock  sync.RWMutex
)

type Result struct {
	StatusCode int
	RemoteHost string
	Error      string
}

// main
func main() {
	Init()

	for i := 0; i < workerNum; i++ {
		startTestWorker()
	}

	// blocking
	startStatsReporter()
}

func startTestWorker() {
	limit := float64(qpsLimit) / float64(workerNum)

	limiter := rate.NewLimiter(rate.Limit(limit), 1)
	go func() {
		for {
			if limiter.Allow() {
				makeRequest()
			}
		}
	}()
}

func makeRequest() {
	//resp, err := testing.HTTPGet(endpoint, nil, nil)
	body := requests[0]
	resp, err := http.Post(endpoint, "application/json", bytes.NewBufferString(body))

	result := &Result{}
	if err != nil {
		result.Error = err.Error()
	}
	if resp != nil {
		result.StatusCode = resp.StatusCode
		if resp.Request != nil {
			result.RemoteHost = resp.Request.Host
		}
	}
	lock.Lock()
	defer lock.Unlock()
	stats = append(stats, result)
}

func startStatsReporter() {
	ticker := time.NewTicker(statsInterval)
	defer ticker.Stop()

	for range ticker.C {
		lock.Lock()
		//TODO get stats
		totalCount := 0
		statusCodeMap := map[int]int{}
		errorMap := map[string]int{}
		for _, s := range stats {
			statusCodeMap[s.StatusCode]++
			totalCount++
			if s.Error != "" {
				errorMap[s.Error]++
			}
		}
		var infos []string
		for code, cnt := range statusCodeMap {
			infos = append(infos, fmt.Sprintf("status code: %v count: %v, percentage:%.2f%%", code, cnt, float64(cnt)/float64(totalCount)*100))
		}
		for e, cnt := range errorMap {
			infos = append(infos, fmt.Sprintf("error [%v] count [%v]", e, cnt))
		}
		fmt.Printf(`
===%v requests for past %v seconds till now %v===:
%s
`, totalCount, statsInterval, time.Now().Format("2006-01-02T15:04:05-07:00"), strings.Join(infos, "\n"),
		)
		//reset
		stats = make([]*Result, 0, int(qpsLimit)*int(statsInterval/time.Second)*2)
		lock.Unlock()
	}
}

func Init() {
	limitStr := os.Getenv("TOTAL_QPS_LIMIT")
	qpsLimit, _ = strconv.ParseInt(limitStr, 10, 64)
	endpoint = os.Getenv("ENDPOINT")
	workers := os.Getenv("WORKERS")
	if workers == "" {
		workerNum = defaultWorkerNum
	} else {
		parsed, _ := strconv.ParseInt(workers, 10, 64)
		workerNum = int(parsed)
	}
	intervalInSecStr := os.Getenv("STATS_INTERVAL_SECONDS")
	if intervalInSecStr == "" {
		statsInterval = defaultStatsIntervalInSec * time.Second
	} else {
		secs, _ := strconv.ParseInt(intervalInSecStr, 10, 64)
		statsInterval = time.Duration(secs) * time.Second
	}

	lock.Lock()
	stats = make([]*Result, 0, int(qpsLimit)*int(statsInterval/time.Second)*2)
	lock.Unlock()

	fmt.Printf("endpoint:%v, qps limit:%v, workers: %v, stats interval: %v\n", endpoint, qpsLimit, workerNum, statsInterval)
}

var requests = []string{
	`{"request_info": {"feature_flag": 102, "request_id": "1_10_1589336671_13-May", "current_time": 1589336671, "city_id": 6, "min_number_drivers_to_filter_threshold": 1, "min_angle_to_filter_threshold": 180, "min_forward_efficiency_threshold": {"6": {"default": 1.02}}}, "bundles": [{"bundle_info": {"code": "IOS-3944887683-1384", "size": 1, "request_time": 1589336488, "restaurant_id": 21, "recipient_id": 940, "pickup": {"location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}, "pickup_time_lb": 1589337324}, "delivery": {"location": {"latitude": -6.093541910599394, "longitude": 106.98402330454529}, "delivery_time_ub": 1589343144, "delivery_buffer_time": 4709}}, "drivers": [{"driver_id": "6877799611", "vehicle_info": {"vehicle_type_id": 999, "vehicle_capacity": 2}, "location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}, "current_step_index": 0, "driver_status": "waiting", "plan_steps": [{"is_pickup_step": true, "bundle_code": "ADR-7112823989-7192", "restaurant_id": 21, "pickup_time_lb": 1589336720, "request_time": 1589335720, "arrive_at": 0, "departure_at": 1589336720, "bundle_size": 1, "location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}}, {"is_pickup_step": false, "bundle_code": "ADR-7112823989-7192", "request_time": 1589335720, "delivery_time_ub": 1589341199, "delivery_buffer_time": 3173, "recipient_id": 385, "restaurant_id": 21, "arrive_at": 1589338026, "departure_at": 1589338106, "bundle_size": 1, "location": {"latitude": -6.225056686619276, "longitude": 107.00824380510859}}], "eta_matrix": [[1306.0, 0.0, 1111.0], [0.0, 1306.0, 739.0], [1306.0, 0.0, 1111.0], [739.0, 1111.0, 0.0]]}, {"driver_id": "4306830741", "vehicle_info": {"vehicle_type_id": 999, "vehicle_capacity": 2}, "location": {"latitude": -6.141539411156077, "longitude": 106.77267286378749}, "current_step_index": -1, "driver_status": "driving", "plan_steps": [{"is_pickup_step": true, "bundle_code": "IOS-0198831876-4710", "restaurant_id": 21, "pickup_time_lb": 1589337641, "request_time": 1589336641, "arrive_at": 88, "departure_at": 1589337641, "bundle_size": 1, "location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}}, {"is_pickup_step": false, "bundle_code": "IOS-0198831876-4710", "request_time": 1589336641, "delivery_time_ub": 1589342186, "delivery_buffer_time": 3506, "recipient_id": 496, "restaurant_id": 21, "arrive_at": 1589338680, "departure_at": 1589338760, "bundle_size": 1, "location": {"latitude": -6.0871706629155415, "longitude": 106.60776863787979}}], "eta_matrix": [[88.0, 960.0, 88.0, 1199.0], [0.0, 1039.0, 0.0, 1111.0], [1039.0, 0.0, 1039.0, 2082.0], [0.0, 1039.0, 0.0, 1111.0], [1111.0, 2082.0, 1111.0, 0.0]]}, {"driver_id": "1056416544", "vehicle_info": {"vehicle_type_id": 999, "vehicle_capacity": 2}, "location": {"latitude": -6.188267939353669, "longitude": 106.77152055842896}, "current_step_index": -1, "driver_status": "driving", "plan_steps": [{"is_pickup_step": true, "bundle_code": "ADR-7747924742-3159", "restaurant_id": 21, "pickup_time_lb": 1589337463, "request_time": 1589336463, "arrive_at": 288, "departure_at": 1589337463, "bundle_size": 1, "location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}}, {"is_pickup_step": false, "bundle_code": "ADR-7747924742-3159", "request_time": 1589336463, "delivery_time_ub": 1589342813, "delivery_buffer_time": 4558, "recipient_id": 892, "restaurant_id": 21, "arrive_at": 1589338255, "departure_at": 1589338335, "bundle_size": 1, "location": {"latitude": -6.187310004043486, "longitude": 106.92314950238702}}], "eta_matrix": [[288.0, 839.0, 288.0, 1287.0], [0.0, 792.0, 0.0, 1111.0], [792.0, 0.0, 792.0, 618.0], [0.0, 792.0, 0.0, 1111.0], [1111.0, 618.0, 1111.0, 0.0]]}, {"driver_id": "6383854745", "vehicle_info": {"vehicle_type_id": 999, "vehicle_capacity": 2}, "location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}, "current_step_index": 0, "driver_status": "waiting", "plan_steps": [{"is_pickup_step": true, "bundle_code": "ADR-8757546005-9457", "restaurant_id": 21, "pickup_time_lb": 1589337798, "request_time": 1589336798, "arrive_at": 0, "departure_at": 1589337798, "bundle_size": 1, "location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}}, {"is_pickup_step": false, "bundle_code": "ADR-8757546005-9457", "request_time": 1589336798, "delivery_time_ub": 1589342500, "delivery_buffer_time": 4603, "recipient_id": 622, "restaurant_id": 21, "arrive_at": 1589337897, "departure_at": 1589337977, "bundle_size": 1, "location": {"latitude": -6.154177530362238, "longitude": 106.77901409357453}}], "eta_matrix": [[99.0, 0.0, 1111.0], [0.0, 99.0, 1183.0], [99.0, 0.0, 1111.0], [1183.0, 1111.0, 0.0]]}, {"driver_id": "8170252727", "vehicle_info": {"vehicle_type_id": 999, "vehicle_capacity": 2}, "location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}, "current_step_index": 0, "driver_status": "waiting", "plan_steps": [{"is_pickup_step": true, "bundle_code": "ADR-3577881612-7559", "restaurant_id": 21, "pickup_time_lb": 1589337138, "request_time": 1589336138, "arrive_at": 0, "departure_at": 1589337138, "bundle_size": 1, "location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}}, {"is_pickup_step": false, "bundle_code": "ADR-3577881612-7559", "request_time": 1589336138, "delivery_time_ub": 1589342409, "delivery_buffer_time": 3808, "recipient_id": 203, "restaurant_id": 21, "arrive_at": 1589338601, "departure_at": 1589338681, "bundle_size": 1, "location": {"latitude": -6.0196050667775856, "longitude": 107.02441930477335}}], "eta_matrix": [[1463.0, 0.0, 1111.0], [0.0, 1463.0, 465.0], [1463.0, 0.0, 1111.0], [465.0, 1111.0, 0.0]]}, {"driver_id": "5021635492", "vehicle_info": {"vehicle_type_id": 999, "vehicle_capacity": 2}, "location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}, "current_step_index": 0, "driver_status": "waiting", "plan_steps": [{"is_pickup_step": true, "bundle_code": "IOS-7963250423-7284", "restaurant_id": 21, "pickup_time_lb": 1589336822, "request_time": 1589335822, "arrive_at": 0, "departure_at": 1589336822, "bundle_size": 1, "location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}}, {"is_pickup_step": false, "bundle_code": "IOS-7963250423-7284", "request_time": 1589335822, "delivery_time_ub": 1589342334, "delivery_buffer_time": 4703, "recipient_id": 848, "restaurant_id": 21, "arrive_at": 1589337631, "departure_at": 1589337711, "bundle_size": 1, "location": {"latitude": -6.2632714558401545, "longitude": 106.71104505777238}}], "eta_matrix": [[809.0, 0.0, 1111.0], [0.0, 809.0, 1778.0], [809.0, 0.0, 1111.0], [1778.0, 1111.0, 0.0]]}, {"driver_id": "3393045231", "vehicle_info": {"vehicle_type_id": 999, "vehicle_capacity": 2}, "location": {"latitude": -6.174273188652419, "longitude": 106.77364575755176}, "current_step_index": -1, "driver_status": "driving", "plan_steps": [{"is_pickup_step": true, "bundle_code": "ADR-7133320568-5999", "restaurant_id": 21, "pickup_time_lb": 1589338101, "request_time": 1589337101, "arrive_at": 211, "departure_at": 1589338101, "bundle_size": 1, "location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}}, {"is_pickup_step": false, "bundle_code": "ADR-7133320568-5999", "request_time": 1589337101, "delivery_time_ub": 1589341864, "delivery_buffer_time": 3728, "recipient_id": 227, "restaurant_id": 21, "arrive_at": 1589338136, "departure_at": 1589338216, "bundle_size": 1, "location": {"latitude": -6.141974043663515, "longitude": 106.79397882327889}}], "eta_matrix": [[211.0, 211.0, 211.0, 1246.0], [0.0, 35.0, 0.0, 1111.0], [35.0, 0.0, 35.0, 1085.0], [0.0, 35.0, 0.0, 1111.0], [1111.0, 1085.0, 1111.0, 0.0]]}, {"driver_id": "1038207331", "vehicle_info": {"vehicle_type_id": 999, "vehicle_capacity": 2}, "location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}, "current_step_index": 0, "driver_status": "waiting", "plan_steps": [{"is_pickup_step": true, "bundle_code": "IOS-7064493203-5408", "restaurant_id": 21, "pickup_time_lb": 1589337292, "request_time": 1589336292, "arrive_at": 0, "departure_at": 1589337292, "bundle_size": 1, "location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}}, {"is_pickup_step": false, "bundle_code": "IOS-7064493203-5408", "request_time": 1589336292, "delivery_time_ub": 1589342233, "delivery_buffer_time": 3841, "recipient_id": 349, "restaurant_id": 21, "arrive_at": 1589338392, "departure_at": 1589338472, "bundle_size": 1, "location": {"latitude": -6.202910515171683, "longitude": 106.97680249721263}}], "eta_matrix": [[1100.0, 0.0, 1111.0], [0.0, 1100.0, 606.0], [1100.0, 0.0, 1111.0], [606.0, 1111.0, 0.0]]}, {"driver_id": "3779904284", "vehicle_info": {"vehicle_type_id": 999, "vehicle_capacity": 2}, "location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}, "current_step_index": 0, "driver_status": "waiting", "plan_steps": [{"is_pickup_step": true, "bundle_code": "ADR-6450127619-2696", "restaurant_id": 21, "pickup_time_lb": 1589337658, "request_time": 1589336658, "arrive_at": 0, "departure_at": 1589337658, "bundle_size": 1, "location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}}, {"is_pickup_step": false, "bundle_code": "ADR-6450127619-2696", "request_time": 1589336658, "delivery_time_ub": 1589341167, "delivery_buffer_time": 2444, "recipient_id": 753, "restaurant_id": 21, "arrive_at": 1589338723, "departure_at": 1589338803, "bundle_size": 1, "location": {"latitude": -6.171264524099298, "longitude": 106.97829563543898}}], "eta_matrix": [[1065.0, 0.0, 1111.0], [0.0, 1065.0, 430.0], [1065.0, 0.0, 1111.0], [430.0, 1111.0, 0.0]]}, {"driver_id": "5351322785", "vehicle_info": {"vehicle_type_id": 999, "vehicle_capacity": 2}, "location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}, "current_step_index": 0, "driver_status": "waiting", "plan_steps": [{"is_pickup_step": true, "bundle_code": "ADR-4869529905-8935", "restaurant_id": 21, "pickup_time_lb": 1589337159, "request_time": 1589336159, "arrive_at": 0, "departure_at": 1589337159, "bundle_size": 1, "location": {"latitude": -6.138927363600913, "longitude": 106.78840161720981}}, {"is_pickup_step": false, "bundle_code": "ADR-4869529905-8935", "request_time": 1589336159, "delivery_time_ub": 1589340994, "delivery_buffer_time": 3256, "recipient_id": 711, "restaurant_id": 21, "arrive_at": 1589337738, "departure_at": 1589337818, "bundle_size": 1, "location": {"latitude": -6.2268181343133815, "longitude": 106.73131028625568}}], "eta_matrix": [[579.0, 0.0, 1111.0], [0.0, 579.0, 1580.0], [579.0, 0.0, 1111.0], [1580.0, 1111.0, 0.0]]}]}]}`,
}
