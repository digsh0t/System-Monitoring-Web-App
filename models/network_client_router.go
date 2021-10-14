package models

import (
	"encoding/hex"
	"errors"
	"strings"
	"time"

	g "github.com/gosnmp/gosnmp"
	"github.com/wintltr/login-api/utils"
)

type InterfaceSNMP struct {
	IfIndex         int    `json:"ifIndex"`
	IfDescription   string `json:"ifDescription"`
	IfType          string `json:"ifType"`
	IfMtu           int    `json:"ifMtu"`
	IfSpeed         uint   `json:"ifSpeed"`
	IfPhysAddress   string `json:"ifPhysAddress"`
	IfAdminStatus   string `json:"ifAdminStatus"`
	IfOperStatus    string `json:"ifOperStatus"`
	IfLastChange    uint32 `json:"ifLastChange"`
	IfInOctets      uint   `json:"ifInOctets"`
	IfInUcastPkts   uint   `json:"ifInUcastPkts"`
	IfInNUcastPkts  uint   `json:"ifInNUcastPkts"`
	IfInDiscards    uint   `json:"ifInDiscards"`
	IfInErrors      uint   `json:"ifInErrors"`
	IfOutOctets     uint   `json:"ifOutOctets"`
	IfOutUcastPkts  uint   `json:"ifOutUcastPkts"`
	IfOutNUcastPkts uint   `json:"ifOutNUcastPkts"`
	IfOutDiscards   uint   `json:"ifOutDiscards"`
	IfOutErrors     uint   `json:"ifOutErrors"`
}

type SystemSNMP struct {
	SysDescr    string `json:"sysDescr"`
	SysObjectID string `json:"sysObjectID"`
	SysUpTime   uint32 `json:"sysUpTime"`
	SysContact  string `json:"sysContact"`
	SysName     string `json:"sysName"`
	SysLocation string `json:"sysLocation"`
	SysServices int    `josn:"sysServices"`
}

type IpSNMP struct {
	IpAdEntIfIndex      int    `json:"ipAdEntIfIndex"`
	IpInterface         string `json:"ipInterface"`
	IpAdEntAddr         string `json:"ipAdEntAddr"`
	IpAdEntNetMask      string `json:"ipAdEntNetMask"`
	IpAdEntBcastAddr    int    `json:"ipAdEntBcastAddr"`
	IpAdEntReasmMaxSize int    `json:"ipAdEntReasmMaxSize"`
}

