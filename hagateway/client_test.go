package hagateway_test

import (
	"log"
	"testing"

	"github.com/NHollmann/ha-gateway/hagateway"
)

func TestEntityMatchName(t *testing.T) {
	cs := []struct {
		entityName string
		checkName  string
		expected   bool
	}{
		{"binary_sensor.light_switch", "binary_sensor.light_switch", true},
		{"binary_sensor.window_contact", "binary_sensor.light_switch", false},
		{"binary_sensor.*", "binary_sensor.light_switch", true},
		{"binary_sensor.*", "sensor.temperature", false},
		{"*", "binary_sensor.light_switch", true},
		{"*", "binary_sensor", true},
		{"*", "", true},
	}

	for _, c := range cs {
		entity := hagateway.Entity{Name: c.entityName}
		if r, e := entity.MatchName(c.checkName), c.expected; r != e {
			log.Printf(
				"Entity{%s}.MatchName(%s) returned %v, expected %v",
				c.entityName, c.checkName, r, e,
			)
			t.Fail()
		}
	}
}

func TestClientCanReadEntity(t *testing.T) {
	testClient := &hagateway.Client{
		Entities: []hagateway.Entity{
			{"binary_sensor.light_switch", false},
			{"binary_sensor.hvac_switch", false},
			{"sensor.*", false},
		},
	}
	cs := []struct {
		entityName string
		expected   bool
	}{
		{"binary_sensor.light_switch", true},
		{"sensor.temperature", true},
		{"sensor.light", true},
		{"binary_sensor.hvac_switch", true},
		{"binary_sensor.window_contact", false},
		{"random_thingy", false},
		{"", false},
	}

	for _, c := range cs {
		if r, e := testClient.CanReadEntity(c.entityName), c.expected; r != e {
			log.Printf(
				"Client.CanReadEntity(%s) returned %v, expected %v",
				c.entityName, r, e,
			)
			t.Fail()
		}
	}
}

func TestClientCanWriteEntity(t *testing.T) {
	testClient := &hagateway.Client{
		Entities: []hagateway.Entity{
			{"binary_sensor.light_switch", true},
			{"binary_sensor.hvac_switch", false},
			{"sensor.*", true},
		},
	}
	cs := []struct {
		entityName string
		expected   bool
	}{
		{"binary_sensor.light_switch", true},
		{"sensor.temperature", true},
		{"sensor.light", true},
		{"binary_sensor.hvac_switch", false},
		{"binary_sensor.window_contact", false},
		{"random_thingy", false},
		{"", false},
	}

	for _, c := range cs {
		if r, e := testClient.CanWriteEntity(c.entityName), c.expected; r != e {
			log.Printf(
				"Client.CanWriteEntity(%s) returned %v, expected %v",
				c.entityName, r, e,
			)
			t.Fail()
		}
	}
}

func TestClientCanWriteEntityDefaultYes(t *testing.T) {
	testClient := &hagateway.Client{
		CanWrite: true,
		Entities: []hagateway.Entity{
			{"binary_sensor.light_switch", true},
			{"binary_sensor.hvac_switch", false},
			{"sensor.*", false},
		},
	}
	cs := []struct {
		entityName string
		expected   bool
	}{
		{"binary_sensor.light_switch", true},
		{"sensor.temperature", true},
		{"sensor.light", true},
		{"binary_sensor.hvac_switch", true},
		{"binary_sensor.window_contact", false},
		{"random_thingy", false},
		{"", false},
	}

	for _, c := range cs {
		if r, e := testClient.CanWriteEntity(c.entityName), c.expected; r != e {
			log.Printf(
				"Client.CanWriteEntity(%s) returned %v, expected %v",
				c.entityName, r, e,
			)
			t.Fail()
		}
	}
}

func TestClients(t *testing.T) {
	ok := true
	clients := hagateway.Clients{}
	ok = ok && clients.Add(&hagateway.Client{Name: "A", TokenHash: "94ee059335e587e501cc4bf90613e0814f00a7b08bc7c648fd865a2af6a22cc2"})
	ok = ok && clients.Add(&hagateway.Client{Name: "B", TokenHash: "567dfae63c1555b57b277ab0fdaef8542ea760b55b56f4b6cf4b1a62ba821c97"})
	ok = ok && clients.Add(&hagateway.Client{Name: "C", TokenHash: "35335e0f21518cdee5d4300c930d9edd55018b4e8b8f5b80aac2f1dffdfac6d9"})
	ok = ok && clients.Add(&hagateway.Client{Name: "D", TokenHash: "bb7edb19fb19a0a455efb2c4d54957b394d8bcf246b478e51ceb2cb44328447f"})
	if !ok {
		t.Fatal("Expected all valid clients to be added")
	}
	if clients.Add(nil) {
		t.Fatal("Expected to block nil client")
	}
	if clients.Add(&hagateway.Client{TokenHash: "94ee059335e587e501cc4bf90613e0814f00a7b08bc7c648fd865a2af6a22cc2"}) {
		t.Fatal("Expected to block double client")
	}

	cs := []struct {
		token       string
		expected    string
		expectedNil bool
	}{
		{"TEST", "A", false},
		{"test", "", true},
		{"abc", "", true},
		{"HomeAssistant", "D", false},
	}

	for _, c := range cs {
		res := clients.FindByToken(c.token)
		if c.expectedNil {
			if res != nil {
				log.Printf("Expected to find no client but found %s", res.Name)
				t.Fail()
			}
		} else {
			if res == nil {
				log.Printf("Expected to find %s  but found no client", c.expected)
				t.Fail()
			} else if res.Name != c.expected {
				log.Printf("Expected to find %s but found %s", c.expected, res.Name)
				t.Fail()
			}
		}
	}
}
