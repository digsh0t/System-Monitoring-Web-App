package models

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

func rev(b []byte) {
	for i := len(b)/2 - 1; i >= 0; i-- {
		j := len(b) - 1 - i
		b[i], b[j] = b[j], b[i]
	}
}

func decodeByte(buf []byte) byte {
	const chars = "BCDFGHJKMPQRTVWXY2346789"
	acc := 0
	for j := 14; j >= 0; j-- {
		acc *= 256
		acc += int(buf[j])
		buf[j] = byte((acc / len(chars)) & 0xFF)
		acc %= len(chars)
	}
	return chars[acc]
}

func binaryKeyToASCII(buf []byte) string {
	var out bytes.Buffer
	for i := 28; i >= 0; i-- {
		if (29-i)%6 == 0 {
			out.WriteByte('-')
			i--
		}
		out.WriteByte(decodeByte(buf))
	}
	outBytes := out.Bytes()
	rev(outBytes)
	return string(outBytes)
}

func (sshConnection SshConnectionInfo) GetWindowsLicenseKey() (string, error) {
	var productKeyOffset = 52

	var regKeyList []RegistryKey
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM registry WHERE key = 'HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion' AND name = 'DigitalProductId4'";`)
	if err != nil {
		return "", err
	}
	regKeyList, err = parseKeyList(result)
	if err != nil {
		return "", err
	}
	if regKeyList == nil {
		return "", nil
	}
	digitalProductID, err := hex.DecodeString(regKeyList[0].Data)
	if err != nil {
		return "", err
	}
	binaryKey := digitalProductID[productKeyOffset:]
	return fmt.Sprint(binaryKeyToASCII(binaryKey)), err
}
