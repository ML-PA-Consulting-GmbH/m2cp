package amqp

import (
	"fmt"
	"m2cp"
	"m2cp/contextplus"
	"m2cp/messages"
	"m2cp/networks/stats"
	"sync"
)

func (t *TestSuite) TestNewSender() {
	ctp := contextplus.NewContextPlus()

	var err error
	var waitGroupInit, waitGroupShutdown sync.WaitGroup

	s, err := NewSender(ctp, &waitGroupInit, &waitGroupShutdown, 100, m2cp.MessageSerializerDefault, &stats.NetworkStats{})
	t.NoError(err)
	t.NotNil(s)

	ctp.Cancel()
}

func (t *TestSuite) TestSend() {
	var err error
	var waitGroupInit, waitGroupShutdown sync.WaitGroup

	ctp := contextplus.NewContextPlus()
	sndr, err := NewSender(ctp, &waitGroupInit, &waitGroupShutdown, 100, m2cp.MessageSerializerDefault, &stats.NetworkStats{})
	t.NoError(err)
	t.NotNil(sndr)

	senderAddress, err := messages.NewAddressWithSubtopic(ctp, "testnode.testapp.local", "testnode/")
	t.NoError(err)
	msg, err := messages.NewSignalMessage(senderAddress, "foo", "bar", m2cp.SignalLevelWarning, m2cp.MessageScopeDevice)
	t.NoError(err)
	t.NotNil(msg)

	err = sndr.Send(msg)
	t.NoError(err)

	ctp.Cancel()
}

// Test_InitSendShutdownRepeat repeatedly sets up a connection, sends a message, shuts down and checks for leaks
// The type of leak we're looking for appears only sometimes, so we need to repeat this test
func (t *TestSuite) TestInitSendShutdownRepeat() {
	const iterations = 10
	for i := 0; i < iterations; i++ {
		fmt.Printf("iteration %d/%d\n", (i + 1), iterations)
		func() {
			var err error
			var waitGroupInit, waitGroupShutdown sync.WaitGroup

			senderInstance, err := NewSender(t.ctp, &waitGroupInit, &waitGroupShutdown, 100, m2cp.MessageSerializerDefault, &stats.NetworkStats{})
			t.NoError(err)
			t.NotNil(senderInstance)

			senderAddress, err := messages.NewAddressWithSubtopic(t.ctp, "testnode.testapp.local", "testnode/")
			t.NoError(err)
			msg, err := messages.NewSignalMessage(senderAddress, "foo", "bar", m2cp.SignalLevelWarning, m2cp.MessageScopeDevice)
			t.NoError(err)
			t.NotNil(msg)

			err = senderInstance.Send(msg)
			t.NoError(err)

			t.ResetContextAndCheckLeaks()

		}()
	}
}
