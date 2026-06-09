import { ref } from 'vue'

const FILTER_OPERATOR_GROUPS = {
  text: [
    { value: '=', label: '=' },
    { value: '!=', label: '!=' },
    { value: 'like', label: 'LIKE' },
    { value: 'not_like', label: 'NOT LIKE' },
    { value: 'contains', label: 'contains' },
    { value: 'does_not_contain', label: 'does not contain' },
    { value: 'starts_with', label: 'starts with' },
    { value: 'does_not_start_with', label: 'does not start with' },
    { value: 'ends_with', label: 'ends with' },
    { value: 'does_not_end_with', label: 'does not end with' },
    { value: 'matches_regexp', label: 'matches RegExp' },
    { value: 'not_matches_regexp', label: 'does not match RegExp' },
    { value: 'in', label: 'IN' },
    { value: 'not_in_or_null', label: 'NOT IN / OR NULL' },
    { value: 'between', label: 'BETWEEN' },
    { value: 'is_null', label: 'IS NULL' },
    { value: 'is_not_null', label: 'IS NOT NULL' },
    { value: 'is_empty', label: 'is empty' },
    { value: 'is_not_empty', label: 'is not empty' },
  ],
  numeric: [
    { value: '=', label: '=' },
    { value: '!=', label: '!=' },
    { value: '>', label: '>' },
    { value: '>=', label: '>=' },
    { value: '<', label: '<' },
    { value: '<=', label: '<=' },
    { value: 'in', label: 'IN' },
    { value: 'not_in_or_null', label: 'NOT IN / OR NULL' },
    { value: 'between', label: 'BETWEEN' },
    { value: 'is_null', label: 'IS NULL' },
    { value: 'is_not_null', label: 'IS NOT NULL' },
  ],
  boolean: [
    { value: '=', label: '=' },
    { value: '!=', label: '!=' },
    { value: 'is_null', label: 'IS NULL' },
    { value: 'is_not_null', label: 'IS NOT NULL' },
  ],
  default: [
    { value: '=', label: '=' },
    { value: '!=', label: '!=' },
    { value: 'is_null', label: 'IS NULL' },
    { value: 'is_not_null', label: 'IS NOT NULL' },
  ],
}

