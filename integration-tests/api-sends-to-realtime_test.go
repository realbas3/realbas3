package integrationtests

import (
	"github.com/go-neutrino/neutrino/log"
	"github.com/go-neutrino/neutrino/models"
	"github.com/go-neutrino/neutrino/realtime-service/client"
	"sync"
	"testing"
	"time"
)

func noop (ev neutrinoclient.NeutrinoEvent, m models.JSON) {
	//dummy func
}

func readMessages(t *testing.T, expectedEv []string, times int,
		cb func(neutrinoclient.NeutrinoEvent, models.JSON)) *sync.WaitGroup {

	wg := &sync.WaitGroup{}
	wg.Add(times)

	go func() {
		for i := 0; i < times; i++ {
			select {
			case ev := <-Data.Event:
				log.Info("Test event:", ev.Code, ev.Data)

				var m models.JSON
				err := m.FromString([]byte(ev.Data))
				if err != nil {
					t.Error(err)
				}

				cb(ev, m)

				foundEv := false
				for _, e := range expectedEv {
					if e == ev.Code {
						foundEv = true
					}
				}

				if !foundEv {
					t.Error("Unexptected event.", ev.Data)
				}
			}

			wg.Done()
		}
	}()

	time.Sleep(time.Second * 1)

	return wg
}

func TestInsertIntoType(t *testing.T) {
	wg := readMessages(t, []string{neutrinoclient.EVENT_CREATE}, 1, noop)

	CreateItem("test", models.JSON{
		"name": "test",
	})

	wg.Wait()
}

func TestUpdateItem(t *testing.T) {
	cb := func (ev neutrinoclient.NeutrinoEvent, m models.JSON) {
		if (ev.Code == neutrinoclient.EVENT_UPDATE) {
			n := m["payload"].(map[string]interface{})["name"]
			if (n != "updated-test") {
				t.Error("Incorrect updated name:", n, "Expected: updated-test")
			}
		}
	}

	wg := readMessages(t, []string{neutrinoclient.EVENT_CREATE, neutrinoclient.EVENT_UPDATE}, 2, cb)

	id := CreateItem("test", models.JSON{
		"name": "test",
	})["_id"].(string)

	UpdateItem("test", id, models.JSON{
		"name": "updated-test",
	})

	wg.Wait()
}
