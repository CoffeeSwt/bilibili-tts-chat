<script setup>
import { ref, onMounted } from 'vue'
import ConfigForm from './components/ConfigForm.vue'
import LogViewer from './components/LogViewer.vue'
import { GetConfig } from '../wailsjs/go/main/App'

const currentView = ref('loading') // loading, config, logs

onMounted(async () => {
  try {
    const config = await GetConfig()
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
    <div class="header">
      <div class="brand">
        <div class="logo-wrapper">
          <img src="./assets/images/logo-universal.png" class="logo-small" />
        </div>
        <span class="brand-text">Bilibili TTS Chat</span>
      </div>
      <button v-if="currentView === 'logs'" class="settings-btn" @click="showSettings" title="设置">
        <span class="icon">⚙️</span>
      </button>
      <button v-if="currentView === 'config' && logs && logs.length > 0" class="back-btn" @click="currentView = 'logs'" title="返回日志">
        <span class="icon">↩️</span>
      </button>
    </div>
    
    <div class="content">
      <transition name="fade" mode="out-in">
        <div v-if="currentView === 'loading'" class="loading" key="loading">
          <div class="spinner"></div>
          <span>正在初始化...</span>
        </div>
        
        <ConfigForm v-else-if="currentView === 'config'" @saved="onSaved" key="config" />
        
        <LogViewer v-else-if="currentView === 'logs'" key="logs" />
      </transition>
    </div>
  </div>
</template>

<style scoped>
.app-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  background-color: var(--bg-color);
  color: var(--text-color);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  height: 60px;
  background-color: var(--surface-color);
  border-bottom: 1px solid var(--border-color);
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  z-index: 10;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-wrapper {
  background: white;
  padding: 4px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-small {
  height: 24px;
  width: 24px;
  object-fit: contain;
}

.brand-text {
  font-weight: 700;
  font-size: 18px;
  background: linear-gradient(90deg, #00AEEC, #FB7299);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  letter-spacing: 0.5px;
}

.settings-btn, .back-btn {
  background: var(--input-bg);
  border: 1px solid var(--border-color);
  color: var(--text-color);
  width: 36px;
  height: 36px;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.settings-btn:hover, .back-btn:hover {
  background: var(--primary-color);
  border-color: var(--primary-color);
  color: white;
  transform: translateY(-1px);
}

.content {
  flex: 1;
  overflow: hidden;
  position: relative;
  display: flex;
  flex-direction: column;
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-muted);
  gap: 20px;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid var(--surface-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

:deep(.config-form) {
  margin: auto;
}
</style>
