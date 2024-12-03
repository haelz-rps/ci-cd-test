package dataextractor

import (
	"strings"

	"github.com/google/uuid"
)

// Connector is a source that has been activated and linked to an Airbyte destination.
// For the moment we assume a single destination.
type Connector struct {
	ID                       *uuid.UUID             `json:"id"`
	SourceId                 *uuid.UUID             `json:"sourceId"`
	WarehouseDatabase        string                 `json:"warehouseDatabase"`
	Name                     string                 `json:"name"`
	ScheduleType             string                 `json:"scheduleType"`
	Tables                   []*Table               `json:"tables"`
	SourceConfiguration      map[string]interface{} `json:"sourceConfiguration"`
	DestinationConfiguration map[string]interface{} `json:"destinationConfiguration"`
	ResourceRequirements     *ResourceRequirements  `json:"resourceRequirements,omitempty"`
	Schedule                 *Schedule              `json:"schedule,omitempty"`
}

func (c Connector) GetWarehouseDatabaseName() string {
	return strings.ReplaceAll(c.SourceId.String(), "-", "")
}

// ResourceRequirements is a struct that represents Kubernetes resources.
type ResourceRequirements struct {
	CpuLimit      *string `json:"cpu_limit,omitempty"`
	CpuRequest    *string `json:"cpu_request,omitempty"`
	MemoryLimit   *string `json:"memory_limit,omitempty"`
	MemoryRequest *string `json:"memory_request,omitempty"`
}

type BasicSchedule struct {
	TimeUnit string `json:"timeUnit"`
	Units    int64  `json:"units"`
}

type CronSchedule struct {
	TimeZone   string `json:"timeZone"`
	Expression string `json:"expression"`
}

type Schedule struct {
	BasicSchedule *BasicSchedule `json:"basicSchedule"`
	CronSchedule  *CronSchedule  `json:"cronSchedule"`
}
