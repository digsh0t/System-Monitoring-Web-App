package models

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
	IfLastChange    string `json:"ifLastChange"`
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
	SysUpTime   string `json:"sysUpTime"`
	SysContact  string `json:"sysContact"`
	SysName     string `json:"sysName"`
	SysLocation string `json:"sysLocation"`
	SysServices int    `josn:"sysServices"`
}

type IpAddrSNMP struct {
	IpAdEntIfIndex      int    `json:"ipAdEntIfIndex"`
	IpInterface         string `json:"ipInterface"`
	IpAdEntAddr         string `json:"ipAdEntAddr"`
	IpAdEntNetMask      string `json:"ipAdEntNetMask"`
	IpAdEntBcastAddr    int    `json:"ipAdEntBcastAddr"`
	IpAdEntReasmMaxSize int    `json:"ipAdEntReasmMaxSize"`
}

type IpNetToMediaSNMP struct {
	IpNetToMediaIfIndex     int    `json:"ipNetToMediaIfIndex"`
	IpInterface             string `json:"ipInterface"`
	IpNetToMediaPhysAddress string `json:"ipNetToMediaPhysAddress"`
	IpNetToMediaNetAddress  string `json:"ipNetToMediaNetAddress"`
	IpNetToMediaType        string `json:"ipNetToMediaType"`
}

type IpRouteSNMP struct {
	IpRouteIfIndex int    `json:"ipRouteIfIndex"`
	IpRouteDest    string `json:"ipRouteDest"`
	IpRouteMetric1 int    `json:"ipRouteMetric1"`
	IpRouteMetric2 int    `json:"ipRouteMetric2"`
	IpRouteMetric3 int    `json:"ipRouteMetric3"`
	IpRouteMetric4 int    `json:"ipRouteMetric4"`
	IpRouteNextHop string `json:"ipRouteNextHop"`
	IpRouteType    string `json:"ipRouteType"`
	IpRouteProto   string `json:"ipRouteProto"`
	IpRouteAge     int    `json:"ipRouteAge"`
	IpRouteMask    string `json:"ipRouteMask"`
	IpRouteMetric5 int    `json:"ipRouteMetric5"`
}

type NetworkJson struct {
	SshConnectionId []int  `json:"sshConnectionId"`
	Host            string `json:"host"`
	Dest            string `json:"dest"`
}

// Get Router Interfaces
func GetNetworkInterfaces(sshConnectionId int) ([]InterfaceSNMP, error) {
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

	// Parse Network interface
	if sshConnection.NetworkType == "router" {
		interfaceSNMPList = ParseRouterInterfaces(array2d)
	} else if sshConnection.NetworkType == "switch" {
		interfaceSNMPList = ParseSwitchInterfaces(array2d)
	}

	return interfaceSNMPList, err

}

func ParseRouterInterfaces(array2d [][]interface{}) []InterfaceSNMP {
	var interfaceSNMPList []InterfaceSNMP
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
				switch rawIfAdminStatus {
				case 1:
					interfaceSNMP.IfAdminStatus = "up"
				case 2:
					interfaceSNMP.IfAdminStatus = "down"
				case 3:
					interfaceSNMP.IfAdminStatus = "testing"
				}
			case 7:
				rawIfOperStatus := array2d[i][y].(int)
				switch rawIfOperStatus {
				case 1:
					interfaceSNMP.IfOperStatus = "up"
				case 2:
					interfaceSNMP.IfOperStatus = "down"
				case 3:
					interfaceSNMP.IfOperStatus = "testing"
				case 4:
					interfaceSNMP.IfOperStatus = "unknown"
				case 5:
					interfaceSNMP.IfOperStatus = "dormant"
				case 6:
					interfaceSNMP.IfOperStatus = "notPresent"
				case 7:
					interfaceSNMP.IfOperStatus = "lowerLayerDown"
				}
			case 8:
				rawIfLastChange := array2d[i][y].(uint32)
				interfaceSNMP.IfLastChange = utils.HundredSecondsToHuman(int(rawIfLastChange))
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
	return interfaceSNMPList
}

