package models

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Syslog struct {
	SyslogPRI      int
	SyslogFacility int
	Timegenerated  string
	Hostname       string
	ProgramName    string
	ProcessId      int
	Message        string
}

type SyslogConnectionInfo struct {
	ID       int    `json:"id"`
	Hostname string `json:"host"`
	ServerIP string `json:"server_ip"`
}
type SyslogPriStat struct {
	Pri0  int `json:"pri_0"`
	Pri1  int `json:"pri_1"`
	Pri2  int `json:"pri_2"`
	Pri3  int `json:"pri_3"`
	Pri4  int `json:"pri_4"`
	Pri5  int `json:"pri_5"`
	Pri6  int `json:"pri_6"`
	Pri7  int `json:"pri_7"`
	Total int `json:"total"`
}

func extractProcessId(input string) (string, int, error) {
	var processId int
	var err error
	var processName string
	input = strings.Trim(input, ":")
	r, _ := regexp.Compile(`\[(.*)\]`)
	tmp := r.FindString(input)
	if tmp != "" {
		tmp = tmp[1 : len(tmp)-1]
	}
	if tmp == "" {
		processId = -1
	} else {
		processId, err = strconv.Atoi(tmp)
		if err != nil {
			return "", -1, err
		}
	}

	processName = strings.Split(input, "[")[0]
	return processName, processId, nil
}

func parseSyslog(rawLogs string) ([]Syslog, error) {
	var processId int
	var programName string
	var err error
	var logs []Syslog
	for _, line := range strings.Split(strings.Trim(rawLogs, "\r\n "), "\n") {
		var log Syslog
		line = strings.Trim(line, "\r\n\t ")
		vars := strings.SplitN(line, ",", 6)
		programName, processId, err = extractProcessId(vars[4])
		if err != nil {
			return nil, err
		}
		log.SyslogPRI, err = strconv.Atoi(vars[0])
		if err != nil {
			return nil, err
		}
		log.SyslogFacility, err = strconv.Atoi(vars[1])
		if err != nil {
			return nil, err
		}
		log.Timegenerated, err = timestampToDate(vars[2])
		if err != nil {
			return nil, err
		}

		log.Hostname = vars[3]
		log.ProgramName = programName
		log.ProcessId = processId
		log.Message = vars[5]
		logs = append(logs, log)
	}
	return logs, nil
}

func timestampToDate(timestamp string) (string, error) {
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return "", err
	}
	return time.Unix(i, 0).String(), err
}

func parseSyslogRowNumbers(rawLogs string) (SyslogPriStat, error) {
	var err error
	var syslogPriStat SyslogPriStat
	for _, line := range strings.Split(strings.Trim(rawLogs, "\r\n "), "\n") {
		var log Syslog
		line = strings.Trim(line, "\r\n\t ")
		vars := strings.SplitN(line, ",", 6)
		log.SyslogPRI, err = strconv.Atoi(vars[0])
		if err != nil {
			return SyslogPriStat{}, err
		}
		switch log.SyslogPRI {
		case 0:
			syslogPriStat.Pri0 += 1
		case 1:
			syslogPriStat.Pri1 += 1
		case 2:
			syslogPriStat.Pri2 += 1
		case 3:
			syslogPriStat.Pri3 += 1
		case 4:
			syslogPriStat.Pri4 += 1
		case 5:
			syslogPriStat.Pri5 += 1
		case 6:
			syslogPriStat.Pri6 += 1
		case 7:
			syslogPriStat.Pri7 += 1
		}
		syslogPriStat.Total += 1
	}
	return syslogPriStat, nil
}

func parseSyslogByPri(rawLogs string, pri int) ([]Syslog, error) {
	var processId int
	var programName string
	var err error
	var logs []Syslog
	for _, line := range strings.Split(strings.Trim(rawLogs, "\r\n "), "\n") {
		var log Syslog
		line = strings.Trim(line, "\r\n\t ")
		vars := strings.SplitN(line, ",", 6)
		programName, processId, err = extractProcessId(vars[4])
		if err != nil {
			return nil, err
		}
		log.SyslogPRI, err = strconv.Atoi(vars[0])
		if err != nil {
			return nil, err
		}
		if log.SyslogPRI != pri {
			log = Syslog{}
			continue
		}
		log.SyslogFacility, err = strconv.Atoi(vars[1])
		if err != nil {
			return nil, err
		}
		log.Timegenerated = vars[2]
		log.Hostname = vars[3]
		log.ProgramName = programName
		log.ProcessId = processId
		log.Message = vars[5]
		logs = append(logs, log)
	}
	return logs, nil
}

