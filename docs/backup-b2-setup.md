# Offsite backups to Backblaze B2 (VPS runbook)

Runs `backend/scripts/backup/backup-prod.sh` nightly on the production VPS and copies the
Postgres dump and uploads archive to a private Backblaze B2 bucket with rclone.

Status: DRAFT. Not installed on the VPS. The cron entry is not added yet.

## Fixed values

| Item | Value |
|---|---|
| Bucket | `auditsphere-prod-backups` (private, dedicated to backups) |
| rclone remote name | `b2` |
| `RCLONE_REMOTE` | `b2:auditsphere-prod-backups` |
| Application key scope | this bucket only; capabilities `listBuckets`, `listFiles`, `writeFiles`, `readFiles` (plus `listAllBucketNames` if the console requires it for the bucket picker) |
| Lifecycle rule | delete after 30 days (see "Lifecycle" below) |

Placeholders used below (never write real values into this file or into git):

- `<KEY_ID>`: the application key ID (Backblaze shows it on the Application Keys page)
- `<APP_KEY>`: the application key secret (shown only once at creation)
- `<REPO_PATH>`: path to the checked-out repo on the VPS (verify with `ls`, e.g. `/opt/risk-based-audit`)

## 1. Install rclone on the VPS

Check first whether a distro package is recent enough. Any recent version works.

```bash
rclone version 2>/dev/null || echo "rclone not installed"
curl -fsSL https://rclone.org/install.sh | bash
# The VPS is AlmaLinux 9 (no apt-get). The install script works on RHEL-family hosts.
rclone version
```

If the apt version is too old, use the official installer from https://rclone.org/install/
(review the script before running it).

Verify the install ran as root (cron runs as root, so the config must be root's).

## 2. Create the remote config (interactive, keys never in shell history)

Run as root on the VPS:

```bash
rclone config
```

Answer the prompts:

1. `n` (new remote)
2. name: `b2`
3. storage: `b2` (Backblaze B2)
4. `account`: `<KEY_ID>` (the application key ID, not the master account ID)
5. `key`: `<APP_KEY>`
6. `endpoint`: leave blank
7. Accept the remaining defaults, then `q` to quit.

Lock down the config file and confirm the remote is visible without printing secrets:

```bash
chmod 600 "$(rclone config file | tail -n1)"
rclone listremotes          # should print: b2:
```

Do not run `rclone config show` in a shared terminal log. It prints the key.

## 3. Test the remote (before the first cron run)

Use a small non-sensitive file, then delete it:

```bash
echo "rclone b2 connectivity test $(date -u +%FT%TZ)" > /tmp/b2-test.txt
rclone copy /tmp/b2-test.txt b2:auditsphere-prod-backups/connectivity-test/ --checksum
rclone ls b2:auditsphere-prod-backups/connectivity-test/
rclone delete b2:auditsphere-prod-backups/connectivity-test/b2-test.txt
rm -f /tmp/b2-test.txt
```

A `403` or `401` here usually means the key is missing a capability, is scoped to a
different bucket, or the `account` field holds the master account ID.

## 4. Run one manual backup (not from cron)

From the repo on the VPS, with the offsite remote set:

```bash
cd <REPO_PATH>
RCLONE_REMOTE=b2:auditsphere-prod-backups backend/scripts/backup/backup-prod.sh
```

Expected: `pg-*.sql.gz`, `uploads-*.tar.gz` and their `.sha256` files appear in
`/root/backups/auditsphere` and in the bucket. Confirm:

```bash
rclone ls b2:auditsphere-prod-backups
```

## 5. Cron line (add only after steps 1-4 pass and the team approves)

Runs nightly at 02:30 server time. The script exits non-zero if the offsite copy fails,
so check the log, not just cron.

```cron
30 2 * * * cd <REPO_PATH> && RCLONE_REMOTE=b2:auditsphere-prod-backups backend/scripts/backup/backup-prod.sh >> /var/log/auditsphere-backup.log 2>&1
```

Add with `crontab -e` as root. Create the log file first with `touch /var/log/auditsphere-backup.log`.

## 6. Restore test (run monthly, and after any change to the backup script)

This pulls the newest offsite copy into a temp directory and restores it into a throwaway
Postgres container. It never touches `rb_audit_postgres` or `audit_uploads`.

```bash
RESTORE_DIR="$(mktemp -d /root/restore-check.XXXXXX)"
rclone copy b2:auditsphere-prod-backups "$RESTORE_DIR" --include "pg-*" --include "uploads-*" --checksum
NEWEST_PG="$(ls -1t "$RESTORE_DIR"/pg-*.sql.gz | head -n1)"
NEWEST_UP="$(ls -1t "$RESTORE_DIR"/uploads-*.tar.gz | head -n1)"
cd <REPO_PATH>
BACKUP_DIR="$RESTORE_DIR" backend/scripts/backup/restore-test.sh "$NEWEST_PG" "$NEWEST_UP"
rm -rf "$RESTORE_DIR"
```

Expected final line: `PASS: backup is restorable`.

## Lifecycle (retention in the bucket)

Set in the Backblaze web console, not by the script (`RETENTION_DAYS` only controls local copies).

Console path (verify the menu names against the live console):

1. Backblaze web console > B2 Cloud Storage > Buckets > `auditsphere-prod-backups`.
2. Lifecycle Settings > Custom rule (leave the file name prefix empty so the rule covers the whole bucket).
3. "Days from uploading to hiding": `30`. "Days from hiding to deleting": `1`.
4. Save.

Effective result: files are hidden 30 days after upload and deleted one day later
(about 31 days of restore points).

## Notes

- Keep `*.sha256` alongside the archives. `restore-test.sh` checks them.
- The app key secret is shown once. Store it in the team password manager, not in git or docs.
- Rotate the application key if it leaks. Create a new key with the same bucket scope,
  run `rclone config` again to update the `b2` remote, delete the old key in the console.
