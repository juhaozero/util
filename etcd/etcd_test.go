package etcd

import (
	"testing"
	"time"
)

func TestConfigWithDefaults(t *testing.T) {
	cfg := Config{}.withDefaults()
	if cfg.DialTimeout != 5*time.Second {
		t.Fatalf("DialTimeout = %v, want 5s", cfg.DialTimeout)
	}

	custom := Config{DialTimeout: time.Second}.withDefaults()
	if custom.DialTimeout != time.Second {
		t.Fatalf("custom DialTimeout overwritten")
	}
}

func TestServiceEventType(t *testing.T) {
	if ServicePut == ServiceDelete {
		t.Fatal("service event types should differ")
	}
}

func TestEtcdClosed(t *testing.T) {
	e := &Etcd{closed: true}
	if err := e.ensureOpen(); err != ErrClosed {
		t.Fatalf("ensureOpen() = %v, want ErrClosed", err)
	}
}

func TestGetServiceListEmpty(t *testing.T) {
	e := &Etcd{discovery: map[string]string{}}
	list := e.GetServiceList()
	if len(list) != 0 {
		t.Fatalf("expected empty service list")
	}
}

func TestSetAndDeleteServiceList(t *testing.T) {
	e := &Etcd{discovery: map[string]string{}}

	var events []ServiceEvent
	e.onChange = func(event ServiceEvent) {
		events = append(events, event)
	}

	e.setServiceList("/services/a", "127.0.0.1:8080")
	e.delServiceList("/services/a")

	if len(events) != 2 {
		t.Fatalf("events = %d, want 2", len(events))
	}
	if events[0].Type != ServicePut || events[0].Value != "127.0.0.1:8080" {
		t.Fatalf("unexpected put event: %+v", events[0])
	}
	if events[1].Type != ServiceDelete {
		t.Fatalf("unexpected delete event: %+v", events[1])
	}

	list := e.GetServiceList()
	if len(list) != 0 {
		t.Fatalf("service list should be empty after delete")
	}
}

func TestServiceRegisterUpsert(t *testing.T) {
	sr := &ServiceRegister{services: make([]serviceEntry, 0)}
	sr.upsertService("k1", "v1")
	sr.upsertService("k1", "v2")

	if len(sr.services) != 1 {
		t.Fatalf("services = %d, want 1", len(sr.services))
	}
	if sr.services[0].val != "v2" {
		t.Fatalf("value = %s, want v2", sr.services[0].val)
	}
}

func TestRegistrarClosed(t *testing.T) {
	sr := &ServiceRegister{closed: true}
	if err := sr.PutService("k", "v"); err != ErrRegistrarClosed {
		t.Fatalf("PutService() = %v, want ErrRegistrarClosed", err)
	}
}
