package testutils

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	// import driver
	_ "github.com/gostor/gotgt/pkg/port/iscsit"
	_ "github.com/gostor/gotgt/pkg/scsi/backingstore"

	"github.com/gostor/gotgt/pkg/config"
	"github.com/gostor/gotgt/pkg/scsi"
	uuid "github.com/satori/go.uuid"
)

type dummyBS struct{}

func (d dummyBS) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, nil
}

func (d dummyBS) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, nil
}

func (d dummyBS) Sync() (int, error) {
	return 0, nil
}

func (d dummyBS) Unmap(i int64, i2 int64) (int, error) {
	return 0, nil
}

func GetTestTarget() (string, string, func(), error) {
	tmpFile, err := ioutil.TempFile("", "gotgt")
	if err != nil {
		return "", "", nil, err
	}

	base := filepath.Base(tmpFile.Name())
	portalIP := "127.0.0.1:3260"
	tgtName := fmt.Sprintf("iqn.2016-09.com.gotgt.gostor:%s", base)
	lhbsName := fmt.Sprintf("file:%s", tmpFile.Name())

	id := uuid.NewV4()
	uid := binary.BigEndian.Uint64(id[:8])
	err = scsi.InitSCSILUMapEx(&config.BackendStorage{
		DeviceID:         uid,
		Path:             lhbsName,
		Online:           true,
		ThinProvisioning: true,
	}, tgtName, uint64(0), dummyBS{})
	if err != nil {
		return "", "", nil, err
	}

	scsiTarget := scsi.NewSCSITargetService()

	targetDriver, err := scsi.NewTargetDriver("iscsi", scsiTarget)
	if err != nil {
		return "", "", nil, err
	}

	conf := config.Config{
		Storages: []config.BackendStorage{
			{
				DeviceID: 1000,
				Path:     lhbsName,
				Online:   true,
			},
		},
		ISCSIPortals: []config.ISCSIPortalInfo{
			{
				ID:     0,
				Portal: portalIP,
			},
		},
		ISCSITargets: map[string]config.ISCSITarget{
			tgtName: {
				TPGTs: map[string][]uint64{
					"1": {0},
				},
				LUNs: map[string]uint64{
					"1": uint64(1000),
				},
			},
		},
	}

	err = targetDriver.NewTarget(tgtName, &conf)
	if err != nil {
		return "", "", nil, err
	}

	go func() {
		if err := targetDriver.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "%+v\n", err)
		}
	}()

	return portalIP, tgtName, func() {
		targetDriver.Close()
		os.Remove(tmpFile.Name())
	}, nil
}
