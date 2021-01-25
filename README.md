# go-os-brick: go-os-brick is Go port of [os-brick](https://github.com/openstack/os-brick)

[![Go Reference](https://pkg.go.dev/badge/github.com/lovi-cloud/go-os-brick.svg)](https://pkg.go.dev/github.com/lovi-cloud/go-os-brick)

## Usage

go-os-brick provide function that connect / disconnect iSCSI volume.

- for multi path
    - `ConnectMultiPathVolume()`
    - `DisconnectVolume()`
- for single path
    - `ConnectSinglePathVolume()`
    - `DisconnectSinglePathVolume()`

### Prepare

go-os-brick execute a some commands. please install before use.

- `iscsiadm(8)`
- `blockdev(8)`
- `qemu-img(1)`
    - if you use `QEMUToRaw()`
- `multipath(8)`
    - if you use multi path

## Testing

### using gostor/gotgt

backend is [gostor/gotgt](https://github.com/gostor/gotgt) via goroutine.
This test need some kernel modules.

```
$ sudo go test -v ./...
```

### using open-iscsi targetd in a host machine

**WARNING: DO NOT EXECUTE YOUR WORKSPACE!!**

This test execute script in a host.

backend is open-iscsi targetd in a host machine.
This test need some kernel modules.

```
$ sudo go test -tags=host -v ./...
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