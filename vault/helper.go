package vault

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/vault/api"
)

func VaultHealth() (string, error) {
	resp, err := http.Get(vaultAddress + "/v1/sys/health")
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// lookup current root generation status
func GenerateRootStatus() (*api.GenerateRootStatusResponse, error) {
	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return nil, err
	}
	client.SetAddress(vaultAddress)
	return client.Sys().GenerateRootStatus()
}

func GenerateRootInit(otp string) (*api.GenerateRootStatusResponse, error) {
	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return nil, err
	}
	client.SetAddress(vaultAddress)
	return client.Sys().GenerateRootInit(otp, "")
}

func GenerateRootUpdate(shard, nonce string) (*api.GenerateRootStatusResponse, error) {
	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return nil, err
	}
	client.SetAddress(vaultAddress)
	return client.Sys().GenerateRootUpdate(shard, nonce)
}

func GenerateRootCancel() error {
	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return err
	}
	client.SetAddress(vaultAddress)
	return client.Sys().GenerateRootCancel()
}

func WriteToCubbyhole(name string, data map[string]interface{}) (interface{}, error) {
	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return nil, err
	}
	client.SetAddress(vaultAddress)
	client.SetToken(vaultToken)
	return vaultClient.Logical().Write("cubbyhole/" + name, data)
}

func ReadFromCubbyhole(name string) (*api.Secret, error) {
	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return nil, err
	}
	client.SetAddress(vaultAddress)
	client.SetToken(vaultToken)
	return vaultClient.Logical().Read("cubbyhole/" + name)
}

func DeleteFromCubbyhole(name string) (*api.Secret, error) {
	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return nil, err
	}
	client.SetAddress(vaultAddress)
	client.SetToken(vaultToken)
	return vaultClient.Logical().Delete("cubbyhole/" + name)
}

func renewServerToken() (err error) {
	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return err
	}
	client.SetAddress(vaultAddress)
	client.SetToken(vaultToken)
	_, err = client.Auth().Token().RenewSelf(0)
	return
}

func WrapData(wrapttl string, data map[string]interface{}) (string, error) {
	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return "", err
	}
	client.SetAddress(vaultAddress)
	client.SetToken(vaultToken)

	client.SetWrappingLookupFunc(func(operation, path string) string {
		return wrapttl
	})

	resp, err := client.Logical().Write("/sys/wrapping/wrap", data)
	if err != nil {
		return "", err
	}
	return resp.WrapInfo.Token, nil
}

func UnwrapData(wrappingToken string) (map[string]interface{}, error) {
	// set up vault client
	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return nil, err
	}
	client.SetAddress(vaultAddress)
	client.SetToken(wrappingToken)

	// make a raw unwrap call. This will use the token as a header
	resp, err := client.Logical().Unwrap("")
	if err != nil {
		return nil, errors.New("Failed to unwrap provided token, revoke it if possible\nReason:" + err.Error())
	}
	return resp.Data, nil
}
