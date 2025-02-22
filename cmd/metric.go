package cmd

import (
	"fmt"
	"strings"

	"github.com/lf-edge/eden/pkg/controller"
	"github.com/lf-edge/eden/pkg/controller/emetric"
	"github.com/lf-edge/eden/pkg/utils"
	"github.com/lf-edge/eve/api/go/metrics"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thediveo/enumflag"
)

var metricTail uint

var metricCmd = &cobra.Command{
	Use:   "metric [field:regexp ...]",
	Short: "Get metrics from a running EVE device",
	Long: `
Scans the ADAM metrics for correspondence with regular expressions requests to json fields.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		assignCobraToViper(cmd)
		viperLoaded, err := utils.LoadConfigFile(configFile)
		if err != nil {
			return fmt.Errorf("error reading config: %s", err.Error())
		}
		if viperLoaded {
			certsIP = viper.GetString("adam.ip")
			adamPort = viper.GetInt("adam.port")
			adamDist = utils.ResolveAbsPath(viper.GetString("adam.dist"))
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctrl, err := controller.CloudPrepare()
		if err != nil {
			log.Fatalf("CloudPrepare: %s", err)
		}
		devFirst, err := ctrl.GetDeviceCurrent()
		if err != nil {
			log.Fatalf("GetDeviceCurrent error: %s", err)
		}
		devUUID := devFirst.GetID()
		follow, err := cmd.Flags().GetBool("follow")
		if err != nil {
			log.Fatalf("Error in get param 'follow'")
		}

		q := make(map[string]string)

		for _, a := range args[0:] {
			s := strings.Split(a, ":")
			q[s[0]] = s[1]
		}

		handleFunc := func(le *metrics.ZMetricMsg) bool {
			if printFields == nil {
				emetric.MetricPrn(le, outputFormat)
			} else {
				emetric.MetricItemPrint(le, printFields).Print()
			}
			return false
		}

		if metricTail > 0 {
			if err = ctrl.MetricChecker(devUUID, q, handleFunc, emetric.MetricTail(metricTail), 0); err != nil {
				log.Fatalf("MetricChecker: %s", err)
			}
		} else {
			if follow {
				// Monitoring of new files
				if err = ctrl.MetricChecker(devUUID, q, handleFunc, emetric.MetricNew, 0); err != nil {
					log.Fatalf("MetricChecker: %s", err)
				}
			} else {
				if err = ctrl.MetricLastCallback(devUUID, q, handleFunc); err != nil {
					log.Fatalf("MetricChecker: %s", err)
				}
			}
		}
	},
}

func metricInit() {
	metricCmd.Flags().UintVar(&metricTail, "tail", 0, "Show only last N lines")
	metricCmd.Flags().StringSliceVarP(&printFields, "out", "o", nil, "Fields to print. Whole message if empty.")
	metricCmd.Flags().BoolP("follow", "f", false, "Monitor changes in selected metrics")
	metricCmd.Flags().Var(
		enumflag.New(&outputFormat, "format", outputFormatIds, enumflag.EnumCaseInsensitive),
		"format",
		"Format to print logs, supports: lines, json")
}
