<template>
  <section class="content-window">
    <header class="content-titlebar">
      <div class="traffic-lights">
        <button type="button" class="red" title="Disconnect" @click="$emit('disconnect')"><X :size="11" /></button>
        <span class="yellow"></span>
        <span class="green"></span>
      </div>
      <strong>{{ windowTitle }}</strong>
      <select v-model="selectedSchema" class="content-db-select" @change="loadTables">
        <option v-for="schema in schemas" :key="schema" :value="schema">{{ schema }}</option>
      </select>
      <nav class="content-tools" aria-label="Table tools">
        <button
          v-for="tool in tools"
          :key="tool.label"
          type="button"
          :class="{ active: tool.label === activeTool }"
          @click="setActiveTool(tool.label)"
        >
          <component :is="tool.icon" :size="24" />
          <span>{{ tool.label }}</span>
        </button>
        <button
          type="button"
          class="theme-toggle"
          :title="theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'"
          @click="$emit('toggle-theme')"
        >
          <component :is="theme === 'dark' ? Sun : Moon" :size="24" />
          <span>{{ theme === 'dark' ? 'Light' : 'Dark' }}</span>
        </button>
      </nav>
    </header>

    <aside class="content-sidebar">
      <label class="sidebar-filter">
        <Search :size="17" />
        <input v-model="tableFilter" placeholder="Filter" />
      </label>

      <div class="sidebar-tables">
        <h2>TABLES</h2>
        <button
          v-for="table in filteredTables"
          :key="table.name"
          type="button"
          :class="{ selected: table.name === selectedTable }"
          @click="selectTable(table.name)"
        >
          <Grid2X2 :size="18" />
          <span>{{ table.name }}</span>
        </button>
      </div>

      <section class="table-information">
        <h2>TABLE INFORMATION</h2>
        <p><span></span>created: {{ tableInformation.created }}</p>
        <p><span></span>engine: PostgreSQL</p>
        <p><span></span>rows: {{ displayRows }}</p>
        <p><span></span>size: {{ tableInformation.size }}</p>
        <p><span></span>schema: {{ selectedSchema || '-' }}</p>
        <p><span></span>columns: {{ tableInformation.columns.length }}</p>
      </section>

      <footer class="content-sidebar-footer">
        <button type="button" title="Add"><Plus :size="17" /></button>
        <button type="button" title="Options"><CircleEllipsis :size="17" /></button>
        <button type="button" title="Refresh" @click="loadTables"><RefreshCw :size="17" /></button>
        <button type="button" title="Preview"><Eye :size="18" /></button>
        <button type="button" title="Resize"><PanelLeftClose :size="16" /></button>
      </footer>
    </aside>

    <main class="content-main">
      <FilterBar
        v-if="activeTool === 'Content'"
        :columns="tableInformation.columns"
        :rules="filterRules"
        :column-kind="columnKind"
        :operator-options-for="filterOperatorOptionsFor"
        :value-mode-for="filterValueModeFor"
        :value-placeholder-for="filterValuePlaceholderFor"
        @add-rule="addFilterRule"
        @apply="applyFilter"
        @remove-rule="removeFilterRule"
        @sync-rule="syncRule"
      />

      <div v-if="activeTool === 'Structure'" class="structure-view">
        <section class="structure-section structure-fields">
          <div class="structure-filterbar">
            <label>
              <Search :size="16" />
              <input v-model="structureFilter" placeholder="Filter" />
            </label>
          </div>
          <div class="structure-grid">
            <table v-if="filteredStructureColumns.length">
              <thead>
                <tr>
                  <th>Field</th>
                  <th>Type</th>
                  <th>Length</th>
                  <th>Allow Null</th>
                  <th>Key</th>
                  <th>Default</th>
                  <th>Extra</th>
                  <th>Encoding</th>
                  <th>Collation</th>
                  <th>Comment</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="column in filteredStructureColumns" :key="column.name">
                  <td>{{ column.name }}</td>
                  <td>{{ column.dataType }}</td>
                  <td>{{ column.length || '-' }}</td>
                  <td><input type="checkbox" :checked="column.nullable" disabled /></td>
                  <td>{{ column.key || '-' }}</td>
                  <td>{{ column.default || 'None' }}</td>
                  <td>{{ column.extra || 'None' }}</td>
                  <td>{{ column.encoding || '-' }}</td>
                  <td>{{ column.collation || '-' }}</td>
                  <td>{{ column.comment || '' }}</td>
                </tr>
              </tbody>
            </table>
            <div v-else class="empty-state">
              <Rows3 :size="34" />
              <span>{{ message || 'No fields loaded' }}</span>
            </div>
          </div>
        </section>

        <section class="structure-section structure-indexes">
          <div class="structure-toolbar">
            <button type="button" title="Add index"><Plus :size="17" /></button>
            <button type="button" title="Remove index"><Minus :size="17" /></button>
            <button type="button" title="Refresh" @click="loadTableInfo"><RefreshCw :size="16" /></button>
            <strong>INDEXES</strong>
          </div>
          <div class="structure-grid">
            <table v-if="filteredIndexes.length">
              <thead>
                <tr>
                  <th>Non_unique</th>
                  <th>Key_name</th>
                  <th>Seq_in_index</th>
                  <th>Column_name</th>
                  <th>Collation</th>
                  <th>Cardinality</th>
                  <th>Sub_part</th>
                  <th>Packed</th>
                  <th>Comment</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="index in filteredIndexes" :key="`${index.keyName}-${index.sequence}-${index.columnName}`">
                  <td>{{ index.nonUnique }}</td>
                  <td>{{ index.keyName }}</td>
                  <td>{{ index.sequence }}</td>
                  <td>{{ index.columnName }}</td>
                  <td>{{ index.collation || '-' }}</td>
                  <td>{{ index.cardinality }}</td>
                  <td>{{ index.subPart || 'NULL' }}</td>
                  <td>{{ index.packed || 'NULL' }}</td>
                  <td>{{ index.comment || '' }}</td>
                </tr>
              </tbody>
            </table>
            <div v-else class="empty-state compact">
              <span>No indexes</span>
            </div>
          </div>
        </section>
      </div>

      <div v-else class="content-grid">
        <table v-if="columns.length">
          <thead>
            <tr>
              <th v-for="column in columns" :key="column" :style="{ width: columnWidth(column) }">
                <span>{{ column }}</span>
                <small>{{ columnType(column) }}</small>
              </th>
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

      <footer class="content-statusbar">
        <div class="status-tools">
          <button type="button" title="Insert row"><Plus :size="17" /></button>
          <button type="button" title="Delete row"><Minus :size="17" /></button>
          <button type="button" title="Duplicate row"><CopyPlus :size="16" /></button>
          <button type="button" title="Refresh" @click="refresh"><RefreshCw :size="16" /></button>
          <button type="button" title="Filter"><Filter :size="16" /></button>
          <label>
            <Search :size="16" />
            <input v-model="columnSearch" placeholder="Filter Columns" />
          </label>
        </div>
        <strong>{{ pageSummary }}</strong>
        <div class="status-pager">
          <button type="button" title="Previous page" :disabled="currentPage <= 1" @click="goToPage(currentPage - 1)">
            <ChevronLeft :size="17" />
          </button>
          <label class="pager-jump">
            <span>Page</span>
            <input v-model.number="currentPage" type="number" min="1" :max="pageCount" @change="goToPage(currentPage)" />
          </label>
          <span class="pager-total">of {{ pageCount }}</span>
          <button type="button" title="Next page" :disabled="currentPage >= pageCount" @click="goToPage(currentPage + 1)">
            <ChevronRight :size="17" />
          </button>
        </div>
      </footer>
    </main>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import {
  Bolt,
  ChevronLeft,
  ChevronRight,
  CircleEllipsis,
  Clock3,
  CopyPlus,
  Eye,
  Filter,
  Grid2X2,
  Info,
  List,
  Minus,
  Moon,
  PanelLeftClose,
  Plus,
  RefreshCw,
  Rows3,
  Search,
  Shuffle,
  Sun,
  TerminalSquare,
  Users,
  X,
} from '@lucide/vue'
import { getSchemas, getTableInfo, getTableRows, getTables, runQuery } from '../api'
import FilterBar from './FilterBar.vue'
import { useTableFilters } from '../composables/useTableFilters'

