package etcdmain

import (
	"flag"

	"github.com/friendlyhank/etcd-hign/net/pkg/flags"

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

	//member
	fs.Var(
		flags.NewUniqueURLsWithExceptions(embed.DefaultListenPeerURLs, ""),
		"listen-peer-urls",
		"List of URLs to listen on for peer traffic.",
	)

	fs.Var(
		flags.NewUniqueURLsWithExceptions(embed.DefaultListenClientURLs, ""), "listen-client-urls",
		"List of URLs to listen on for client traffic.",
	)

	fs.StringVar(&cfg.ec.Name, "name", cfg.ec.Name, "Human-readable name for this member.")
	fs.StringVar(&cfg.ec.InitialCluster, "initial-cluster", cfg.ec.InitialCluster, "Initial cluster configuration for bootstrapping.")
	return cfg
}

func (cfg *config) parse(arguments []string) error {
	perr := cfg.cf.flagSet.Parse(arguments)
	switch perr {
	case nil:
	}
	var err error
	//解析url等参数
	err = cfg.configFromCmdLine()
	return err
}

func (cfg *config) configFromCmdLine() error {
	cfg.ec.LPUrls = flags.UniqueURLsFromFlag(cfg.cf.flagSet, "listen-peer-urls")
	cfg.ec.LCUrls = flags.UniqueURLsFromFlag(cfg.cf.flagSet, "listen-client-urls")
	return cfg.validate()
}

func (cfg *config) validate() error {
	err := cfg.ec.Validate()
	return err
}
