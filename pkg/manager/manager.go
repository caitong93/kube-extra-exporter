package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/caicloud/nirvana/log"
	"github.com/caitong93/kube-extra-exporter/pkg/info"
	"github.com/caitong93/kube-extra-exporter/pkg/network"
	"github.com/caitong93/kube-extra-exporter/pkg/pod"

	v1 "k8s.io/api/core/v1"
)

var (
	hostRootfsPath = "/rootfs"
)

type Manager struct {
	podLister            pod.Lister
	networkStatsProvider network.StatsProvider

	containersLock sync.Mutex
	pods           map[string]*podData
}

func New(podLister pod.Lister) (*Manager, error) {
	return &Manager{
		pods:                 make(map[string]*podData),
		podLister:            podLister,
		networkStatsProvider: network.NewStatsProvider(),
	}, nil
}

func (m *Manager) findPodByUID(UID string) (*v1.Pod, error) {
	pods, err := m.podLister.List()
	if err != nil {
		return nil, err
	}
	for _, po := range pods {
		if string(po.UID) == UID {
			return po, nil
		}
	}
	return nil, fmt.Errorf("pod not found")
}

func (m *Manager) Run(ctx context.Context) error {
	renewPods := func() error {
		pods, err := m.podLister.List()
		if err != nil {
			log.Error("Err list pods:", err)
		}
		newPods := make(map[string]*podData)
		for _, po := range pods {
			if po.Status.Phase != v1.PodRunning {
				continue
			}

			UID := string(po.UID)
			data := newPodData(po)
			for _, cont := range po.Status.ContainerStatuses {
				if err := data.addContainer(cont.ContainerID); err != nil {
					return err
				}
			}

			newPods[UID] = data
		}

		// log.Infof("refresh pods %v", pretty.Sprint(newPods))
		m.containersLock.Lock()
		m.pods = newPods
		m.containersLock.Unlock()
		return nil
	}

	if err := renewPods(); err != nil {
		log.Errorf("Err refresh pod infos: %v", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				if err := renewPods(); err != nil {
					log.Errorf("Err refresh pod infos: %v", err)
				}
			}
		}
	}()

	return nil
}

func (m *Manager) ListStats() ([]*info.Stats, error) {
	m.containersLock.Lock()
	defer m.containersLock.Unlock()

	infos := []*info.Stats{}
	for _, pod := range m.pods {
		stat := &info.Stats{
			PodName:   pod.Name,
			Namespace: pod.Namespace,
		}

		// Fill network stats
		netStat, err := m.networkStatsProvider.GetStats(hostRootfsPath, pod.onePid())
		if err != nil {
			log.Errorf("err get network stats for pod %v: %v", pod.Name, err)
			continue
			// return nil, err
		}
		stat.Network = netStat

		infos = append(infos, stat)
	}

	return infos, nil
}