func ParseSwitchInterfaces(array2d [][]interface{}) []InterfaceSNMP {
	var interfaceSNMPList []InterfaceSNMP
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
				rawIfLastChange := array2d[i][y].(uint32)
				interfaceSNMP.IfLastChange = utils.HundredSecondsToHuman(int(rawIfLastChange))
			case 9:
				interfaceSNMP.IfInOctets = array2d[i][y].(uint)
			case 10:
				interfaceSNMP.IfInUcastPkts = array2d[i][y].(uint)
			case 11:
				interfaceSNMP.IfInDiscards = array2d[i][y].(uint)
			case 12:
				interfaceSNMP.IfInErrors = array2d[i][y].(uint)
			case 14:
				interfaceSNMP.IfOutOctets = array2d[i][y].(uint)
			case 15:
				interfaceSNMP.IfOutUcastPkts = array2d[i][y].(uint)
			case 16:
				interfaceSNMP.IfOutDiscards = array2d[i][y].(uint)
			case 17:
				interfaceSNMP.IfOutErrors = array2d[i][y].(uint)

			}
		}
		interfaceSNMPList = append(interfaceSNMPList, interfaceSNMP)
	}
	return interfaceSNMPList
}

// Get Router Interfaces
func GetMapIndexInterfaceName(sshConnectionId int) (map[int]string, error) {
	var (
		err error
	)
	mapIndexName := make(map[int]string)
	// Get Hostname
	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return mapIndexName, errors.New("fail to get ssh connection")
	}

	// Connect SNMP
	params, err := ConnectSNMP(*sshConnection)
	if err != nil {
		return mapIndexName, errors.New("fail to connect SNMP")
	}
	defer params.Conn.Close()

	// Get max row index
	oidsList := []string{".1.3.6.1.2.1.2.1.0"}
	result, err := params.Get(oidsList)
	if err != nil {
		return mapIndexName, errors.New("fail to get numbers of interfaces")
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
		return mapIndexName, errors.New("fail to get information")
	}

	rowNumber := len(array2d)

	for i := 0; i < rowNumber; i++ {
		index := array2d[i][0].(int)
		interfaceName := array2d[i][1].(string)
		mapIndexName[index] = interfaceName

	}

	return mapIndexName, err

}

