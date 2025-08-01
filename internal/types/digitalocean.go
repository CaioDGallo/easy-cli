package types

import "github.com/digitalocean/godo"

type DigitalOceanEnvVars struct {
	AppEnvs       map[string]godo.AppVariableDefinition
	ComponentEnvs map[string]godo.AppVariableDefinition
}