const props = defineProps({
  sessionId: { type: String, required: true },
  connectionName: { type: String, required: true },
  theme: { type: String, default: 'light' },
})

defineEmits(['disconnect', 'toggle-theme'])

const schemas = ref([])
const tables = ref([])
const selectedSchema = ref('public')
const selectedTable = ref('')
const tableFilter = ref('')
const activeTool = ref('Content')
const columns = ref([])
const rows = ref([])
const rowCount = ref(0)
const totalRows = ref(0)
const message = ref('')
const currentPage = ref(1)
const pageSize = ref(1000)
const lastQueryFiltered = ref(false)
const columnSearch = ref('')
const structureFilter = ref('')
const tableInformation = ref({ created: '-', rows: 0, size: '-', columns: [], indexes: [] })
const {
  filterRules,
  resetFilterRules,
  syncRule,
  addFilterRule,
  removeFilterRule,
  buildWherePredicate,
  columnKind,
  filterOperatorOptionsFor,
  filterValueModeFor,
  filterValuePlaceholderFor,
} = useTableFilters({
  getColumnType: columnType,
  getDefaultColumn: () => tableInformation.value.columns[0]?.name || '',
})

const tools = [
  { label: 'Structure', icon: Shuffle },
  { label: 'Content', icon: List },
  { label: 'Relations', icon: CopyPlus },
  { label: 'Triggers', icon: Bolt },
  { label: 'Table Info', icon: Info },
  { label: 'Query', icon: TerminalSquare },
  { label: 'Table History', icon: Clock3 },
  { label: 'Users', icon: Users },
  { label: 'Console', icon: Rows3 },
]

