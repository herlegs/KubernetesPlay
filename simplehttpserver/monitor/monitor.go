package monitor

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	dto2 "gitlab.myteksi.net/gophers/go/end/common/k8s/metricserver/dto"

	"github.com/herlegs/KubernetesPlay/k8s"

	"gitlab.myteksi.net/gophers/go/end/tools/pipelinemeta/dto"
	"gitlab.myteksi.net/gophers/go/end/tools/pipelinemeta/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	namespace        = "sprinkler8"
	allNamespacePods = "apis/metrics.k8s.io/v1beta1/pods"
	namespacePods    = "apis/metrics.k8s.io/v1beta1/namespaces/%s/pods"
)

func InitResouceMonitor(env string) {
	clientset, err := k8s.NewInClusterClient()
	if err != nil {
		fmt.Printf("error creating in-cluster client: %v\n", err)
	}
	pipelines := meta.GetPipelineMeta()[env]

	pipelineInfoMap := map[string]*dto.PipelineInfo{}
	for _, p := range pipelines {
		pipelineInfoMap[p.Name] = p
	}

	go func() {
		seconds, _ := strconv.ParseInt(os.Getenv("INTERVAL_SECOND"), 10, 64)
		if seconds <= 0 {
			seconds = 60
		}
		ticker := time.NewTicker(time.Second * time.Duration(seconds))
		defer ticker.Stop()
		AnalyzePipelineMeta(clientset, pipelineInfoMap)
		for {
			select {
			case <-ticker.C:
				AnalyzePipelineMeta(clientset, pipelineInfoMap)
			}
		}
	}()
}

// AnalyzePipelineMeta ...
func AnalyzePipelineMeta(clientset *kubernetes.Clientset, pipelineInfoMap map[string]*dto.PipelineInfo) {
	/*
		for testing
	*/
	//endpoint := os.Getenv("ENDPOINT")
	//bytes, err := clientset.RESTClient().Get().AbsPath(endpoint).DoRaw()
	//fmt.Printf("err:%v,res:%v", err, string(bytes))
	//clientset.CoreV1().
	testAPI(clientset)

	//UpdateActualResource(clientset, pipelineInfoMap)

	//AnalyzeMergePipelines(pipelineInfoMap)

	//AnalyzeOverResource(pipelineInfoMap)
}

// testAPI
func testAPI(clientset *kubernetes.Clientset) {
	l, err := clientset.CoreV1().Pods("sprinkler8").List(metav1.ListOptions{
		//LabelSelector: "app=helloapp",
	})
	if err != nil {
		fmt.Printf("error calling api: %v\n", err)
	}

	fmt.Printf("try listing result...\n")
	l.Items = l.Items[:8]
	bytes, err := json.Marshal(l)
	if err != nil {
		fmt.Printf("err marshal: %v\n", err)
	} else {
		fmt.Printf("%v\n", string(bytes))
	}
}

// UpdateActualResource ...
func UpdateActualResource(clientset *kubernetes.Clientset, pipelineInfoMap map[string]*dto.PipelineInfo) {
	podMetricList, err := GetPodMetricList(clientset, namespace)
	if err != nil {
		fmt.Printf("error getting pods metric: %v\n", err)
		return
	}

	for _, item := range podMetricList.Items {
		for _, c := range item.Containers {
			// container name is pipeline name (except for other non-app containers)
			pipelineName := c.Name
			if pipelineInfoMap[pipelineName] == nil {
				continue
			}
			cpu := dto.ParseCPU(c.Usage.CPU)
			mem := dto.ParseMemory(c.Usage.Memory)

			if pipelineInfoMap[pipelineName].ActualTotal == nil {
				pipelineInfoMap[pipelineName].ActualTotal = &dto.Resource{}
			}

			pipelineInfoMap[pipelineName].ActualTotal.CPU += cpu
			pipelineInfoMap[pipelineName].ActualTotal.Memory += mem
			pipelineInfoMap[pipelineName].ActualPods++
		}
	}
}

