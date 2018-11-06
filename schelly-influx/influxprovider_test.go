package main

import (
	"flag"
	"testing"

	"go.uber.org/zap"
)

func TestRegisterFlags(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	backuper := InfluxBackuper{}

	sugar.Infof("Starting TestRegisterFlags...")

	err := backuper.RegisterFlags()

	if err != nil {
		t.Errorf("Error registering flags: %s", err)
	}

	requiredFlags := []string{"file", "dbname", "host", "port", "username", "password"}
	for _, f := range requiredFlags {
		if flag.Lookup(f) == nil {
			t.Errorf("Imperative flag not defined: %s", f)
		} else {
			sugar.Infof("File flag: %s", flag.Lookup(f))
		}
	}
}
