package etcdmain

import (
	"flag"

	"github.com/friendlyhank/etcd-hign/net/embed"
)

type config struct {
	ec embed.Config
	cf configFlags
}

type configFlags struct {
	flagSet *flag.FlagSet
}

func newConfig() *config {
	cfg := &config{
		ec: *embed.NewConfig(),
	}
	cfg.cf = configFlags{
		flagSet: flag.NewFlagSet("etcd", flag.ContinueOnError),
	}

	fs := cfg.cf.flagSet
	fs.StringVar(&cfg.ec.Name, "name", cfg.ec.Name, "Human-readable name for this member.")
	fs.StringVar(&cfg.ec.InitialCluster, "initial-cluster", cfg.ec.InitialCluster, "Initial cluster configuration for bootstrapping.")
	return cfg
}

func (cfg *config) parse(arguments []string) error {
	perr := cfg.cf.flagSet.Parse(arguments)
	switch perr {
	case nil:
	}
	return nil
}
