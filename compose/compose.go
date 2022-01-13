package compose

import (
	"io/ioutil"
	"log"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type ComposeSpec struct {
	Services map[string]interface{}
	Version  string
}

type ServiceSpec struct {
	Restart    string   `yaml:"restart"`
	Image      string   `yaml:"image"`
	Volumes    []string `yaml:"volumes"`
	Entrypoint []string `yaml:"entrypoint"`
	WorkingDir string   `yaml:"working_dir"`
	Ports      []string `yaml:"ports"`
}

func Create() error {
	yamlFile, err := ioutil.ReadFile("config/harness-delegate.yml")
	if err != nil {
		return errors.Wrap(err, "failed to read config/harness-delegate.yml")
	}

	c := &ComposeSpec{}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return errors.Wrap(err, "failed to parse config/harness-delegate.yml")
	}

	for _, svc := range c.Services {
		s, ok := svc.(map[interface{}]interface{})
		if ok {
			s["network_mode"] = "host"
			svc = s
		}
	}
	c.Services["drone-runner-aws"] = getRunnerSpec()

	d, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalln(err)
	}
	err = ioutil.WriteFile("docker-compose.yml", d, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write output docker compose file")
	}
	return nil
}

func getRunnerSpec() ServiceSpec {
	return ServiceSpec{
		Restart:    "unless-stopped",
		Image:      "drone/drone-runner-aws",
		Volumes:    []string{".:/runner"},
		Entrypoint: []string{"/bin/drone-runner-aws", "delegate"},
		WorkingDir: "/runner",
		Ports:      []string{"3000:3000"},
	}
}
