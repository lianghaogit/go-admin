<template>
  <view class="tree">
    <view
      v-for="node in flatNodes"
      :key="node.key"
      class="tree-row"
      :style="{ paddingLeft: node.depth * 28 + 'rpx' }"
      @tap="toggle(node)"
    >
      <view class="tree-box" :class="{ checked: checkedSet.has(node.key), single: single }">
        <text v-if="checkedSet.has(node.key)" class="tree-tick">✓</text>
      </view>
      <text class="tree-label">{{ node.label }}</text>
      <text v-if="node.tag" class="tree-tag">{{ node.tag }}</text>
    </view>
    <view v-if="!flatNodes.length" class="tree-empty">暂无数据</view>
  </view>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  treeData: { type: Array, default: () => [] },
  checkedKeys: { type: Array, default: () => [] },
  valueKey: { type: String, default: 'menuId' },
  labelKey: { type: String, default: 'menuName' },
  single: { type: Boolean, default: false }
})

const emit = defineEmits(['update:checkedKeys'])

// 拍平树并预计算父子关系，便于级联勾选
const treeIndex = computed(() => {
  const rows = []
  const keyToValue = new Map()
  const parentMap = new Map()
  const childMap = new Map()

  function collectKeys(nodes) {
    const out = []
    ;(nodes || []).forEach((n) => {
      const v = n[props.valueKey]
      if (v === undefined || v === null) return
      out.push(String(v))
      out.push(...collectKeys(n.children))
    })
    return out
  }

  function walk(nodes, depth, parentKey) {
    ;(nodes || []).forEach((n) => {
      const value = n[props.valueKey]
      if (value === undefined || value === null) return
      const key = String(value)
      keyToValue.set(key, value)
      if (parentKey !== null) parentMap.set(key, parentKey)
      rows.push({
        key,
        label: n[props.labelKey] || '未命名',
        depth,
        tag: n.menuType === 'F' ? '按钮' : ''
      })
      if (n.children && n.children.length) {
        walk(n.children, depth + 1, key)
        childMap.set(key, collectKeys(n.children))
      }
    })
  }
  walk(props.treeData, 0, null)
  return { rows, keyToValue, parentMap, childMap }
})

const flatNodes = computed(() => treeIndex.value.rows)
const checkedSet = computed(() => new Set((props.checkedKeys || []).map(String)))

function toggle(node) {
  const { keyToValue, parentMap, childMap } = treeIndex.value
  const next = new Set(checkedSet.value)
  if (props.single) {
    if (next.has(node.key)) {
      next.delete(node.key)
    } else {
      next.clear()
      next.add(node.key)
    }
  } else if (next.has(node.key)) {
    // 取消：同时取消所有子孙
    next.delete(node.key)
    ;(childMap.get(node.key) || []).forEach((k) => next.delete(k))
  } else {
    // 选中：同时选中所有子孙与祖先，保证菜单层级完整
    next.add(node.key)
    ;(childMap.get(node.key) || []).forEach((k) => next.add(k))
    let parent = parentMap.get(node.key)
    while (parent) {
      next.add(parent)
      parent = parentMap.get(parent)
    }
  }
  emit('update:checkedKeys', Array.from(next).map((k) => (keyToValue.has(k) ? keyToValue.get(k) : k)))
}
</script>

<style scoped>
.tree { border: 1rpx solid #e6ebf2; border-radius: 8rpx; background: #fafbfc; max-height: 520rpx; overflow-y: auto; }
.tree-row { display: flex; align-items: center; padding: 14rpx 16rpx; border-bottom: 1rpx solid #f0f3f7; }
.tree-row:last-child { border-bottom: none; }
.tree-box { width: 34rpx; height: 34rpx; border: 1rpx solid #c6d0dc; border-radius: 6rpx; background: #fff; display: flex; align-items: center; justify-content: center; flex-shrink: 0; margin-right: 16rpx; }
.tree-box.checked { background: #409eff; border-color: #409eff; }
.tree-box.single.checked { border-radius: 50%; }
.tree-tick { color: #fff; font-size: 22rpx; line-height: 1; }
.tree-label { flex: 1; font-size: 25rpx; color: #303133; word-break: break-all; }
.tree-tag { font-size: 20rpx; color: #909399; background: #f0f3f7; border-radius: 4rpx; padding: 2rpx 10rpx; flex-shrink: 0; }
.tree-empty { padding: 40rpx 0; text-align: center; color: #b3bcc8; font-size: 24rpx; }
</style>
