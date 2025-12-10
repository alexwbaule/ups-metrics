# ups-metrics
Get metrics from SMS Brazil Wifi UPS

[![Latest Release](https://github.com/alexwbaule/ups-metrics/actions/workflows/build-release-binaries.yml/badge.svg?branch=main)](https://github.com/alexwbaule/ups-metrics/actions/workflows/build-release-binaries.yml)

## Configuration

### Log Destinations

The application supports multiple log destinations through an explicit switch mechanism. You **must** specify the log type in your configuration.

#### Supported Log Types

- `gelf` - Send logs to Graylog using GELF protocol
- `victorialogs` - Send logs to VictoriaLogs using HTTP API

#### Configuration Structure

```yaml
logs:
  type: "gelf"  # REQUIRED: Must be "gelf" or "victorialogs"
  gelf:
    address: graylog
    port: 12201
  victorialogs:
    address: victoria-logs
    port: 9428
    username: admin
    password: secret
    timeout: 30s
```

#### Switch Behavior

- **`type: "gelf"`** - Uses only Graylog configuration, ignores VictoriaLogs settings
- **`type: "victorialogs"`** - Uses only VictoriaLogs configuration, ignores Graylog settings
- **Missing or invalid `type`** - Application will fail to start with a clear error message

### VictoriaLogs Integration

VictoriaLogs is a high-performance log storage system that provides:

- Fast log ingestion and querying
- Efficient storage compression
- HTTP API for log submission
- Structured JSON log format

#### VictoriaLogs Configuration Parameters

| Parameter | Required | Description | Example |
|-----------|----------|-------------|---------|
| `address` | Yes | VictoriaLogs server hostname/IP | `victoria-logs` |
| `port` | Yes | VictoriaLogs HTTP API port | `9428` |
| `username` | No | Authentication username | `admin` |
| `password` | No | Authentication password | `secret` |
| `timeout` | No | HTTP request timeout | `30s` |

#### Log Format

Logs sent to VictoriaLogs are formatted as structured JSON with the following fields:

```json
{
  "timestamp": "2023-12-10T15:30:45Z",
  "level": "info",
  "message": "UPS notification processed",
  "source": "ups-metrics",
  "metadata": {
    "device_id": "ups-001",
    "notification_id": "12345",
    "battery_level": 85
  }
}
```

### Migration Guide

#### From Graylog-only to Explicit Switch

If you have an existing configuration that only uses Graylog:

**Before:**
```yaml
logs:
  gelf:
    address: graylog
    port: 12201
```

**After:**
```yaml
logs:
  type: "gelf"  # Add this required field
  gelf:
    address: graylog
    port: 12201
```

#### Migrating to VictoriaLogs

To switch from Graylog to VictoriaLogs:

1. Update your configuration:
```yaml
logs:
  type: "victorialogs"  # Change from "gelf" to "victorialogs"
  victorialogs:
    address: your-victoria-logs-server
    port: 9428
    username: your-username  # Optional
    password: your-password  # Optional
    timeout: 30s
```

2. Ensure VictoriaLogs is running and accessible
3. Restart the ups-metrics service

#### Error Handling

The application will fail to start with descriptive error messages in these cases:

- Missing `logs.type` field
- Invalid `logs.type` value (not "gelf" or "victorialogs")
- Missing destination configuration for the specified type
- Invalid destination configuration (missing required fields)

### Complete Configuration Example

See `conf/config.sample.yaml` for a complete configuration example with all supported options.
