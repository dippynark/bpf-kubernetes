package bpf

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	//"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/iovisor/gobpf/elf"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

const (
	cgroupSkbMetricsAssetName   = "cgroup_skb_metrics.o"
	cgroupSkbProgramSectionName = "cgroup/skb"
	cgroupV2Root                = "/sys/fs/cgroup/unified"
	countMapSectionName         = "maps/count"
	mapSectionPrefix            = "maps/"
	defaultFilePerm             = 0700
	pingHost                    = "www.google.com"
	curlAddress                 = "http://www.google.com/"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
	module   *elf.Module
}

func NewController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller) (*Controller, error) {

	// load bpf program
	module, err := Load(cgroupSkbMetricsAssetName, map[string]elf.SectionParams{})
	if err != nil {
		return nil, err
	}

	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
		module:   module,
	}, nil
}

func (c *Controller) processNextItem() bool {

	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.sync(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

func (c *Controller) sync(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	// Note that you also have to check the uid if you have a local controlled resource, which
	// is dependent on the actual instance, to detect that a Pod was recreated with the same name
	pod := obj.(*v1.Pod)

	// only handle pods on local node
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		return errors.New("NODE_NAME not set")
	}
	if pod.Spec.NodeName != nodeName {
		return nil
	}

	// build cgroup path
	podCgroupPathPrefix := "kubepods.slice/kubepods-"
	if pod.Status.QOSClass != v1.PodQOSGuaranteed {
		qosClass := strings.ToLower(string(pod.Status.QOSClass))
		podCgroupPathPrefix = filepath.Join(podCgroupPathPrefix+qosClass+".slice", "kubepods-"+qosClass+"-")
	}
	podCgroupPath := filepath.Join(cgroupV2Root, podCgroupPathPrefix+"pod"+strings.ReplaceAll(string(pod.ObjectMeta.UID), "-", "_")+".slice")

	// retrieve bpf program
	cgroupProgram := c.module.CgroupProgram(cgroupSkbProgramSectionName)
	if cgroupProgram == nil {
		return fmt.Errorf("failed to retrieve cgroup program %s", cgroupSkbProgramSectionName)
	}

	for i, container := range pod.Status.ContainerStatuses {

		containerCgroupPath := filepath.Join(podCgroupPath, strings.Replace(container.ContainerID, "://", "-", 1)+".scope")

		// attach bpf program to container cgroup
		err = elf.AttachCgroupProgram(cgroupProgram, containerCgroupPath, elf.EgressType)
		if err != nil {
			return err
		}
		klog.Infof("attached bpf program to pod %s container %s cgroup", pod.GetName(), pod.Status.ContainerStatuses[i].Name)
	}

	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Controller) handleErr(err error, key interface{}) {

	if err == nil {
		c.queue.Forget(key)
		return
	}

	c.queue.AddRateLimited(key)
	runtime.HandleError(err)
}

func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	klog.Info("starting controller")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	klog.Info("stopping pod controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func CgroupSkbMetrics() error {

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		if home := os.Getenv("HOME"); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// create the pod watcher
	podListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", v1.NamespaceDefault, fields.Everything())

	// create the workqueue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	indexer, informer := cache.NewIndexerInformer(podListWatcher, &v1.Pod{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	}, cache.Indexers{})

	controller, err := NewController(queue, indexer, informer)
	if err != nil {
		return err
	}

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	// watch count map
	countMapName := strings.TrimPrefix(countMapSectionName, mapSectionPrefix)
	countMap := controller.module.Map(countMapName)
	for {
		key := 0
		var value uint64
		err = controller.module.LookupElement(countMap, unsafe.Pointer(&key), unsafe.Pointer(&value))
		if err != nil {
			return err
		}
		klog.Infof("packet count: %d", value)
		time.Sleep(time.Second * 5)
	}

}
