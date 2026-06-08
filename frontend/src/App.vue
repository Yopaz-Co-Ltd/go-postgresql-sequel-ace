<template>
  <main class="shell">
    <ConnectionView
      v-if="!sessionId"
      :busy="busy"
      :message="message"
      :theme="theme"
      @test="handleTest"
      @connect="handleConnect"
      @toggle-theme="toggleTheme"
    />

    <WorkspaceView
      v-else
      :session-id="sessionId"
      :connection-name="connectionName"
      :theme="theme"
      @disconnect="handleDisconnect"
      @toggle-theme="toggleTheme"
    />
  </main>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import ConnectionView from './components/ConnectionView.vue'
import WorkspaceView from './components/WorkspaceView.vue'
import { connect, testConnection } from './api'

const THEME_STORAGE_KEY = 'sequel-ace-theme'

const sessionId = ref('')
const connectionName = ref('')
const busy = ref(false)
const message = ref('')
const theme = ref('light')

function applyTheme(value) {
  theme.value = value
  document.documentElement.dataset.theme = value
  localStorage.setItem(THEME_STORAGE_KEY, value)
}

function toggleTheme() {
  applyTheme(theme.value === 'dark' ? 'light' : 'dark')
}

onMounted(() => {
  const stored = localStorage.getItem(THEME_STORAGE_KEY)
  applyTheme(stored === 'dark' ? 'dark' : 'light')
})

async function handleTest(form) {
  busy.value = true
  message.value = ''
  try {
    await testConnection(form)
    message.value = 'Connection OK'
  } catch (error) {
    message.value = error.message
  } finally {
    busy.value = false
  }
}

async function handleConnect(form) {
  busy.value = true
  message.value = ''
  try {
    const result = await connect(form)
    sessionId.value = result.sessionId
    connectionName.value = form.name || `${form.host}:${form.port}`
  } catch (error) {
    message.value = error.message
  } finally {
    busy.value = false
  }
}

function handleDisconnect() {
  sessionId.value = ''
  connectionName.value = ''
}
</script>