func GetClientSyslog(logBasePath string, sshConnectionId int, date string) ([]Syslog, error) {
	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return nil, err
	}
	logPath := logBasePath + "/" + sshConnection.HostSSH + "/" + date + ".log"
	dat, err := os.ReadFile(logPath)
	if err != nil {
		return nil, err
	}
	logs, err := parseSyslog(string(dat))
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func GetClientSyslogByPri(logBasePath string, sshConnectionId int, date string, pri int) ([]Syslog, error) {
	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return nil, err
	}
	logPath := logBasePath + "/" + sshConnection.HostSSH + "/" + date + ".log"
	dat, err := os.ReadFile(logPath)
	if err != nil {
		return nil, err
	}
	logs, err := parseSyslogByPri(string(dat), pri)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func GetTotalSyslogRows(logBasePath string, sshConnectionId int, date string) ([]Syslog, error) {
	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return nil, err
	}
	logPath := logBasePath + "/" + sshConnection.HostSSH + "/" + date + ".log"
	dat, err := os.ReadFile(logPath)
	if err != nil {
		return nil, err
	}
	logs, err := parseSyslog(string(dat))
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func GetClientSyslogPriStat(logBasePath string, sshConnectionId int, date string) (SyslogPriStat, error) {
	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return SyslogPriStat{}, err
	}
	logPath := logBasePath + "/" + sshConnection.HostSSH + "/" + date + ".log"
	dat, err := os.ReadFile(logPath)
	if err != nil {
		return SyslogPriStat{}, err
	}
	logNumbers, err := parseSyslogRowNumbers(string(dat))
	if err != nil {
		return SyslogPriStat{}, err
	}
	return logNumbers, nil
}

func GetAllClientSyslogPriStat(logBasePath string, date string) (SyslogPriStat, error) {
	var syslogPriStat, tmpStat SyslogPriStat
	var err error
	sshConnectionList, err := GetAllSSHConnection()
	if err != nil {
		return SyslogPriStat{}, err
	}
	for _, sshConnection := range sshConnectionList {
		tmpStat, err = GetClientSyslogPriStat("/var/log/remotelogs", sshConnection.SSHConnectionId, date)
		if err != nil {
			if strings.Contains(err.Error(), "no such file or directory") {
				err = nil
				continue
			}
			return SyslogPriStat{}, err
		}
		syslogPriStat.Pri0 += tmpStat.Pri0
		syslogPriStat.Pri1 += tmpStat.Pri1
		syslogPriStat.Pri2 += tmpStat.Pri2
		syslogPriStat.Pri3 += tmpStat.Pri3
		syslogPriStat.Pri4 += tmpStat.Pri4
		syslogPriStat.Pri5 += tmpStat.Pri5
		syslogPriStat.Pri6 += tmpStat.Pri6
		syslogPriStat.Pri7 += tmpStat.Pri7
		syslogPriStat.Total += tmpStat.Total
		tmpStat = SyslogPriStat{}
	}
	return syslogPriStat, err
}

func GetAllClientSyslog(logBasePath string, date string) ([]Syslog, error) {
	var syslogRows, tmpRows []Syslog
	var err error
	sshConnectionList, err := GetAllSSHConnection()
	if err != nil {
		return nil, err
	}
	for _, sshConnection := range sshConnectionList {
		tmpRows, err = GetClientSyslog(logBasePath, sshConnection.SSHConnectionId, date)
		// if err != nil {
		// 	return nil, err
		// }
		syslogRows = append(syslogRows, tmpRows...)
		tmpRows = nil
	}
	syslogRows, err = sortSyslogByDate(syslogRows)
	return syslogRows, err
}

func sortSyslogByDate(logRows []Syslog) ([]Syslog, error) {
	var logTime1, logTime2 time.Time
	var err error
	var layoutNumber string = "2006-01-02 15:04:05 -0700 -07"
	sort.SliceStable(logRows, func(i, j int) bool {
		logTime1, err = time.Parse(layoutNumber, logRows[i].Timegenerated)
		if err != nil {
			return false
		}
		logTime2, err = time.Parse(layoutNumber, logRows[j].Timegenerated)
		if err != nil {
			return false
		}
		return logTime1.After(logTime2)
	})
	return logRows, err
}

