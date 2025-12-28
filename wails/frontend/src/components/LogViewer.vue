<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { EventsOn } from '../../wailsjs/runtime/runtime'

const logs = ref([])
const logContainer = ref(null)
const autoScroll = ref(true)

const maxLogs = 1000

onMounted(() => {
  // Listen for log events
  EventsOn('log', (log) => {
    logs.value.push(log)
    if (logs.value.length > maxLogs) {
      logs.value.shift()
    }
    
    if (autoScroll.value) {
      nextTick(() => {
        scrollToBottom()
      })
    }
  })
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
</script>

<template>
  <div class="log-viewer">
    <div class="toolbar">
      <span class="title">日志监控</span>
      <div class="actions">
        <label><input type="checkbox" v-model="autoScroll"> 自动滚动</label>
        <button class="btn-clear" @click="clearLogs">清空</button>
      </div>
    </div>
    
    <div class="logs" ref="logContainer">
      <div v-for="(log, index) in logs" :key="index" class="log-entry">
        <span class="time">{{ log.timestamp.split(' ')[1] }}</span>
        <span class="level" :class="getLevelClass(log.level)">[{{ log.level }}]</span>
        <span class="location">{{ log.location }}</span>
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
  background: #1e1e1e;
  border-radius: 8px;
  overflow: hidden;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #2d2d2d;
  border-bottom: 1px solid #3d3d3d;
}

.title {
  font-weight: bold;
  color: #eee;
}

.actions {
  display: flex;
  gap: 15px;
  align-items: center;
  font-size: 12px;
}

.btn-clear {
  padding: 2px 8px;
  background: #444;
  border: none;
  border-radius: 3px;
  color: #ddd;
  cursor: pointer;
}

.btn-clear:hover {
  background: #555;
}

.logs {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
  text-align: left;
}

.log-entry {
  margin-bottom: 4px;
  line-height: 1.4;
  word-break: break-all;
}

.time {
  color: #888;
  margin-right: 8px;
}

.level {
  font-weight: bold;
  margin-right: 8px;
  min-width: 45px;
  display: inline-block;
  text-align: center;
}

.level.info { color: #4fc1ff; }
.level.warn { color: #cca700; }
.level.error { color: #f14c4c; }
.level.debug { color: #9cdcfe; }

.location {
  color: #569cd6;
  margin-right: 8px;
}

.message {
  color: #d4d4d4;
}

/* Custom Scrollbar */
.logs::-webkit-scrollbar {
  width: 8px;
}
.logs::-webkit-scrollbar-track {
  background: #1e1e1e;
}
.logs::-webkit-scrollbar-thumb {
  background: #444;
  border-radius: 4px;
}
.logs::-webkit-scrollbar-thumb:hover {
  background: #555;
}
</style>
