package controller

import (
	"bytes"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"text/template"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/pkg/rulefmt"
	"github.com/dbsystel/prometheus-config-controller/prometheus"
	"k8s.io/api/core/v1"
)

type Controller struct {
	logger log.Logger
	p      prometheus.APIClient
}

// create new Controller instance
func New(p prometheus.APIClient, logger log.Logger) *Controller {
	controller := &Controller{}
	controller.logger = logger
	controller.p = p
	return controller
}

// do something when a configmap created
func (c *Controller) Create(obj interface{}) {
	configmapObj := obj.(*v1.ConfigMap)
	id, _     := configmapObj.Annotations["prometheus.net/id"]
	rule, _   := configmapObj.Annotations["prometheus.net/rule"]
	job, _    := configmapObj.Annotations["prometheus.net/job"]
	promConfig, _ := configmapObj.Annotations["prometheus.net/config"]
	key, _    := configmapObj.Annotations["prometheus.net/key"]
	isPrometheusRule, _   := strconv.ParseBool(rule)
	isPrometheusJob, _    := strconv.ParseBool(job)
	isPrometheusConfig, _ := strconv.ParseBool(promConfig)
	prometheusId,_ := strconv.Atoi(id)

	if prometheusId == c.p.Id {
		var err error

		if isPrometheusRule && c.validateRules(configmapObj) {
			err = c.createRules(configmapObj)
		} else if isPrometheusJob && c.validateJobs(configmapObj) {
			c.createJobs(configmapObj)
			err = c.buildConfig()
		} else if isPrometheusConfig && key == c.p.Key && c.validateConfig(configmapObj) {
			c.createConfig(configmapObj)
			err = c.buildConfig()
		} else {
			if !isPrometheusRule && !isPrometheusConfig && !isPrometheusJob {
				err = errors.New("unknown config type")
			} else {
				err = errors.New("invalid config data")
			}
		}

		if err == nil {
			err, _ = c.p.Reload()
			if err != nil {
				level.Error(c.logger).Log("msg", "Failed to reload prometheus.yml",
					"err", err.Error(),
					"namespace", configmapObj.Namespace,
					"name", configmapObj.Name,
				)
			} else {
				level.Info(c.logger).Log("msg", "Succeeded: Reloaded Prometheus")
			}
		} else if err.Error() == "unknown config type" {
			level.Debug(c.logger).Log("msg", "Skipping configmap:" + configmapObj.Name, "namespace", configmapObj.Namespace)
		} else {
			level.Error(c.logger).Log("msg", "Failed to create", "namespace", configmapObj.Namespace, "name", configmapObj.Name)
		}
	} else {
		level.Debug(c.logger).Log("msg", "Skipping configmap:" + configmapObj.Name)
	}
}

// validate rules and save them into storage
func (c *Controller) createRules(configmapObj *v1.ConfigMap) error {
	for k, v := range configmapObj.Data {
		level.Info(c.logger).Log(
			"msg", "Creating rule: " + k,
			"namespace", configmapObj.Namespace,
			"name", configmapObj.Name,
			)
		re,_ := regexp.Compile("^groups:(\\s*.*)*")
		if !re.MatchString(v) {
			v = "groups:\n" + v
		}

		path := c.p.ConfigPath + "/rules/"
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err = os.MkdirAll(path, 0766)
			if err != nil {
				level.Error(c.logger).Log("msg", "Failed to create directory", "err", err)
			}
		}

		filename := path + configmapObj.Namespace + "-" + configmapObj.Name + "-"
		ioErr := ioutil.WriteFile(filename + k, []byte(v), 0644)
		if ioErr != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to create rules",
				"name", configmapObj.Name,
				"namespace", configmapObj.Namespace,
				"err", ioErr,
			)
			return ioErr
		}
	}
	return nil
}

