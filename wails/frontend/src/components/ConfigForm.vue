<script setup>
import { reactive, onMounted } from 'vue'
import { GetConfig, SaveConfig } from '../../wailsjs/go/main/App'

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
    // Ensure types are correct
    const configToSave = {
      ...form,
      volume: parseInt(form.volume),
      max_user_data_len: 1000, // default or preserve
      cleanup_interval: 30,
      speech_rate: 0,
      assistant_memory_size: 10
    }
    // We should ideally merge with existing config to not lose other fields, 
    // but GetConfig already loaded them. 
    // Wait, form only has a few fields. I should store the full config.
    
    // Better approach:
    const currentConfig = await GetConfig()
    Object.assign(currentConfig, form)
    currentConfig.volume = parseInt(form.volume)
    
    await SaveConfig(currentConfig)
    emit('saved')
  } catch (e) {
    state.error = '保存失败: ' + e
  } finally {
    state.loading = false
  }
}
</script>

<template>
  <div class="config-form">
    <h2>设置</h2>
    
    <div class="form-group">
      <label>直播间身份码 (必填)</label>
      <input v-model="form.room_id_code" placeholder="输入B站直播身份码" />
      <small>请在B站直播开放平台获取身份码</small>
    </div>

    <div class="form-group">
      <label>音量 ({{ form.volume }}%)</label>
      <input type="range" v-model="form.volume" min="0" max="100" />
    </div>

    <div class="form-group">
      <label>助手名称</label>
      <input v-model="form.assistant_name" placeholder="小助手" />
    </div>

    <div class="form-group">
      <label>直播间描述</label>
      <textarea v-model="form.room_description" placeholder="描述你的直播间，帮助AI理解上下文"></textarea>
    </div>

    <div class="form-group checkbox">
      <label>
        <input type="checkbox" v-model="form.use_llm_replay" />
        启用AI回复
      </label>
    </div>

    <div v-if="state.error" class="error">{{ state.error }}</div>

    <button @click="save" :disabled="state.loading">
      {{ state.loading ? '保存中...' : '保存并启动' }}
    </button>
  </div>
</template>

<style scoped>
.config-form {
  max-width: 500px;
  margin: 0 auto;
  padding: 20px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
}

.form-group {
  margin-bottom: 15px;
  text-align: left;
}

label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
}

input[type="text"], textarea {
  width: 100%;
  padding: 8px;
  border-radius: 4px;
  border: 1px solid #444;
  background: #222;
  color: white;
}

textarea {
  min-height: 80px;
}

.checkbox label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: normal;
  cursor: pointer;
}

small {
  display: block;
  margin-top: 4px;
  color: #888;
  font-size: 0.8em;
}

.error {
  color: #ff4444;
  margin-bottom: 15px;
}

button {
  width: 100%;
  padding: 10px;
  background: #42b983;
  border: none;
  border-radius: 4px;
  color: white;
  font-weight: bold;
  cursor: pointer;
}

button:disabled {
  background: #555;
  cursor: not-allowed;
}
</style>
