package ratelimiter

import (
	"container/list"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var _ zerolog.Hook = (*RateLimiter)(nil)

type KeyExtractor interface {
	ExtractKey(e *zerolog.Event, level zerolog.Level, message string) string
}

type KeyConfig struct {
	Extractor        KeyExtractor
	SizeLimit        int
	sizeLimitEnabled bool
}

type RateLimiter struct {
	key                *KeyConfig
	keys               map[string]struct{}
	lock               sync.RWMutex
	queue              *list.List
	currentBlock       []string
	blockSize          time.Duration
	windowSize         time.Duration
	queueLength        int // Maximum queue length
	defaultBlockLength int // Maximum block length
}

func New(opt ...Option) (*RateLimiter, error) {
	cfg := NewDefaultConfig()
	for _, o := range opt {
		o(cfg)
	}
	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}
	rl := &RateLimiter{
		key:                &cfg.KeyConfig,
		queue:              list.New(),
		keys:               make(map[string]struct{}, cfg.MinBlockSize*cfg.queueLength),
		currentBlock:       make([]string, 0, 100),
		blockSize:          cfg.BlockSize,
		windowSize:         cfg.WindowSize,
		queueLength:        cfg.queueLength,
		defaultBlockLength: cfg.MinBlockSize,
	}
	go rl.start()
	return rl, nil
}

func (rl *RateLimiter) Run(e *zerolog.Event, level zerolog.Level, message string) {
	var key string
	if rl.key.Extractor != nil {
		key = rl.key.Extractor.ExtractKey(e, level, message)
	} else if message != "" {
		key = level.String() + "_" + message
	}
	if key == "" {
		return
	}
	if rl.key.sizeLimitEnabled && len(key) > rl.key.SizeLimit {
		key = key[:rl.key.SizeLimit]
	}
	if rl.block(key) {
		e.Discard()
		return
	}
}

func (rl *RateLimiter) block(key string) bool {
	rl.lock.RLock()
	if _, exists := rl.keys[key]; exists {
		rl.lock.RUnlock()
		return true
	}

	rl.lock.RUnlock()
	rl.lock.Lock()

	rl.keys[key] = struct{}{}
	rl.currentBlock = append(rl.currentBlock, key)

	rl.lock.Unlock()

	return false
}

func (rl *RateLimiter) start() {
	ticker := time.NewTicker(rl.blockSize)
	defer ticker.Stop()

	for range ticker.C {
		rl.queue.PushBack(rl.currentBlock)
		rl.lock.Lock()
		rl.currentBlock = make([]string, 0, max(rl.defaultBlockLength, len(rl.currentBlock)))
		for rl.queue.Len() >= rl.queueLength {
			front := rl.queue.Front()
			block, ok := front.Value.([]string)
			if ok {
				for _, key := range block {
					delete(rl.keys, key)
				}
			}
			rl.queue.Remove(front)
		}
		rl.lock.Unlock()
	}
}