// save jobs into storage
func (c *Controller) createJobs(configmapObj *v1.ConfigMap) {
	var err error
	for k, v := range configmapObj.Data {
		level.Info(c.logger).Log(
			"msg", "Creating job: " + k,
			"namespace", configmapObj.Namespace,
			"name", configmapObj.Name,
		)

		path := c.p.ConfigPath + "/jobs/"
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err = os.MkdirAll(path, 0766)
			if err != nil {
				level.Error(c.logger).Log("msg", "Failed to create directory", "err", err)
			}
		}

		filename := path + configmapObj.Namespace + "-" + configmapObj.Name + "-"

		err = ioutil.WriteFile(filename + k, []byte(v), 0644)
		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to create job",
				"name", configmapObj.Name,
				"namespace", configmapObj.Namespace,
				"err", err,
			)
		}
	}
}

// save config template into storage
func (c *Controller) createConfig(configmapObj *v1.ConfigMap) {
	for k, v := range configmapObj.Data {
		level.Info(c.logger).Log(
			"msg", "Creating config: " + k,
			"namespace", configmapObj.Namespace,
			"name", configmapObj.Name,
		)
		err := ioutil.WriteFile(c.p.ConfigTemplate, []byte(v), 0644)
		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to create config",
				"name", configmapObj.Name,
				"namespace", configmapObj.Namespace,
				"err", err,
			)
		}
	}
}

// do something when a configmap updated
func (c *Controller) Update(oldobj, newobj interface{}) {
	newConfigmapObj := newobj.(*v1.ConfigMap)
	oldConfigmapObj := oldobj.(*v1.ConfigMap)
	newId, _  := newConfigmapObj.Annotations["prometheus.net/id"]
	oldId, _  := oldConfigmapObj.Annotations["prometheus.net/id"]
	rule, _   := newConfigmapObj.Annotations["prometheus.net/rule"]
	job, _    := newConfigmapObj.Annotations["prometheus.net/job"]
	promConfig, _ := newConfigmapObj.Annotations["prometheus.net/config"]
	key, _    := newConfigmapObj.Annotations["prometheus.net/key"]
	isPrometheusRule, _   := strconv.ParseBool(rule)
	isPrometheusJob, _    := strconv.ParseBool(job)
	isPrometheusConfig, _ := strconv.ParseBool(promConfig)
	newPrometheusId, _ := strconv.Atoi(newId)
	oldPrometheusId, _ := strconv.Atoi(oldId)

	if newPrometheusId == oldPrometheusId && noDifference(oldConfigmapObj, newConfigmapObj) {
		level.Debug(c.logger).Log("msg", "Skipping automatically updated configmap:" + newConfigmapObj.Name)
		return
	}

	var err error

	if isPrometheusRule && c.validateRules(newConfigmapObj) {
		if oldPrometheusId == c.p.Id{
			c.deleteRules(oldConfigmapObj)
		}
		if newPrometheusId == c.p.Id {
			err = c.createRules(newConfigmapObj)
		}
	} else if isPrometheusJob  {
		if oldPrometheusId == c.p.Id {
			c.deleteJobs(oldConfigmapObj)
		}
		if newPrometheusId == c.p.Id {
			if c.validateJobs(newConfigmapObj) {
				c.createJobs(newConfigmapObj)
			} else {
				level.Info(c.logger).Log("msg", "The new job is not valid. Recover the old one.")
				c.createJobs(oldConfigmapObj)
			}

		}
		err = c.buildConfig()
	} else if isPrometheusConfig && key == c.p.Key  && c.validateConfig(newConfigmapObj){
		c.createConfig(newConfigmapObj)
		err = c.buildConfig()
	} else {
		if !isPrometheusRule && !isPrometheusConfig && !isPrometheusJob {
			err = errors.New("unknown config type")
		} else {
			err = errors.New("invalid config data")
		}
	}

	if err == nil {
		err, _ = c.p.Reload()
		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to reload prometheus.yml",
				"err", err.Error(),
				"namespace", newConfigmapObj.Namespace,
				"name", newConfigmapObj.Name,
				)
		} else {
			level.Info(c.logger).Log("msg", "Succeeded: Reloaded Prometheus", "namespace")
		}
	} else if err.Error() == "unknown config type" {
		level.Debug(c.logger).Log("msg", "Skipping configmap:" + newConfigmapObj.Name)
	} else {
		level.Error(c.logger).Log("msg", "Failed to create",
			"namespace", newConfigmapObj.Namespace,
			"name", newConfigmapObj.Name,
			)
	}


}

