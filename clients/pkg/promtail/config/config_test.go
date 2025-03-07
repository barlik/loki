package config

import (
	"fmt"
	"net/url"
	"testing"

	dskitflagext "github.com/grafana/dskit/flagext"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/grafana/loki/clients/pkg/promtail/client"

	"github.com/grafana/loki/pkg/util/flagext"
)

const testFile = `
clients:
  - external_labels:
      cluster: dev1
    url: https://1:shh@example.com/loki/api/v1/push
  - external_labels:
      cluster: prod1
    url: https://1:shh@example.com/loki/api/v1/push
scrape_configs:
  - job_name: kubernetes-pods-name
    kubernetes_sd_configs:
      - role: pod
  - job_name: system
    static_configs:
    - targets:
      - localhost
      labels:
        job: varlogs
limits_config:
  readline_rate: 100
  readline_burst: 200
options:
  stream_lag_labels: foo
`

func Test_Load(t *testing.T) {
	var dst Config
	err := yaml.Unmarshal([]byte(testFile), &dst)
	require.Nil(t, err)
}

func Test_RateLimitLoad(t *testing.T) {
	var dst Config
	err := yaml.Unmarshal([]byte(testFile), &dst)
	require.Nil(t, err)
	config := dst.LimitsConfig
	require.Equal(t, float64(100), config.ReadlineRate)
	require.Equal(t, 200, config.ReadlineBurst)
}

func TestConfig_Setup(t *testing.T) {
	for i, tt := range []struct {
		in       Config
		expected Config
	}{
		{
			Config{
				ClientConfig: client.Config{
					ExternalLabels: flagext.LabelSet{LabelSet: model.LabelSet{"foo": "bar"}},
				},
				ClientConfigs: []client.Config{
					{
						ExternalLabels: flagext.LabelSet{LabelSet: model.LabelSet{"client1": "1"}},
					},
					{
						ExternalLabels: flagext.LabelSet{LabelSet: model.LabelSet{"client2": "2"}},
					},
				},
				Options: Options{
					StreamLagLabels: []string{},
				},
			},
			Config{
				ClientConfig: client.Config{
					ExternalLabels: flagext.LabelSet{LabelSet: model.LabelSet{"foo": "bar"}},
				},
				ClientConfigs: []client.Config{
					{
						ExternalLabels: flagext.LabelSet{LabelSet: model.LabelSet{"client1": "1", "foo": "bar"}},
					},
					{
						ExternalLabels: flagext.LabelSet{LabelSet: model.LabelSet{"client2": "2", "foo": "bar"}},
					},
				},
				Options: Options{
					StreamLagLabels: []string{},
				},
			},
		},
		{
			Config{
				ClientConfig: client.Config{
					ExternalLabels: flagext.LabelSet{LabelSet: model.LabelSet{"foo": "bar"}},
					URL:            dskitflagext.URLValue{URL: mustURL("http://foo")},
				},
				ClientConfigs: []client.Config{
					{
						ExternalLabels: flagext.LabelSet{LabelSet: model.LabelSet{"client1": "1"}},
					},
					{
						ExternalLabels: flagext.LabelSet{LabelSet: model.LabelSet{"client2": "2"}},
					},
				},
				Options: Options{
					StreamLagLabels: []string{},
				},
			},
			Config{
				ClientConfig: client.Config{
					ExternalLabels: flagext.LabelSet{LabelSet: model.LabelSet{"foo": "bar"}},
					URL:            dskitflagext.URLValue{URL: mustURL("http://foo")},
				},
				ClientConfigs: []client.Config{
					{
						ExternalLabels: flagext.LabelSet{LabelSet: model.LabelSet{"client1": "1", "foo": "bar"}},
					},
					{
						ExternalLabels: flagext.LabelSet{LabelSet: model.LabelSet{"client2": "2", "foo": "bar"}},
					},
					{
						ExternalLabels: flagext.LabelSet{LabelSet: model.LabelSet{"foo": "bar"}},
						URL:            dskitflagext.URLValue{URL: mustURL("http://foo")},
					},
				},
				Options: Options{
					StreamLagLabels: []string{},
				},
			},
		},
	} {
		tt := tt
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			tt.in.Setup()
			require.Equal(t, tt.expected, tt.in)
		})
	}
}

func mustURL(u string) *url.URL {
	res, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	return res
}
