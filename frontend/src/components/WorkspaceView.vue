<template>
  <section class="workspace">
    <aside class="sidebar">
      <div class="window-controls">
        <span class="red"></span>
        <span class="yellow"></span>
        <span class="green"></span>
      </div>

      <div class="connection-title">
        <Database :size="18" />
        <span>{{ connectionName }}</span>
      </div>

      <select v-model="selectedSchema" class="schema-select" @change="loadTables">
        <option v-for="schema in schemas" :key="schema" :value="schema">{{ schema }}</option>
      </select>

      <div class="table-list">
        <button
          v-for="table in tables"
          :key="table.name"
          type="button"
          :class="{ selected: table.name === selectedTable }"
          @click="selectTable(table.name)"
        >
          <Table2 :size="16" />
          <span>{{ table.name }}</span>
        </button>
      </div>
    </aside>

    <div class="main-pane">
      <header class="toolbar">
        <div class="toolbar-group">
          <button type="button" title="Refresh" @click="refresh"><RefreshCw :size="18" /></button>
          <button type="button" title="Run SQL" @click="executeSql"><Play :size="18" /></button>
        </div>
        <div class="title-stack">
          <strong>{{ selectedTable || 'SQL Console' }}</strong>
          <span>{{ rowCount }} rows</span>
        </div>
        <button type="button" title="Disconnect" @click="$emit('disconnect')"><LogOut :size="18" /></button>
      </header>

      <div class="sql-editor">
        <textarea v-model="sql" spellcheck="false"></textarea>
      </div>

      <div class="data-grid">
        <table v-if="columns.length">
          <thead>
            <tr>
              <th v-for="column in columns" :key="column">{{ column }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, index) in rows" :key="index">
              <td v-for="column in columns" :key="column">{{ formatCell(row[column]) }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else class="empty-state">
          <Rows3 :size="34" />
          <span>{{ message || 'No rows loaded' }}</span>
        </div>
      </div>

      <footer class="statusbar">
        <span>{{ message || 'Ready' }}</span>
      </footer>
    </div>
  </section>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { Database, LogOut, Play, RefreshCw, Rows3, Table2 } from '@lucide/vue'
import { getSchemas, getTableRows, getTables, runQuery } from '../api'

const props = defineProps({
  sessionId: { type: String, required: true },
  connectionName: { type: String, required: true },
})

defineEmits(['disconnect'])

const schemas = ref([])
const tables = ref([])
const selectedSchema = ref('public')
const selectedTable = ref('')
const columns = ref([])
const rows = ref([])
const rowCount = ref(0)
const message = ref('')
const sql = ref('select now();')

onMounted(async () => {
  await loadSchemas()
})

async function loadSchemas() {
  try {
    schemas.value = await getSchemas(props.sessionId)
    selectedSchema.value = schemas.value.includes('public') ? 'public' : schemas.value[0] || ''
    await loadTables()
  } catch (error) {
    message.value = error.message
  }
}

async function loadTables() {
  if (!selectedSchema.value) return
  try {
    tables.value = await getTables(props.sessionId, selectedSchema.value)
    if (tables.value.length) {
      await selectTable(tables.value[0].name)
    }
  } catch (error) {
    message.value = error.message
  }
}

async function selectTable(table) {
  selectedTable.value = table
  sql.value = `select * from "${selectedSchema.value}"."${table}" limit 200;`
  await refresh()
}

async function refresh() {
  if (!selectedSchema.value || !selectedTable.value) return
  try {
    const result = await getTableRows(props.sessionId, selectedSchema.value, selectedTable.value)
    setResult(result)
    message.value = `Loaded ${result.count} rows`
  } catch (error) {
    message.value = error.message
  }
}

async function executeSql() {
  try {
    const result = await runQuery(props.sessionId, sql.value)
    setResult(result)
    selectedTable.value = ''
    message.value = `Query returned ${result.count} rows`
  } catch (error) {
    message.value = error.message
  }
}

function setResult(result) {
  columns.value = result.columns || []
  rows.value = (result.rows || []).map((row) => (typeof row === 'string' ? JSON.parse(row) : row))
  rowCount.value = result.count || 0
}

function formatCell(value) {
  if (value === null || value === undefined) return 'NULL'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}
</script>
