package contextplus

import (
	"fmt"
	"m2cp"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
)

// simulation of a worker thread
type branch struct {
	name   string
	ctp    m2cp.ContextPlus
	status string
	t      *testing.T
}

func (b *branch) run() {
	b.status = "running"
	for {
		select {
		case <-b.ctp.Done():
			b.t.Logf("Branch %s canceled", b.name)
			b.status = "canceled"
			return
		case <-time.After(30 * time.Second): // This timeout can be adjusted
			b.t.Errorf("Branch %s was not canceled", b.name)
			b.ctp.Cancel() // Ensure cleanup if not canceled
			b.status = "timeout"
			return
		default:
			b.t.Logf("Branch %s running..\n", b.name)
			b.ctp.Sleep(100 * time.Millisecond)
		}
	}
}

type TestSuite struct {
	suite.Suite
	wg            sync.WaitGroup
	branches      []branch
	branchCounter int
}

func (t *TestSuite) SetupSuite() {
	fmt.Println(">>> From SetupSuite")
}

func (t *TestSuite) TearDownSuite() {
	fmt.Println(">>> From TearDownSuite")
}

func (t *TestSuite) SetupTest() {
	fmt.Println("-- From SetupTest")
	t.wg = sync.WaitGroup{}
	t.branchCounter = 0
}

func (t *TestSuite) TearDownTest() {
	fmt.Println("-- From TearDownTest")
	defer goleak.VerifyNone(t.T())
}

func TestCompleteSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (t *TestSuite) newBranch(ctp m2cp.ContextPlus) branch {
	name := fmt.Sprintf("t%d", t.branchCounter)
	return t.newNamedBranch(name, ctp)
}

func (t *TestSuite) TestNewIContextPlus() {
	ctp := NewContextPlus()
	t.NotNil(ctp)
	ctp.Cancel()
}

func (t *TestSuite) newNamedBranch(name string, ctp m2cp.ContextPlus) branch {
	t.branchCounter++
	b := branch{
		name,
		ctp,
		"starting",
		t.T(),
	}
	return b
}

func (t *TestSuite) runBranches() {
	for i := range t.branches {
		b := &t.branches[i]
		t.wg.Add(1)
		go func() {
			defer t.wg.Done()
			b.run()
		}()
	}
}

func (t *TestSuite) countBranchesRunning() int {
	count := 0
	for _, b := range t.branches {
		if b.status == "running" {
			count++
		}
	}
	return count
}

func (t *TestSuite) TestEssentialLeaks() {
	ctp := NewContextPlus()
	assert.NotNil(t.T(), ctp)
	ctp.Cancel()
}

func (t *TestSuite) TestCancelRoot() {
	ctp := NewContextPlus()

	t.branches = []branch{
		t.newBranch(ctp.Branch()),
		t.newBranch(ctp.Branch()),
		t.newBranch(ctp.Branch()),
		t.newBranch(ctp.Branch()),
		t.newBranch(ctp.Branch()),
	}
	t.runBranches()

	ctp.Sleep(1200 * time.Millisecond)
	ctp.Cancel()
	t.wg.Wait()
}

func (t *TestSuite) TestCancelSimple() {
	ctp := NewContextPlus()
	ctpA := ctp.Branch()

	ctp.Sleep(200 * time.Millisecond)
	ctpA.Cancel()

	ctp.Sleep(200 * time.Millisecond)
	t.False(ctp.IsCancelled(), "ctp should not be cancelled")

	ctp.Cancel()
}

func (t *TestSuite) TestCancelBranchSimple() {
	t.T().Skip("Test sometimes fails due to timing issues. Needs investigation.")

	ctp := NewContextPlus()
	ctpA := ctp.Branch()

	t.branches = []branch{
		t.newNamedBranch("ctp", ctp),
		t.newNamedBranch("ctpA", ctpA),
	}
	t.runBranches()

	ctp.Sleep(500 * time.Millisecond)
	t.Equal(2, t.countBranchesRunning())

	ctpA.Cancel()
	ctp.Sleep(500 * time.Millisecond)
	t.Equal(1, t.countBranchesRunning())

	ctp.Cancel()
	ctp.Sleep(500 * time.Millisecond)
	t.Equal(0, t.countBranchesRunning())

	t.wg.Wait()
}

func (t *TestSuite) TestCancelBranch() {
	t.T().Skip("Test sometimes fails due to timing issues. Needs investigation.")

	ctp := NewContextPlus()
	ctpA := ctp.Branch()
	ctpA1 := ctpA.Branch()
	ctpA2 := ctpA.Branch()
	ctpB := ctp.Branch()
	ctpB1 := ctpB.Branch()
	ctpB2 := ctpB.Branch()

	t.branches = []branch{
		t.newNamedBranch("ctp-1", ctp),
		t.newNamedBranch("ctp-2", ctp),
		t.newNamedBranch("ctpA-1", ctpA),
		t.newNamedBranch("ctpA-2", ctpA),
		t.newNamedBranch("ctpA1-1", ctpA1),
		t.newNamedBranch("ctpA1-2", ctpA1),
		t.newNamedBranch("ctpA2-1", ctpA2),
		t.newNamedBranch("ctpA2-2", ctpA2),
		t.newNamedBranch("ctpB-1", ctpB),
		t.newNamedBranch("ctpB-2", ctpB),
		t.newNamedBranch("ctpB1-1", ctpB1),
		t.newNamedBranch("ctpB1-2", ctpB1),
		t.newNamedBranch("ctpB2-1", ctpB2),
		t.newNamedBranch("ctpB2-2", ctpB2),
	}
	t.runBranches()

	ctp.Sleep(500 * time.Millisecond)
	t.Equal(14, t.countBranchesRunning())

	ctpA.Cancel()
	ctp.Sleep(500 * time.Millisecond)
	t.Equal(8, t.countBranchesRunning())

	ctpB1.Cancel()
	ctp.Sleep(500 * time.Millisecond)
	t.Equal(6, t.countBranchesRunning())

	ctpB.Cancel()
	ctp.Sleep(500 * time.Millisecond)
	t.Equal(2, t.countBranchesRunning())

	ctp.Cancel()
	ctp.Sleep(500 * time.Millisecond)
	t.Equal(0, t.countBranchesRunning())

	t.wg.Wait()
}

