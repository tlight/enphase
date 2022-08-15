package enphase

import (
	"testing"
)

func TestGetProduction(t *testing.T) {
	e := &Envoy{URL: "http://192.168.1.14"}
	obj, err := e.GetProduction()
	if err != nil {
		t.Fatal(err)
	} else {
		//		t.Logf("%f", obj.Consumption[0].WNow)
		t.Logf("%s", obj)
	}
}
