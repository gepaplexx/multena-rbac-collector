package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"strings"
)

type Configuration struct {
	Kubeconfig            string
	ReconcileLoopInterval int
	LogLevel              string
	Namespace             string
	ConfigMapName         string
	Port                  int
}

func ReadConfig() Configuration {

	viper.SetConfigName("config")
	viper.AddConfigPath("config")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("fatal error with config file: %v", err)
	}
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()
	conf := Configuration{
		Kubeconfig:            viper.GetString("kubeconfig"),
		ReconcileLoopInterval: viper.GetInt("reconcile_loop_interval"),
		LogLevel:              viper.GetString("log_level"),
		Namespace:             viper.GetString("namespace"),
		ConfigMapName:         viper.GetString("config_map_name"),
		Port:                  viper.GetInt("port"),
	}
	if conf.Kubeconfig == "" {
		conf.Kubeconfig = defaultKubeconfig()
	}
	return conf
}

func defaultKubeconfig() string {
	return filepath.Join(homedir.HomeDir(), ".kube", "config")
}
