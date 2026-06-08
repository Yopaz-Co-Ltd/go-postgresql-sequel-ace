<template>
  <section class="connection-screen">
    <header class="login-titlebar">
      <div class="traffic-lights">
        <span class="red"></span>
        <span class="yellow"></span>
        <span class="green"></span>
      </div>
      <strong>Sequel Ace</strong>
      <div class="database-picker">Choose Database... <ChevronDown :size="14" /></div>
      <nav class="login-tools" aria-label="Database tools">
        <button v-for="tool in tools" :key="tool.label" type="button" :class="{ active: tool.label === 'Table History' }">
          <component :is="tool.icon" :size="24" />
          <span>{{ tool.label }}</span>
        </button>
      </nav>
    </header>

    <aside class="favorites-sidebar">
      <div class="quick-connect"><Zap :size="13" /> QUICK CONNECT</div>
      <h2>FAVORITES</h2>
      <div class="favorites-list">
        <button
          v-for="favorite in favorites"
          :key="favorite.id"
          type="button"
          :class="{ selected: favorite.id === selectedFavoriteId }"
          @click="selectFavorite(favorite.id)"
        >
          <Database :size="15" />
          <span>{{ favorite.name }}</span>
        </button>
      </div>
      <div class="favorites-footer">
        <button type="button" title="Options"><CircleEllipsis :size="17" /></button>
        <button type="button" title="Add folder"><FolderPlus :size="18" /></button>
        <button type="button" title="Add favorite" @click="addFavorite"><Plus :size="18" /></button>
        <button type="button" title="Sidebar"><PanelLeftClose :size="16" /></button>
      </div>
    </aside>

    <div class="login-stage">
      <p class="connection-heading">Enter connection details below, or choose a favorite</p>

      <div class="connection-stack">
        <div class="connection-panel">
          <div class="tabs" role="tablist">
            <button class="tab active" type="button">TCP/IP</button>
            <button class="tab" type="button">Socket</button>
            <button class="tab" type="button">SSH</button>
            <button class="tab" type="button">AWS IAM</button>
          </div>

          <form class="connection-form" @submit.prevent="connectCurrent">
            <label>
              <span>Name:</span>
              <input v-model="model.name" autocomplete="off" />
            </label>

            <div class="swatches" aria-label="Favorite color">
              <X class="swatch-x" :size="15" />
              <button
                v-for="color in colors"
                :key="color"
                :class="{ selected: color === model.color }"
                :style="{ backgroundColor: color }"
                type="button"
                @click="model.color = color"
              />
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
              <span>Time Zone:</span>
              <select v-model="model.timeZone">
                <option>Use Server Time Zone</option>
                <option>UTC</option>
                <option>Asia/Ho_Chi_Minh</option>
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
          <button class="help-button" type="button" title="Help"><CircleHelp :size="20" /></button>
          <button class="primary connect-button" type="button" :disabled="busy" @click="connectCurrent">Connect</button>
          <button type="button" @click="addFavorite">Add to Favorites</button>
          <button type="button" @click="saveFavorite">Save changes</button>
          <button type="button" :disabled="busy" @click="testCurrent">Test connection</button>
        </div>

        <p v-if="statusText" class="status-line">{{ statusText }}</p>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  Bolt,
  ChevronsLeftRightEllipsis,
  ChevronDown,
  CircleEllipsis,
  CircleHelp,
  Clock3,
  Database,
  FolderPlus,
  Info,
  List,
  PanelLeftClose,
  Plus,
  Rows3,
  Shuffle,
  TerminalSquare,
  Users,
  X,
  Zap,
} from '@lucide/vue'

const props = defineProps({
  busy: Boolean,
  message: { type: String, default: '' },
})

const emit = defineEmits(['connect', 'test'])

const storageKey = 'postgresql-client-favorites'

const model = reactive({
  id: '',
  name: 'quan.pro.vn',
  host: '127.0.0.1',
  username: 'root',
  password: 'postgres',
  database: '',
  port: 3306,
  sslMode: 'prefer',
  timeZone: 'Use Server Time Zone',
  color: '#8a63a0',
})

const flags = reactive({
  local: false,
  cleartext: false,
  requireSsl: false,
})

const colors = ['#a84d3f', '#b57b24', '#a59b31', '#659449', '#477f9f', '#82629b', '#898989']
const favorites = ref([])
const selectedFavoriteId = ref('')
const localMessage = ref('')
const statusText = computed(() => props.message || localMessage.value)

const tools = [
  { label: 'Structure', icon: Shuffle },
  { label: 'Content', icon: List },
  { label: 'Relations', icon: ChevronsLeftRightEllipsis },
  { label: 'Triggers', icon: Bolt },
  { label: 'Table Info', icon: Info },
  { label: 'Query', icon: TerminalSquare },
  { label: 'Table History', icon: Clock3 },
  { label: 'Users', icon: Users },
  { label: 'Console', icon: Rows3 },
]

