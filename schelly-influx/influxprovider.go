package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/flaviostutz/schelly-webhook/schellyhook"
	"go.uber.org/zap"
)

var dataStringSeparator string

var fileName *string //output file

// backups directory where the backup files will be placed
var backupsDir *string

// General options:

// Options controlling the output content:
var retention *string // retetion policy.
var shard *string     // Shard ID of the shard to be backed up
var start *string     // Include all points starting with the specified timestamp. RFC3339 format
var end *string       // Exclude all results after the specified timestamp. RFC3339 format.
var since *string     // Perform an incremental backup after the specified timestamp RFC3339 format.

// Connection options:
var database *string // database to dump
var host *string     // database server host or socket directory
var port *int        // database server port number

//InfluxBackuper sample backuper
type InfluxBackuper struct{}

//Init check the pg_dump version
func (sb InfluxBackuper) Init() error {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	dataStringSeparator = "---"

	info, err := schellyhook.ExecShell("influx --version")
	if err != nil {
		sugar.Errorf("Couldn't retrieve influx version. err=%s", err)
		return err
	} else {
		sugar.Infof(info)
	}

	if *backupsDir == "" {
		return fmt.Errorf("backup-dir arg must be defined")
	}
	if *host == "" {
		return fmt.Errorf("`database host` (-host) arg must be set. It can be an IP address or a domain name")
	}
	if *port <= 0 {
		return fmt.Errorf("`database port` (-port) arg must be a valid value, such as 5432")
	}
	if *database == "" {
		return fmt.Errorf("`database` (-database) arg must be set")
	}

	err = mkDirs(*backupsDir)
	if err != nil {
		return fmt.Errorf("Error creating backups `base-dir`. error: %s", err)
	}

	sugar.Infof("InfluxDB Provider ready to work. Version: %s", info)

	return nil
}

//RegisterFlags register command line flags
func (sb InfluxBackuper) RegisterFlags() error {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	// General options:
	backupsDir = flag.String("backup-dir", "/var/backups/database", "--backup-dir=PATH -> output file path")

	// Options controlling the output content:
	retention = flag.String("retention", "", "-retention -> Retention policy for the backup. If not specified, the default is to use all retention policies. If specified, then -database is required")
	shard = flag.String("shard", "", "-shard -> Shard ID of the shard to be backed up")
	start = flag.String("start", "", "-start -> Include all points starting with the specified timestamp (RFC3339 format).")
	end = flag.String("end", "", "-end -> Exclude all results after the specified timestamp (RFC3339 format).")
	since = flag.String("since", "", "-since -> Perform an incremental backup after the specified timestamp RFC3339 format.")

	// Connection options:
	database = flag.String("database", "", "-database=DBNAME -> database to dump")
	host = flag.String("host", "", "--host=HOSTNAME -> database server host or socket directory")
	port = flag.Int("port", 8088, "--port=PORT -> database server port number")

	// flag.Parse() //invoked by the hook
	sugar.Infof("Flags registration completed")

	return nil
}

//CreateNewBackup creates a new backup
func (sb InfluxBackuper) CreateNewBackup(apiID string, timeout time.Duration, shellContext *schellyhook.ShellContext) error {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	sugar.Infof("CreateNewBackup() apiID=%s timeout=%d s", apiID, timeout.Seconds)
	sugar.Infof("Running InfluxDB backup")

	dumpID := time.Now().Format("20060102150405")

	thisBackupDir := *backupsDir + "/" + apiID + dataStringSeparator + dumpID

	retentionString := ""
	if *retention != "" {
		retentionString = " -retention" + *retention
	}
	shardString := ""
	if *shard != "" {
		shardString = " -shard=" + *shard
	}
	startString := ""
	if *start != "" {
		startString = " -start" + *start
	}
	endString := ""
	if *end != "" {
		endString = " -end" + *end
	}
	sinceString := ""
	if *since != "" {
		sinceString = " -since" + *since
	}

	backupCommand := "influxd backup -portable -database=" + *database + " -host=" + *host + ":" + strconv.Itoa(*port) + retentionString + shardString + startString + endString + sinceString + " " + thisBackupDir
	sugar.Debugf("Executing influxd backup command: %s", backupCommand)
	out, err := schellyhook.ExecShellTimeout(backupCommand, timeout, shellContext)

	if err != nil {
		status := (*shellContext).CmdRef.Status()
		if status.Exit == -1 {
			sugar.Warnf("InfluxProvider influxd backup command timeout enforced (%d seconds)", (status.StopTs-status.StartTs)/1000000000)
		}
		sugar.Debugf("InfluxProvider backup error. out=%s; err=%s", out, err.Error())
		errorFileBytes := []byte(dumpID)
		errorFilePath := resolveErrorFilePath(apiID)
		err := ioutil.WriteFile(errorFilePath, errorFileBytes, 0600)
		if err != nil {
			sugar.Errorf("Error writing .error file for %s. err: %s", apiID, err)
			return err
		}

		return err
	}

	sugar.Debugf("InfluxDB backup started. Output log:")
	sugar.Debugf(out)

	sugar.Infof("InfluxDB backup launched")
	return nil
}

