package amqp

import (
	"fmt"
	"m2cp"
	"m2cp/networks/stats"
	"m2cp/tools"
	"strings"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	DebugSender = false // set to true to enable super-verbose logging of sending
)

const (
	maxAllowedQueueSize = 1000
	maxSendAckAmount    = 20
	maxSendAckAge       = time.Second * 1
)

type sender struct {
	amqpClient           *amqpClient
	ctp                  m2cp.ContextPlus
	sendLock             sync.Mutex
	sendQueue            sendJobQueue
	sendAcks             sendJobQueue
	shuttingDown         bool
	mutex                sync.Mutex
	connectedAtLeastOnce bool
	waitGroupInit        *sync.WaitGroup
	waitGroupShutdown    *sync.WaitGroup
	closed               bool
	maxQueueSize         int
	messageSerializer    m2cp.MessageSerializer
	stats                *stats.NetworkStats
}

func NewSender(
	ctpParent m2cp.ContextPlus,
	waitGroupInit *sync.WaitGroup,
	waitGroupShutdown *sync.WaitGroup,
	sendQueueSize int,
	messageSerializer m2cp.MessageSerializer,
	stats *stats.NetworkStats,
) (m2cp.Sender, error) {
	if sendQueueSize <= 0 {
		sendQueueSize = maxAllowedQueueSize
	}
	if sendQueueSize < 2 {
		return nil, fmt.Errorf("sendQueueSize must be at least 2")
	}
	stats.Sender.QueueSend.Max.Set(uint64(sendQueueSize))
	stats.Sender.QueueAck.Max.Set(uint64(sendQueueSize))

	ctp := ctpParent.BranchWithName("networks.amqp.Sender")

	s := &sender{
		amqpClient:        nil,
		ctp:               ctp,
		sendQueue:         newSendJobQueue(&stats.Sender.QueueSend),
		sendAcks:          newSendJobQueue(&stats.Sender.QueueAck),
		shuttingDown:      false,
		mutex:             sync.Mutex{},
		waitGroupInit:     waitGroupInit,
		waitGroupShutdown: waitGroupShutdown,
		maxQueueSize:      sendQueueSize,
		messageSerializer: messageSerializer,
		stats:             stats,
	}

	s.waitGroupInit.Add(1)

	go s.workerThread()

	return s, nil
}

func (s *sender) Send(m m2cp.Message) error {
	s.sendLock.Lock()
	defer s.sendLock.Unlock()

	// one message can generate more than one send job, because copies might be sent to log or legacy topics
	sendJobs, err := newSendJobs(m, s.messageSerializer)
	if err != nil {
		return err
	}
	if len(sendJobs) == 0 {
		return fmt.Errorf("no send jobs")
	}
	// we need enough slots in the queue
	// fmt.Printf("send queue: %d/%d, ack queue: %d/%d\n", s.sendQueue.Len(), s.maxQueueSize, s.sendAcks.Len(), maxSendAckAmount)

	if (s.sendQueue.Len() + len(sendJobs)) > s.maxQueueSize {
		return fmt.Errorf("queue full")
	}
	for _, sj := range sendJobs {
		s.queueSendJob(sj)
	}
	return nil
}

func (s *sender) queueSendJob(sendJob *SendJob) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.shuttingDown {
		return
	}
	//s.ctp.LogDebug("queueing message: %v", sendJob)
	s.sendQueue.Enqueue(sendJob)
	if DebugSender {
		s.ctp.LogDebug("queued origin=%s topic=%s", sendJob.Origin, sendJob.Topic)
	}
}

func (s *sender) SendByeAndCloseQueue(byeMessages []SendJob) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.shuttingDown = true
	s.sendQueue.Clear()
	for i := range byeMessages {
		s.sendQueue.Enqueue(&byeMessages[i])
	}
}

func (s *sender) Stop() {
	s.shuttingDown = true
	s.ctp.Cancel()
	for !s.closed {
		s.ctp.Sleep(10 * time.Millisecond)
	}
}

func (s *sender) workerThread() {
	var err error
	s.waitGroupShutdown.Add(1)
	defer func() {
		s.doneInit()
		s.closed = true
		s.waitGroupShutdown.Done()
	}()

	// init amqp connection and setup exchanges
	for {
		if s.amqpClient, err = newAmqpConnection(s.ctp); err == nil {
			break // initialization successful
		}
		s.ctp.Sleep(1 * time.Second)
		if s.ctp.IsCancelled() {
			return
		}
	}

	// main loop
	for {
		err = s.work()
		if err != nil {
			s.ctp.LogError(err.Error())
		}
		if s.ctp.IsCancelled() {
			s.stats.Sender.Connected = false
			return
		}
		s.ctp.Sleep(1 * time.Second)
	}
}

var expirationDataAndSignal = fmt.Sprintf("%d", (48 * time.Hour).Milliseconds())
var expirationCommandAndResponse = fmt.Sprintf("%d", (60 * time.Second).Milliseconds())

func (s *sender) doneInit() {
	if !s.connectedAtLeastOnce {
		s.waitGroupInit.Done()
		s.connectedAtLeastOnce = true
	}
}