onMounted(() => {
  favorites.value = loadFavorites()
  if (favorites.value.length) {
    const preferred = favorites.value.find((favorite) => favorite.name === 'quan.pro.vn')
    selectFavorite((preferred || favorites.value[0]).id)
  }
})

watch(
  () => flags.requireSsl,
  (enabled) => {
    if (enabled && model.sslMode === 'disable') {
      model.sslMode = 'require'
    }
  },
)

function connectCurrent() {
  emit('connect', toConnectionPayload())
}

function testCurrent() {
  localMessage.value = ''
  emit('test', toConnectionPayload())
}

function addFavorite() {
  const favorite = {
    ...toStoredFavorite(),
    id: crypto.randomUUID(),
    name: model.name.trim() || model.host || 'New Favorite',
  }
  favorites.value = [...favorites.value, favorite]
  selectedFavoriteId.value = favorite.id
  model.id = favorite.id
  persistFavorites()
  localMessage.value = 'Favorite added'
}

function saveFavorite() {
  if (!selectedFavoriteId.value) {
    addFavorite()
    return
  }

  const next = toStoredFavorite()
  favorites.value = favorites.value.map((favorite) =>
    favorite.id === selectedFavoriteId.value ? { ...next, id: selectedFavoriteId.value } : favorite,
  )
  persistFavorites()
  localMessage.value = 'Favorite saved'
}

function selectFavorite(id) {
  const favorite = favorites.value.find((item) => item.id === id)
  if (!favorite) return
  selectedFavoriteId.value = id
  Object.assign(model, favorite)
  flags.requireSsl = favorite.sslMode === 'require' || favorite.sslMode === 'verify-ca' || favorite.sslMode === 'verify-full'
}

function toConnectionPayload() {
  return {
    name: model.name,
    host: model.host,
    username: model.username,
    password: model.password,
    database: model.database,
    port: Number(model.port) || 5432,
    sslMode: flags.requireSsl && model.sslMode === 'disable' ? 'require' : model.sslMode,
  }
}

function toStoredFavorite() {
  return {
    id: model.id,
    name: model.name.trim() || model.host || 'New Favorite',
    host: model.host.trim(),
    username: model.username.trim(),
    password: model.password,
    database: model.database.trim(),
    port: Number(model.port) || 5432,
    sslMode: flags.requireSsl && model.sslMode === 'disable' ? 'require' : model.sslMode,
    timeZone: model.timeZone,
    color: model.color,
  }
}

function loadFavorites() {
  try {
    const raw = JSON.parse(localStorage.getItem(storageKey) || '[]')
    if (Array.isArray(raw) && raw.length) return raw
  } catch {
    localStorage.removeItem(storageKey)
  }

  const seed = [
    createFavorite('127.0.0.1', { host: '127.0.0.1', username: 'postgres', password: 'postgres', database: 'sample_store', port: 5432, sslMode: 'disable' }),
    createFavorite('mory-dev', { host: '127.0.0.1', username: 'postgres', database: 'postgres', port: 5432 }),
    createFavorite('real.reddy.vn'),
    createFavorite('real.chipmunk.vn'),
    createFavorite('itech_iot'),
    createFavorite('New Favorite'),
    createFavorite('spispi'),
    createFavorite('php-dev'),
    createFavorite('chipmunk'),
    createFavorite('shopbay.cloud'),
    createFavorite('yopaz.dev'),
    createFavorite('127.0.0.1 docker', { host: '127.0.0.1', username: 'postgres', password: 'postgres', database: 'sample_store', port: 5432, sslMode: 'disable' }),
    createFavorite('quanlv'),
    createFavorite('goldwin-hub'),
    createFavorite('golang-ci'),
    createFavorite('web3'),
    createFavorite('itech'),
    createFavorite('goldwin'),
    createFavorite('lumine-v2'),
    createFavorite('quan.pro.vn', { host: '127.0.0.1', username: 'root', password: 'postgres', port: 3306 }),
    createFavorite('3307', { host: '127.0.0.1', username: 'root', port: 3307 }),
  ]
  localStorage.setItem(storageKey, JSON.stringify(seed))
  return seed
}

function createFavorite(name, overrides = {}) {
  return {
    id: crypto.randomUUID(),
    name,
    host: overrides.host || name,
    username: overrides.username || 'postgres',
    password: overrides.password || '',
    database: overrides.database || '',
    port: overrides.port || 5432,
    sslMode: overrides.sslMode || 'prefer',
    timeZone: 'Use Server Time Zone',
    color: overrides.color || '#8a63a0',
  }
}

function persistFavorites() {
  localStorage.setItem(storageKey, JSON.stringify(favorites.value))
}
</script>
