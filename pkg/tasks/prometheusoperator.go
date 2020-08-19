// Copyright 2018 The Cluster Monitoring Operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tasks

import (
	"github.com/openshift/cluster-monitoring-operator/pkg/client"
	"github.com/openshift/cluster-monitoring-operator/pkg/manifests"
	"github.com/pkg/errors"
)

type PrometheusOperatorTask struct {
	client  *client.Client
	factory *manifests.Factory
}

func NewPrometheusOperatorTask(client *client.Client, factory *manifests.Factory) *PrometheusOperatorTask {
	return &PrometheusOperatorTask{
		client:  client,
		factory: factory,
	}
}

func (t *PrometheusOperatorTask) Run() error {
	// Remove existing ValidatingWebhookConfigration resource in case we're downgrading from 4.6.
	// See https://bugzilla.redhat.com/show_bug.cgi?id=1868304
	err := t.client.DeleteValidatingWebhookConfiguration("prometheusrules.openshift.io")
	if err != nil {
		return errors.Wrap(err, "reconciling Prometheus Operator ValidatingWebhookConfiguration failed")
	}

	sa, err := t.factory.PrometheusOperatorServiceAccount()
	if err != nil {
		return errors.Wrap(err, "initializing Prometheus Operator ServiceAccount failed")
	}

	err = t.client.CreateOrUpdateServiceAccount(sa)
	if err != nil {
		return errors.Wrap(err, "reconciling Prometheus Operator ServiceAccount failed")
	}

	cr, err := t.factory.PrometheusOperatorClusterRole()
	if err != nil {
		return errors.Wrap(err, "initializing Prometheus Operator ClusterRole failed")
	}

	err = t.client.CreateOrUpdateClusterRole(cr)
	if err != nil {
		return errors.Wrap(err, "reconciling Prometheus Operator ClusterRole failed")
	}

	crb, err := t.factory.PrometheusOperatorClusterRoleBinding()
	if err != nil {
		return errors.Wrap(err, "initializing Prometheus Operator ClusterRoleBinding failed")
	}

	err = t.client.CreateOrUpdateClusterRoleBinding(crb)
	if err != nil {
		return errors.Wrap(err, "reconciling Prometheus Operator ClusterRoleBinding failed")
	}

	svc, err := t.factory.PrometheusOperatorService()
	if err != nil {
		return errors.Wrap(err, "initializing Prometheus Operator Service failed")
	}

	err = t.client.CreateOrUpdateService(svc)
	if err != nil {
		return errors.Wrap(err, "reconciling Prometheus Operator Service failed")
	}

	clusterNamespaces, err := t.client.NamespacesToMonitor()
	if err != nil {
		return errors.Wrap(err, "listing namespaces to monitor failed")
	}

	d, err := t.factory.PrometheusOperatorDeployment(clusterNamespaces)
	if err != nil {
		return errors.Wrap(err, "initializing Prometheus Operator Deployment failed")
	}

	err = t.client.CreateOrUpdateDeployment(d)
	if err != nil {
		return errors.Wrap(err, "reconciling Prometheus Operator Deployment failed")
	}

	err = t.client.WaitForPrometheusOperatorCRDsReady()
	if err != nil {
		return errors.Wrap(err, "waiting for Prometheus CRDs to become available failed")
	}

	smpo, err := t.factory.PrometheusOperatorServiceMonitor()
	if err != nil {
		return errors.Wrap(err, "initializing Prometheus Operator ServiceMonitor failed")
	}

	err = t.client.CreateOrUpdateServiceMonitor(smpo)
	return errors.Wrap(err, "reconciling Prometheus Operator ServiceMonitor failed")
}
