package store

import (
    "context"
    "fmt"
    "os"
    "testing"
    "time"
)

func TestSessionRevocationAndMembershipIntegration(t *testing.T) {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" { t.Skip("DATABASE_URL not set for PostgreSQL tests") }
    ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
    defer cancel()

    s, err := Open(ctx, dsn)
    if err != nil { t.Fatal(err) }
    defer s.Close()
    slug := fmt.Sprintf("sessions-%d", time.Now().UnixNano())
    email := slug + "@example.invalid"
    orgID, err := s.BootstrapV2(ctx, "Sessions Test", slug, "Test Admin", email, "test-secure-password-2026", false)
    if err != nil { t.Fatal(err) }
    identity, err := s.LoginIdentity(ctx, email, slug)
    if err != nil { t.Fatal(err) }
    defer func() {
        _, _ = s.pool.Exec(context.Background(), `DELETE FROM organizations WHERE id=$1::uuid`, orgID)
        _, _ = s.pool.Exec(context.Background(), `DELETE FROM users WHERE id=$1::uuid`, identity.UserID)
    }()

    first := SessionDigest("signed-jwt-test-one")
    second := SessionDigest("signed-jwt-test-two")
    if first == second || len(first) != 64 { t.Fatal("different bearer tokens require distinct SHA-256 hashes") }

    active, err := s.SessionActive(ctx, first, orgID, identity.UserID, "admin")
    if err != nil || !active { t.Fatalf("new token should be active: active=%t err=%v", active, err) }
    if err := s.RevokeSession(ctx, first, orgID, identity.UserID, time.Now().Add(6*time.Hour)); err != nil { t.Fatal(err) }
    if err := s.RevokeSession(ctx, first, orgID, identity.UserID, time.Now().Add(6*time.Hour)); err != nil { t.Fatalf("repeated logout must be idempotent: %v", err) }
    active, err = s.SessionActive(ctx, first, orgID, identity.UserID, "admin")
    if err != nil || active { t.Fatalf("revoked token was accepted: active=%t err=%v", active, err) }
    active, err = s.SessionActive(ctx, second, orgID, identity.UserID, "admin")
    if err != nil || !active { t.Fatalf("logging out one device should not revoke another: active=%t err=%v", active, err) }

    // A role change invalidates already issued claims immediately.
    _, err = s.pool.Exec(ctx, `UPDATE organization_memberships SET role='manager',updated_at=now() WHERE organization_id=$1::uuid AND user_id=$2::uuid`, orgID, identity.UserID)
    if err != nil { t.Fatal(err) }
    active, err = s.SessionActive(ctx, second, orgID, identity.UserID, "admin")
    if err != nil || active { t.Fatalf("stale admin claim retained access: active=%t err=%v", active, err) }
    active, err = s.SessionActive(ctx, second, orgID, identity.UserID, "manager")
    if err != nil || !active { t.Fatalf("current manager role should be valid: active=%t err=%v", active, err) }

    // A membership suspension and user deactivation both take effect immediately.
    _, err = s.pool.Exec(ctx, `UPDATE organization_memberships SET status='disabled' WHERE organization_id=$1::uuid AND user_id=$2::uuid`, orgID, identity.UserID)
    if err != nil { t.Fatal(err) }
    active, err = s.SessionActive(ctx, second, orgID, identity.UserID, "manager")
    if err != nil || active { t.Fatalf("disabled membership was accepted: active=%t err=%v", active, err) }
    _, err = s.pool.Exec(ctx, `UPDATE organization_memberships SET status='active' WHERE organization_id=$1::uuid AND user_id=$2::uuid`, orgID, identity.UserID)
    if err != nil { t.Fatal(err) }
    _, err = s.pool.Exec(ctx, `UPDATE users SET status='disabled' WHERE id=$1::uuid`, identity.UserID)
    if err != nil { t.Fatal(err) }
    active, err = s.SessionActive(ctx, second, orgID, identity.UserID, "manager")
    if err != nil || active { t.Fatalf("disabled user was accepted: active=%t err=%v", active, err) }
}
