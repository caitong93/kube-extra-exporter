package manager

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/caicloud/nirvana/log"

	v1 "k8s.io/api/core/v1"
)

type podData struct {
	Name       string
	Namespace  string
	UID        string
	qos        v1.PodQOSClass
	Containers []*containerData
}

func newPodData(po *v1.Pod) *podData {
	return &podData{
		Name:      po.Name,
		Namespace: po.Namespace,
		UID:       string(po.UID),
		qos:       po.Status.QOSClass,
	}
}

func (pd *podData) addContainer(ID string) error {
	newCont, err := newContainerData(pd.qos, pd.UID, ID)
	if err != nil {
		return err
	}
	pd.Containers = append(pd.Containers, newCont)
	return nil
}

func (pd *podData) getContainer(ID string) *containerData {
	for _, cont := range pd.Containers {
		if cont.ID == ID {
			return cont
		}
	}
	return nil
}

func (pd *podData) deleteContainer(ID string) {
	if pd.getContainer(ID) == nil {
		return
	}
	for i, cont := range pd.Containers {
		if cont.ID == ID {
			tmp := pd.Containers
			pd.Containers = append([]*containerData{}, tmp[:i]...)
			pd.Containers = append(pd.Containers, tmp[i+1:]...)
		}
	}
}

func (pd *podData) onePid() int {
	for _, cont := range pd.Containers {
		if len(cont.Pids) > 0 {
			return cont.Pids[0]
		}
	}
	return -1
}

type containerData struct {
	ID   string
	Pids []int
}

// Remove cri prefix, e.g. docker://999a54e3e9eb3c1bf58c96788850aa03a47d3e3c009da9ecae8d2edfdba5a328
func parseContainerID(ID string) string {
	i := strings.Index(ID, "://")
	if i < 0 {
		return ID
	}
	return ID[i+3:]
}

func newContainerData(qos v1.PodQOSClass, podUID, containerID string) (*containerData, error) {
	containerID = parseContainerID(containerID)
	cgroupPath, err := resolveCgroupPath(qos, podUID, containerID)
	if err != nil {
		return nil, fmt.Errorf("err get cgroup path for podID=%v, containerID=%v: %v", podUID, containerID, err)
	}
	pids, err := parseCgroupTasks(cgroupPath)
	if err != nil {
		return nil, fmt.Errorf("err parse cgroup tasks: %v", err)
	}

	return &containerData{
		ID:   containerID,
		Pids: pids,
	}, nil
}

func resolveCgroupPath(qos v1.PodQOSClass, podUID, contID string) (string, error) {
	var cPath string

	switch qos {
	case v1.PodQOSGuaranteed:
		cPath = path.Join(hostRootfsPath, fmt.Sprintf("/sys/fs/cgroup/cpu/kubepods/pod%s/%s", podUID, contID))
	case v1.PodQOSBestEffort:
		cPath = path.Join(hostRootfsPath, fmt.Sprintf("/sys/fs/cgroup/cpu/kubepods/besteffort/pod%s/%s", podUID, contID))
	case v1.PodQOSBurstable:
		cPath = path.Join(hostRootfsPath, fmt.Sprintf("/sys/fs/cgroup/cpu/kubepods/burstable/pod%s/%s", podUID, contID))
	}

	if _, err := os.Stat(cPath); err == nil {
		return cPath, nil
	} else {
		return "", err
	}

	return "", fmt.Errorf("invalid qos %v", qos)
}

func parseCgroupTasks(cgroupPath string) ([]int, error) {
	data, err := ioutil.ReadFile(path.Join(cgroupPath, "tasks"))
	if err != nil {
		return nil, err
	}

	pids := []int{}
	lines := strings.Split(string(data), "\n")
	for _, s := range lines {
		if len(s) == 0 {
			continue
		}
		pid, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("err parse pid %v: %v", s, err)
		}
		pids = append(pids, pid)
	}

	if len(pids) == 0 {
		log.Warningf("pid not found under %v, tasks content: %v", cgroupPath, string(data))
	}

	return pids, nil
}
