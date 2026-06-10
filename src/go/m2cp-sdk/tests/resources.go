package tests

import (
	"fmt"
	"sync"
	"time"
)

var (
	resources     map[string]*sync.Mutex
	resourcesLock sync.Mutex
)

func LockResource(name string) {
	fmt.Println("Locking resource", name)

	if resources == nil {
		resourcesLock.Lock()
		resources = make(map[string]*sync.Mutex)
		resourcesLock.Unlock()
	}

	resourcesLock.Lock()
	if _, ok := resources[name]; !ok {
		resources[name] = &sync.Mutex{}
	}
	r := resources[name]
	resourcesLock.Unlock()

	r.Lock()
	fmt.Println("Resource", name, "locked")
}

func FreeResource(name string) {
	fmt.Println("Freeing resource", name)

	resourcesLock.Lock()
	defer resourcesLock.Unlock()

	if r, ok := resources[name]; ok {
		r.Unlock()
	}
}

func GetCoapUDPTestPort() (string, func()) {

	LockResource("test-port")
	time.Sleep(500 * time.Millisecond) // to avoid port conflicts in tests

	return "5688", func() {
		FreeResource("test-port")
	}

}

func GetCoapDTLSTestPort() (string, func()) {
	LockResource("test-port")
	time.Sleep(500 * time.Millisecond) // to avoid port conflicts in tests

	return "5684", func() {
		FreeResource("test-port")
	}
}

func GetCoapUDPAndDTLSTestPort() (string, string, func()) {

	LockResource("test-port")
	time.Sleep(500 * time.Millisecond) // to avoid port conflicts in tests

	return "5683", "5684", func() {
		FreeResource("test-port")
	}
}
