package enphase

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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

type Production struct {
	Production  []*ProductionInfo `json:"production"`
	Consumption []*ProductionInfo `json:"consumption"`
	Storage     []*ProductionInfo `json:"storage"`
}

type ProductionInfo struct {
	Type             string  `json:"type"`
	ActiveCount      int     `json:"activeCount"`      // 1
	MeasurementType  string  `json:"measurementType"`  // "production",
	ReadingTime      int64   `json:"readingTime"`      // 1660454270,
	WNow             float32 `json:"wNow"`             // 5261.947
	WhLifetime       float32 `json:"whLifetime"`       // 30015.385,
	VarhLeadLifetime float32 `json:"varhLeadLifetime"` // 3655.084,
	VarhLagLifetime  float32 `json:"varhLagLifetime"`  // 14644.192,
	VahLifetime      float32 `json:"vahLifetime"`      // 44573.333,
	RmsCurrent       float32 `json:"rmsCurrent"`       // 22.295,
	RmsVoltage       float32 `json:"rmsVoltage"`       // 715.657,
	ReactPwr         float32 `json:"reactPwr"`         // -744.02,
	ApprntPwr        float32 `json:"apprntPwr"`        // 5319.274,
	PwrFactor        float32 `json:"pwrFactor"`        // 0.99,
	WhToday          float32 `json:"whToday"`          // 16796.385,
	WhLastSevenDays  float32 `json:"whLastSevenDays"`  // 29147.385,
	VahToday         float32 `json:"vahToday"`         // 20645.333,
	VarhLeadToday    float32 `json:"varhLeadToday"`    // 2802.084,
	VarhLagToday     float32 `json:"varhLagToday"`     // 3578.192
}

func (e *Envoy) GetProduction() (*Production, error) {
	url := e.URL + "/production.json"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	reader := strings.NewReader(string(body))
	obj := &Production{}
	err = json.NewDecoder(reader).Decode(obj)
	if err != nil {
		log.Fatalln(err)
	}
	return obj, nil
	//payload, err := DecodeJSON(json)
}

func (p *Production) String() string {
	return fmt.Sprintf("+%4.0f -%4.0f =%4.0f",
		p.Production[1].WNow,
		p.Consumption[0].WNow,
		p.Consumption[0].WNow-p.Production[1].WNow)
}

type StreamMeter struct {
	Production struct {
		A StreamMeterInfo `json:"ph-a"`
		B StreamMeterInfo `json:"ph-b"`
		C StreamMeterInfo `json:"ph-c"`
	} `json:"production"`
	Net struct {
		A StreamMeterInfo `json:"ph-a"`
		B StreamMeterInfo `json:"ph-b"`
		C StreamMeterInfo `json:"ph-c"`
	} `json:"net-consumption"`
	Consumption struct {
		A StreamMeterInfo `json:"ph-a"`
		B StreamMeterInfo `json:"ph-b"`
		C StreamMeterInfo `json:"ph-c"`
	} `json:"total-consumption"`
}

func (s *StreamMeter) String() string {
	return fmt.Sprintf("+%4.0f -%4.0f =%4.0f",
		s.Production.A.Power+s.Production.B.Power+s.Production.C.Power,
		s.Consumption.A.Power+s.Consumption.B.Power+s.Consumption.C.Power,
		s.Net.A.Power+s.Net.B.Power+s.Net.C.Power)
}

type StreamMeterInfo struct {
	Power       float32 `json:"p"`
	Q           float32 `json:"q"` //?
	S           float32 `json:"s"` //?
	Voltage     float32 `json:"v"`
	Current     float32 `json:"i"`
	PowerFactor float32 `json:"pf"`
	Frequency   float32 `json:"f"`
}

func NewEnvoy(s, u, p, url string) *Envoy {
	return &Envoy{
		SerialNumber:      s,
		InstallerUsername: u,
		InstallerPassword: p,
		URL:               url,
	}
}

func (e *Envoy) GetStreamMeter(handler func(*StreamMeter)) error {
	url := e.URL + "/stream/meter"
	client := &http.Client{
		//Timeout: 30 * time.Second,
		Transport: &digest.Transport{
			Username: e.InstallerUsername,
			Password: e.InstallerPassword,
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return err
		}
		re := regexp.MustCompile("data: (.*)") // extract json from string (everything after data:
		matches := re.FindStringSubmatch(string(line))
		if len(matches) == 1 {
			reader := strings.NewReader(matches[1])
			obj := &StreamMeter{}
			err := json.NewDecoder(reader).Decode(obj)
			if err != nil {
				return err
			}
			handler(obj)
		}
	}
}

func DefaultStreamMeterHandler(obj *StreamMeter) {
	fmt.Printf("Solar: %4.0f, Net %4.0f, Consumption: %4.0f (A %4.0f B %4.0f C %4.0f)\n",
		obj.Production.A.Power+obj.Production.B.Power+obj.Production.C.Power,
		(obj.Production.A.Power+obj.Production.B.Power+obj.Production.C.Power)-
			(obj.Consumption.A.Power+obj.Consumption.B.Power+obj.Consumption.C.Power),
		obj.Consumption.A.Power+obj.Consumption.B.Power+obj.Consumption.C.Power,
		obj.Consumption.A.Power,
		obj.Consumption.B.Power,
		obj.Consumption.C.Power)
}