func TestSystemContext01(t *testing.T) {
	ctp := NewContextPlus()
	assert.NotNil(t, ctp)
}

func TestIsCanceled01(t *testing.T) {
	ctp := NewContextPlus()
	assert.NotNil(t, ctp)
	done := ctp.IsCancelled()
	assert.False(t, done)
}

func TestIsCanceled02(t *testing.T) {
	ctp := NewContextPlus()
	assert.NotNil(t, ctp)
	ctp.Cancel()
	done := ctp.IsCancelled()
	assert.True(t, done)
}

func TestSetLogLevel01(t *testing.T) {
	ctp := NewContextPlus()
	assert.NotNil(t, ctp)
	levels := []m2cp.LogLevel{
		m2cp.LogLevelDebug,
		m2cp.LogLevelInfo,
		m2cp.LogLevelWarn,
		m2cp.LogLevelError,
		m2cp.LogLevelFatal}
	for _, expectedLevel := range levels {
		ctp.SetLogLevel(expectedLevel)
		actualLevel := ctp.GetLogLevel()
		assert.Equal(t, expectedLevel, actualLevel)
	}
}

func TestLogOnDifferentLevels(t *testing.T) {
	ctp := NewContextPlus()
	assert.NotNil(t, ctp)
	ctp.LogDebug("debug")
	ctp.LogInfo("info")
	ctp.LogWarn("warn")
	ctp.LogError("error")
	// ctp.LogFatal("fatal") // Calls os.Exit(1)
	// TODO: bad test due to no assertions. But improves test coverage.
}

func (t *TestSuite) TestPrintALogMessage() {
	ctp := NewContextPlus()
	ctp.SetModule("foo")
	ctp.LogInfo("hello")
	ctp2 := ctp.BranchWithName("bar")
	ctp2.LogInfo("world")
	ctp.Cancel()
}

func (t *TestSuite) TestSleepCancelled() {
	ctp := NewContextPlus()
	start := time.Now()
	go func() {
		ctp.Sleep(10 * time.Second)
	}()
	ctp.Cancel()
	elapsed := time.Since(start)
	t.True(elapsed < 1*time.Second)
}

func (t *TestSuite) TestSleepLong() {
	ctp := NewContextPlus()
	start := time.Now()
	ctp.Sleep(3 * time.Second)
	elapsed := time.Since(start)
	t.True(elapsed > 1*time.Second)
	ctp.Cancel()
}

func (t *TestSuite) TestSetModule() {
	ctp := NewContextPlus()
	ctp.SetModule("testcase12345")

	log := captureLogOutput(ctp, func(ctp m2cp.ContextPlus) {
		ctp.LogInfo("foo")
	})

	t.Contains(log, "testcase12345")
	ctp.Cancel()
}

func (t *TestSuite) TestSetTaskId() {

	ctp := NewContextPlus()
	ctp.SetTaskId("abcde")

	log := captureLogOutput(ctp, func(ctp m2cp.ContextPlus) {
		ctp.LogInfo("foo")
	})

	t.Contains(log, "abcde")
	ctp.Cancel()

}

func (t *TestSuite) TestSetDevice() {
	ctp := NewContextPlus()
	ctp.SetDeviceId("nicedevice")

	log := captureLogOutput(ctp, func(ctp m2cp.ContextPlus) {
		ctp.LogInfo("foo")
	})

	t.Contains(log, "nicedevice")
	ctp.Cancel()
}

func (t *TestSuite) TestSetUser() {
	ctp := NewContextPlus()
	ctp.SetUserId("special1234")

	log := captureLogOutput(ctp, func(ctp m2cp.ContextPlus) {
		ctp.LogInfo("foo")
	})

	t.Contains(log, "special1234")
	ctp.Cancel()
}

func (t *TestSuite) TestBranchWithTimeout01() {
	ctp := NewContextPlus()
	t.NotNil(ctp)
	ctp2 := ctp.BranchWithTimeout(1000 * time.Millisecond)
	t.NotNil(ctp2)

	ctp2Cancelled := false

	run := func(name string, ctx m2cp.ContextPlus) {
		for {
			select {
			case <-ctx.Done():
				fmt.Print("Branch '" + name + "' canceled")
				ctp2Cancelled = true
				return
			default:
				ctx.Sleep(time.Millisecond)
			}
		}
	}
	go run("ctp2", ctp2)

	ctp.Sleep(1500 * time.Millisecond)

	t.True(ctp2Cancelled)
	t.False(ctp.IsCancelled())
	ctp.Cancel()
}

func captureLogOutput(ctp m2cp.ContextPlus, f func(ctp m2cp.ContextPlus)) string {
	r, w, _ := os.Pipe()
	ctp.SetLogOutput(w)

	f(ctp)
	ctp.LogInfo("bar")
	ctp.SetLogOutput(os.Stdout)

	_ = w.Close()

	buf := make([]byte, 10240) // 10k buffer
	var out []byte
	n, _ := r.Read(buf)
	if n > 0 {
		out = append(out, buf[:n]...)
	}

	_ = r.Close()

	return string(out)
}
