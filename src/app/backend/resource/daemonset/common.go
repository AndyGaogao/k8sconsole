package daemonset

import (
	"github.com/wzt3309/k8sconsole/src/app/backend/resource/common"
	"github.com/wzt3309/k8sconsole/src/app/backend/resource/dataselect"
	"github.com/wzt3309/k8sconsole/src/app/backend/resource/event"
	apps "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
)

// The code below allows to perform complex data section on Daemon Set

type DaemonSetCell apps.DaemonSet

func (self DaemonSetCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Name)
	case dataselect.CreationTimestampProperty:
		return dataselect.StdComparableTime(self.ObjectMeta.CreationTimestamp.Time)
	case dataselect.NamespaceProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Namespace)
	default:
		// if name is not supported then just return a constant dummy value, sort will have no effect.
		return nil
	}
}

func toCells(std []apps.DaemonSet) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = DaemonSetCell(std[i])
	}
	return cells
}

func fromCells(cells []dataselect.DataCell) []apps.DaemonSet {
	std := make([]apps.DaemonSet, len(cells))
	for i := range std {
		std[i] = apps.DaemonSet(cells[i].(DaemonSetCell))
	}
	return std
}

func getStatus(list *apps.DaemonSetList, pods []v1.Pod, events []v1.Event) common.ResourceStatus {
	info := common.ResourceStatus{}
	if list == nil {
		return info
	}

	for _, daemonSet := range list.Items {
		matchingPods := common.FilterPodsByControllerRef(&daemonSet, pods)
		podInfo := common.GetPodInfo(daemonSet.Status.CurrentNumberScheduled,
			&daemonSet.Status.DesiredNumberScheduled, matchingPods)
		warnings := event.GetPodsEventWarnings(events, matchingPods)

		if len(warnings) > 0 {
			info.Failed++
		} else if podInfo.Pending > 0 {
			info.Pending++
		} else {
			info.Running++
		}
	}

	return info
}
