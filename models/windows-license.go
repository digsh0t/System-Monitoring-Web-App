package models

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

type windowsLicense struct {
	ProductName string `json:"product_name"`
	ProductKey  string `json:"product_key"`
	ProductId   string `json:"product_id"`
}

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

func (sshConnection SshConnectionInfo) GetWindowsLicenseKey() (windowsLicense, error) {
	var productKeyOffset = 52
	var license windowsLicense

	var regKeyList []RegistryKey
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM registry WHERE key = 'HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion' AND name = 'DigitalProductId'";`)
	if err != nil {
		return license, err
	}
	regKeyList, err = parseKeyList(result)
	if err != nil {
		return license, err
	}
	if regKeyList == nil {
		return license, nil
	}
	digitalProductID, err := hex.DecodeString(regKeyList[0].Data)
	if err != nil {
		return license, err
	}
	binaryKey := digitalProductID[productKeyOffset:]
	license.ProductName = "Windows 10 Pro"
	license.ProductKey = fmt.Sprint(binaryKeyToASCII(binaryKey))
	return license, err
}

func (sshConnection SshConnectionInfo) GetWindowsVmwareProductKey() (windowsLicense, error) {
	var regKeyList []RegistryKey
	var license windowsLicense
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT name,path FROM registry WHERE key = 'HKEY_LOCAL_MACHINE\SOFTWARE\WOW6432Node\VMware, Inc.\VMware Workstation';`)
	if err != nil {
		return license, err
	}
	regKeyList, err = parseKeyList(result)
	if err != nil {
		return license, err
	}
	if regKeyList == nil {
		return license, nil
	}
	for _, key := range regKeyList {
		if strings.Contains(key.Name, "License.") {
			tmpString := key.Name
			result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT name,path,data FROM registry WHERE key = 'HKEY_LOCAL_MACHINE\SOFTWARE\WOW6432Node\VMware, Inc.\VMware Workstation\` + key.Name + `';`)
			if err != nil {
				return license, err
			}
			regKeyList = nil
			regKeyList, err = parseKeyList(result)
			if err != nil {
				return license, err
			}
			if regKeyList == nil {
				return license, nil
			}
			for _, key := range regKeyList {
				if key.Name == "ProductID" {
					license.ProductName = key.Data + " " + tmpString
				}
				if key.Name == "Serial" {
					license.ProductKey = key.Data
				}
			}
			return license, err
		}
	}
	return license, err
}

func (sshConnection SshConnectionInfo) GetWindowsProductKey() (windowsLicense, error) {
	var regKeyList []RegistryKey
	var license windowsLicense
	tmpInfo, _ := sshConnection.GetDetailSSHConInfo()
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM registry WHERE key = 'HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion\SoftwareProtectionPlatform';"`)
	if err != nil {
		return license, err
	}
	regKeyList, err = parseKeyList(result)
	if err != nil {
		return license, err
	}
	if regKeyList == nil {
		return license, nil
	}
	for _, key := range regKeyList {
		if key.Name == "BackupProductKeyDefault" {
			license.ProductName = tmpInfo.OsName
			license.ProductKey = key.Data
			return license, err
		}
	}
	return license, err
}

func (sshConnection SshConnectionInfo) GetAllWindowsLicense() ([]windowsLicense, error) {
	var tmpLicense windowsLicense
	var licenseList []windowsLicense
	var err error
	tmpLicense, err = sshConnection.GetWindowsLicenseKey()
	if err != nil {
		return nil, err
	}
	licenseList = append(licenseList, tmpLicense)
	tmpLicense, err = sshConnection.GetWindowsVmwareProductKey()
	if err != nil {
		return nil, err
	}
	licenseList = append(licenseList, tmpLicense)
	return licenseList, err
}