const filteredTables = computed(() => {
  const needle = tableFilter.value.trim().toLowerCase()
  if (!needle) return tables.value
  return tables.value.filter((table) => table.name.toLowerCase().includes(needle))
})
const filteredStructureColumns = computed(() => {
  const needle = structureFilter.value.trim().toLowerCase()
  const list = tableInformation.value.columns || []
  if (!needle) return list
  return list.filter((column) => {
    return [column.name, column.dataType, column.key, column.default, column.extra, column.comment]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(needle))
  })
})
const filteredIndexes = computed(() => {
  const needle = structureFilter.value.trim().toLowerCase()
  const list = tableInformation.value.indexes || []
  if (!needle) return list
  return list.filter((index) => {
    return [index.keyName, index.columnName, index.comment]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(needle))
  })
})

const displayRows = computed(() => {
  const estimatedRows = Number(tableInformation.value.rows || 0)
  return formatNumber(estimatedRows > 0 ? estimatedRows : rowCount.value || 0)
})
const pageCount = computed(() => Math.max(1, Math.ceil((totalRows.value || 0) / pageSize.value)))
const pageSummary = computed(() => {
  const visible = rowCount.value || 0
  const total = totalRows.value || 0
  if (!total) return '0 rows'
  const noun = visible === 1 ? 'row' : 'rows'
  return lastQueryFiltered.value
    ? `${visible} ${noun} of ${total} matches filter`
    : `${visible} ${noun} of ${total} rows`
})
const windowTitle = computed(() => `(PostgreSQL) ${props.connectionName}/${selectedSchema.value}/${selectedTable.value || ''}`)

onMounted(async () => {
  await loadSchemas()
})

async function loadSchemas() {
  try {
    schemas.value = (await getSchemas(props.sessionId)) || []
    selectedSchema.value = schemas.value.includes('public') ? 'public' : schemas.value[0] || ''
    await loadTables()
  } catch (error) {
    message.value = error.message
  }
}

