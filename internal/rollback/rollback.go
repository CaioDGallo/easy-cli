package rollback

import (
	"context"
	"fmt"

	"github.com/CaioDGallo/easy-cli/internal/logger"
	"github.com/sirupsen/logrus"
)

type Action struct {
	Name     string
	Rollback func(ctx context.Context) error
}

type Manager struct {
	actions []Action
}

func NewManager() *Manager {
	return &Manager{
		actions: make([]Action, 0),
	}
}

func (m *Manager) AddAction(name string, rollbackFn func(ctx context.Context) error) {
	m.actions = append(m.actions, Action{
		Name:     name,
		Rollback: rollbackFn,
	})
}

func (m *Manager) ExecuteRollback(ctx context.Context) error {
	log := logger.WithFields(logrus.Fields{
		"component": "rollback",
		"actions":   len(m.actions),
	})

	log.Info("Starting rollback process")

	var rollbackErrors []error

	for i := len(m.actions) - 1; i >= 0; i-- {
		action := m.actions[i]
		actionLog := log.WithField("action", action.Name)

		actionLog.Info("Executing rollback action")

		if err := action.Rollback(ctx); err != nil {
			actionLog.WithError(err).Error("Rollback action failed")
			rollbackErrors = append(rollbackErrors, fmt.Errorf("failed to rollback %s: %w", action.Name, err))
		} else {
			actionLog.Info("Rollback action completed successfully")
		}
	}

	if len(rollbackErrors) > 0 {
		log.WithField("errors", len(rollbackErrors)).Error("Rollback completed with errors")
		return fmt.Errorf("%d rollback actions failed: %v", len(rollbackErrors), rollbackErrors)
	}

	log.Info("Rollback completed successfully")
	return nil
}
