## Configuration via JSON files

This folder contains example configuration files for the server and the agent.

### Examples

- Server: `config/server.json`

`
{
    "connection": {
        "host": "localhost:8080"
    },
    "repository": {
        "fs": {
            "store_interval": 300,
            "file_storage_path": "./test.db",
            "restore": true
        },
        "db": {
            "database_dsn": ""
        }
    },
    "sign": {
        "key": "test_key"
    },
    "audit": {
        "audit_file": "./test.jsonl",
        "audit_url": ""
    },
    "encryption": {
        "crypto_key": "./test.key"
    },
    "log_level": "debug"
}
`

- Agent: `config/agent.json`

`
{
    "connection": {
        "host": "localhost:8080"
    },
    "agent": {
        "poll_interval": 2,
        "report_interval": 5,
        "enable_gzip": true,
        "enable_get_metrics": false,
        "rate_limit": 10
    },
    "sign": {
        "key": "test_key"
    },
    "encryption": {
        "crypto_key": "./test.key"
    },
    "log_level": "info"
}
`