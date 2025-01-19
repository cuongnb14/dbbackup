# Usage

**List ENV**

```shell
PG_HOST: db host
PG_PORT: db port
PG_USER: db user
PG_PASSWORD: db password
PG_SSLMODE: ssl mode

ENCRYPT_KEY: key to encrypt backup file

# Azure Blob Storage for remote backup
ABS_URL: abs url
ABS_CONTAINER: abs container
ABS_SAS: abs token
```

# Run now

```sh
docker exec -it dbbackup backup --now
```
