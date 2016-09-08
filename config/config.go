package config

import (
	"flag"
	"os/user"
)

func LoadGlogConfig() {
	// We can configure the flags here rather than from command line arguments
	flag.Parse();

	// TODO: set the log dir later.
	u, _ := user.Current()
	flag.Lookup("log_dir").Value.Set(u.HomeDir);
	//flag.Lookup("logtostderr").Value.Set("true");

	// This seem to be not mandatory.
	//defer glog.Flush();
}
