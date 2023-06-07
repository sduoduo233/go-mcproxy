package core_test

import (
	"mcproxy/core"
	"testing"
)

func TestSRV(t *testing.T) {
	tests := make(map[string]string)
	tests["mc.hypixel.net:25565"] = "mc.hypixel.net:25565"
	tests["blocksmc.com"] = "ccc.blocksmc.com.:25565"
	tests["mc.hypixel.net"] = "mc.hypixel.net:25565"

	for k, v := range tests {
		addr, err := core.Resolve(k)
		if err != nil {
			t.Error(k, err)
			return
		}

		if addr != v {
			t.Errorf("%s: %s != %s", k, addr, v)
			return
		}
	}

}
