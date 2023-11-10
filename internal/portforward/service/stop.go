package service

import (
	"context"
	"fmt"
	"os"
)

func (s *Service) Stop() (err error) {
	s.startStopMutex.Lock()
	defer s.startStopMutex.Unlock()

	s.portMutex.RLock()
	serviceNotRunning := s.port == 0
	s.portMutex.RUnlock()
	if serviceNotRunning {
		// TODO replace with goservices.ErrAlreadyStopped
		return nil
	}

	s.logger.Info("stopping")

	s.keepPortCancel()
	<-s.keepPortDoneCh

	return s.cleanup()
}

func (s *Service) cleanup() (err error) {
	s.portMutex.Lock()
	defer s.portMutex.Unlock()

	err = s.portAllower.RemoveAllowedPort(context.Background(), s.port)
	if err != nil {
		return fmt.Errorf("blocking previous port in firewall: %w", err)
	}

	if s.settings.ListeningPort != 0 {
		ctx := context.Background()
		const listeningPort = 0 // 0 to clear the redirection
		err = s.portAllower.RedirectPort(ctx, s.settings.Interface, s.port, listeningPort)
		if err != nil {
			return fmt.Errorf("removing previous port redirection in firewall: %w", err)
		}
	}

	s.port = 0

	filepath := s.settings.Filepath
	s.logger.Info("removing port file " + filepath)
	err = os.Remove(filepath)
	if err != nil {
		return fmt.Errorf("removing port file: %w", err)
	}

	return nil
}