// Get Router Interfaces
func GetRouterInterfaces(sshConnectionId int) ([]InterfaceSNMP, error) {
	var (
		interfaceSNMPList []InterfaceSNMP
		err               error
	)
	// Get Hostname
	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return interfaceSNMPList, errors.New("fail to get ssh connection")
	}

	// Connect SNMP
	params, err := ConnectSNMP(*sshConnection)
	if err != nil {
		return interfaceSNMPList, errors.New("fail to connect SNMP")
	}
	defer params.Conn.Close()

	// Get max row index
	oidsList := []string{".1.3.6.1.2.1.2.1.0"}
	result, err := params.Get(oidsList)
	if err != nil {
		return interfaceSNMPList, errors.New("fail to get numbers of interfaces")
	}

	rowNum := result.Variables[0].Value.(int)

	// Initial array 2 dimensions
	array2d := make([][]interface{}, rowNum)
	for i := range array2d {
		array2d[i] = make([]interface{}, 22)
	}

	// Get Interfaces Information
	oids := ".1.3.6.1.2.1.2.2.1"
	column := 0
	row := 0
	err = params.Walk(oids, func(dataUnit g.SnmpPDU) error {
		switch dataUnit.Type {
		case g.OctetString:
			if column == 5 {
				encodedString := hex.EncodeToString(dataUnit.Value.([]byte))
				var rawMacAddress string
				for k, v := range encodedString {
					if k == 2 || k == 4 || k == 6 || k == 8 || k == 10 {
						rawMacAddress += "-"
					}
					rawMacAddress += string(v)
				}
				array2d[row][column] = strings.ToUpper(rawMacAddress)
			} else {
				bytes := dataUnit.Value.([]byte)
				array2d[row][column] = string(bytes)
			}
		default:
			array2d[row][column] = dataUnit.Value
		}
		row++
		if row == rowNum {
			column++
			row = 0
		}
		return err
	})
	if err != nil {
		return interfaceSNMPList, errors.New("fail to get information")
	}

	rowNumber := len(array2d)
	columnNumber := len(array2d[0])

	for i := 0; i < rowNumber; i++ {
		var interfaceSNMP InterfaceSNMP
		for y := 0; y < columnNumber; y++ {
			switch y {
			case 0:
				interfaceSNMP.IfIndex = array2d[i][y].(int)
			case 1:
				interfaceSNMP.IfDescription = array2d[i][y].(string)
			case 2:
				rawIfType := array2d[i][y].(int)
				interfaceSNMP.IfType = utils.ReferenceIfTypeRecord(rawIfType)
			case 3:
				interfaceSNMP.IfMtu = array2d[i][y].(int)
			case 4:
				interfaceSNMP.IfSpeed = array2d[i][y].(uint)
			case 5:
				interfaceSNMP.IfPhysAddress = array2d[i][y].(string)
			case 6:
				rawIfAdminStatus := array2d[i][y].(int)
				if rawIfAdminStatus == 1 {
					interfaceSNMP.IfAdminStatus = "up"
				} else {
					interfaceSNMP.IfAdminStatus = "down"
				}
			case 7:
				rawIfOperStatus := array2d[i][y].(int)
				if rawIfOperStatus == 1 {
					interfaceSNMP.IfOperStatus = "up"
				} else {
					interfaceSNMP.IfOperStatus = "down"
				}
			case 8:
				interfaceSNMP.IfLastChange = array2d[i][y].(uint32)
			case 9:
				interfaceSNMP.IfInOctets = array2d[i][y].(uint)
			case 10:
				interfaceSNMP.IfInUcastPkts = array2d[i][y].(uint)
			case 11:
				interfaceSNMP.IfInNUcastPkts = array2d[i][y].(uint)
			case 12:
				interfaceSNMP.IfInDiscards = array2d[i][y].(uint)
			case 13:
				interfaceSNMP.IfInErrors = array2d[i][y].(uint)
			case 15:
				interfaceSNMP.IfOutOctets = array2d[i][y].(uint)
			case 16:
				interfaceSNMP.IfOutUcastPkts = array2d[i][y].(uint)
			case 17:
				interfaceSNMP.IfOutNUcastPkts = array2d[i][y].(uint)
			case 18:
				interfaceSNMP.IfOutDiscards = array2d[i][y].(uint)
			case 19:
				interfaceSNMP.IfOutErrors = array2d[i][y].(uint)

			}
		}

		interfaceSNMPList = append(interfaceSNMPList, interfaceSNMP)

	}

	return interfaceSNMPList, err

}

// Get Router System Info
func GetRouterSystem(sshConnectionId int) (SystemSNMP, error) {
	var (
		systemSNMP SystemSNMP
		err        error
	)
	// Get Hostname
	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return systemSNMP, errors.New("fail to get ssh connection")
	}

	// Connect SNMP
	params, err := ConnectSNMP(*sshConnection)
	if err != nil {
		return systemSNMP, errors.New("fail to connect SNMP")
	}
	defer params.Conn.Close()

	// Get max row index
	oidsList := []string{".1.3.6.1.2.1.1.1.0", ".1.3.6.1.2.1.1.2.0", ".1.3.6.1.2.1.1.3.0", ".1.3.6.1.2.1.1.4.0", ".1.3.6.1.2.1.1.5.0", ".1.3.6.1.2.1.1.6.0", ".1.3.6.1.2.1.1.7.0"}
	result, err := params.Get(oidsList)
	if err != nil {
		return systemSNMP, errors.New("fail to get object identifiers")
	}

	for index, variable := range result.Variables {
		switch index {
		case 0:
			systemSNMP.SysDescr = string(variable.Value.([]byte))
		case 1:
			systemSNMP.SysObjectID = variable.Value.(string)
		case 2:
			systemSNMP.SysUpTime = variable.Value.(uint32)
		case 3:
			systemSNMP.SysContact = string(variable.Value.([]byte))
		case 4:
			systemSNMP.SysName = string(variable.Value.([]byte))
		case 5:
			systemSNMP.SysLocation = string(variable.Value.([]byte))
		case 6:
			systemSNMP.SysServices = variable.Value.(int)
		}

	}

	return systemSNMP, err

}