func (s *sender) work() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	amqpConnection, err := s.amqpClient.getAmqpConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to AMQP: %s", err.Error())
	}

	amqpChannel, err := amqpConnection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %s", err.Error())
	}

	// Put the channel in confirm mode
	err = amqpChannel.Confirm(false)
	if err != nil {
		return fmt.Errorf("failed to put channel in confirm mode: %v", err)
	}

	// Listen to success/failure events
	failedChan := amqpChannel.NotifyReturn(make(chan amqp.Return))
	confirmChan := amqpChannel.NotifyPublish(make(chan amqp.Confirmation))

	s.stats.Sender.Connected = true

	// Closing procedure (this was a tricky one to get right, any deviation from this can cause non-zero probability of deadlock):
	// 1. Set termination flag to true to avoid double execution of closing procedure
	// 2. Start thread for closing amqp (important: this needs to be done concurrently as we need to keep listening to the response channels!)
	// 3. Wait for response channels to close, which will be indicated by them being set to nil
	// 4. Now we can break out of the loop and return
	closingInitiated := false
	closeConnection := func() {
		if !closingInitiated {
			closingInitiated = true
			go func() {
				s.ctp.LogDebug("trying to close amqp connection")
				s.stats.Sender.Connected = false
				if err := amqpConnection.Close(); err != nil {
					s.ctp.LogError(fmt.Sprintf("error closing channel: %s", err.Error()))
				}
				s.ctp.LogDebug("closed amqp connection")
			}()
		}
	}

	errorCount := uint64(0)
	registerError := func(err error) {
		s.stats.Sender.Errors.Inc()
		errorCount++
		s.ctp.LogError("error: %s", err.Error())
		if errorCount > 10 {
			s.ctp.LogError("too many errors, closing connection")
			closeConnection()
		}
	}

	publishCount := uint64(0)
	var job *SendJob
	var sleepMultiplier int64 = 1
	const sleepMultiplierMax = 10

	s.doneInit()

	for {

		select {
		case confirm, ok := <-confirmChan:
			if ok {
				if err = s.sendAcks.RemoveJobWithDeliveryTag(confirm.DeliveryTag); err != nil {
					s.ctp.LogWarn("no send job with delivery tag %d: %s", confirm.DeliveryTag, err.Error())
				} else {
					s.stats.Sender.Acks.Inc()
				}
			} else {
				// this happens, if the channel is closed
				confirmChan = nil
			}
		case failed, ok := <-failedChan:
			if ok {
				registerError(fmt.Errorf("failed to publish a message: %s (ReplyCode %d)", failed.ReplyText, failed.ReplyCode))
			} else {
				// this happens, if the channel is closed
				failedChan = nil
			}
			// we used to close the connection here - but that's bad, we first want to send all messages - we now only close the connection once all messages are sent!
			/*		case <-s.ctp.Done():
					// context will send infinite cancellation events - so this case will be called repeatedly.
					if !closingInitiated {
						s.ctp.LogDebug("context cancelled, closing connection")
						closeConnection()
					}
			*/
		default:
			// if acks take too long, then something is wrong
			if job = s.sendAcks.Peek(); job != nil {
				if tools.TimeSystemRunning()-job.DeliveryAttempt > maxSendAckAge {
					s.ctp.LogError("message not acked for %d seconds, considering %d unacked messages as lost, resetting amqp channel", int(maxSendAckAge.Seconds()), s.sendAcks.Len())
					// we don't know if the messages were sent or not, so we have to requeue them
					for {
						job = s.sendAcks.DequeueEnd()
						if job == nil {
							break
						}
						s.sendQueue.EnqueueFront(job)
					}
					// let's close the channel and crash the sender - a new one will be started
					_ = amqpChannel.Close()
				}
			}

			// don't send too many messages at once - wait for acks before sending more
			if s.sendAcks.Len() > maxSendAckAmount {
				s.ctp.Sleep(10 * time.Millisecond)
				continue
			}

			job = s.sendQueue.Peek()
			if job == nil {
				// nothing to do?
				// we only close the connection once all messages are sent!
				if s.ctp.IsCancelled() && !closingInitiated {
					s.ctp.LogDebug("context cancelled, closing connection")
					closeConnection()
				}
				s.ctp.Sleep(time.Duration(sleepMultiplier) * 10 * time.Millisecond)
				sleepMultiplier = min(sleepMultiplier+1, sleepMultiplierMax)

			} else {
				// we have a message to send

				sleepMultiplier = 1

				//s.ctp.LogDebug("send: %s@%s", job.RoutingKey, job.Exchange)
				publishCount++
				publishing := amqp.Publishing{
					ContentType:   "application/octet-stream",
					Body:          job.Message,
					CorrelationId: job.CorrelationId,
				}
				if job.MessageType == m2cp.MessageTypeData || job.MessageType == m2cp.MessageTypeSignal {
					publishing.DeliveryMode = amqp.Persistent
					publishing.Expiration = expirationDataAndSignal
				}
				if job.MessageType == m2cp.MessageTypeCommand {
					publishing.Expiration = expirationCommandAndResponse
					tokens := strings.Split(job.Origin, ".")
					if len(tokens) > 2 {
						publishing.ReplyTo = fmt.Sprintf("response.%s.%s", tokens[len(tokens)-2], tokens[len(tokens)-1])
					}
				}
				var deliveryTag *amqp.DeferredConfirmation
				if deliveryTag, err = amqpChannel.PublishWithDeferredConfirm(job.Exchange, job.RoutingKey, false, false, publishing); err != nil {
					registerError(err)
					s.ctp.Sleep(100 * time.Millisecond)
				} else {
					s.stats.Sender.Transmitted.Inc()
					job.DeliveryTag = deliveryTag.DeliveryTag
					job.DeliveryAttempt = tools.TimeSystemRunning()
					if DebugSender {
						s.ctp.LogDebug("sent origin=%s topic=%s", job.Origin, job.Topic)
					}
					_ = s.sendQueue.Dequeue()
					s.sendAcks.Enqueue(job)
				}
			}
		}

		// The closing procedure leads to confirmChan and failedChan being set to nil, which will cause the loop to break
		if confirmChan == nil && failedChan == nil {
			s.stats.Sender.Connected = false
			s.ctp.LogDebug("connection closed, returning")
			return nil
		}
	}

}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