// AnalyzeMergePipelines ...
func AnalyzeMergePipelines(pipelineInfoMap map[string]*dto.PipelineInfo) {
	streamKeyMap := map[string][]*dto.PipelineInfo{}
	for _, info := range pipelineInfoMap {
		streamKeyMap[info.StreamsKey] = append(streamKeyMap[info.StreamsKey], info)
	}
	logs := "\n===Merging Pipelines===\n"
	savedCPU, savedMem := dto.CPU(0), dto.Memory(0)
	for key, pipelines := range streamKeyMap {
		if len(pipelines) <= 1 {
			continue
		}
		totalC, totalM := dto.CPU(0), dto.Memory(0)
		maxC, maxM := dto.CPU(0), dto.Memory(0)
		var costingNotes []string
		for _, p := range pipelines {
			if p.ActualTotal == nil {
				continue
			}
			totalC += p.ActualTotal.CPU
			totalM += p.ActualTotal.Memory
			if p.ActualTotal.CPU > maxC {
				maxC = p.ActualTotal.CPU
			}
			if p.ActualTotal.Memory > maxM {
				maxM = p.ActualTotal.Memory
			}
			costingNotes = append(costingNotes, fmt.Sprintf("%v [%v]CPU(mCore) [%v]Mem(Mb)", p.Name, p.ActualTotal.CPU/dto.MCore, p.ActualTotal.Memory/dto.MB))
		}
		savedCPU += totalC - maxC
		savedMem += totalM - maxM
		logs += fmt.Sprintf("save [%v]CPU(mCore) and [%v]Mem(Mb) consuming streams[%v] for merging (%v)\n", (totalC-maxC)/dto.MCore, (totalM-maxM)/dto.MB, key, strings.Join(costingNotes, ","))
	}
	logs += fmt.Sprintf("\nsaved total [%v]CPU(mCore) and [%v]Mem(Mb)\n", savedCPU/dto.MCore, savedMem/dto.MB)
	fmt.Printf(logs)
}

// AnalyzeOverResource ...
func AnalyzeOverResource(pipelineInfoMap map[string]*dto.PipelineInfo) {
	logs := "\n===Adjusting Resources===\n"
	for _, p := range pipelineInfoMap {
		if p.ActualPods > 0 {
			avgCPU := p.ActualTotal.CPU / dto.CPU(p.ActualPods)
			avgMem := p.ActualTotal.Memory / dto.Memory(p.ActualPods)

			if float64(avgCPU)/float64(p.Requested.CPU) < 0.5 {
				logs += fmt.Sprintf("[%v] request CPU [%v]mCore too high compared to used [%v]mCore\n", p.Name, p.Requested.CPU/dto.MCore, avgCPU/dto.MCore)
			}

			if float64(avgMem)/float64(p.Requested.Memory) < 0.5 {
				logs += fmt.Sprintf("[%v] request Memory [%v]Mb too high compared to used [%v]Mb\n", p.Name, p.Requested.Memory/dto.MB, avgMem/dto.MB)
			}
		}
	}
	fmt.Printf(logs)
}

// GetPodMetricList returns the pods metrics in a namespace (first param)
// if namespace is not provided, will get from all namespaces
func GetPodMetricList(clientset *kubernetes.Clientset, namespaces ...string) (*dto2.PodMetricsList, error) {
	var queryPath string
	if len(namespaces) == 0 {
		queryPath = allNamespacePods
	} else {
		queryPath = fmt.Sprintf(namespacePods, namespaces[0])
	}
	data, err := clientset.RESTClient().Get().AbsPath(queryPath).DoRaw()
	if err != nil {
		return nil, fmt.Errorf("error querying metric server: %v", err.Error())
	}
	metrics := &dto2.PodMetricsList{}
	err = json.Unmarshal(data, metrics)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling metric result: %v", err.Error())
	}
	return metrics, nil
}
