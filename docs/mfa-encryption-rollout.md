# MFA secret encryption rollout

Goal: encrypt every stored TOTP secret (`mfa_setups.secret_key`) with
AES-256-GCM (`enc:v1:...`, bound to the user id) in each tenant database.

## How it behaves

- With no key set, secrets are stored and read as plaintext (current prod and training state).
- With `MFA_ENCRYPTION_KEY` set, new secrets are written encrypted. Existing
  plaintext rows (no `enc:` prefix) are still read as-is, so setting the key
  does not lock anyone out. Existing rows stay plaintext until migrated.
- An encrypted row that no configured key opens makes that user's login fail
  closed (no bypass, no data change).
- Email-MFA rows have an empty secret and are skipped.
- Backup/recovery codes are not persisted by auth-service (only returned in a
  DTO), so there is nothing else to migrate.

## Command

    auth migrate encrypt-mfa-secrets          # dry run: prints counts only
    auth migrate encrypt-mfa-secrets --apply  # writes

- Dry run is the default. Output has counts only (total, empty,
  already_encrypted, legacy_encrypted = plaintext to encrypt, rotated, unreadable), never secrets.
- One transaction with the rows locked. Each rewritten row is read back,
  decrypted and compared to the original before commit; any mismatch rolls back everything.
- If any row is unreadable with the configured keys (wrong key), the whole
  run rolls back and nothing is changed.
- Idempotent: a second run finds only already-encrypted rows and changes nothing.
- Tenant selection: it works on the database that auth-service's `DB_*`
  settings point to (log field `database`). Run it once per tenant, from that
  tenant's auth container, so env and database always match. Each tenant gets
  its own key; do not reuse the key across tenants. No seeders run.

## Rollout order (training first, then prod)

1. Generate a key on the server, not in git, not in chat:
   `openssl rand -base64 32`
2. Put it in the stack's server-side env file (the same untracked file that
   holds the other secrets) as `MFA_ENCRYPTION_KEY=...`. Make the file
   root-only. Back the key up in the password manager now (see "Key lost").
3. Make sure the auth-service image contains this command (deploy with the
   normal per-service steps; do not use a path that runs seeders on prod).
4. Restart auth-service so it picks up the key. Check the startup log no
   longer prints the "MFA_ENCRYPTION_KEY is not set" warning.
5. Check a real TOTP login still works (plaintext rows must still work).
6. Take a database backup of `mfa_setups` (or the whole auth DB).
7. Dry run in the auth container: `auth migrate encrypt-mfa-secrets`.
   Expect `unreadable=0` and `legacy_encrypted` equal to the number of TOTP users.
8. Apply: `auth migrate encrypt-mfa-secrets --apply`. Run the dry run again;
   expect `legacy_encrypted=0`.
9. Verify a real login with a TOTP code for a migrated user, plus a new MFA enrollment.
10. Do all of the above on training first. Only after step 9 passes there, repeat on prod.

## Key lost or changed wrongly

- Encrypted secrets cannot be recovered without the key. Affected users
  cannot pass TOTP and must re-enroll MFA (an admin resets their MFA, then
  they enroll again). Keep the key in the password manager and in the server env file.
- Never replace the key by overwriting it: for rotation set the new key as
  `MFA_ENCRYPTION_KEY`, move the previous one to `MFA_ENCRYPTION_KEY_OLD`,
  restart, run `--apply` (re-encrypts old-key rows), then remove the old key.

## Rollback

- Before the migration (steps 1-5): remove the key and restart. Nothing was
  rewritten, so nothing changes.
- After the migration: do not remove the key, since encrypted rows would
  become unreadable and fail closed. Restore the step 6 backup of
  `mfa_setups` (rows written before the migration are plaintext and work
  without a key). Any MFA enrollments made after the backup must be redone.

## Caveats

- If `APP_ENV` is `production`/`prod` or `MFA_ENCRYPTION_REQUIRED=true`,
  auth-service refuses to start without a key. Prod currently starts without
  one, so check which of these is set before restarting.
- Deployment and env files belong to devops; this doc does not change them.
