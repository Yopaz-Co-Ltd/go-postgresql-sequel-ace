<template>
  <main class="shell">
    <ConnectionView
      v-if="!sessionId"
      :busy="busy"
      :message="message"
      @test="handleTest"
      @connect="handleConnect"
    />

    <WorkspaceView
      v-else
      :session-id="sessionId"
      :connection-name="connectionName"
      @disconnect="handleDisconnect"
    />
  </main>
</template>

<script setup>
import { ref } from 'vue'
import ConnectionView from './components/ConnectionView.vue'
import WorkspaceView from './components/WorkspaceView.vue'
import { connect, testConnection } from './api'

const sessionId = ref('')
const connectionName = ref('')
const busy = ref(false)
const message = ref('')

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
