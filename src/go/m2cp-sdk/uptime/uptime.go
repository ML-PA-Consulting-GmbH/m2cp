package uptime

import (
	"fmt"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"m2cp"
	coap_client "m2cp/coap/coap-client"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"
)

var appName string

var (
	lastConfirmationTime      *time.Time
	lastConfirmationTimeMutex = sync.Mutex{}

	currentRuntimeRequested     = make(map[string]time.Time)
	currentRuntimeRequestedLock = sync.Mutex{}

	agentRunning      bool
	agentRunningMutex = sync.Mutex{}
)

func resetRequestedRuntimes() {
	currentRuntimeRequestedLock.Lock()
	defer currentRuntimeRequestedLock.Unlock()

	currentRuntimeRequested = make(map[string]time.Time)

	lastConfirmationTimeMutex.Lock()
	defer lastConfirmationTimeMutex.Unlock()
	lastConfirmationTime = nil
}

// RequestAppRuntime requests a runtime for the application. The "identifier" allows a single app to request multiple runtimes.
// This can be useful, when submodules of an application need to request different runtimes.
// You can leave the field empty, if you don't need it.
func RequestAppRuntime(ctx m2cp.ContextPlus, identifier string, seconds int) {
	arch, _ := os.LookupEnv("SNAP_ARCH")
	if arch == "amd64" {
		// no need to request runtime on amd64
		return
	}

	if seconds < 0 {
		seconds = 0
	}

	t := time.Now().Add(time.Second * time.Duration(seconds))
	setRuntimeRequested(identifier, t)

	// reset last confirmation time to trigger a new request
	lastConfirmationTimeMutex.Lock()
	lastConfirmationTime = nil
	lastConfirmationTimeMutex.Unlock()

	agentRunningMutex.Lock()
	defer agentRunningMutex.Unlock()

	if appName == "" {
		var isSet bool
		if appName, isSet = os.LookupEnv("SNAP_NAME"); !isSet {
			appName = fmt.Sprintf("anonymous_app_%d", rand.Int())
		}
	}

	if !agentRunning {
		agentRunning = true
		go agentLoop(ctx)
	}

	// requests for 0 seconds are very important! they are often send at service shutdown...
	// they _must_ be delivered... at least we have to try hard!
	if seconds == 0 {
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if v, ok := currentRuntimeRequested[identifier]; !ok {
				// the request is gone
				break
			} else {
				// a request for our identifier is still in the outbox
				if v != t {
					// a new request has overwritten our request - we're done
					break
				} else {
					// the request is still in the outbox - let's wait a bit
					time.Sleep(100 * time.Millisecond)
					continue
				}
			}
		}
		// Exit after 5 seconds of trying
	}

}

func setRuntimeRequested(identifier string, t time.Time) {
	currentRuntimeRequestedLock.Lock()
	defer currentRuntimeRequestedLock.Unlock()
	currentRuntimeRequested[identifier] = t
}

func getRuntimeRequest(identifier string) (time.Time, bool) {
	currentRuntimeRequestedLock.Lock()
	defer currentRuntimeRequestedLock.Unlock()
	t, ok := currentRuntimeRequested[identifier]
	return t, ok
}

func getRuntimeRequestIdentifiers() []string {
	currentRuntimeRequestedLock.Lock()
	defer currentRuntimeRequestedLock.Unlock()
	identifiers := make([]string, 0, len(currentRuntimeRequested))
	for k := range currentRuntimeRequested {
		identifiers = append(identifiers, k)
	}
	return identifiers
}

func safeDeleteRuntimeRequest(identifier string, t time.Time) {
	currentRuntimeRequestedLock.Lock()
	defer currentRuntimeRequestedLock.Unlock()
	if v, ok := currentRuntimeRequested[identifier]; ok && v == t {
		delete(currentRuntimeRequested, identifier)
	}
}

// GetLastConfirmationTime returns the time of the last contact with the uptime service
func GetLastConfirmationTime() *time.Time {
	lastConfirmationTimeMutex.Lock()
	defer lastConfirmationTimeMutex.Unlock()
	return lastConfirmationTime
}

var uptimeManagerIp = "[::1]:5683"

func SetUptimeManagerIp(ip string) {
	uptimeManagerIp = ip
}

func agentLoop(ctp m2cp.ContextPlus) {
	defer func() {
		agentRunningMutex.Lock()
		agentRunning = false
		agentRunningMutex.Unlock()
	}()

	for !ctp.IsCancelled() {
		for _, identifier := range getRuntimeRequestIdentifiers() {
			err := doUptimeRequest(ctp, identifier)
			if err != nil {
				ctp.Sleep(1 * time.Second)
			}
		}

		ctp.Sleep(1 * time.Second)
	}
}

func doUptimeRequest(ctp m2cp.ContextPlus, identifier string) error {

	runtimeRequest, ok := getRuntimeRequest(identifier)
	if !ok {
		// already gone
		return nil
	}

	// do an uptime request
	seconds := int(math.Ceil(runtimeRequest.Sub(time.Now()).Seconds()))
	if seconds < 0 {
		seconds = 0
	}
	//ctp.LogDebug("Requesting %d seconds uptime", seconds)

	client, err := coap_client.NewClient(ctp, uptimeManagerIp)
	if err != nil {
		return fmt.Errorf("failed to create CoAP client for uptime request: %s", err)
	}
	res, err := client.Put("/request_uptime/1", fmt.Sprintf(
		"app=%s&seconds=%d", fmt.Sprintf("%s_%s", appName, identifier), seconds), nil)

	if ctp.IsCancelled() {
		return fmt.Errorf("coap request for uptime failed: context cancelled")
	}

	if err != nil {
		return fmt.Errorf("failed to request uptime - CoAP call failed: %s", err)
	} else if res.GetResponseCode() != codes.Changed {
		return fmt.Errorf("failed to request uptime - unexpected CoAP response code: %s (%d)", res.GetResponseCode(), res.GetResponseCode())
	}

	// mark the time of the last confirmation to avoid to frequent requests
	t := time.Now()
	lastConfirmationTimeMutex.Lock()
	lastConfirmationTime = &t
	lastConfirmationTimeMutex.Unlock()

	safeDeleteRuntimeRequest(identifier, runtimeRequest)

	return nil
}
