# kubectl-db-restore

`kubectl-db-restore` is a [Krew](https://krew.sigs.k8s.io/) plugin for restoring databases running in Kubernetes, directly from your terminal using `kubectl`.

This plugin supports cloud-native database restoration workflows via Kubernetes Jobs. It is designed to be engine-extensible (currently supports ClickHouse and a placeholder for PostgreSQL) and integrates seamlessly into Kubernetes-native workflows.

---

## 🔧 Features

- ⚡ Fast and CLI-native
- 📦 Restore databases from S3-compatible backups
- 🔁 Extensible engine system (e.g., ClickHouse, PostgreSQL)
- 🧪 Dry-run support
- 🔐 Secret-based credential resolution from Kubernetes Secret
- 🛠️ Runs restore commands as Kubernetes Jobs

---

## 📦 Installation (via Krew)

Ensure [Krew](https://krew.sigs.k8s.io/docs/user-guide/setup/install/) is installed:

```sh
kubectl krew install restore
```

Then invoke the plugin using:

```
kubectl db-restore
```

---

🧠 Supported Engines
Engine	Status
[ClickHouse](doc/clickhouse.md)	✅ Fully Supported
PostgreSQL	⚠️ Not yet implemented

🚀 Example

```
export CLICKHOUSE_AWS_S3_ENDPOINT_URL_BACKUP="https://backup-name.s3.region-name.amazonaws.com"

kubectl db-restore database \
  --engine clickhouse \
  --backup-name daily-backup-2025-06-16 \
  --database my_db \
  --namespace analytics \
  --service-name clickhouse-svc \
  --secret-ref CLICKHOUSE_USER=clickhouse-secrets:user \
  --secret-ref CLICKHOUSE_PASSWORD=clickhouse-secrets:password \
  --secret-ref AWS_ACCESS_KEY_ID=aws-creds:access-key-id \
  --secret-ref AWS_SECRET_ACCESS_KEY=aws-creds:secret-access-key \
```

These credentials are used to formulate the restore SQL and run it via clickhouse-client inside a Kubernetes Job.

🧪 Dry Run Mode

You can preview the generated SQL without actually executing it by passing the --dry-run flag:

```
kubectl db-restore database \
  --dry-run \
  --engine clickhouse ...
```

## 🧩 Contributing

Pull requests are welcome! Feel free to open issues or feature requests.
