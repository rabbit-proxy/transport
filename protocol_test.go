package transport

import (
	"testing"
)

func TestAuthType_Marshal(t *testing.T) {
	buffer := GetBuffer()
	defer PutBuffer(buffer)

	auth := AuthType{
		Token: "helloworldJack",
		FuncFlag: FunctionTagMux,
	}

	n, err := auth.Marshal(buffer)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("used length:%d, buffer:%v", n, buffer[:n])

	cmp := AuthType{}
	_, err = cmp.UnMarshal(buffer)
	if err != nil {
		t.Fatal(err)
	}

	if cmp!=auth {
		t.Fatal("unmarshal not equal")
	}
}


func TestConnectType_Marshal(t *testing.T) {
	buffer := GetBuffer()
	defer PutBuffer(buffer)

	origin := ConnectType{
		ProtocolType: TransProtocolTypeTcp,
		AddressType: AddressEnumTypeDomain,
		Host:        "qiaohong.org",
		Port:        443,
	}

	n, err := origin.Marshal(buffer)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("used length:%d, buffer:%v", n, buffer[:n])

	cmp := ConnectType{}
	n, err = cmp.UnMarshal(buffer)

	if err != nil {
		t.Fatal(err)
	}

	if cmp!= origin {
		t.Fatalf("unmarshal not equal，origin:%v, cmp:%v", origin, cmp)
	}
}
