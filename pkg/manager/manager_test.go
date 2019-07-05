package manager

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockPodLister struct {
	pods []*v1.Pod
}

func (p *mockPodLister) List() ([]*v1.Pod, error) {
	return p.pods, nil
}

type testData struct {
	pods   []*v1.Pod
	pids   map[string][]int
	expect map[string]*podData
}

func TestManager(t *testing.T) {
	saved := hostRootfsPath
	defer func() {
		hostRootfsPath = saved
	}()

	cases := []testData{
		{
			pods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
						UID:  types.UID("1952d77-996a-11e9-81b0-0242ac110002"),
					},
					Status: v1.PodStatus{
						QOSClass: v1.PodQOSGuaranteed,
						Phase:    v1.PodRunning,
						ContainerStatuses: []v1.ContainerStatus{
							{
								ContainerID: "2726ab85f748125d79e5d64544a632f29f864b79e5905ac8f3398c42ba6a9b3e",
							},
							{
								ContainerID: "b78989809beb9310b39404391b36981da0dc31ed6e4ea9fb8c7cbbcccaddb691",
							},
						},
					},
				},
			},
			pids: map[string][]int{
				"/sys/fs/cgroup/cpu/kubepods/pod1952d77-996a-11e9-81b0-0242ac110002/2726ab85f748125d79e5d64544a632f29f864b79e5905ac8f3398c42ba6a9b3e": []int{1},
				"/sys/fs/cgroup/cpu/kubepods/pod1952d77-996a-11e9-81b0-0242ac110002/b78989809beb9310b39404391b36981da0dc31ed6e4ea9fb8c7cbbcccaddb691": []int{2, 3},
			},
			expect: map[string]*podData{
				"1952d77-996a-11e9-81b0-0242ac110002": &podData{
					Name: "foo",
					UID:  "1952d77-996a-11e9-81b0-0242ac110002",
					Containers: []*containerData{
						{
							ID:   "2726ab85f748125d79e5d64544a632f29f864b79e5905ac8f3398c42ba6a9b3e",
							Pids: []int{1},
						},
						{
							ID:   "b78989809beb9310b39404391b36981da0dc31ed6e4ea9fb8c7cbbcccaddb691",
							Pids: []int{2, 3},
						},
					},
				},
			},
		},
	}

	for _, cas := range cases {
		mgr, err := New(&mockPodLister{cas.pods})
		if err != nil {
			t.Error(err)
		}

		func() {
			// Set up fake rootfs
			tmpDir, err := ioutil.TempDir("", "kube-extra-exporter")
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("Temp dir %v", tmpDir)
			defer os.RemoveAll(tmpDir)

			hostRootfsPath = tmpDir
			for subPath, pids := range cas.pids {
				fullPath := path.Join(hostRootfsPath, subPath)
				if err := os.MkdirAll(fullPath, 0777); err != nil {
					t.Fatal(err)
				}
				var buf bytes.Buffer
				for _, pid := range pids {
					buf.WriteString(fmt.Sprintf("%v\n", pid))
				}
				ioutil.WriteFile(path.Join(fullPath, "tasks"), buf.Bytes(), 0777)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if err := mgr.Run(ctx); err != nil {
				t.Error(err)
			}

			time.Sleep(1 * time.Second)
			if len(cas.expect) != len(mgr.pods) {
				t.Errorf("Result length not equal, expect\n%#v,\ngot\n%#v\n", cas.expect, mgr.pods)
			}
			for _, po := range cas.expect {
				p, ok := mgr.pods[po.UID]
				if !ok {
					t.Error("Not found")
					continue
				}
				if po.Name != p.Name {
					t.Error("Pod name")
				}
				if !reflect.DeepEqual(po.Containers, p.Containers) {
					t.Errorf("Containers, expect\n%#v,\ngot\n%#v\n", po.Containers, p.Containers)
				}
			}
		}()
	}
}

func TestParseContainerID(t *testing.T) {
	ID := "docker://999a54e3e9eb3c1bf58c96788850aa03a47d3e3c009da9ecae8d2edfdba5a328"
	result := parseContainerID(ID)
	expect := "999a54e3e9eb3c1bf58c96788850aa03a47d3e3c009da9ecae8d2edfdba5a328"
	if result != expect {
		t.Errorf("expect %s, got %s", expect, result)
	}
}
