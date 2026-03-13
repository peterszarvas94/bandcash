package auth

import (
	"fmt"
	"testing"
	"time"
)

func TestLoginThrottleBlocksEmailCooldown(t *testing.T) {
	throttle := newLoginThrottle()
	now := time.Unix(1_700_000_000, 0)

	if !throttle.allow("10.0.0.1", "user@example.com", now) {
		t.Fatal("expected first request to pass")
	}

	if throttle.allow("10.0.0.2", "user@example.com", now.Add(10*time.Second)) {
		t.Fatal("expected second request within cooldown to be blocked")
	}

	if !throttle.allow("10.0.0.2", "user@example.com", now.Add(loginEmailCooldown+time.Second)) {
		t.Fatal("expected request after cooldown to pass")
	}
}

func TestLoginThrottleBlocksIPEmailLimit(t *testing.T) {
	throttle := newLoginThrottle()
	base := time.Unix(1_700_000_000, 0)
	email := "user@example.com"
	ip := "10.0.0.1"

	for i := 0; i < loginIPEmailLimit; i++ {
		now := base.Add(time.Duration(i) * (loginEmailCooldown + time.Second))
		if !throttle.allow(ip, email, now) {
			t.Fatalf("expected request %d to pass", i+1)
		}
	}

	blockedAt := base.Add(time.Duration(loginIPEmailLimit) * (loginEmailCooldown + time.Second))
	if throttle.allow(ip, email, blockedAt) {
		t.Fatal("expected request above ip+email limit to be blocked")
	}
}

func TestLoginThrottleBlocksEmailGlobalLimit(t *testing.T) {
	throttle := newLoginThrottle()
	base := time.Unix(1_700_000_000, 0)
	email := "user@example.com"

	for i := 0; i < loginEmailGlobalLimit; i++ {
		now := base.Add(time.Duration(i) * (loginEmailCooldown + time.Second))
		ip := fmt.Sprintf("10.0.0.%d", i+1)
		if !throttle.allow(ip, email, now) {
			t.Fatalf("expected request %d to pass", i+1)
		}
	}

	blockedAt := base.Add(time.Duration(loginEmailGlobalLimit) * (loginEmailCooldown + time.Second))
	if throttle.allow("10.0.0.99", email, blockedAt) {
		t.Fatal("expected request above email global limit to be blocked")
	}
}
