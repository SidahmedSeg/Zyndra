package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/store"
)

// Pool manages a pool of workers that process jobs
type Pool struct {
	store         *store.DB
	config        *config.Config
	workers       []*Worker
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	buildWorker   *BuildWorker
	rollbackWorker *RollbackWorker
}

// NewPool creates a new worker pool
func NewPool(store *store.DB, cfg *config.Config) (*Pool, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize build worker
	buildWorker, err := NewBuildWorker(store, cfg)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create build worker: %w", err)
	}

	// Initialize rollback worker
	rollbackWorker := NewRollbackWorker(store, cfg)

	pool := &Pool{
		store:          store,
		config:         cfg,
		ctx:            ctx,
		cancel:         cancel,
		buildWorker:    buildWorker,
		rollbackWorker: rollbackWorker,
	}

	return pool, nil
}

// Start starts the worker pool
func (p *Pool) Start(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		worker := &Worker{
			id:            i + 1,
			pool:          p,
			buildWorker:   p.buildWorker,
			rollbackWorker: p.rollbackWorker,
		}
		p.workers = append(p.workers, worker)
		p.wg.Add(1)
		go worker.run()
	}

	log.Printf("Started worker pool with %d workers", numWorkers)
}

// Stop stops the worker pool
func (p *Pool) Stop() {
	log.Println("Stopping worker pool...")
	p.cancel()
	p.wg.Wait()
	if p.buildWorker != nil {
		p.buildWorker.Close()
	}
	log.Println("Worker pool stopped")
}

// Worker represents a single worker in the pool
type Worker struct {
	id            int
	pool          *Pool
	buildWorker   *BuildWorker
	rollbackWorker *RollbackWorker
}

// run runs the worker loop
func (w *Worker) run() {
	defer w.pool.wg.Done()

	ticker := time.NewTicker(5 * time.Second) // Poll every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-w.pool.ctx.Done():
			return
		case <-ticker.C:
			w.processNextJob()
		}
	}
}

// processNextJob processes the next available job
func (w *Worker) processNextJob() {
	ctx := context.Background()

	// Get next job using SKIP LOCKED
	job, err := w.pool.store.GetNextJob(ctx)
	if err != nil {
		log.Printf("Worker %d: Error getting next job: %v", w.id, err)
		return
	}

	if job == nil {
		// No jobs available
		return
	}

	log.Printf("Worker %d: Processing job %s (type: %s)", w.id, job.ID, job.Type)

	// Mark job as processing
	if err := w.pool.store.StartJob(ctx, job.ID); err != nil {
		log.Printf("Worker %d: Error starting job: %v", w.id, err)
		return
	}

	// Process job based on type
	var processErr error
	switch job.Type {
	case "build":
		processErr = w.processBuildJob(ctx, job)
	case "rollback":
		processErr = w.processRollbackJob(ctx, job)
	default:
		processErr = fmt.Errorf("unknown job type: %s", job.Type)
	}

	// Update job status
	if processErr != nil {
		// Increment attempts
		w.pool.store.IncrementJobAttempts(ctx, job.ID)

		// Check if max attempts reached
		if job.Attempts+1 >= job.MaxAttempts {
			w.pool.store.FailJob(ctx, job.ID, processErr.Error())
			log.Printf("Worker %d: Job %s failed after %d attempts: %v", w.id, job.ID, job.Attempts+1, processErr)
		} else {
			// Requeue job
			w.pool.store.UpdateJobStatus(ctx, job.ID, "queued")
			log.Printf("Worker %d: Job %s requeued (attempt %d/%d): %v", w.id, job.ID, job.Attempts+1, job.MaxAttempts, processErr)
		}
	} else {
		w.pool.store.CompleteJob(ctx, job.ID)
		log.Printf("Worker %d: Job %s completed successfully", w.id, job.ID)
	}
}

// processBuildJob processes a build job
func (w *Worker) processBuildJob(ctx context.Context, job *store.Job) error {
	// Extract deployment ID from payload
	deploymentIDStr, ok := job.Payload["deployment_id"].(string)
	if !ok {
		return fmt.Errorf("missing deployment_id in job payload")
	}

	deploymentID, err := uuid.Parse(deploymentIDStr)
	if err != nil {
		return fmt.Errorf("invalid deployment_id: %w", err)
	}

	// Process build
	return w.buildWorker.ProcessBuildJob(ctx, deploymentID)
}

// processRollbackJob processes a rollback job
func (w *Worker) processRollbackJob(ctx context.Context, job *store.Job) error {
	return w.rollbackWorker.ProcessRollbackJob(ctx, job)
}

