package models

import (
	"errors"
	"strings"
	"time"

	"github.com/wintltr/login-api/database"
)

type PackageInstalledInfo struct {
	PackageId       int
	PackageName     string
	PackageDate     string
	SSHConnectionId int
}

type PackageInfo struct {
	Host    []string `json:"host"`
	File    string   `json:"file"`
	Mode    string   `json:"mode"`
	Package string   `json:"package"`
	Link    string   `json:"link"`
}

func AddPackage(recapStructList []RecapInfo, pkgName string) (bool, error) {
	var err error = nil
	var result bool = true

	for _, recap := range recapStructList {

		// changed > 0 means installing package successfully
		if recap.Changed > 0 {
			var packageInfo PackageInstalledInfo
			packageInfo.PackageName = pkgName
			currentTime := time.Now()
			packageInfo.PackageDate = currentTime.Format("2006-01-02 15:04:05")
			recap.ClientName = strings.TrimSpace(recap.ClientName)
			SshConnectionInfo, err := GetSSHConnectionFromHostName(recap.ClientName)
			if err != nil {
				return false, err
			}
			packageInfo.SSHConnectionId = SshConnectionInfo.SSHConnectionId
			result, err = InsertPackageToDB(packageInfo)
			if err != nil {
				return false, err
			}

		}
	}
	return result, err

}

func InsertPackageToDB(Package PackageInstalledInfo) (bool, error) {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO package_installed (pkg_name,pkg_date,pkg_host_id) VALUES (?,?,?)")
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(Package.PackageName, Package.PackageDate, Package.SSHConnectionId)
	if err != nil {
		return false, err
	}

	return true, err

}

func RemovePackage(recapStructList []RecapInfo, pkgName string) (bool, error) {
	var err error
	var result bool = true

	for _, recap := range recapStructList {

		// changed > 0 means installing package successfully
		if recap.Changed > 0 {
			var packageInfo PackageInstalledInfo
			packageInfo.PackageName = pkgName
			recap.ClientName = strings.TrimSpace(recap.ClientName)
			SshConnectionInfo, err := GetSSHConnectionFromHostName(recap.ClientName)
			if err != nil {
				return false, err
			}
			packageInfo.SSHConnectionId = SshConnectionInfo.SSHConnectionId
			result, err = DeletePackageFromDB(packageInfo)
			if err != nil {
				return false, err
			}

		}
	}
	return result, err

}

func DeletePackageFromDB(Package PackageInstalledInfo) (bool, error) {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM package_installed WHERE pkg_host_id = ? AND pkg_name = ?")
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(Package.SSHConnectionId, Package.PackageName)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if rows == 0 {
		return false, errors.New("no SSH Connections with this ID exists")
	}
	return true, err
}

func GetAllPackageFromHostID(HostId int) ([]PackageInstalledInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	var PackageList []PackageInstalledInfo
	selDB, err := db.Query("SELECT * FROM package_installed WHERE pkg_host_id = ?", HostId)
	if err != nil {
		return PackageList, err
	}

	var Package PackageInstalledInfo
	for selDB.Next() {
		var id int
		var name, date, hostId string

		err = selDB.Scan(&id, &name, &date, &hostId)
		if err != nil {
			return PackageList, err
		}
		Package.PackageId = id
		Package.PackageName = name
		Package.PackageDate = date
		Package.SSHConnectionId = HostId
		PackageList = append(PackageList, Package)
	}

	return PackageList, err

}

func GetAllPackage() ([]PackageInstalledInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	var PackageList []PackageInstalledInfo
	selDB, err := db.Query("SELECT * FROM package_installed")
	if err != nil {
		return PackageList, err
	}

	var Package PackageInstalledInfo
	for selDB.Next() {
		var id, hostId int
		var name, date string

		err = selDB.Scan(&id, &name, &date, &hostId)
		if err != nil {
			return PackageList, err
		}
		Package.PackageId = id
		Package.PackageName = name
		Package.PackageDate = date
		Package.SSHConnectionId = hostId
		PackageList = append(PackageList, Package)
	}

	return PackageList, err

}
