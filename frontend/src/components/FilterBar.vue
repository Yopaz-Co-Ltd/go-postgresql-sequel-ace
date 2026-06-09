<template>
  <div class="content-filterbar">
    <div class="filter-rules">
      <div v-for="(rule, index) in rules" :key="rule.id" class="filter-rule">
        <input v-model="rule.enabled" class="filter-check" type="checkbox" />
        <select v-model="rule.column" @change="$emit('sync-rule', rule)">
          <option v-for="column in columns" :key="column.name" :value="column.name">
            {{ column.name }}
          </option>
        </select>
        <select v-model="rule.operator" @change="$emit('sync-rule', rule)">
          <option v-for="option in operatorOptionsFor(rule.column)" :key="option.value" :value="option.value">
            {{ option.label }}
          </option>
        </select>
        <div class="filter-value-group" :class="valueModeFor(rule)">
          <template v-if="valueModeFor(rule) === 'single'">
            <input
              v-if="columnKind(rule.column) !== 'boolean'"
              v-model="rule.value"
              class="filter-input"
              :placeholder="valuePlaceholderFor(rule)"
            />
            <select v-else v-model="rule.value" class="filter-input">
              <option value="">select value</option>
              <option value="true">true</option>
              <option value="false">false</option>
            </select>
          </template>
          <template v-else-if="valueModeFor(rule) === 'range'">
            <input v-model="rule.start" class="filter-input" placeholder="from" />
            <input v-model="rule.end" class="filter-input" placeholder="to" />
          </template>
        </div>
        <button type="button" title="Remove filter" @click="$emit('remove-rule', index)"><Minus :size="16" /></button>
        <button type="button" title="Add filter" @click="$emit('add-rule', index)"><Plus :size="16" /></button>
      </div>
    </div>
    <button class="apply-filter" type="button" @click="$emit('apply')">Apply Filter(s)</button>
  </div>
</template>

<script setup>
import { Minus, Plus } from '@lucide/vue'

defineProps({
  columns: { type: Array, default: () => [] },
  rules: { type: Array, default: () => [] },
  columnKind: { type: Function, required: true },
  operatorOptionsFor: { type: Function, required: true },
  valueModeFor: { type: Function, required: true },
  valuePlaceholderFor: { type: Function, required: true },
})

defineEmits(['add-rule', 'apply', 'remove-rule', 'sync-rule'])
</script>