// check if rules are valid
func (c *Controller) validateRules(configmapObj *v1.ConfigMap) bool {
	for k, v := range configmapObj.Data {
		re,_ := regexp.Compile("^groups:(\\s*.*)*")
		if !re.MatchString(v) {
			v = "groups:\n" + v
		}
		_, fmtErr := rulefmt.Parse([]byte(v))
		if fmtErr != nil {
			for _, err := range fmtErr {
				level.Error(c.logger).Log(
					"msg", "Invalid rule: " + k,
					"name", configmapObj.Name,
					"namespace", configmapObj.Namespace,
					"err", err,
				)
			}
			return false
		}
	}
	return true
}

// check if jobs are valid
func (c *Controller) validateJobs(configmapObj *v1.ConfigMap) bool {
	configTemplate := "scrape_configs:\n{{ .Jobs }}"

	t, err := template.New("prometheus.yaml").Parse(configTemplate)
	if err != nil {
		level.Error(c.logger).Log("msg", "Failed to parse template", "err", err.Error())
	}

	for k, v := range configmapObj.Data {
		var prometheusConfig prometheus.PrometheusConfig
		prometheusConfig.Jobs = c.readJobs() + v
		var tpl bytes.Buffer
		err = t.Execute(&tpl, prometheusConfig)
		_, configErr := config.Load(tpl.String())
		if configErr != nil {
			level.Error(c.logger).Log(
				"msg", "Invalid job: " + k,
				"name", configmapObj.Name,
				"namespace", configmapObj.Namespace,
				"err", configErr,
			)
			return false
		}

	}
	return true
}

// check if config template is valid
func (c *Controller) validateConfig(configmapObj *v1.ConfigMap) bool {
	for k, v := range configmapObj.Data {
		t, err := template.New("prometheus.yaml").Parse(v)
		if err != nil {
			level.Error(c.logger).Log("msg", "Failed to parse template", "err", err.Error())
		}

		var prometheusConfig prometheus.PrometheusConfig
		prometheusConfig.Jobs = c.readJobs()
		var tpl bytes.Buffer
		err = t.Execute(&tpl, prometheusConfig)
		_, configErr := config.Load(tpl.String())
		if configErr != nil {
			level.Error(c.logger).Log(
				"msg", "Invalid Config: " + k,
				"name", configmapObj.Name,
				"namespace", configmapObj.Namespace,
				"err", configErr,
			)
			return false
		}
	}
	return true
}

// do something when a configmap deleted
func (c *Controller) Delete(obj interface{}) {
	configmapObj := obj.(*v1.ConfigMap)
	id, _    := configmapObj.Annotations["prometheus.net/id"]
	rule, _ := configmapObj.Annotations["prometheus.net/rule"]
	job, _    := configmapObj.Annotations["prometheus.net/job"]
	isPrometheusRule, _   := strconv.ParseBool(rule)
	isPrometheusJob, _    := strconv.ParseBool(job)
	prometheusId,_ := strconv.Atoi(id)

	if prometheusId == c.p.Id {
		var err error
		if isPrometheusRule {
			c.deleteRules(configmapObj)
		} else if isPrometheusJob {
			c.deleteJobs(configmapObj)
			err = c.buildConfig()
		} else {
			err = errors.New("unknown config type")
		}

		if err == nil {
			err, _ = c.p.Reload()
			if err != nil {
				level.Error(c.logger).Log(
					"msg", "Failed to reload prometheus.yml",
					"err", err.Error(),
					"namespace", configmapObj.Namespace,
					"name", configmapObj.Name,
					)
			} else {
				level.Info(c.logger).Log("msg", "Succeeded: Reloaded Prometheus")
			}
		} else if err.Error() == "unknown config type" {
			level.Debug(c.logger).Log("msg", "Skipping configmap:" + configmapObj.Name)
		} else {
			level.Error(c.logger).Log("msg", "Failed to delete", "namespace", configmapObj.Namespace, "name", configmapObj.Name)
		}
	} else {
		level.Debug(c.logger).Log("msg", "Skipping configmap:" + configmapObj.Name)
	}
}

