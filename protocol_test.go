package transport

import (
	"go-rabbit-client/common/utils"
	"testing"
)

func TestAuthType_Marshal(t *testing.T) {
	buffer := utils.GetBuffer()
	defer utils.PutBuffer(buffer)

	auth := AuthType{
		AccessToken: "helloworldJack",
	}

	_, err := auth.Marshal(buffer)
	if err != nil {
		t.Error(err)
	}
	cmp := AuthType{}
	_, err = cmp.UnMarshal(buffer)
	if err != nil {
		t.Error(err)
	}

	if cmp!=auth {
		t.Error("unmarshal not equal")
	}
}

func TestConnectType_Marshal(t *testing.T) {
	buffer := utils.GetBuffer()
	defer utils.PutBuffer(buffer)

	connect := ConnectType{
		AddressType: AddressEnumTypeDomain,
		Host:        "qiaohong.org",
		Port:        443,
	}

	_, err := connect.Marshal(buffer)
	if err != nil {
		t.Error(err)
	}
	cmp := ConnectType{}
	_, err = cmp.UnMarshal(buffer)
	if err != nil {
		t.Error(err)
	}

	if cmp!=connect {
		t.Error("unmarshal not equal")
	}
}
