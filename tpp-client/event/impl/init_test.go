package event_impl_test

import (
	test_util "github.com/sgerogia/sol-stablecoin/tpp-client/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"os"
	"testing"
)

type TestingInfo struct {
	chainInfo *test_util.ChainInfo
	bankInfo  *test_util.NatwestSandboxInfo
	l         *zap.SugaredLogger
	o         *observer.ObservedLogs
}

// --- Package variable for child tests ---
var testingCtx = TestingInfo{}

func TestMain(m *testing.M) {

	// global setup
	chain, err := test_util.DeployProvableGBPAndCreateAccounts()
	if err != nil {
		os.Exit(-1)
	}
	testingCtx.chainInfo = chain
	testingCtx.bankInfo = test_util.GetNatwestSandboxInfo()
	if testingCtx.bankInfo == nil {
		os.Exit(-1)
	}

	// logger and log observer
	core, obs := observer.New(zap.InfoLevel)
	testingCtx.l = zap.New(core).Sugar()
	testingCtx.o = obs

	// Run test suites
	exitVal := m.Run()

	// clean up here

	// ...and exit test suite
	os.Exit(exitVal)
}