// Get Router Interfaces
func GetRouterIP(sshConnectionId int) ([]IpSNMP, error) {
	var (
		ipSNMPList    []IpSNMP
		tmpIpSNMPList []IpSNMP
		err           error
	)
	// Get Hostname
	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return ipSNMPList, errors.New("fail to get ssh connection")
	}

	// Connect SNMP
	params, err := ConnectSNMP(*sshConnection)
	if err != nil {
		return ipSNMPList, errors.New("fail to connect SNMP")
	}
	defer params.Conn.Close()

	// Get max row index
	rowNum := 0
	oids := ".1.3.6.1.2.1.4.20.1.1"
	err = params.Walk(oids, func(dataUnit g.SnmpPDU) error {
		rowNum++
		return err
	})
	if err != nil {
		return ipSNMPList, errors.New("fail to get numbers of interfaces")
	}

	// Initial array 2 dimensions
	array2d := make([][]interface{}, rowNum)
	for i := range array2d {
		array2d[i] = make([]interface{}, 5)
	}

	// Get Interfaces Information
	oids = ".1.3.6.1.2.1.4.20.1"
	column := 0
	row := 0
	err = params.Walk(oids, func(dataUnit g.SnmpPDU) error {
		switch dataUnit.Type {
		case g.OctetString:
			bytes := dataUnit.Value.([]byte)
			array2d[row][column] = string(bytes)
		default:
			array2d[row][column] = dataUnit.Value
		}
		row++
		if row == rowNum {
			column++
			row = 0
		}
		return err
	})
	if err != nil {
		return ipSNMPList, errors.New("fail to get information")
	}

	rowNumber := len(array2d)
	columnNumber := len(array2d[0])

	for i := 0; i < rowNumber; i++ {
		var ipSNMP IpSNMP
		for y := 0; y < columnNumber; y++ {
			switch y {
			case 0:
				ipSNMP.IpAdEntAddr = array2d[i][y].(string)
			case 1:
				ipSNMP.IpAdEntIfIndex = array2d[i][y].(int)
			case 2:
				ipSNMP.IpAdEntNetMask = array2d[i][y].(string)
			case 3:
				ipSNMP.IpAdEntBcastAddr = array2d[i][y].(int)
			case 4:
				if array2d[i][y] != nil {
					ipSNMP.IpAdEntReasmMaxSize = array2d[i][y].(int)
				}

			}
		}
		tmpIpSNMPList = append(tmpIpSNMPList, ipSNMP)
	}

	interfaceSNMPList, err := GetRouterInterfaces(sshConnectionId)
	if err != nil {
		return ipSNMPList, errors.New("fail to get interfaces")
	}

	for _, interfaces := range interfaceSNMPList {
		var ipSNMP IpSNMP
		ipSNMP.IpAdEntIfIndex = interfaces.IfIndex
		ipSNMP.IpInterface = interfaces.IfDescription
		for _, tmpIpSNMP := range tmpIpSNMPList {
			if interfaces.IfIndex == tmpIpSNMP.IpAdEntIfIndex {
				ipSNMP.IpAdEntAddr = tmpIpSNMP.IpAdEntAddr
				ipSNMP.IpAdEntNetMask = tmpIpSNMP.IpAdEntNetMask
				ipSNMP.IpAdEntBcastAddr = tmpIpSNMP.IpAdEntBcastAddr
				ipSNMP.IpAdEntReasmMaxSize = tmpIpSNMP.IpAdEntReasmMaxSize
			}
		}
		ipSNMPList = append(ipSNMPList, ipSNMP)

	}

	return ipSNMPList, err

}

func ConnectSNMP(sshConnection SshConnectionInfo) (*g.GoSNMP, error) {
	var err error

	// build our own GoSNMP struct, rather than using g.Default
	params := &g.GoSNMP{
		Target:        sshConnection.HostSSH,
		Port:          161,
		Version:       g.Version3,
		SecurityModel: g.UserSecurityModel,
		MsgFlags:      g.AuthPriv,
		Timeout:       time.Duration(30) * time.Second,
		SecurityParameters: &g.UsmSecurityParameters{UserName: "snmpUser",
			AuthenticationProtocol:   g.MD5,
			AuthenticationPassphrase: "snmpP@ssword",
			PrivacyProtocol:          g.DES,
			PrivacyPassphrase:        "snmpP@ssword",
		},
	}
	err = params.Connect()
	if err != nil {
		return params, err
	}
	return params, err
}
