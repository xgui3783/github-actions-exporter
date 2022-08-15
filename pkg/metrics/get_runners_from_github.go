package metrics

import (
	"context"
	"github-actions-exporter/pkg/config"
	"github-actions-exporter/pkg/skipQuery"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	runnersGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "github_runner_status",
			Help: "runner status",
		},
		[]string{"repo", "os", "name", "id", "busy"},
	)
)

// getRunnersFromGithub - return information about runners and their status for a specific repo
func getRunnersFromGithub() {

	namespace := "getRunnersFromGithub"
	_, ok := skipQuery.SkipMap[namespace];
	if !ok {
		skipQuery.SkipMap[namespace] = 0
	}

	for {
		for _, repo := range config.Github.Repositories.Value() {
			r := strings.Split(repo, "/")
			resp, _, err := client.Actions.ListRunners(context.Background(), r[0], r[1], nil)
			if err != nil {
				skipQuery.SkipMap[namespace] = skipQuery.SkipMap[namespace] + 1
				log.Printf("ListRunners error for %s: %s", repo, err.Error())
			} else {
				for _, runner := range resp.Runners {
					if runner.GetStatus() == "online" {
						runnersGauge.WithLabelValues(repo, *runner.OS, *runner.Name, strconv.FormatInt(runner.GetID(), 10), strconv.FormatBool(runner.GetBusy())).Set(1)
					} else {
						runnersGauge.WithLabelValues(repo, *runner.OS, *runner.Name, strconv.FormatInt(runner.GetID(), 10), strconv.FormatBool(runner.GetBusy())).Set(0)
					}
				}
			}
		}

		if skipQuery.SkipMap[namespace] > skipQuery.Attempts {
			break
		}
		time.Sleep(time.Duration(config.Github.Refresh) * time.Second)
	}
}