//GetAllBackups returns all backups from underlaying backuper. optional for Schelly
func (sb InfluxBackuper) GetAllBackups() ([]schellyhook.SchellyResponse, error) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	sugar.Debugf("GetAllBackups")
	files, err := ioutil.ReadDir(*backupsDir)

	if err != nil {
		return nil, err
	}

	backups := make([]schellyhook.SchellyResponse, 0)
	for _, fileName := range files {

		id := strings.Split(fileName.Name(), dataStringSeparator)[0]
		dataID := strings.Split(fileName.Name(), dataStringSeparator)[1]

		backupFilePath := *backupsDir + "/" + fileName.Name()
		_, err = os.Open(backupFilePath)
		if err != nil {
			return nil, err
		}

		sizeMB := retriveBackupSize(backupFilePath)

		sugar.Debugf("Found and opened backup file: %s", backupFilePath)
		status := "available"

		sr := schellyhook.SchellyResponse{
			ID:      id,
			DataID:  dataID,
			Status:  status,
			Message: backupFilePath,
			SizeMB:  float64(sizeMB),
		}
		backups = append(backups, sr)
	}

	return backups, nil
}

//GetBackup get an specific backup along with status
func (sb InfluxBackuper) GetBackup(apiID string) (*schellyhook.SchellyResponse, error) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	sugar.Debugf("GetBackup apiID=%s", apiID)

	dirs, err := ioutil.ReadDir(*backupsDir)

	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {

		if strings.Contains(dir.Name(), apiID) && strings.Contains(dir.Name(), dataStringSeparator) {
			id := strings.Split(dir.Name(), dataStringSeparator)[0]
			dataID := strings.Split(dir.Name(), dataStringSeparator)[1]
			sizeMB := retriveBackupSize(*backupsDir + "/" + dir.Name())
			status := "available"

			return &schellyhook.SchellyResponse{
				ID:      id,
				DataID:  dataID,
				Status:  status,
				Message: dir.Name(),
				SizeMB:  float64(sizeMB),
			}, nil
		}
	}

	sugar.Debugf("dumpID not found for apiId %s.", apiID)

	return nil, nil
}

//DeleteBackup removes current backup from underlaying backup storage
func (sb InfluxBackuper) DeleteBackup(apiID string) error {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	sugar.Debugf("DeleteBackup apiID=%s", apiID)

	dirs, err := ioutil.ReadDir(*backupsDir)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		sugar.Debugf("Backup File <Loop>: %s", dir.Name())
		if strings.Contains(dir.Name(), apiID) && strings.Contains(dir.Name(), dataStringSeparator) {
			if _, err := os.Stat(*backupsDir + "/" + dir.Name()); err == nil {
				sugar.Debugf("Found file for apiID reference: %s", apiID)
				err1 := os.RemoveAll(*backupsDir + "/" + dir.Name())
				if err1 != nil {
					return err1
				}
			}
		}
	}

	sugar.Debugf("Delete apiID %s successful", apiID)
	return nil

}

func retriveBackupSize(dir string) int64 {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	files, err := ioutil.ReadDir(dir)

	if err != nil {
		sugar.Error("error retrieving file size: ", err)
		return 0
	}

	for _, file := range files {
		if strings.Contains(file.Name(), "tar.gz") {
			return file.Size()
		}
	}
	return 0
}

func resolveErrorFilePath(apiID string) string {
	return *backupsDir + "/" + apiID + ".err"
}

func mkDirs(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}
