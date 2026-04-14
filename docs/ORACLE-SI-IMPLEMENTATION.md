# Oracle SI Implementation (ERA-63369)

## Summary
Added support for Oracle Single Instance (SI) database provisioning and cloning in the NDB Operator, following the same patterns as MongoDB, MSSQL, MySQL, and PostgreSQL.

## Changes Made

### 1. Core Implementation Files

#### `ndb_api/interfaces.go`
- Added `OracleRequestAppender struct{}` implementing the `RequestAppender` interface

#### `ndb_api/db_helpers.go`
- Implemented `appendProvisioningRequest` for Oracle
- Default action arguments:
  - `listener_port: "1521"` (standard Oracle port)
  - `db_password`: from reqData
  - `database_names`: from Database spec
  - `auto_tune_staging_drive: "true"`

#### `ndb_api/clone_helpers.go`
- Implemented `appendCloningRequest` for Oracle
- Default action arguments:
  - `vm_name`: database name
  - `dbserver_description`: auto-generated
  - `db_password`: from reqData
- LCM config support enabled (expiry/refresh via `additionalArguments`)

#### `ndb_api/common_helpers.go`
- Updated `GetRequestAppender()` switch to include Oracle case
- Returns `OracleRequestAppender` for `DATABASE_TYPE_ORACLE`
- Updated error message to include "oracle" in supported types

### 2. Constants and Validation

#### `common/constants.go`
- Added `DATABASE_DEFAULT_PORT_ORACLE = 1521`
- Updated `DATABASE_TYPES = "mssql, mysql, postgres, mongodb, oracle"`

#### `common/util/additionalArguments.go`
- Added Oracle support in `GetAllowedAdditionalArgumentsForClone()`:
  - LCM fields: `expireInDays`, `expiryDateTimezone`, `deleteDatabase`
  - Refresh fields: `refreshInDays`, `refreshTime`, `refreshDateTimezone`
- Added Oracle support in `GetAllowedAdditionalArgumentsForDatabase()`:
  - `listener_port` (action argument)
- Updated function documentation to mention Oracle

### 3. Unit Tests

#### `ndb_api/db_helpers_test.go`
- Added test case for Oracle in `TestGetRequestAppenderByType`:
  ```go
  {databaseType: common.DATABASE_TYPE_ORACLE,
      expected: &OracleRequestAppender{},
  }
  ```

## Feature Parity with Other Engines

| Feature | Postgres | MySQL | MongoDB | MSSQL | Oracle |
|---------|----------|-------|---------|--------|--------|
| Provisioning | ✅ | ✅ | ✅ | ✅ | ✅ |
| Cloning | ✅ | ✅ | ✅ | ✅ | ✅ |
| LCM expiry | ✅ | ✅ | ✅ | ✅ | ✅ |
| LCM refresh | ✅ | ✅ | ✅ | ✅ | ✅ |
| Default port | 5432 | 3306 | 27017 | 1433 | 1521 |
| Software profile | Optional | Optional | Optional | Required | Required |

## Oracle-Specific Notes

1. **Closed-source engine**: Like MSSQL, Oracle requires a software profile to be explicitly provided (validated in `profile_helpers.go` lines 66-72).

2. **Listener port**: Default is 1521 (standard Oracle port), overridable via `additionalArguments`.

3. **Action arguments**: Minimal set (similar to MySQL/PostgreSQL) - no complex Windows-specific arguments like MSSQL.

4. **LCM support**: Full support for clone lifecycle management (expiry and refresh) via `additionalArguments`.

## Testing

### Unit tests
All Oracle-related unit tests pass with the existing test infrastructure.

### Manual testing required
1. **Provisioning**: Create Database CR with `type: oracle`, verify NDB provisioning request
2. **Cloning**: Create clone Database CR with `type: oracle`, verify clone request with LCM args
3. **Profile validation**: Verify software profile requirement is enforced for Oracle

## Example YAML

### Provisioning Oracle Database

```yaml
apiVersion: ndb.nutanix.com/v1alpha1
kind: Database
metadata:
  name: oracle-test-db
  namespace: default
spec:
  ndbRef: ndb-191
  isClone: false
  databaseInstance:
    type: oracle
    name: oracle-test-db
    databaseNames: ["testdb"]
    credentialSecret: db-secret-oracle
    clusterName: "production-cluster"
    timezone: "UTC"
    size: 50
    profiles:
      software:
        name: "ORACLE_19c_OOB"  # Required for Oracle (closed-source)
      compute:
        name: "DEFAULT_OOB_COMPUTE"
      network:
        name: "DEFAULT_OOB_ORACLE_NETWORK"
      dbParam:
        name: "DEFAULT_ORACLE_PARAMS"
    timeMachine:
      sla: "BRASS_SLA"
      dailySnapshotTime: "12:00:00"
      snapshotsPerDay: 1
    additionalArguments:
      listener_port: "1521"
```

### Cloning Oracle Database with LCM

```yaml
apiVersion: ndb.nutanix.com/v1alpha1
kind: Database
metadata:
  name: oracle-clone-test
  namespace: default
spec:
  ndbRef: ndb-191
  isClone: true
  clone:
    type: oracle
    name: oracle-clone-test
    credentialSecret: db-secret-oracle
    sourceDatabaseName: oracle-test-db
    clusterName: "production-cluster"
    timezone: "UTC"
    profiles:
      software:
        name: "ORACLE_19c_OOB"
      compute:
        name: "DEFAULT_OOB_COMPUTE"
      network:
        name: "DEFAULT_OOB_ORACLE_NETWORK"
      dbParam:
        name: "DEFAULT_ORACLE_PARAMS"
    additionalArguments:
      expireInDays: "7"
      expiryDateTimezone: "UTC"
      deleteDatabase: "true"
```

## Files Modified

```
 common/constants.go                |  3 ++-
 common/util/additionalArguments.go | 16 +++++++++++++++-
 ndb_api/clone_helpers.go           | 29 +++++++++++++++++++++++++++++
 ndb_api/common_helpers.go          |  4 +++-
 ndb_api/db_helpers.go              | 25 +++++++++++++++++++++++++
 ndb_api/db_helpers_test.go         |  3 +++
 ndb_api/interfaces.go              |  3 +++
 7 files changed, 80 insertions(+), 3 deletions(-)
```

## Commit Message

```
feat: Add Oracle SI support for provisioning and cloning (ERA-63369)

Implement Oracle Single Instance database support following the same
patterns as MongoDB, MSSQL, MySQL, and PostgreSQL:

- Add OracleRequestAppender with provision and clone request builders
- Set default listener port to 1521 (standard Oracle port)
- Enable LCM lifecycle management for Oracle clones (expiry/refresh)
- Update constants, validation, and unit tests for Oracle
- Oracle requires explicit software profile (closed-source engine)

Closes ERA-63369
```

## Related Jira

**ERA-63369**: Add Oracle SI support to NDB Operator

**Scope**: Oracle Single Instance provisioning and cloning on NDB, matching feature parity with existing supported database engines.
