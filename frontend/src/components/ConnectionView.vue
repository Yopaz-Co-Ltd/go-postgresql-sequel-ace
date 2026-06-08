<template>
  <section class="connection-screen">
    <div class="connection-panel">
      <div class="tabs" role="tablist">
        <button class="tab active" type="button">TCP/IP</button>
        <button class="tab" type="button">Socket</button>
        <button class="tab" type="button">SSH</button>
        <button class="tab" type="button">AWS IAM</button>
      </div>

      <form class="connection-form" @submit.prevent="$emit('connect', model)">
        <label>
          <span>Name:</span>
          <input v-model="model.name" autocomplete="off" />
        </label>

        <div class="swatches" aria-label="Favorite color">
          <X class="swatch-x" :size="24" />
          <button v-for="color in colors" :key="color" :style="{ backgroundColor: color }" type="button" />
        </div>

        <label>
          <span>Host:</span>
          <input v-model="model.host" autocomplete="off" />
        </label>

        <label>
          <span>Username:</span>
          <input v-model="model.username" autocomplete="username" />
        </label>

        <label>
          <span>Password:</span>
          <input v-model="model.password" type="password" autocomplete="current-password" />
        </label>

        <label>
          <span>Database:</span>
          <input v-model="model.database" placeholder="optional" autocomplete="off" />
        </label>

        <label>
          <span>Port:</span>
          <input v-model.number="model.port" type="number" min="1" max="65535" />
        </label>

        <label>
          <span>SSL Mode:</span>
          <select v-model="model.sslMode">
            <option value="prefer">Prefer</option>
            <option value="disable">Disable</option>
            <option value="require">Require</option>
            <option value="verify-ca">Verify CA</option>
            <option value="verify-full">Verify Full</option>
          </select>
        </label>

        <label class="check-row">
          <input v-model="flags.local" type="checkbox" />
          <span>Allow LOCAL_DATA_INFILE (insecure)</span>
        </label>
        <label class="check-row">
          <input v-model="flags.cleartext" type="checkbox" />
          <span>Enable Cleartext plugin (insecure)</span>
        </label>
        <label class="check-row">
          <input v-model="flags.requireSsl" type="checkbox" />
          <span>Require SSL</span>
        </label>
      </form>
    </div>

    <div class="connection-actions">
      <button class="help-button" type="button" title="Help"><CircleHelp :size="24" /></button>
      <button class="primary connect-button" type="button" :disabled="busy" @click="$emit('connect', model)">
        Connect
      </button>
      <button type="button">Add to Favorites</button>
      <button type="button">Save changes</button>
      <button type="button" :disabled="busy" @click="$emit('test', model)">Test connection</button>
    </div>

    <p v-if="message" class="status-line">{{ message }}</p>
  </section>
</template>

<script setup>
import { reactive, watch } from 'vue'
import { CircleHelp, X } from '@lucide/vue'

defineProps({
  busy: Boolean,
  message: { type: String, default: '' },
})

defineEmits(['connect', 'test'])

const model = reactive({
  name: 'Local PostgreSQL',
  host: '127.0.0.1',
  username: 'postgres',
  password: '',
  database: 'postgres',
  port: 5432,
  sslMode: 'prefer',
})

const flags = reactive({
  local: false,
  cleartext: false,
  requireSsl: false,
})

const colors = ['#a84d3f', '#b57b24', '#a59b31', '#659449', '#477f9f', '#82629b', '#898989']

watch(
  () => flags.requireSsl,
  (enabled) => {
    if (enabled && model.sslMode === 'disable') {
      model.sslMode = 'require'
    }
  },
)
</script>
