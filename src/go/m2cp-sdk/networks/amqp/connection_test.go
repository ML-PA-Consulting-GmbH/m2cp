package amqp

import (
	"m2cp/contextplus"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (t *TestSuite) TestAmqpMisconfigurationAndHealing() {
	var err error

	// create connection - this will create the exchanges
	ctp := contextplus.NewContextPlus()
	c, err := newAmqpConnection(ctp)
	t.NoError(err)
	t.NotNil(c)

	// make an extra connection for messing things up
	conn, err := c.Dial(ctp)
	t.NoError(err)
	ch, err := conn.Channel()
	t.NoError(err)

	// intentionally misconfigure an exchange
	err = ch.ExchangeDelete("M2CPSDK-Signals", false, false)
	t.NoError(err)
	err = ch.ExchangeDeclare("M2CPSDK-Signals", "topic", false, false, false, false, nil)
	t.NoError(err)

	// close everything
	_ = ch.Close()
	_ = conn.Close()
	ctp.Cancel()

	// init new connection - exchanges should be fixed
	ctp = contextplus.NewContextPlus()
	c, err = newAmqpConnection(ctp)
	t.NoError(err)
	t.NotNil(c)

	ctp.Cancel()
}

func (t *TestSuite) TestAmqpMisconfigurationAndHealingDurability() {
	var err error

	// create connection - this will create the exchanges
	c, err := newAmqpConnection(t.ctp)
	t.NoError(err)
	t.NotNil(c)

	var con *amqp.Connection
	var ch *amqp.Channel

	exchange := "test-reconfigure"
	topic := "topic"

	con, ch = t.getConAndChannel(c, c.url)
	t.NoError(ch.ExchangeDelete(exchange, false, false))
	_ = ch.Close()
	_ = con.Close()

	con, ch = t.getConAndChannel(c, c.url)
	t.NoError(ch.ExchangeDeclare(exchange, topic, true, false, false, false, nil))
	_ = ch.Close()
	_ = con.Close()

	// ATTENTION: Passive declaration doesn't throw an error if durability mismatches!
	con, ch = t.getConAndChannel(c, c.url)
	t.NoError(ch.ExchangeDeclarePassive(exchange, topic, false, false, false, false, nil))
	_ = ch.Close()
	_ = con.Close()

	//con, ch = t.getConAndChannel(c.url)
	//t.NoError(ch.ExchangeDeclare(exchange, topic, true, false, false, false, nil))
	//_ = ch.Close()
	//_ = con.Close()

}

func (t *TestSuite) getConAndChannel(c *amqpClient, url string) (*amqp.Connection, *amqp.Channel) {
	// make an extra connection for messing things up
	conn, err := c.Dial(c.ctp)
	t.NoError(err)
	ch, err := conn.Channel()
	t.NoError(err)

	return conn, ch
}