// remove rule files from storage
func (c *Controller) deleteRules(configmapObj *v1.ConfigMap) {
	for k := range configmapObj.Data {
		level.Info(c.logger).Log(
			"msg", "Deleting rule: " + k,
			"namespace", configmapObj.Namespace,
			"name", configmapObj.Name,
		)
		filename := configmapObj.Namespace + "-" + configmapObj.Name + "-"
		ioErr := os.Remove(c.p.ConfigPath + "/rules/" + filename + k)
		if ioErr != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to delete rules",
				"name", configmapObj.Name,
				"namespace", configmapObj.Namespace,
				"err", ioErr,
			)
			break
		}
	}
}

// format config file from jobs files and config template
func (c *Controller) buildConfig() error {
	configTemplate, err := ioutil.ReadFile(c.p.ConfigTemplate)
	if err != nil {
		level.Error(c.logger).Log("msg", "Failed to read template", "err", err.Error(), "file", c.p.ConfigTemplate)
	}

	jobs := c.readJobs()

	var prometheusConfig prometheus.PrometheusConfig
	prometheusConfig.Jobs = jobs

	t, err := template.New("prometheus.yaml").Parse(string(configTemplate))
	if err != nil {
		level.Error(c.logger).Log("msg", "Failed to parse template", "err", err.Error())
	}

	f, err := os.Create(c.p.ConfigPath + "/prometheus.yml")
	if err != nil {
		level.Error(c.logger).Log("msg", "failed to create prometheus.yaml", "err", err.Error())
	}
	defer f.Close()
	err = t.Execute(f, prometheusConfig)
	if err != nil {
		level.Error(c.logger).Log("err", err.Error())
	}


	return err
}

// read job files from storage
func (c *Controller) readJobs() string {
	jobfiles, err := filepath.Glob(c.p.ConfigPath + "/jobs/*")
	if err != nil {
		level.Error(c.logger).Log("msg", "Failed to read jobs", "err", err.Error())
	}
	jobs := ""
	for _, jobfile := range jobfiles {
		job, err := ioutil.ReadFile(jobfile)
		if err != nil {
			level.Error(c.logger).Log("msg", "Failed to read job", "job", jobfile, "err", err.Error())
		}
		jobs = jobs + string(job) + "\n"
	}
	return jobs
}

// delete job files from storage
func (c *Controller) deleteJobs(configmapObj *v1.ConfigMap) {
	var err error
	for k := range configmapObj.Data {
		level.Info(c.logger).Log(
			"msg", "Deleting job: " + k,
			"namespace", configmapObj.Namespace,
			"name", configmapObj.Name,
		)
		filename := configmapObj.Namespace + "-" + configmapObj.Name + "-"
		err = os.Remove(c.p.ConfigPath + "/jobs/" + filename + k)
		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to delete job",
				"name", configmapObj.Name,
				"namespace", configmapObj.Namespace,
				"err", err,
			)
		}
	}
}

// are two configmaps same
func noDifference(newConfigMap *v1.ConfigMap, oldConfigMap *v1.ConfigMap) bool {
	if len(newConfigMap.Data) != len(oldConfigMap.Data) {
		return false
	}
	for k, v := range newConfigMap.Data {
		if v != oldConfigMap.Data[k]{
			return false
		}
	}
	if len(newConfigMap.Annotations) != len(oldConfigMap.Annotations) {
		return false
	}
	for k, v := range newConfigMap.Annotations {
		if v != oldConfigMap.Annotations[k] {
			return false
		}
	}
    return true
}
