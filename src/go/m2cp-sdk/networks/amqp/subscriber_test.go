package amqp

import (
	"fmt"
	"m2cp"
	"m2cp/contextplus"
	"m2cp/tools"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/goleak"
)

// TestNewSignalSubscriber tests the creation and destruction of a signal subscriber - no leaks expected
func (t *TestSuite) TestNewSignalSubscriber() {
	var waitGroupInit, waitGroupShutdown sync.WaitGroup

	ctp := contextplus.NewContextPlus()
	topics := []string{"#"}
	callbackSignal := func(_ m2cp.SignalMessage, _ m2cp.Acknowledger) {}

	subs, err := NewSubscriberSignalsAck(m2cp.SubscriptionOptions{Context: ctp}, topics, callbackSignal, &waitGroupInit, &waitGroupShutdown)
	t.NoError(err)
	t.NotNil(subs)

	ctp.Cancel()
}

// TestNewDataSubscriber tests the creation and destruction of a data subscriber - no leaks expected
func (t *TestSuite) TestNewDataSubscriber() {
	DebugSender = true

	var waitGroupInit, waitGroupShutdown sync.WaitGroup
	ctp := contextplus.NewContextPlus()

	topics := []string{"#"}
	callbackData := func(_ m2cp.DataMessage, _ m2cp.Acknowledger) {}
	subs, err := NewSubscriberDataAck(m2cp.SubscriptionOptions{Context: ctp}, topics, callbackData, &waitGroupInit, &waitGroupShutdown)
	t.NoError(err)
	t.NotNil(subs)

	ctp.Cancel()
}

func (t *TestSuite) TestNewDataSubscriberPersistent() {
	DebugSender = true

	var waitGroupInit, waitGroupShutdown sync.WaitGroup
	ctp := contextplus.NewContextPlus()

	topics := []string{"#"}
	callbackData := func(_ m2cp.DataMessage, _ m2cp.Acknowledger) {}
	subs, err := NewSubscriberDataAck(m2cp.SubscriptionOptions{Context: ctp, PersistenceId: tools.StrPtr("test-persistent-data")}, topics, callbackData, &waitGroupInit, &waitGroupShutdown)
	t.NoError(err)
	t.NotNil(subs)

	ctp.Cancel()
}

func (t *TestSuite) TestNewCommandSubscriber() {
	// this test creates and shuts down a subscriber. It uses different delays between creation and shutdown
	// to increase the chance of finding a leak. Delay is raised exponentially.
	for i := 0; i < 32; i += i + 1 {
		var waitGroupInit, waitGroupShutdown sync.WaitGroup
		domain := newTestDomain("cmds")
		callbackCmd := func(msg m2cp.CommandMessage) {
			str := fmt.Sprintf("%s", msg.GetHeader())
			t.ctp.LogDebug(str)
		}
		subs, err := NewSubscriberCommands(m2cp.SubscriptionOptions{Context: t.ctp}, domain, callbackCmd, &waitGroupInit, &waitGroupShutdown)
		t.NoError(err)
		t.NotNil(subs)

		t.ctp.Sleep(time.Duration(i) * 100 * time.Millisecond)

		t.ctp.Cancel()
		goleak.VerifyNone(t.T())
		t.ctp = contextplus.NewContextPlus()
	}
}

func (t *TestSuite) TestNewResponseSubscriber() {
	// this test creates and shuts down a subscriber. It uses different delays between creation and shutdown
	// to increase the chance of finding a leak. Delay is raised exponentially.
	for i := 0; i < 32; i += i + 1 {
		var waitGroupInit, waitGroupShutdown sync.WaitGroup

		domainResponder := newTestDomain("rsp")
		callbackRes := func(msg m2cp.ResponseMessage) {
			str := fmt.Sprintf("%s", msg.GetHeader())
			t.ctp.LogDebug(str)
		}
		subs, err := NewSubscriberResponses(m2cp.SubscriptionOptions{Context: t.ctp}, domainResponder, callbackRes, &waitGroupInit, &waitGroupShutdown)
		t.NoError(err)
		t.NotNil(subs)

		t.ctp.Sleep(time.Duration(i) * 100 * time.Millisecond)

		t.ctp.Cancel()
		goleak.VerifyNone(t.T())
		t.ctp = contextplus.NewContextPlus()
	}
}

// Test_InitSubscribeShutdownRepeat repeatedly sets up a connection, subscribes something, shuts down and checks for leaks
// The type of leak we're looking for appears only sometimes, so we need to repeat this test
func (t *TestSuite) Test_InitSubscribeShutdownRepeat() {
	const iterations = 10
	for i := 0; i < iterations; i++ {
		fmt.Printf("iteration %d/%d\n", (i + 1), iterations)
		func() {
			var err error
			var waitGroupInit, waitGroupShutdown sync.WaitGroup

			domain := newTestDomain("")
			callbackCmd := func(msg m2cp.CommandMessage) {
				str := fmt.Sprintf("%s", msg.GetHeader())
				t.ctp.LogDebug(str)
			}
			subs, err := NewSubscriberCommands(m2cp.SubscriptionOptions{
				Context: t.ctp,
			}, domain, callbackCmd, &waitGroupInit, &waitGroupShutdown)
			t.NoError(err)
			t.NotNil(subs)

			subs.Stop()

			t.ctp.Cancel()

			goleak.VerifyNone(t.T())

			t.ctp = contextplus.NewContextPlus()

		}()
	}
}

type testDomain struct {
	snapName   string
	deviceName string
}

func (td *testDomain) GetAppName() string {
	panic("implement me")
}

func (td *testDomain) GetName() string {
	return td.snapName
}

func (td *testDomain) GetDeviceName() string {
	return td.deviceName
}

func newTestDomain(prefix string) m2cp.Domain {
	if prefix == "" {
		prefix = "test"
	}
	randomName := fmt.Sprintf("%s-%s", prefix, randomString(8))
	return &testDomain{
		snapName:   randomName,
		deviceName: fmt.Sprintf("dev-%s", randomString(8)),
	}
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
