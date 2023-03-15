# backup-restore-tool

Backup Restore Tool is a command line application that perform backup or restore of Postgres database, downloading or uploading backup to S3 compatible storage.
The Backup Restore Tool can also be run as a Docker container.

## Usage

To backup the database and save it in Amazon S3, run:
```bash
backup-restore-tool backup -d "[DB_CONNECTION_STRING]" --storage-object-key=my-backup --storage-bucket=buckups \
--storage-access-key "[AWS_KEY_ID]" --storage-secret-key "[AWS_SECRET_KEY]" --storage-region us-east-1
```

To restore the database, run:
```bash
backup-restore-tool restore -d "[DB_CONNECTION_STRING]" --storage-object-key=my-backup --storage-bucket=buckups \
--storage-access-key "[AWS_KEY_ID]" --storage-secret-key "[AWS_SECRET_KEY]" --storage-region us-east-1
```

For detailed description of available parameters, run:
```bash
backup-restore-tool [COMMAND] -h
```

All options can be provided as environment variables. Make sure to replace `-` with `_`, capitalize option name and add `BRT` prefix:
```
BRT_[CAPITALIZED_OPTION_NAME]
```

## Development

Install the tool locally:
```bash
make install
```

Build docker image:
```bash
make build-image
```
