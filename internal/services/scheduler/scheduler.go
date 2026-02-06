package scheduler

import (
	"log"
	"time"

	"jjmc/internal/models"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	// Import manager to execute tasks on instances
	// We might need to inject the manager to avoid circular deps, or use an interface
)

type Scheduler struct {
	DB       *gorm.DB
	Cron     *cron.Cron
	EntryMap map[string]cron.EntryID // Map ScheduleID to CronEntryID
	// We need a way to execute actions on instances.
	// To avoid circular dependency with 'manager' package, we will likely define an interface
	// or pass a function executioner.
	TaskExecutor func(instanceID string, taskType string, payload string) error
}

func NewScheduler(db *gorm.DB, executor func(instanceID string, taskType string, payload string) error) *Scheduler {
	return &Scheduler{
		DB:           db,
		Cron:         cron.New(),
		EntryMap:     make(map[string]cron.EntryID),
		TaskExecutor: executor,
	}
}

func (s *Scheduler) Start() {
	s.Cron.Start()
	s.LoadSchedules()
}

func (s *Scheduler) Stop() {
	s.Cron.Stop()
}

func (s *Scheduler) LoadSchedules() {
	var schedules []models.Schedule
	if err := s.DB.Find(&schedules).Error; err != nil {
		log.Printf("Failed to load schedules: %v", err)
		return
	}

	for _, schedule := range schedules {
		if schedule.Enabled {
			s.AddJob(schedule)
		}
	}
}

func (s *Scheduler) AddJob(schedule models.Schedule) error {
	// Remove existing if any (update case)
	if entryID, exists := s.EntryMap[schedule.ID]; exists {
		s.Cron.Remove(entryID)
	}

	if !schedule.Enabled {
		return nil
	}

	entryID, err := s.Cron.AddFunc(schedule.CronExpression, func() {
		log.Printf("Executing schedule %s for instance %s", schedule.Name, schedule.InstanceID)

		// Update LastRun
		s.DB.Model(&models.Schedule{}).Where("id = ?", schedule.ID).Update("last_run", time.Now().Unix())

		if err := s.TaskExecutor(schedule.InstanceID, schedule.Type, schedule.Payload); err != nil {
			log.Printf("Failed to execute schedule %s: %v", schedule.Name, err)
		}
	})

	if err != nil {
		return err
	}

	s.EntryMap[schedule.ID] = entryID
	return nil
}

func (s *Scheduler) RemoveJob(scheduleID string) {
	if entryID, exists := s.EntryMap[scheduleID]; exists {
		s.Cron.Remove(entryID)
		delete(s.EntryMap, scheduleID)
	}
}
