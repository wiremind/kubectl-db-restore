# Usage Guide: kubectl-restore

This document covers how to use the `kubectl restore` plugin effectively within your Kubernetes environments

## Usage
The following assumes you have the plugin installed via

```shell
kubectl krew install kubectl-restore
```

## ⚙️ Command Overview

```
kubectl restore database [flags]
```

This command launches a Kubernetes Job that runs a database restore process. It is currently available for ClickHouse, with support for other engines planned.

### 📌 Required Flags
Flag	Description
--engine	Database engine name (e.g., clickhouse)
--backup-name	Name of the backup to restore from
--database	Name of the database to restore
--service-name	K8s service name of the DB (for the Job)
--namespace	Kubernetes namespace (default: default)

All of the above flags are required except for --namespace (which defaults to default).

### 🧪 Optional Flags
Flag	Description
--dry-run	Print the SQL query and exit

🧾 Example

```
kubectl restore database \
  --engine clickhouse \
  --backup-name my-daily-backup \
  --database example_db \
  --namespace my-ns \
  --service-name clickhouse-service
```
To simulate the job without actually running it:

```
kubectl restore database ... --dry-run
```

### 🧠 Job Lifecycle & Monitoring

The plugin will:

Create a Kubernetes Job with the restore command.

Wait for the Job to succeed or fail.

Output logs and instructions on failure.

In case of failure, helpful logs and error messages will be printed. You can inspect the job manually using:

```
kubectl get jobs -n <namespace>
kubectl logs job/<job-name> -n <namespace>
```

### 🧩 Extensibility
New engines can be added by implementing the Engine interface in Go and registering it via RegisterEngine.

### 🆘 Troubleshooting
Missing environment variable error: Ensure all required variables are exported before running.

Job failed unexpectedly: Use --dry-run to debug the restore SQL.

Job not created: Make sure you have sufficient RBAC permissions to create Jobs in the namespace.

## 👥 Community & Support
File issues and discuss improvements via GitHub: wiremind/kubectl-restore
