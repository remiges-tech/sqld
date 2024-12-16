# sqlc-gen-tablename Plugin

This plugin generates `TableName()` methods for SQLC generated Go structs. These methods return the name of the database table corresponding to the struct. The plugin was specifically developed to meet the requirements of the `sqld` project, which needs table name information for its dynamic SQL query building and execution.

## Building the Plugin

Clone the repository and build the plugin executable:

```bash
git clone https://github.com/remiges-tech/sqld.git
cd sqld/sqlc-gen-tablename
go build
```

This will create the sqlc-gen-tablename executable in the current directory. Wherever you keep this executable will be ${SQLC_PLUGIN_PATH} in your `sqlc.yaml` file. So, if you have kept it in `/usr/local/bin`, your `sqlc.yaml` file will look like this:

```yaml
plugins:
  - name: sqlc-gen-tablename
    process:
      cmd: /usr/local/bin/sqlc-gen-tablename
```

Check an example in `test/sqlc.yaml`