package osbrick

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
)

const iqnFilePath = "/etc/iscsi/initiatorname.iscsi"

func GetIQN(ctx context.Context) (string, error) {
	b, err := ioutil.ReadFile(iqnFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read iqn file: %w", err)
	}

	tmp := strings.TrimSpace(string(b))
	words := strings.Split(tmp, "=")
	if len(words) < 2 {
		return "", fmt.Errorf("failed to parse iqn file: %w", err)
	}

	return words[1], nil
}
