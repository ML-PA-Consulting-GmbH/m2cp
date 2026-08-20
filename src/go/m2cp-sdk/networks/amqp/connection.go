package amqp

import (
	"errors"
	"fmt"
	"m2cp"
	"os"
	"strconv"
	"strings"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

var exchangeSetupLock sync.Mutex

var (
	amqpURL      string
	amqpHostLock sync.Mutex
)

type amqpClient struct {
	url               string
	port              int
	declaredExchanges bool
	ctp               m2cp.ContextPlus
}

func newAmqpConnection(ctx m2cp.ContextPlus) (*amqpClient, error) {
	ctx = ctx.Branch()
	ctx = ctx.SetModule("networks.amqp.connection")
	amqpPort := 5672

	if m2cpVirtualDevice, exists := os.LookupEnv("M2CP_VIRTUAL_DEVICE"); exists && m2cpVirtualDevice != "" {
		if virtualDeviceBasePort, err := strconv.Atoi(m2cpVirtualDevice); err != nil {
			return nil, fmt.Errorf("failed parsing M2CP_VIRTUAL_DEVICE: %w", err)
		} else {
			amqpPort = virtualDeviceBasePort + 1
			ctx.LogDebug("M2CP_VIRTUAL_DEVICE is set, using port %d for AMQP", amqpPort)
		}
	}

	c := &amqpClient{
		url:  amqpURL,
		port: amqpPort,
		ctp:  ctx,
	}

	return c, nil
}

// Dial wraps amqp.Dial with automatic IPv6/IPv4 fallback and singleton caching
func (c *amqpClient) Dial(ctp m2cp.ContextPlus) (*amqp.Connection, error) {
	amqpHostLock.Lock()
	if c.url == "" {

		// Try IPv6 first
		url := fmt.Sprintf("amqp://[::1]:%d/", c.port)
		conn, err := amqp.Dial(url)
		if err == nil {
			amqpURL = url
			c.url = url
			ctp.LogDebug("AMQP connection successful using IPv6 (::1)")
			amqpHostLock.Unlock()
			return conn, nil
		}

		// Fall back to IPv4
		url = fmt.Sprintf("amqp://127.0.0.1:%d/", c.port)
		conn, err = amqp.Dial(url)
		if err == nil {
			amqpURL = url
			c.url = url
			ctp.LogDebug("AMQP connection successful using IPv4 (127.0.0.1)")
			amqpHostLock.Unlock()
			return conn, nil
		}

		amqpHostLock.Unlock()
		return nil, fmt.Errorf("failed to connect to AMQP on both IPv6 and IPv4: %s", err.Error())
	}
	amqpHostLock.Unlock()

	return amqp.Dial(c.url)
}

func (c *amqpClient) getAmqpConnection() (*amqp.Connection, error) {
	if !c.declaredExchanges {
		exchangeSetupLock.Lock()
		defer exchangeSetupLock.Unlock()

		err := c.declareExchanges()
		if err != nil {
			return nil, err
		}

		c.ctp.LogDebug("exchanges declared")
		c.declaredExchanges = true
	}
	conn, err := c.Dial(c.ctp)
	if err != nil {
		msg := fmt.Sprintf("failed to connect to AMQP: %s", err.Error())
		//c.ctp.LogError(msg)
		return nil, errors.New(msg)
	}
	return conn, nil
}

func (c *amqpClient) declareExchanges() error {
	exchanges := map[string]string{
		ExchangeNameData:    ExchangeTypeData,
		ExchangeNameSignals: ExchangeTypeSignals,
		ExchangeNameLog:     ExchangeTypeLog,
		//ExchangeNameTest:    ExchangeTypeTest,
		// NOTE: There's no RPC exchange. Each node has its own RPC queue with direct delivery
	}

	var ch *amqp.Channel
	conn, err := c.Dial(c.ctp)
	if err != nil {
		return fmt.Errorf("failed to connect to AMQP: %w", err)
	}
	defer func() {
		if ch != nil && !ch.IsClosed() {
			if err := ch.Close(); err != nil {
				c.ctp.LogError(fmt.Sprintf("error closing channel: %s", err.Error()))
			}
		}
		if conn != nil && !conn.IsClosed() {
			if err := conn.Close(); err != nil {
				c.ctp.LogError(fmt.Sprintf("error closing connection: %s", err.Error()))
			}
		}
	}()

	prepareChannel := func() error {
		if ch == nil || ch.IsClosed() {
			if ch, err = conn.Channel(); err != nil {
				return fmt.Errorf("failed to open a channel: %w", err)
			}
		}
		return nil
	}

	declareExchange := func(exchangeName, exchangeType string) error {
		retryCount := 0
		maxTries := 2
		for retryCount < maxTries {

			// let's check, if the exchange already exists
			if err = prepareChannel(); err != nil {
				return err
			}
			if ch.IsClosed() {
				return fmt.Errorf("channel is closed")
			}

			// now try to declare the missing exchange
			const durable = true
			if err = ch.ExchangeDeclarePassive(exchangeName, exchangeType, durable, false, false, false, nil); err == nil {
				c.ctp.LogDebug("exchange '%s' (%s, durable=%v): found, valid.", exchangeName, exchangeType, durable)
				return nil
			}
			c.ctp.LogDebug("exchange '%s' (%s): missing, creating.", exchangeName, exchangeType)

			// after a failed passive declaration the channel might be closed - reopen it
			if ch.IsClosed() {
				if err = prepareChannel(); err != nil {
					return err
				}
				if ch.IsClosed() {
					return fmt.Errorf("channel is closed")
				}
			}

			// now try to declare the missing exchange
			if err = ch.ExchangeDeclare(exchangeName, exchangeType, durable, false, false, false, nil); err == nil {
				c.ctp.LogDebug("exchange '%s' (%s, durable=%v): created, valid.", exchangeName, exchangeType, durable)
				return nil
			}
			//if err = ch.ExchangeDeclare(exchangeName, exchangeType, false, false, false, false, nil); err == nil {
			//	c.ctp.LogDebug("exchange '%s' (%s): missing, created.", exchangeName, exchangeType)
			//	return nil
			//}
			msg := fmt.Sprintf("exchange '%s' (%s): missing, failed creating - will attempt deletion and recreation of channel. error: %s", exchangeName, exchangeType, err.Error())
			c.ctp.LogWarn(msg)

			// check if it's a known error pattern
			isFixable := strings.Contains(err.Error(), `PRECONDITION_FAILED - inequivalent arg 'type' for exchange`)
			// we want to upgrade to durable - not downgrade to non-durable
			isFixable = isFixable || (strings.Contains(err.Error(), `PRECONDITION_FAILED - inequivalent arg 'durable' for exchange`) && strings.Contains(err.Error(), `received 'true' but current is 'false'`))
			if !isFixable {
				c.ctp.LogWarn("exchange declaration error does NOT match expected patterns. will NOT attempt to delete and recreate exchange. error: %s", err.Error())
				return err
			}

			// try to delete misconfigured exchange - will be recreated on next loop
			if err = prepareChannel(); err != nil {
				return err
			}
			if err = ch.ExchangeDelete(exchangeName, false, false); err != nil {
				c.ctp.LogError("exchange: '%s': failed to delete exchange. error: %s", exchangeName, err.Error())
			}
			c.ctp.LogDebug("exchange '%s' deleted.", exchangeName)

			retryCount++
		}
		msg := fmt.Sprintf("failed declaring exchange '%s' as type '%s'", exchangeName, exchangeType)
		c.ctp.LogError(msg)
		return errors.New(msg)
	}

	for exchangeName, exchangeType := range exchanges {
		if err = declareExchange(exchangeName, exchangeType); err != nil {
			return err
		}
	}
	return nil
}
