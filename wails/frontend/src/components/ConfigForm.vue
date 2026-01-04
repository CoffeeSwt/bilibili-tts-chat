<script setup>
import { reactive, onMounted } from 'vue'
import { GetConfig, SaveConfig, RestartApp } from '../../wailsjs/go/main/App'

const emit = defineEmits(['saved'])

const form = reactive({
  room_id_code: '',
  volume: 50,
  room_description: '',
  assistant_name: '小助手',
  use_llm_replay: true
})

const state = reactive({
  loading: false,
  error: ''
})

onMounted(async () => {
  try {
    const config = await GetConfig()
    if (config) {
      Object.assign(form, config)
    }
  } catch (e) {
    state.error = '加载配置失败: ' + e
  }
})

const save = async () => {
  if (!form.room_id_code) {
    state.error = '请输入直播间身份码'
    return
  }
  
  state.loading = true
  state.error = ''
  
  try {
    const currentConfig = await GetConfig()
    Object.assign(currentConfig, form)
    currentConfig.volume = parseInt(form.volume)
    
    await SaveConfig(currentConfig)
    await RestartApp()
    
    emit('saved')
  } catch (e) {
    state.error = '保存失败: ' + e
  } finally {
    state.loading = false
  }
}
</script>

<template>
  <div class="config-container">
    <div class="config-card">
      <div class="card-header">
        <h2>应用设置</h2>
        <p class="subtitle">配置直播间信息与助手行为</p>
      </div>
      
      <div class="form-content">
        <div class="form-group">
          <label>直播间身份码 <span class="required">*</span></label>
          <div class="input-wrapper">
            <input v-model="form.room_id_code" placeholder="输入 B 站直播身份码" type="text" />
            <div class="input-focus-border"></div>
          </div>
          <small>请在 B 站直播开放平台获取身份码</small>
        </div>

        <div class="form-group">
          <div class="label-row">
            <label>系统音量</label>
            <span class="value-badge">{{ form.volume }}%</span>
          </div>
          <input type="range" v-model="form.volume" min="0" max="100" class="slider" />
        </div>

        <!-- <div class="form-group">
          <label>助手名称</label>
          <div class="input-wrapper">
            <input v-model="form.assistant_name" placeholder="例如：小助手" type="text" />
            <div class="input-focus-border"></div>
          </div>
        </div>

        <div class="form-group">
          <label>直播间描述</label>
          <div class="input-wrapper">
            <textarea v-model="form.room_description" placeholder="描述你的直播间，帮助 AI 更好地理解上下文..."></textarea>
            <div class="input-focus-border"></div>
          </div>
        </div>

        <div class="form-group switch-group">
          <label class="switch-label">
            <span>启用 AI 智能回复</span>
            <small>根据弹幕内容生成拟人化回复</small>
          </label>
          <label class="switch">
            <input type="checkbox" v-model="form.use_llm_replay">
            <span class="slider-round"></span>
          </label>
        </div> -->

        <div v-if="state.error" class="error-msg">
          <span class="error-icon">⚠️</span> {{ state.error }}
        </div>
      </div>

      <div class="card-footer">
        <button @click="save" :disabled="state.loading" class="save-btn">
          <span v-if="state.loading" class="btn-spinner"></span>
          {{ state.loading ? '正在保存并重启...' : '保存配置并启动' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.config-container {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  width: 100%;
  padding: 20px;
  box-sizing: border-box;
}

.config-card {
  background: var(--surface-color);
  border-radius: 16px;
  width: 100%;
  max-width: 480px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
  border: 1px solid var(--border-color);
  overflow: hidden;
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.card-header {
  padding: 25px 30px;
  border-bottom: 1px solid var(--border-color);
  background: linear-gradient(to right, rgba(0, 174, 236, 0.05), transparent);
}

.card-header h2 {
  margin: 0;
  font-size: 24px;
  color: var(--primary-color);
}

.subtitle {
  margin: 5px 0 0 0;
  color: var(--text-muted);
  font-size: 14px;
}

.form-content {
  padding: 30px;
}

.form-group {
  margin-bottom: 24px;
}

.form-group:last-child {
  margin-bottom: 0;
}

label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: var(--text-color);
  font-size: 14px;
}

.required {
  color: var(--error-color);
}

.input-wrapper {
  position: relative;
}

input[type="text"], textarea {
  width: 100%;
  padding: 12px 16px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  background: var(--input-bg);
  color: white;
  font-size: 14px;
  transition: all 0.2s;
  box-sizing: border-box;
  outline: none;
}

input[type="text"]:focus, textarea:focus {
  border-color: var(--primary-color);
  background: #1a1b26;
  box-shadow: 0 0 0 3px rgba(0, 174, 236, 0.15);
}

textarea {
  min-height: 100px;
  resize: vertical;
  line-height: 1.5;
}

small {
  display: block;
  margin-top: 6px;
  color: var(--text-muted);
  font-size: 12px;
}

/* Range Slider */
.label-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.value-badge {
  background: var(--primary-color);
  color: white;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: bold;
}

.slider {
  -webkit-appearance: none;
  width: 100%;
  height: 6px;
  border-radius: 3px;
  background: var(--border-color);
  outline: none;
}

.slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--primary-color);
  cursor: pointer;
  transition: transform 0.2s;
}

.slider::-webkit-slider-thumb:hover {
  transform: scale(1.2);
}

/* Switch */
.switch-group {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--input-bg);
  padding: 15px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.switch-label {
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.switch {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 26px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider-round {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--border-color);
  transition: .4s;
  border-radius: 34px;
}

.slider-round:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 4px;
  bottom: 4px;
  background-color: white;
  transition: .4s;
  border-radius: 50%;
}

input:checked + .slider-round {
  background-color: var(--primary-color);
}

input:checked + .slider-round:before {
  transform: translateX(24px);
}

.error-msg {
  background: rgba(247, 118, 142, 0.1);
  border: 1px solid rgba(247, 118, 142, 0.2);
  color: var(--error-color);
  padding: 10px;
  border-radius: 8px;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.card-footer {
  padding: 20px 30px;
  border-top: 1px solid var(--border-color);
  background: var(--input-bg);
}

.save-btn {
  width: 100%;
  padding: 14px;
  background: linear-gradient(135deg, #00AEEC, #007aff);
  border: none;
  border-radius: 8px;
  color: white;
  font-weight: 700;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.save-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 174, 236, 0.3);
}

.save-btn:disabled {
  background: var(--border-color);
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.btn-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}
</style>