func (sshConnection SshConnectionInfo) SetupSyslogWindows(serverIp string, configFilePath string) error {
	var command string
	newConfig := `Panic Soft
	#NoFreeOnExit TRUE
	
	define ROOT     C:\Program Files (x86)\nxlog
	define CERTDIR  %ROOT%\cert
	define CONFDIR  %ROOT%\conf
	define LOGDIR   %ROOT%\data
	define LOGFILE  %LOGDIR%\nxlog.log
	LogFile %LOGFILE%
	
	Moduledir %ROOT%\modules
	CacheDir  %ROOT%\data
	Pidfile   %ROOT%\data\nxlog.pid
	SpoolDir  %ROOT%\data
	
	<Extension _syslog>
		Module      xm_syslog
	</Extension>
	
	<Extension _charconv>
		Module      xm_charconv
		AutodetectCharsets iso8859-2, utf-8, utf-16, utf-32
	</Extension>
	
	<Extension _exec>
		Module      xm_exec
	</Extension>
	
	<Extension _fileop>
		Module      xm_fileop
	
		# Check the size of our log file hourly, rotate if larger than 5MB
		<Schedule>
			Every   1 hour
			Exec    if (file_exists('%LOGFILE%') and \
					   (file_size('%LOGFILE%') >= 5M)) \
						file_cycle('%LOGFILE%', 8);
		</Schedule>
	
		# Rotate our log file every week on Sunday at midnight
		<Schedule>
			When    @weekly
			Exec    if file_exists('%LOGFILE%') file_cycle('%LOGFILE%', 8);
		</Schedule>
	</Extension>
	
	# Snare compatible example configuration
	# Collecting event log
	 <Input in>
		 Module      im_msvistalog
	 </Input>
	# Buffering log if TLS/SSL out is unreachable
	 <Processor buffer>
		Module      pm_buffer
		Type        Disk
	
		# 40 MiB buffer
		MaxSize     40960
	
		# Generate warning message at 20 MiB
		WarnLimit   20480
	 </Processor>
	# Converting events to Snare format and sending them out over TCP syslog
	<Output out>
		Module  om_tcp
		Host    ` + serverIp + `
		Port    514
		Exec    to_syslog_bsd();
	</Output>
	# 
	# Connect input 'in' to output 'out'
	 <Route 1>
		 Path        in => buffer => out
	 </Route>
	
	`
	if configFilePath == "" {
		configFilePath = `C:\Program Files (x86)\nxlog\conf\nxlog.conf`
	}
	command = "(" + writeNewFileCMDCommand(newConfig) + ") > " + `"` + configFilePath + `"`
	output, err := sshConnection.RunCommandFromSSHConnectionUseKeys(command)
	if err != nil {
		return err
	}
	command += newConfig
	if strings.Trim(output, "\r\n\t ") != "" {
		return errors.New(output)
	}
	return err
}

func writeNewFileCMDCommand(text string) string {
	var command string = "echo "
	for _, line := range strings.Split(text, "\n") {
		line = strings.ReplaceAll(line, "(", "^(")
		line = strings.ReplaceAll(line, ")", "^)")
		line = strings.ReplaceAll(line, "&", "^&")
		line = strings.ReplaceAll(line, "<", "^<")
		line = strings.ReplaceAll(line, ">", "^>")
		command += strings.Trim(line, "\r\n\t ") + `& echo.`
	}
	return command
}

func (sshConnection SshConnectionInfo) SetupSyslogRsyslog(serverIp string, configFilePath string) (string, error) {
	var command string
	newConfig := `
# /etc/rsyslog.conf configuration file for rsyslog
#
# For more information install rsyslog-doc and see
# /usr/share/doc/rsyslog-doc/html/configuration/index.html
#
# Default logging rules can be found in /etc/rsyslog.d/50-default.conf


#################
#### MODULES ####
#################

module(load="imuxsock") # provides support for local system logging
#module(load="immark")  # provides --MARK-- message capability

# provides UDP syslog reception
module(load="imudp")
input(type="imudp" port="514")

# provides TCP syslog reception
#module(load="imtcp")
#input(type="imtcp" port="514")

# provides kernel logging support and enable non-kernel klog messages
module(load="imklog" permitnonkernelfacility="on")

###########################
#### GLOBAL DIRECTIVES ####
###########################

#
# Use traditional timestamp format.
# To enable high precision timestamps, comment out the following line.
#
$ActionFileDefaultTemplate RSYSLOG_TraditionalFileFormat

# Filter duplicated messages
$RepeatedMsgReduction on

#
# Set the default permissions for all log files.
#
$FileOwner syslog
$FileGroup adm
$FileCreateMode 0640
$DirCreateMode 0755
$Umask 0022
$PrivDropToUser syslog
$PrivDropToGroup syslog

#
# Where to place spool and state files
#
$WorkDirectory /var/spool/rsyslog

#
# Include all config files in /etc/rsyslog.d/
#
$IncludeConfig /etc/rsyslog.d/*.conf
*.* @@` + serverIp + `:514          # Use @@ for TCP protocol
`
	if configFilePath == "" {
		configFilePath = `/etc/rsyslog.conf`
	}
	command = `echo -e '` + newConfig + `' > "` + configFilePath + `"`
	output, err := sshConnection.RunCommandFromSSHConnectionUseKeys(command)
	if err != nil {
		return output, err
	}
	command += newConfig
	if strings.Trim(output, "\r\n\t ") != "" {
		return output, errors.New(output)
	}
	return output, err
}