async function loadTables() {
  if (!selectedSchema.value) return
  try {
    tables.value = (await getTables(props.sessionId, selectedSchema.value)) || []
    if (!tables.value.length) {
      const fallbackSchema = await findSchemaWithTables()
      if (fallbackSchema && fallbackSchema !== selectedSchema.value) {
        selectedSchema.value = fallbackSchema.schema
        tables.value = fallbackSchema.tables
      }
    }
    if (tables.value.length) {
      const current = tables.value.find((table) => table.name === selectedTable.value)
      await selectTable((current || tables.value[0]).name)
    }
  } catch (error) {
    message.value = error.message
  }
}

async function findSchemaWithTables() {
  for (const schema of schemas.value) {
    if (schema === selectedSchema.value) continue
    const schemaTables = (await getTables(props.sessionId, schema)) || []
    if (schemaTables.length) return { schema, tables: schemaTables }
  }
  return null
}

async function selectTable(table) {
  selectedTable.value = table
  await loadTableInfo()
  if (activeTool.value === 'Content') {
    await loadRows(1)
  }
}

async function loadTableInfo() {
  if (!selectedSchema.value || !selectedTable.value) return
  try {
    const info = await getTableInfo(props.sessionId, selectedSchema.value, selectedTable.value)
    tableInformation.value = {
      created: '-',
      rows: info.rows || 0,
      size: info.size || '-',
      columns: info.columns || [],
      indexes: info.indexes || [],
    }
    resetFilterRules()
    currentPage.value = 1
    totalRows.value = 0
    rowCount.value = 0
    lastQueryFiltered.value = false
  } catch (error) {
    message.value = error.message
  }
}

async function setActiveTool(tool) {
  if (!['Structure', 'Content'].includes(tool)) return
  activeTool.value = tool
  if (tool === 'Structure') {
    await loadTableInfo()
    return
  }
  if (tool === 'Content' && selectedTable.value && !columns.value.length) {
    await refresh()
  }
}

async function refresh() {
  await loadRows(currentPage.value)
}

async function goToPage(page) {
  const nextPage = Math.max(1, Math.min(Number(page) || 1, pageCount.value))
  await loadRows(nextPage)
}

async function loadRows(page = 1) {
  if (!selectedSchema.value || !selectedTable.value) return
  try {
    const { predicate, error } = buildWherePredicate()
    if (error) {
      message.value = error
      return
    }

    const limit = pageSize.value
    const offset = Math.max(0, (page - 1) * limit)
    let result

    if (predicate) {
      lastQueryFiltered.value = true
      const table = `${quoteIdent(selectedSchema.value)}.${quoteIdent(selectedTable.value)}`
      result = await runQuery(props.sessionId, `select * from ${table} where ${predicate}`, {
        limit,
        offset,
      })
    } else {
      lastQueryFiltered.value = false
      result = await getTableRows(props.sessionId, selectedSchema.value, selectedTable.value, limit, page)
    }

    currentPage.value = page
    setResult(result)
    totalRows.value = Number(
      predicate
        ? (result.total || result.count || 0)
        : (result.total || tableInformation.value.rows || result.count || 0),
    )
    message.value = predicate
      ? `Loaded ${result.count} rows, ${totalRows.value} matches filter`
      : `Loaded ${result.count} rows`
  } catch (error) {
    message.value = error.message
  }
}

async function applyFilter() {
  await goToPage(1)
}

function setResult(result) {
  const rawColumns = result.columns || []
  const needle = columnSearch.value.trim().toLowerCase()
  columns.value = needle ? rawColumns.filter((column) => column.toLowerCase().includes(needle)) : rawColumns
  rows.value = (result.rows || []).map((row) => (typeof row === 'string' ? JSON.parse(row) : row))
  rowCount.value = result.count || 0
}

function columnType(column) {
  return tableInformation.value.columns.find((item) => item.name === column)?.dataType || ''
}

function columnWidth(column) {
  const type = columnType(column)
  if (type.includes('TIMESTAMP')) return '210px'
  if (type.includes('INT') || type.includes('NUMERIC')) return '96px'
  return '180px'
}

function formatCell(value) {
  if (value === null || value === undefined) return 'NULL'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

function formatNumber(value) {
  return new Intl.NumberFormat('de-DE').format(value)
}

function quoteIdent(value) {
  return `"${String(value).replaceAll('"', '""')}"`
}
</script>
