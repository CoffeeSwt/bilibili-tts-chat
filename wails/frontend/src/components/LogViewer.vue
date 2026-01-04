<script setup>
import { ref, onMounted, nextTick, onUnmounted } from 'vue'
import { EventsOn } from '../../wailsjs/runtime/runtime'

const logs = ref([])
const logContainer = ref(null)
const autoScroll = ref(true)
const heartbeatStatus = ref('inactive') // inactive, active, warning
const lastHeartbeatTime = ref(0)
const heartbeatTimer = ref(null)

const maxLogs = 1000

onMounted(() => {
  // Listen for log events
  EventsOn('log', (log) => {
    logs.value.push(log)
    if (logs.value.length > maxLogs) {
      logs.value.shift()
    }
    
    // Any incoming log (which are filtered important events) counts as activity
    heartbeatStatus.value = 'active'
    lastHeartbeatTime.value = Date.now()
    
    if (autoScroll.value) {
      nextTick(() => {
        scrollToBottom()
      })
    }
  })

  // Listen for heartbeat events
  EventsOn('heartbeat', () => {
    heartbeatStatus.value = 'active'
    lastHeartbeatTime.value = Date.now()
  })

  // Check heartbeat status periodically
  heartbeatTimer.value = setInterval(() => {
    const now = Date.now()
    if (lastHeartbeatTime.value > 0) {
      const diff = now - lastHeartbeatTime.value
      if (diff > 60000) { // > 60s no heartbeat
        heartbeatStatus.value = 'inactive'
      } else if (diff > 30000) { // > 30s no heartbeat
        heartbeatStatus.value = 'warning'
      } else {
        heartbeatStatus.value = 'active'
      }
    }
  }, 5000)
})

onUnmounted(() => {
  if (heartbeatTimer.value) {
    clearInterval(heartbeatTimer.value)
  }
})

const scrollToBottom = () => {
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight
  }
}

const clearLogs = () => {
  logs.value = []
}

const getLevelClass = (level) => {
  switch (level) {
    case 'INFO': return 'info'
    case 'WARN': return 'warn'
    case 'ERROR': return 'error'
    case 'DEBUG': return 'debug'
    default: return ''
  }
}

const getStatusText = (status) => {
  switch(status) {
    case 'active': return '运行中'
    case 'warning': return '连接不稳定'
    case 'inactive': return '连接断开'
    default: return '未连接'
  }
}
</script>

<template>
  <div class="log-viewer">
    <div class="toolbar">
      <div class="status-group">
        <div class="status-indicator-wrapper" :class="heartbeatStatus">
          <div class="status-dot"></div>
          <div class="status-ping"></div>
        </div>
        <span class="status-text">{{ getStatusText(heartbeatStatus) }}</span>
      </div>
      
      <div class="actions">
        <label class="toggle-scroll">
          <input type="checkbox" v-model="autoScroll">
          <span class="toggle-text">自动滚动</span>
        </label>
        <button class="btn-clear" @click="clearLogs" title="清空日志">
          <span class="icon">🗑️</span>
        </button>
      </div>
    </div>
    
    <div class="logs" ref="logContainer">
      <div v-if="logs.length === 0" class="empty-state">
        <span>暂无日志，等待连接...</span>
      </div>
      <div v-for="(log, index) in logs" :key="index" class="log-entry" :class="getLevelClass(log.level)">
        <span class="time">{{ log.timestamp.split(' ')[1] }}</span>
        <span class="level-tag">{{ log.level }}</span>
        <span class="message">{{ log.message }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.log-viewer {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  background: var(--bg-color);
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 20px;
  background: var(--surface-color);
  border-bottom: 1px solid var(--border-color);
  height: 50px;
  box-sizing: border-box;
}

.status-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.status-indicator-wrapper {
  position: relative;
  width: 12px;
  height: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--text-muted);
  z-index: 2;
  transition: all 0.3s;
}

.status-ping {
  position: absolute;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background: inherit;
  opacity: 0.5;
  z-index: 1;
}

/* Active State */
.active .status-dot { background: var(--success-color); box-shadow: 0 0 5px var(--success-color); }
.active .status-ping { 
  background: var(--success-color);
  animation: ping 2s cubic-bezier(0, 0, 0.2, 1) infinite;
}

/* Warning State */
.warning .status-dot { background: var(--warning-color); box-shadow: 0 0 5px var(--warning-color); }

/* Inactive State */
.inactive .status-dot { background: var(--error-color); }

@keyframes ping {
  75%, 100% {
    transform: scale(2.5);
    opacity: 0;
  }
}

.status-text {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color);
}

.actions {
  display: flex;
  gap: 15px;
  align-items: center;
}

.toggle-scroll {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-muted);
  user-select: none;
}

.toggle-scroll input {
  accent-color: var(--primary-color);
}

.btn-clear {
  background: transparent;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-clear:hover {
  background: var(--input-bg);
  border-color: var(--error-color);
}

.btn-clear .icon {
  font-size: 14px;
}

.logs {
  flex: 1;
  overflow-y: auto;
  padding: 10px 0;
  font-family: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-muted);
  font-style: italic;
}

.log-entry {
  display: flex;
  align-items: flex-start;
  padding: 4px 20px;
  line-height: 1.6;
  border-left: 2px solid transparent;
}

.log-entry:hover {
  background: rgba(255, 255, 255, 0.02);
}

.time {
  color: var(--text-muted);
  margin-right: 12px;
  font-size: 12px;
  flex-shrink: 0;
}

.level-tag {
  font-weight: bold;
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 4px;
  margin-right: 12px;
  min-width: 40px;
  text-align: center;
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.1);
}

.message {
  color: var(--text-color);
  word-break: break-all;
}

/* Log Level Colors */
.info .level-tag { color: #7aa2f7; background: rgba(122, 162, 247, 0.1); }
.info .message { color: #c0caf5; }

.warn .level-tag { color: #e0af68; background: rgba(224, 175, 104, 0.1); }
.warn .message { color: #e0af68; }
.warn { border-left-color: #e0af68; background: rgba(224, 175, 104, 0.05); }

.error .level-tag { color: #f7768e; background: rgba(247, 118, 142, 0.1); }
.error .message { color: #f7768e; }
.error { border-left-color: #f7768e; background: rgba(247, 118, 142, 0.05); }

.debug .level-tag { color: #9aa5ce; background: rgba(154, 165, 206, 0.1); }
.debug .message { color: #9aa5ce; }
</style>