// Get Router System Info
func GetNetworkSystem(sshConnectionId int) (SystemSNMP, error) {
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
			rawSysUpTime := variable.Value.(uint32)
			systemSNMP.SysUpTime = utils.HundredSecondsToHuman(int(rawSysUpTime))
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
func GetNetworkIPAddr(sshConnectionId int) ([]IpAddrSNMP, error) {
	var (
		ipSNMPList    []IpAddrSNMP
		tmpIpSNMPList []IpAddrSNMP
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
		fmt.Println(err.Error())
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
		var ipSNMP IpAddrSNMP
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

	interfaceSNMPList, err := GetNetworkInterfaces(sshConnectionId)
	if err != nil {
		return ipSNMPList, errors.New("fail to get interfaces")
	}

	for _, interfaces := range interfaceSNMPList {
		var ipSNMP IpAddrSNMP
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
	var (
		err    error
		params *g.GoSNMP
	)

	// Get SNMP Credential
	snmpCredential, err := GetSNMPCredentialFromSshConnectionId(sshConnection.SSHConnectionId)
	if err != nil {
		return params, err
	}

	// build our own GoSNMP struct, rather than using g.Default
	params = &g.GoSNMP{
		Target:        sshConnection.HostSSH,
		Port:          161,
		Version:       g.Version3,
		SecurityModel: g.UserSecurityModel,
		MsgFlags:      g.AuthPriv,
		Timeout:       time.Duration(30) * time.Second,
		SecurityParameters: &g.UsmSecurityParameters{UserName: snmpCredential.AuthUsername,
			AuthenticationProtocol:   g.MD5,
			AuthenticationPassphrase: snmpCredential.AuthPassword,
			PrivacyProtocol:          g.DES,
			PrivacyPassphrase:        snmpCredential.PrivPassword,
		},
	}
	err = params.Connect()
	if err != nil {
		return params, err
	}
	return params, err
}

// Get Router Ip Net To Media
func GetNetworkIPNetToMedia(sshConnectionId int) ([]IpNetToMediaSNMP, error) {
	var (
		ipNetToMediaSNMPList []IpNetToMediaSNMP
		err                  error
	)
	// Get Hostname
	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return ipNetToMediaSNMPList, errors.New("fail to get ssh connection")
	}

	// Connect SNMP
	params, err := ConnectSNMP(*sshConnection)
	if err != nil {
		return ipNetToMediaSNMPList, errors.New("fail to connect SNMP")
	}
	defer params.Conn.Close()

	// Get max row index
	oidsList := ".1.3.6.1.2.1.4.22.1.1"
	rowNum := 0
	err = params.Walk(oidsList, func(dataUnit g.SnmpPDU) error {
		rowNum++
		return err
	})
	if err != nil {
		return ipNetToMediaSNMPList, errors.New("fail to get number of rows")
	}

	// Initial array 2 dimensions
	array2d := make([][]interface{}, rowNum)
	for i := range array2d {
		array2d[i] = make([]interface{}, 4)
	}

	// Get IpNetToMedia Information
	oids := ".1.3.6.1.2.1.4.22.1"
	column := 0
	row := 0
	err = params.Walk(oids, func(dataUnit g.SnmpPDU) error {
		switch dataUnit.Type {
		case g.OctetString:
			if column == 1 {
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
		return ipNetToMediaSNMPList, errors.New("fail to get information")
	}

	mapIndexInterfaceName, err := GetMapIndexInterfaceName(sshConnectionId)
	if err != nil {
		return ipNetToMediaSNMPList, errors.New("fail to get map index name information")
	}

	rowNumber := len(array2d)
	columnNumber := len(array2d[0])

	for i := 0; i < rowNumber; i++ {
		var ipNetToMediaSNMP IpNetToMediaSNMP
		for y := 0; y < columnNumber; y++ {
			switch y {
			case 0:
				ipNetToMediaSNMP.IpNetToMediaIfIndex = array2d[i][y].(int)
			case 1:
				ipNetToMediaSNMP.IpNetToMediaPhysAddress = array2d[i][y].(string)
			case 2:
				ipNetToMediaSNMP.IpNetToMediaNetAddress = array2d[i][y].(string)
			case 3:
				ipNetToMediaType := array2d[i][y].(int)
				switch ipNetToMediaType {
				case 1:
					ipNetToMediaSNMP.IpNetToMediaType = "other"
				case 2:
					ipNetToMediaSNMP.IpNetToMediaType = "invalid"
				case 3:
					ipNetToMediaSNMP.IpNetToMediaType = "dynamic"
				case 4:
					ipNetToMediaSNMP.IpNetToMediaType = "static"
				}
			}
		}
		ipNetToMediaSNMP.IpInterface = mapIndexInterfaceName[ipNetToMediaSNMP.IpNetToMediaIfIndex]
		ipNetToMediaSNMPList = append(ipNetToMediaSNMPList, ipNetToMediaSNMP)

	}

	return ipNetToMediaSNMPList, err

}

// Get Router route
func GetNetworkIPRoute(sshConnectionId int) ([]IpRouteSNMP, error) {
	var (
		ipRouteSNMPList []IpRouteSNMP
		err             error
	)

	// Get Hostname
	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return ipRouteSNMPList, errors.New("fail to get ssh connection")
	}

	// Exception: Juniper not supported
	if sshConnection.NetworkType == "router" && sshConnection.NetworkOS == "junos" {
		return ipRouteSNMPList, errors.New("this function is not supported on the device")
	}
	// Connect SNMP
	params, err := ConnectSNMP(*sshConnection)
	if err != nil {
		return ipRouteSNMPList, errors.New("fail to connect SNMP")
	}
	defer params.Conn.Close()

	// Get max row index
	oidsList := ".1.3.6.1.2.1.4.21.1.1"
	rowNum := 0
	err = params.Walk(oidsList, func(dataUnit g.SnmpPDU) error {
		rowNum++
		return err
	})
	if err != nil {
		return ipRouteSNMPList, errors.New("fail to get number of rows")
	}

	// No record
	if rowNum == 0 {
		return ipRouteSNMPList, err
	}
	// Initial array 2 dimensions
	array2d := make([][]interface{}, rowNum)
	for i := range array2d {
		array2d[i] = make([]interface{}, 13)
	}

	// Get Route Information
	oids := ".1.3.6.1.2.1.4.21.1"
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
		return ipRouteSNMPList, errors.New("fail to get information")
	}

	rowNumber := len(array2d)
	columnNumber := len(array2d[0])

	if sshConnection.NetworkOS != "vyos" {
		for i := 0; i < rowNumber; i++ {
			var ipRouteSNMP IpRouteSNMP
			for y := 0; y < columnNumber; y++ {
				switch y {
				case 0:
					ipRouteSNMP.IpRouteDest = array2d[i][y].(string)
				case 1:
					ipRouteSNMP.IpRouteIfIndex = array2d[i][y].(int)
				case 2:
					ipRouteSNMP.IpRouteMetric1 = array2d[i][y].(int)
				case 3:
					ipRouteSNMP.IpRouteMetric2 = array2d[i][y].(int)
				case 4:
					ipRouteSNMP.IpRouteMetric3 = array2d[i][y].(int)
				case 5:
					ipRouteSNMP.IpRouteMetric4 = array2d[i][y].(int)
				case 6:
					ipRouteSNMP.IpRouteNextHop = array2d[i][y].(string)
				case 7:
					ipRouteType := array2d[i][y].(int)
					switch ipRouteType {
					case 1:
						ipRouteSNMP.IpRouteType = "other"
					case 2:
						ipRouteSNMP.IpRouteType = "invalid"
					case 3:
						ipRouteSNMP.IpRouteType = "direct"
					case 4:
						ipRouteSNMP.IpRouteType = "indirect"
					}
				case 8:
					ipRouteProto := array2d[i][y].(int)
					ipRouteSNMP.IpRouteProto = utils.ReferenceIpRouteProtoRecord(ipRouteProto)
				case 9:
					ipRouteSNMP.IpRouteAge = array2d[i][y].(int)
				case 10:
					ipRouteSNMP.IpRouteMask = array2d[i][y].(string)
				case 11:
					ipRouteSNMP.IpRouteMetric5 = array2d[i][y].(int)

				}
			}
			ipRouteSNMPList = append(ipRouteSNMPList, ipRouteSNMP)

		}
	} else {
		for i := 0; i < rowNumber; i++ {
			var ipRouteSNMP IpRouteSNMP
			for y := 0; y < columnNumber; y++ {
				switch y {
				case 0:
					ipRouteSNMP.IpRouteDest = array2d[i][y].(string)
				case 1:
					ipRouteSNMP.IpRouteIfIndex = array2d[i][y].(int)
				case 2:
					ipRouteSNMP.IpRouteMetric1 = array2d[i][y].(int)
				case 3:
					ipRouteSNMP.IpRouteNextHop = array2d[i][y].(string)
				case 4:
					ipRouteType := array2d[i][y].(int)
					switch ipRouteType {
					case 1:
						ipRouteSNMP.IpRouteType = "other"
					case 2:
						ipRouteSNMP.IpRouteType = "invalid"
					case 3:
						ipRouteSNMP.IpRouteType = "direct"
					case 4:
						ipRouteSNMP.IpRouteType = "indirect"
					}
				case 5:
					ipRouteProto := array2d[i][y].(int)
					ipRouteSNMP.IpRouteProto = utils.ReferenceIpRouteProtoRecord(ipRouteProto)
				case 6:
					ipRouteSNMP.IpRouteMask = array2d[i][y].(string)
				}

			}
			ipRouteSNMPList = append(ipRouteSNMPList, ipRouteSNMP)

		}
	}

	return ipRouteSNMPList, err

}

// config static route
func TestPingNetworkDevices(networkJson NetworkJson) ([]string, error) {
	var (
		outputList []string
		err        error
	)
	// Get Hostname from Id
	for _, id := range networkJson.SshConnectionId {
		sshConnection, err := GetSSHConnectionFromId(id)
		if err != nil {
			return outputList, errors.New("fail to parse id")
		}

		networkJson.Host = sshConnection.HostNameSSH

		// Marshal and run playbook
		ciscoJsonMarshal, err := json.Marshal(networkJson)
		if err != nil {
			return outputList, err
		}
		var filepath string
		if sshConnection.NetworkOS == "ios" {
			filepath = "./yamls/network_client/cisco/cisco_test_ping.yml"
		} else if sshConnection.NetworkOS == "vyos" {
			filepath = "./yamls/network_client/vyos/vyos_test_ping.yml"
		} else if sshConnection.NetworkOS == "junos" {
			filepath = "./yamls/network_client/juniper/juniper_test_ping.yml"
		}
		output, err := RunAnsiblePlaybookWithjson(filepath, string(ciscoJsonMarshal))
		if err != nil {
			return outputList, errors.New("fail to load yaml file")
		}
		outputList = append(outputList, output)
	}
	return outputList, err
}