func SetupRsyslogServer(configFilePath string) (string, error) {
	var command string
	newConfig := `
# rsyslog configuration file

# For more information see /usr/share/doc/rsyslog-*/rsyslog_conf.html
# or latest version online at http://www.rsyslog.com/doc/rsyslog_conf.html 
# If you experience problems, see http://www.rsyslog.com/doc/troubleshoot.html

#### MODULES ####

module(load="imuxsock" 	  # provides support for local system logging (e.g. via logger command)
	   SysSock.Use="off") # Turn off message reception via local log socket; 
			  # local messages are retrieved through imjournal now.
module(load="imjournal" 	    # provides access to the systemd journal
	   StateFile="imjournal.state") # File to store the position in the journal
#module(load="imklog") # reads kernel messages (the same are read from journald)
#module(load="immark") # provides --MARK-- message capability

# Provides UDP syslog reception
# for parameters see http://www.rsyslog.com/doc/imudp.html
module(load="imudp") # needs to be done just once
input(type="imudp" port="514")

# Provides TCP syslog reception
# for parameters see http://www.rsyslog.com/doc/imtcp.html
module(load="imtcp") # needs to be done just once
input(type="imtcp" port="514")

$template remote-incoming-logs, "/var/log/remotelogs/%FROMHOST-IP%/%$day%-%$month%-%$year%.log"

$template precise,"%syslogpriority%,%syslogfacility%,%$now-unixtimestamp%,%HOSTNAME%,%syslogtag%,%msg%\n"

*.* ?remote-incoming-logs;precise
#### GLOBAL DIRECTIVES ####

# Where to place auxiliary files
global(workDirectory="/var/lib/rsyslog")

# Use default timestamp format
module(load="builtin:omfile" Template="RSYSLOG_TraditionalFileFormat")

# Include all config files in /etc/rsyslog.d/
include(file="/etc/rsyslog.d/*.conf" mode="optional")

#### RULES ####

# Log all kernel messages to the console.
# Logging much else clutters up the screen.
#kern.*                                                 /dev/console

# Log anything (except mail) of level info or higher.
# Don't log private authentication messages!
*.info;mail.none;authpriv.none;cron.none                /var/log/messages

# The authpriv file has restricted access.
auth,authpriv.*                                              /var/log/secure

# Log all the mail messages in one place.
mail.*                                                  -/var/log/maillog


# Log cron stuff
cron.*                                                  /var/log/cron

# Everybody gets emergency messages
*.emerg                                                 :omusrmsg:*

# Save news errors of level crit and higher in a special file.
uucp,news.crit                                          /var/log/spooler

# Save boot messages also to boot.log
local7.*                                                /var/log/boot.log


# ### sample forwarding rule ###
#action(type="omfwd"  
# An on-disk queue is created for this action. If the remote host is
# down, messages are spooled to disk and sent when it is up again.
#queue.filename="fwdRule1"       # unique name prefix for spool files
#queue.maxdiskspace="1g"         # 1gb space limit (use as much as possible)
#queue.saveonshutdown="on"       # save messages to disk on shutdown
#queue.type="LinkedList"         # run asynchronously
#action.resumeRetryCount="-1"    # infinite retries if host is down
# Remote Logging (we use TCP for reliable delivery)
# remote_host is: name/ip, e.g. 192.168.0.1, port optional e.g. 10514
#Target="remote_host" Port="XXX" Protocol="tcp")`

	if configFilePath == "" {
		configFilePath = `/etc/rsyslog.conf`
	}
	newConfig = strings.ReplaceAll(newConfig, `\`, `\\`)
	newConfig = strings.ReplaceAll(newConfig, `'`, `'\''`)
	command = `echo -e '` + newConfig + `' > ` + configFilePath
	cmd := exec.Command("bash", "-c", command)
	var out, errbuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errbuf
	err := cmd.Run()
	output := out.String()
	stderr := errbuf.String()
	if stderr != "" {
		log.Println(stderr)
	}
	command += newConfig
	if err != nil {
		return string(output), err
	}
	if strings.Trim(string(output), "\r\n\t ") != "" {
		return string(output), errors.New(string(output))
	}
	return string(output), err
}
