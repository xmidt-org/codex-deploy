package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	db "github.com/xmidt-org/codex-db"
	"github.com/xmidt-org/voynicrypto"
	"github.com/xmidt-org/wrp-go/wrp"
	"github.com/yugabyte/gocql"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

type gungnirFeature struct {
	db   db.Inserter
	host string
	resp *http.Response
}

func handleGungnir(s *godog.Suite) {
	gungnirFeature := gungnirFeature{
		db: GetInserter(),
	}

	s.BeforeScenario(gungnirFeature.resetResponse)

	s.Step(`^Gungnir is at "([^"]*)"$`, gungnirFeature.setHost)
	s.Step(`^I send "(GET|POST|PUT|DELETE)" request to "([^"]*)" with ID "(.*)"$`, gungnirFeature.iSendrequestTo)
	s.Step(`^the response code should be (\d+)$`, gungnirFeature.theResponseCodeShouldBe)
	s.Step(`^the response should match json:$`, gungnirFeature.theResponseShouldMatchJSON)
	s.Step(`^the data:$`, gungnirFeature.theData)

}

func (a *gungnirFeature) setHost(host string) error {
	a.host = host
	return nil
}

func (a *gungnirFeature) resetResponse(interface{}) {
	a.resp = nil
	// Truncate table
	session, err := gocql.NewCluster("localhost:9042").CreateSession()
	if err != nil {
		panic(err)
	}
	err = session.Query("TRUNCATE TABLE devices.events").Exec()
	if err != nil {
		panic(err)
	}
	session.Close()
}

func (a *gungnirFeature) theData(data *gherkin.DataTable) error {
	head := data.Rows[0].Cells
	rows := []db.Record{}
	for i := 1; i < len(data.Rows); i++ {
		record := db.Record{
			Type:      db.State,
			BirthDate: time.Now().UnixNano(),
			DeathDate: time.Now().Add(time.Hour).UnixNano(),
			RowID:     gocql.TimeUUID().String(),
		}
		event := wrp.Message{
			Type:        4,
			Source:      "dns:talaria",
			Destination: "",
			ContentType: "json",
			Headers:     []string{},
			Metadata:    map[string]string{},
			Payload:     []byte{},
			PartnerIDs:  []string{},
		}
		holdType := ""

		for n, cell := range data.Rows[i].Cells {
			switch head[n].Value {
			case "deviceID":
				record.DeviceID = cell.Value
			case "type":
				holdType = cell.Value
			case "birthdate":
				birth, err := strconv.ParseInt(cell.Value, 10, 64)
				if err != nil {
					return err
				}
				record.BirthDate = birth
			case "deathdate":
				death, err := strconv.ParseInt(cell.Value, 10, 64)
				if err != nil {
					return err
				}
				record.DeathDate = death
			case "payload":
				payload, err := base64.StdEncoding.DecodeString(cell.Value)
				if err != nil {
					return err
				}
				event.Payload = payload
			case "partnerID":
				event.PartnerIDs = []string{cell.Value}
			case "transaction_uuid":
				event.TransactionUUID = cell.Value
			case "metadata":
				metadata := map[string]string{}
				err := json.Unmarshal([]byte(cell.Value), &metadata)
				if err != nil {
					return err
				}
				event.Metadata = metadata

			default:
				return fmt.Errorf("unexpected column name: %s", head[n].Value)
			}

		}
		event.Destination = fmt.Sprintf("event:device-status/%s/%s", record.DeviceID, holdType)
		encrypter := voynicrypto.NOOP{}
		encyptedData, nonce, err := encrypter.EncryptMessage(wrp.MustEncode(&event, wrp.Msgpack))
		if err != nil {
			return err
		}
		record.Data = encyptedData
		record.Nonce = nonce
		record.Alg = string(encrypter.GetAlgorithm())
		record.KID = encrypter.GetKID()
		rows = append(rows, record)
	}
	return a.db.InsertRecords(rows...)
}

func (a *gungnirFeature) iSendrequestTo(method, endpoint string, deviceID string) (err error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s/api/v1/device/%s%s", a.host, deviceID, endpoint), nil)
	req.Header.Add("Authorization", "Basic YXV0aEhlYWRlcjp0ZXN0")
	if err != nil {
		return
	}

	// handle panic
	defer func() {
		switch t := recover().(type) {
		case string:
			err = fmt.Errorf(t)
		case error:
			err = t
		}
	}()

	resp, err := http.DefaultClient.Do(req)
	a.resp = resp
	return err
}
func (a *gungnirFeature) theResponseCodeShouldBe(code int) error {
	if code != a.resp.StatusCode {
		if a.resp.StatusCode >= 400 {
			data, _ := ioutil.ReadAll(a.resp.Body)
			fmt.Println(a.resp.Request.URL)
			return fmt.Errorf("expected response code to be: %d, but actual is: %d, response message: %s", code, a.resp.StatusCode, string(data))
		}
		fmt.Println(a.resp.Request.URL)
		return fmt.Errorf("expected response code to be: %d, but actual is: %d, with error: %s", code, a.resp.StatusCode, a.resp.Header.Get("x-codex"))
	}
	return nil
}

func (a *gungnirFeature) theResponseShouldMatchJSON(body *gherkin.DocString) (err error) {
	var expected, actual interface{}

	// re-encode expected response
	if err = json.Unmarshal([]byte(body.Content), &expected); err != nil {
		return
	}

	// re-encode actual response too
	data, err := ioutil.ReadAll(a.resp.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, &actual); err != nil {
		return
	}

	// the matching may be adapted per different requirements.
	if !reflect.DeepEqual(expected, actual) {
		fmt.Println(body.ContentType)
		fmt.Println(string(data))
		return fmt.Errorf("expected JSON does not match actual, %v vs. %v", expected, actual)
	}
	return nil
}
