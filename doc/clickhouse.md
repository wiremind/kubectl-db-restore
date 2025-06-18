# 📚 ClickHouse Guide for `kubectl-restore`

This guide explains how to restore ClickHouse databases using the `kubectl restore` plugin.

---

## ✅ Required Flags

| Flag             | Description                                  |
|------------------|----------------------------------------------|
| `--engine`       | Must be `clickhouse`                         |
| `--backup-name`  | Name of the backup to restore from           |
| `--database`     | Target database name                         |
| `--service-name` | K8s service pointing to the ClickHouse pods  |
| `--namespace`    | Kubernetes namespace (default: `default`)    |

---

## 🔐 Required Variables

The following **must be defined** either via environment variables **or** via `--secret-ref`:

- `CLICKHOUSE_USER`
- `CLICKHOUSE_PASSWORD`
- `CLICKHOUSE_AWS_S3_ENDPOINT_URL_BACKUP`
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`

### Via `--secret-ref` Example

```
kubectl restore database \
  --engine clickhouse \
  --backup-name my-daily-backup \
  --database example_db \
  --namespace analytics \
  --service-name clickhouse-service \
  --secret-ref CLICKHOUSE_USER=clickhouse-secrets:user \
  --secret-ref CLICKHOUSE_PASSWORD=clickhouse-secrets:password \
  --secret-ref CLICKHOUSE_AWS_S3_ENDPOINT_URL_BACKUP=s3-secrets:endpoint \
  --secret-ref AWS_ACCESS_KEY_ID=aws-secrets:access \
  --secret-ref AWS_SECRET_ACCESS_KEY=aws-secrets:secret
```

## 🧪 Dry Run Mode

To preview the SQL that would be executed:

```
kubectl restore database ... --dry-run
```
This logs the SQL restore process without creating a Job.

## 🔄 Job Lifecycle

The restore consists of three sequential Kubernetes Jobs:
    1. Drop existing DB (if exists)
    2. Create the DB fresh
    3. Run RESTORE SQL from S3

Each step uses the `clickhouse-client` tool within a container (`clickhouse/clickhouse-server:25.5-alpine`).

## 🧩 Developer Notes

The ClickHouse engine implementation lives under:

```
pkg/engine/clickhouse.go
```
You can add new engines by implementing the Engine interface and calling RegisterEngine.