package modules

import (
	"github.com/redhat-gpe/agnostics/api/v1"
	"testing"
)

func TestLabelPredicates(t *testing.T) {
	var (
		labels map[string]string
		clouds map[string]v1.Cloud
		result map[string]v1.Cloud
	)
	clouds = map[string]v1.Cloud{
		"openstack-1": v1.Cloud{
			Name: "openstack-1",
			Labels: map[string]string{
				"type": "osp",
				"region": "na",
				"purpose": "development",
			},
		},
		"openstack-2": v1.Cloud{
			Name: "openstack-2",
			Labels: map[string]string{
				"type": "osp",
				"region": "emea",
				"purpose": "ILT",
			},
		},
		"openstack-3": v1.Cloud{
			Name: "openstack-3",
			Labels: map[string]string{
				"type": "osp",
				"region": "emea",
				"purpose": "ELT",
			},
		},
	}

	labels = map[string]string{
		"type": "osp",
	}
	result = LabelPredicates(clouds, labels)
	if len(result) != len(clouds) {
		t.Error(clouds, labels, result)
	}

	labels = map[string]string{
		"region": "emea",
	}
	result = LabelPredicates(clouds, labels)
	if len(result) != 2 {
		t.Error(clouds, labels, result)
	}
	if _, ok := result["openstack-2"]; ! ok {
		t.Error(clouds, labels, result)
	}
	if _, ok := result["openstack-3"]; ! ok {
		t.Error(clouds, labels, result)
	}

	labels = map[string]string{
		"region": "emea",
		"purpose": "ILT",
	}
	result = LabelPredicates(clouds, labels)
	if len(result) != 1 {
		t.Error(clouds, labels, result)
	}
	if _, ok := result["openstack-2"]; ! ok {
		t.Error(clouds, labels, result)
	}

	labels = map[string]string{
		"region": "emea",
		"purpose": "ILT",
		"foo": "bar",
	}
	result = LabelPredicates(clouds, labels)
	if len(result) != 0 {
		t.Error(clouds, labels, result)
	}
}

func TestLabelPriorities(t *testing.T) {
	var (
		preferences map[string]string
		clouds map[string]v1.Cloud
		result []v1.Cloud
	)

	clouds = map[string]v1.Cloud{
		"openstack-1": v1.Cloud{
			Name: "openstack-1",
			Labels: map[string]string{
				"type": "osp",
				"region": "na",
				"purpose": "development",
			},
		},
		"openstack-2": v1.Cloud{
			Name: "openstack-2",
			Labels: map[string]string{
				"type": "osp",
				"region": "apac",
				"purpose": "ILT",
			},
		},
		"openstack-3": v1.Cloud{
			Name: "openstack-3",
			Labels: map[string]string{
				"type": "osp",
				"region": "emea",
				"purpose": "ELT",
			},
		},
		"openstack-4": v1.Cloud{
			Name: "openstack-4",
			Labels: map[string]string{
				"type": "osp",
				"region": "sa",
				"purpose": "ELT",
			},
		},
		"openstack-5": v1.Cloud{
			Name: "openstack-5",
			Labels: map[string]string{
				"type": "osp",
				"region": "emea",
				"purpose": "development",
			},
		},
		"openstack-6": v1.Cloud{
			Name: "openstack-6",
			Labels: map[string]string{
				"type": "osp",
				"region": "na",
				"purpose": "ILT",
			},
		},
	}
	preferences = map[string]string{
		"region": "sa",
	}

	result = LabelPriorities(clouds, preferences)

	if len(result) != len(clouds) {
		t.Error(preferences, result)
	}
	if result[0].Name != "openstack-4" {
		t.Error(preferences, result)
	}

	preferences = map[string]string{
		"region": "na",
		"purpose": "development",
	}

	result = LabelPriorities(clouds, preferences)

	if len(result) != len(clouds) {
		t.Error(preferences, result)
	}
	if result[0].Name != "openstack-1" {
		t.Error(preferences, result)
	}
	if result[1].Name != "openstack-5" && result[1].Name != "openstack-6" {
		t.Error(preferences, result)
	}
	if result[2].Name != "openstack-5" && result[2].Name != "openstack-6" {
		t.Error(preferences, result)
	}
}
