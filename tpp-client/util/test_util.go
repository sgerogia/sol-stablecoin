package test_util

import (
	"github.com/sgerogia/hello-stablecoin/tpp-client/bank"
	"os"
)

const INSTITUTION = "natwest-sandbox"
const AMOUNT = "1"

func Payer() *bank.AccountDetails {
	return &bank.AccountDetails{
		Name:          "John Doe",
		SortCode:      "500000",
		AccountNumber: "12345601",
	}
}

func Receiver() *bank.AccountDetails {
	return &bank.AccountDetails{
		Name:          "ProvableGBP Limited",
		SortCode:      "500000",
		AccountNumber: "87654301",
	}
}

type NatwestSandboxInfo struct {
	ClientId         string
	ClientSecret     string
	RedirectUrl      string
	CustomerUsername string
}

// GetNatwestSandboxInfo returns the information needed to connect to the Natwest sandbox or nil if the environment variables are not set
func GetNatwestSandboxInfo() *NatwestSandboxInfo {

	if !isEnvVarSet("NATWEST_SANDBOX_CLIENT_ID") ||
		!isEnvVarSet("NATWEST_SANDBOX_CLIENT_SECRET") ||
		!isEnvVarSet("NATWEST_SANDBOX_REDIRECT_URL") ||
		!isEnvVarSet("NATWEST_SANDBOX_CUSTOMER_USERNAME") {
		return nil
	}
	return &NatwestSandboxInfo{
		ClientId:         os.Getenv("NATWEST_SANDBOX_CLIENT_ID"),
		ClientSecret:     os.Getenv("NATWEST_SANDBOX_CLIENT_SECRET"),
		RedirectUrl:      os.Getenv("NATWEST_SANDBOX_REDIRECT_URL"),
		CustomerUsername: os.Getenv("NATWEST_SANDBOX_CUSTOMER_USERNAME"),
	}
}

func isEnvVarSet(e string) bool {
	return os.Getenv(e) != ""
}
