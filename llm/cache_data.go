package llm

import (
	"sync"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
)

var (
	cache_event_Data []string
	onceCache        sync.Once
)

func AddCacheEventData(eventData string) {
	onceCache.Do(func() {
		cache_event_Data = make([]string, 0)
	})
	if len(cache_event_Data) >= config.GetAssistantMemorySize() {
		cache_event_Data = cache_event_Data[1:]
	}
	cache_event_Data = append(cache_event_Data, eventData)
}

func GetCacheEventData() []string {
	onceCache.Do(func() {
		cache_event_Data = make([]string, 0)
	})
	return cache_event_Data
}
