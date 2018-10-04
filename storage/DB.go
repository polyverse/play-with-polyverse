package storage

import (
	"github.com/polyverse/play-with-polyverse/pwd/types"
)

type DB struct {
	Sessions         map[string]*types.Session         `json:"sessions"`
	Instances        map[string]*types.Instance        `json:"instances"`
	Clients          map[string]*types.Client          `json:"clients"`
	WindowsInstances map[string]*types.WindowsInstance `json:"windows_instances"`
	LoginRequests    map[string]*types.LoginRequest    `json:"login_requests"`
	Users            map[string]*types.User            `json:"user"`
	Playgrounds      map[string]*types.Playground      `json:"playgrounds"`

	WindowsInstancesBySessionId map[string][]string `json:"windows_instances_by_session_id"`
	InstancesBySessionId        map[string][]string `json:"instances_by_session_id"`
	ClientsBySessionId          map[string][]string `json:"clients_by_session_id"`
	UsersByProvider             map[string]string   `json:"users_by_providers"`
}
