<script setup>
import { ref, onMounted } from 'vue'
import ConfigForm from './components/ConfigForm.vue'
import LogViewer from './components/LogViewer.vue'
import { GetConfig } from '../wailsjs/go/main/App'

const currentView = ref('loading') // loading, config, logs

onMounted(async () => {
  try {
    const config = await GetConfig()
    // If room_id_code is present, assume configured. 
    // You can also add a specific flag in config if needed.
    if (config && config.room_id_code) {
      currentView.value = 'logs'
    } else {
      currentView.value = 'config'
    }
  } catch (e) {
    console.error('Failed to load config:', e)
    currentView.value = 'config'
  }
})

const onSaved = () => {
  currentView.value = 'logs'
}

const showSettings = () => {
  currentView.value = 'config'
}
</script>

<template>
  <div class="app-container">
    <div class="header" v-if="currentView === 'logs'">
      <div class="brand">
        <img src="./assets/images/logo-universal.png" class="logo-small" />
        <span>Bilibili TTS Chat</span>
      </div>
      <button class="settings-btn" @click="showSettings">⚙️ 设置</button>
    </div>
    
    <div class="content">
      <div v-if="currentView === 'loading'" class="loading">
        <div class="spinner"></div>
        <span>正在初始化...</span>
      </div>
      
      <ConfigForm v-else-if="currentView === 'config'" @saved="onSaved" />
      
      <LogViewer v-else-if="currentView === 'logs'" />
    </div>
  </div>
</template>

<style>
/* Reset and global styles */
html, body { 
  margin: 0; 
  padding: 0; 
  width: 100%; 
  height: 100%; 
  background-color: #1a1a1a; 
  color: #ffffff;
  font-family: 'Nunito', sans-serif;
}
#app {
  width: 100%;
  height: 100%;
}
</style>

<style scoped>
.app-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  box-sizing: border-box;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 20px;
  background: #252526;
  border-bottom: 1px solid #333;
  height: 50px;
  box-sizing: border-box;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: bold;
}

.logo-small {
  height: 24px;
}

.settings-btn {
  background: transparent;
  border: 1px solid #444;
  color: #ccc;
  padding: 4px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.settings-btn:hover {
  background: #333;
  border-color: #666;
}

.content {
  flex: 1;
  overflow: hidden;
  position: relative;
  background: #1e1e1e;
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #888;
  gap: 15px;
}

.spinner {
  width: 30px;
  height: 30px;
  border: 3px solid #333;
  border-top-color: #42b983;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Ensure components take full space */
:deep(.log-viewer), :deep(.config-form) {
  height: 100%;
  box-sizing: border-box;
}

/* Center config form */
:deep(.config-form) {
  margin: 20px auto;
  height: auto;
  max-height: calc(100% - 40px);
  overflow-y: auto;
}
</style>