export function useTableFilters({ getColumnType, getDefaultColumn }) {
  const filterRules = ref([])
  let filterRuleSeed = 0

  function resetFilterRules() {
    filterRules.value = [createFilterRule()]
  }

  function createFilterRule(column = getDefaultColumn()) {
    const nextColumn = column || getDefaultColumn()
    const options = filterOperatorOptionsFor(nextColumn)
    const operator =
      columnKind(nextColumn) === 'text' && options.some((option) => option.value === 'contains')
        ? 'contains'
        : options[0]?.value || 'contains'

    return {
      id: `${Date.now()}-${++filterRuleSeed}`,
      column: nextColumn,
      operator,
      enabled: true,
      value: '',
      start: '',
      end: '',
    }
  }

  function syncRule(rule) {
    const options = filterOperatorOptionsFor(rule.column)
    if (!options.some((option) => option.value === rule.operator)) {
      rule.operator = options[0]?.value || 'contains'
    }
    if (filterValueModeFor(rule) !== 'range') {
      rule.start = ''
      rule.end = ''
    }
    if (filterValueModeFor(rule) === 'none') {
      rule.value = ''
    }
  }

  function addFilterRule(index) {
    filterRules.value.splice(index + 1, 0, createFilterRule())
  }

  function removeFilterRule(index) {
    if (filterRules.value.length === 1) {
      filterRules.value[0] = createFilterRule()
      return
    }
    filterRules.value.splice(index, 1)
  }

  function buildWherePredicate() {
    const predicates = []

    for (const rule of filterRules.value) {
      if (!rule.enabled) continue

      const result = buildRulePredicate(rule)
      if (result.error) return result
      if (result.predicate) predicates.push(result.predicate)
    }

    if (!predicates.length) return { predicate: '' }
    return { predicate: predicates.length === 1 ? predicates[0] : `(${predicates.join(' AND ')})` }
  }

  function columnKind(column) {
    const type = getColumnType(column).toUpperCase()
    if (!type) return 'default'
    if (type.includes('CHAR') || type.includes('TEXT') || type.includes('UUID') || type.includes('JSON')) return 'text'
    if (type.includes('BOOL')) return 'boolean'
    if (
      type.includes('INT') ||
      type.includes('NUMERIC') ||
      type.includes('DECIMAL') ||
      type.includes('REAL') ||
      type.includes('DOUBLE') ||
      type.includes('FLOAT')
    ) {
      return 'numeric'
    }
    if (type.includes('DATE') || type.includes('TIME')) return 'numeric'
    return 'default'
  }

  function filterOperatorOptionsFor(column) {
    return FILTER_OPERATOR_GROUPS[columnKind(column)]
  }

  function filterValueModeFor(rule) {
    if (rule.operator === 'between') return 'range'
    if (operatorNeedsNoValue(rule.operator)) return 'none'
    return 'single'
  }

  function filterValuePlaceholderFor(rule) {
    if (rule.operator === 'in' || rule.operator === 'not_in_or_null') return 'comma-separated values'
    if (rule.operator === 'matches_regexp' || rule.operator === 'not_matches_regexp') return 'regex'
    if (rule.operator === 'like' || rule.operator === 'not_like') return 'pattern'
    return 'value'
  }

  function buildRulePredicate(rule) {
    if (!rule.column) return { error: 'Select a field for each filter row' }

    const column = quoteIdent(rule.column)
    const operator = rule.operator
    const rawValue = String(rule.value || '').trim()
    const rawStart = String(rule.start || '').trim()
    const rawEnd = String(rule.end || '').trim()
    const value = escapeSql(rawValue)
    const startValue = escapeSql(rawStart)
    const endValue = escapeSql(rawEnd)
    const columnTypeName = getColumnType(rule.column).toUpperCase()
    const kind = columnKind(rule.column)
    const textPredicate = (sqlOperator, pattern) => `${column}::text ${sqlOperator} '${pattern}'`
    const typedPredicate = (sqlOperator, rhs = `'${value}'`) => `${column} ${sqlOperator} ${rhs}`

    if (operatorNeedsNoValue(operator)) {
      return {
        predicate:
          operator === 'is_null'
            ? `${column} IS NULL`
            : operator === 'is_not_null'
              ? `${column} IS NOT NULL`
              : operator === 'is_empty'
                ? `${column} = ''`
                : `${column} <> ''`,
      }
    }

    if (operator === 'between') {
      if (!rawStart || !rawEnd) return { predicate: '' }
      return { predicate: `${column} BETWEEN '${startValue}' AND '${endValue}'` }
    }

    if (kind === 'boolean') {
      if (rawValue !== 'true' && rawValue !== 'false') return { predicate: '' }
      return { predicate: `${column} ${operator} ${rawValue}` }
    }

    if (!value) return { predicate: '' }

    if (operator === 'contains') return { predicate: textPredicate('ilike', `%${value}%`) }
    if (operator === 'does_not_contain') return { predicate: `not (${textPredicate('ilike', `%${value}%`)})` }
    if (operator === 'starts_with') return { predicate: textPredicate('ilike', `${value}%`) }
    if (operator === 'does_not_start_with') return { predicate: `not (${textPredicate('ilike', `${value}%`)})` }
    if (operator === 'ends_with') return { predicate: textPredicate('ilike', `%${value}`) }
    if (operator === 'does_not_end_with') return { predicate: `not (${textPredicate('ilike', `%${value}`)})` }
    if (operator === 'like') return { predicate: textPredicate('like', value) }
    if (operator === 'not_like') return { predicate: textPredicate('not like', value) }
    if (operator === 'matches_regexp') return { predicate: textPredicate('~', value) }
    if (operator === 'not_matches_regexp') return { predicate: textPredicate('!~', value) }

    if (operator === 'in' || operator === 'not_in_or_null') {
      const values = rawValue
        .split(',')
        .map((item) => item.trim())
        .filter(Boolean)
        .map((item) => `'${escapeSql(item)}'`)
      if (!values.length) return { predicate: '' }

      const clause = `${column} ${operator === 'in' ? 'IN' : 'NOT IN'} (${values.join(', ')})`
      return { predicate: operator === 'in' ? clause : `(${clause} OR ${column} IS NULL)` }
    }

    if (['=', '!=', '>', '>=', '<', '<='].includes(operator)) {
      return {
        predicate:
          columnTypeName.includes('CHAR') || columnTypeName.includes('TEXT') || columnTypeName.includes('UUID')
            ? textPredicate(operator, value)
            : typedPredicate(operator),
      }
    }

    return { predicate: textPredicate('ilike', `%${value}%`) }
  }

  return {
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
  }
}

function operatorNeedsNoValue(operator) {
  return ['is_null', 'is_not_null', 'is_empty', 'is_not_empty'].includes(operator)
}

function quoteIdent(value) {
  return `"${String(value).replaceAll('"', '""')}"`
}

function escapeSql(value) {
  return String(value).replaceAll("'", "''")
}
