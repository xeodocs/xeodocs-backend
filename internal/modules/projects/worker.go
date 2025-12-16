package projects

import (
	"log"
	"time"

	"github.com/xeodocs/xeodocs-backend/internal/modules/configurations"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	"github.com/xeodocs/xeodocs-backend/internal/shared/github"
)

type SyncWorker struct {
	projectsRepo *Repository
	configRepo   *configurations.Repository
	ghService    *github.Service
	config       *config.Config
}

func NewSyncWorker(pRepo *Repository, cRepo *configurations.Repository, gh *github.Service, cfg *config.Config) *SyncWorker {
	return &SyncWorker{
		projectsRepo: pRepo,
		configRepo:   cRepo,
		ghService:    gh,
		config:       cfg,
	}
}

func (w *SyncWorker) Start() {
	go func() {
		log.Println("Starting Project Sync Worker...")

		for {
			w.runSync()

			// Get wake interval configuration
			wakeInterval := 1 * time.Hour // Default
			configItem, err := w.configRepo.GetByKey("worker_wake_interval")
			if err == nil && configItem != nil {
				if d, err := time.ParseDuration(configItem.Value); err == nil {
					wakeInterval = d
				}
			}

			time.Sleep(wakeInterval)
		}
	}()
}

func (w *SyncWorker) runSync() {
	projects, err := w.projectsRepo.GetActive()
	if err != nil {
		log.Printf("Error fetching active projects for sync: %v", err)
		return
	}

	// Get interval configuration
	interval := 24 * time.Hour // Default
	configItem, err := w.configRepo.GetByKey("project_sync_interval")
	if err == nil && configItem != nil {
		if d, err := time.ParseDuration(configItem.Value); err == nil {
			interval = d
		}
	}

	for _, p := range projects {
		if shouldCheck(p, interval) {
			if err := w.syncProject(&p); err != nil {
				log.Printf("Error syncing project %s: %v", p.Slug, err)
			}
		}
	}
}

func shouldCheck(p Project, interval time.Duration) bool {
	if !p.LastRepoCheckAt.Valid {
		return true
	}
	return time.Since(p.LastRepoCheckAt.Time) >= interval
}

func (w *SyncWorker) syncProject(p *Project) error {
	// 1. Get latest commit from source (UPSTREAM)
	// SourceRepoURL is the upstream URL.
	latestCommit, err := w.ghService.GetLatestCommit(p.SourceRepoURL, p.SourceBranch)
	if err != nil {
		return err
	}

	// 2. Check if new commit
	if latestCommit == p.LastCommitHash {
		// Nothing new, but update check time
		return w.projectsRepo.UpdateSyncStatus(p.ID, time.Now(), latestCommit)
	}

	log.Printf("New commit found for project %s: %s (old: %s). Syncing fork...", p.Slug, latestCommit, p.LastCommitHash)

	// 3. Sync Fork
	// Determine fork name based on environment
	forkName := w.getForkName(p.Slug)

	// We fork to `GITHUB_OWNER/forkName`.
	// The SyncFork method expects the fork name (slug) and branch.
	if err := w.ghService.SyncFork(forkName, p.SourceBranch); err != nil {
		return err
	}

	// 4. Update DB
	return w.projectsRepo.UpdateSyncStatus(p.ID, time.Now(), latestCommit)
}

func (w *SyncWorker) getForkName(slug string) string {
	switch w.config.Environment {
	case "development":
		return "development-" + slug
	case "staging":
		return "staging-" + slug
	default: // production
		return slug
	}
}
