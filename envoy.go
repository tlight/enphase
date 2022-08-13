package enphase

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/icholy/digest"
)

type Envoy struct {
	SerialNumber      string
	InstallerUsername string
	InstallerPassword string
	URL               string
}

type PhaseInfo struct {
	Power       float32 `json:"p"`
	Q           float32 `json:"q"` //?
	S           float32 `json:"s"` //?
	Voltage     float32 `json:"v"`
	Current     float32 `json:"i"`
	PowerFactor float32 `json:"pf"`
	Frequency   float32 `json:"f"`
}

type Payload struct {
	Production struct {
		A PhaseInfo `json:"ph-a"`
		B PhaseInfo `json:"ph-b"`
		C PhaseInfo `json:"ph-c"`
	} `json:"production"`
	Net struct {
		A PhaseInfo `json:"ph-a"`
		B PhaseInfo `json:"ph-b"`
		C PhaseInfo `json:"ph-c"`
	} `json:"net-consumption"`
	Consumption struct {
		A PhaseInfo `json:"ph-a"`
		B PhaseInfo `json:"ph-b"`
		C PhaseInfo `json:"ph-c"`
	} `json:"total-consumption"`
}

// curl --digest --user installer:FEb5Dafd http://192.168.1.14/stream/meter
func NewEnvoy(s, u, p, url string) *Envoy {
	return &Envoy{
		SerialNumber:      s,
		InstallerUsername: u,
		InstallerPassword: p,
		URL:               url,
	}
}

func DefaultHandler(payload *Payload) {
	fmt.Printf("Solar: %4.0f, Net %4.0f, Consumption: %4.0f (A %4.0f B %4.0f C %4.0f)\n",
		payload.Production.A.Power+payload.Production.B.Power+payload.Production.C.Power,
		(payload.Production.A.Power+payload.Production.B.Power+payload.Production.C.Power)-
			(payload.Consumption.A.Power+payload.Consumption.B.Power+payload.Consumption.C.Power),
		payload.Consumption.A.Power+payload.Consumption.B.Power+payload.Consumption.C.Power,
		payload.Consumption.A.Power,
		payload.Consumption.B.Power,
		payload.Consumption.C.Power)
}

func (e *Envoy) Stream(handler func(*Payload)) {
	url := e.URL + "/stream/meter"

	client := &http.Client{
		//Timeout: 30 * time.Second,
		Transport: &digest.Transport{
			Username: e.InstallerUsername,
			Password: e.InstallerPassword,
		},
	}
	res, err := client.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	//resp, err := http.Get(url)
	reader := bufio.NewReader(res.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			log.Fatalln(err)
		}
		json, err := GetJSON(string(line))

		//fmt.Println(">" + string(line) + "<")
		if err != nil {

		} else {
			//	fmt.Println(">" + json + "<")
			payload, err := DecodeJSON(json)
			if err != nil {
				log.Println(err)
			} else {
				// log.Println(payload)
				if handler != nil {
					handler(payload)
				} else {
					DefaultHandler(payload)
				}

			}

		}
	}
}

func GetJSON(s string) (string, error) {
	re := regexp.MustCompile("data: (.*)")
	json := re.FindStringSubmatch(s)
	if len(json) == 0 {
		return "", errors.New("no matching data")
	} else {
		return json[1], nil
	}

}

func DecodeJSON(s string) (*Payload, error) {
	payload := &Payload{}
	reader := strings.NewReader(s)

	err := json.NewDecoder(reader).Decode(payload)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return payload, nil
}
