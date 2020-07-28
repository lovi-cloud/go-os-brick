# go-os-brick

go-os-brick is Go port of [os-brick](https://github.com/openstack/os-brick)

## Testing

### using gostor/gotgt

backend is [gostor/gotgt](https://github.com/gostor/gotgt) via goroutine.
This test need some kernel modules.

```
$ sudo go test -v ./...
```

### using open-iscsi targetd in a docker container

backend is Docker container contain open-iscsi targetd.
This test need some kernel modules.

please see detail in `test/docker/Dockerfile`.

```
$ sudo go test -tags=container -v ./...
```

### using real iSCSI target endpoint

you can use real iSCSI target as backend for testing.

please set environment value

- `OS_BRICK_TEST_PORTAL_ADDRESS`
- `OS_BRICK_TEST_TARGET_IQN`

```
$ export OS_BRICK_TEST_PORTAL_ADDRESS="192.0.2.1"
$ export ...
$ sudo go test -v ./...
```