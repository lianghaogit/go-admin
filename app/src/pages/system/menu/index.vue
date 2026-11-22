<template>
  <view class="page">
    <view class="top-bar">
      <view v-if="$hasPerm('system:menu:add')" class="add-btn" @tap="openForm()">＋ 新增菜单</view>
    </view>

    <view v-if="rows.length" class="menu-list">
      <view
        v-for="r in rows"
        :key="r.node.menuId"
        class="menu-item"
        :style="{ paddingLeft: 28 + r.level * 40 + 'rpx' }"
      >
        <text class="arrow" @tap="toggle(r.node)">{{ hasChild(r.node) ? (expanded.has(r.node.menuId) ? '▾' : '▸') : '' }}</text>
        <view class="menu-main" @tap="toggle(r.node)">
          <view class="menu-line">
            <text class="menu-name">{{ r.node.menuName }}</text>
            <text class="type-tag" :class="r.node.menuType">{{ typeText(r.node.menuType) }}</text>
            <text class="status" :class="r.node.status === '0' ? 'on' : 'off'" @tap.stop="toggleStatus(r.node)">{{ r.node.status === '0' ? '正常' : '停用' }}</text>
          </view>
          <view class="menu-line sub">
            <text class="menu-path">{{ r.node.path || '—' }}</text>
            <text v-if="r.node.perms" class="menu-perms">{{ r.node.perms }}</text>
          </view>
        </view>
        <view v-if="$hasPerm('system:menu:edit')" class="op" @tap="openForm(r.node)">✎</view>
        <view v-if="$hasPerm('system:menu:remove')" class="op danger" @tap="onDelete(r.node)">✕</view>
      </view>
      <view v-if="loading" class="load-hint">
        <text class="load-spin"></text>
        <text>加载中…</text>
      </view>
    </view>
    <view v-else-if="!loading" class="empty">暂无菜单数据</view>
    <view v-else class="empty">加载中…</view>

    <!-- 新增/编辑弹层 -->
    <view v-if="showForm" class="mask" @tap="closeForm">
      <view class="sheet" @tap.stop>
        <view class="sheet-head">
          <view class="sheet-avatar">{{ (form.menuName || 'M').charAt(0).toUpperCase() }}</view>
          <view class="sheet-head-main">
            <text class="sheet-title">{{ form.menuId ? '编辑菜单' : '新增菜单' }}</text>
            <text class="sheet-sub">{{ form.menuId ? '调整菜单路由与访问权限' : '添加一个目录、页面或按钮' }}</text>
          </view>
          <view class="sheet-close" @tap="closeForm">×</view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">基本信息</text>
          </view>
          <view class="field">
            <text class="field-label">上级菜单</text>
            <picker :range="parentOptions" range-key="label" :value="parentIndex" @change="onParentChange">
              <view class="field-select">
                <text class="field-select-text">{{ parentOptions[parentIndex].label }}</text>
                <text class="field-select-arrow">›</text>
              </view>
            </picker>
          </view>
          <view class="field">
            <text class="field-label">菜单类型</text>
            <picker :range="typeOptions" range-key="label" :value="typeIndex" @change="onTypeChange">
              <view class="field-select">
                <text class="field-select-text">{{ typeOptions[typeIndex].label }}</text>
                <text class="field-select-arrow">›</text>
              </view>
            </picker>
          </view>
          <view class="field">
            <text class="field-label">菜单名称</text>
            <input v-model="form.menuName" class="field-input" placeholder="必填，如 用户管理" placeholder-class="ph" />
          </view>
          <view class="field">
            <text class="field-label">显示排序</text>
            <input v-model="form.orderNum" class="field-input" type="number" placeholder="数字越小越靠前" placeholder-class="ph" />
          </view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">路由与权限</text>
          </view>
          <view v-if="form.menuType !== 'F'" class="field">
            <text class="field-label">路由地址</text>
            <input v-model="form.path" class="field-input" placeholder="相对路径，如 user" placeholder-class="ph" />
          </view>
          <view v-if="form.menuType === 'C'" class="field">
            <text class="field-label">组件路径</text>
            <input v-model="form.component" class="field-input" placeholder="如 system/user/index" placeholder-class="ph" />
          </view>
          <view v-if="form.menuType !== 'M'" class="field">
            <text class="field-label">权限标识</text>
            <input v-model="form.perms" class="field-input" placeholder="如 system:user:list" placeholder-class="ph" />
          </view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">显示设置</text>
          </view>
          <view class="field">
            <text class="field-label">图标</text>
            <input v-model="form.icon" class="field-input" placeholder="图标名，如 user" placeholder-class="ph" />
          </view>
          <view class="field">
            <text class="field-label">是否显示</text>
            <picker :range="visibleOptions" range-key="label" :value="visibleIndex" @change="onVisibleChange">
              <view class="field-select">
                <text class="field-select-text">{{ visibleOptions[visibleIndex].label }}</text>
                <text class="field-select-arrow">›</text>
              </view>
            </picker>
          </view>
          <view class="field-card">
            <view class="field-main">
              <text class="field-label field-label-inline">菜单状态</text>
              <text class="field-desc">停用后该菜单对已授权用户隐藏</text>
            </view>
            <switch :checked="form.status === '0'" color="#3a6fe0" style="transform: scale(0.7)" @change="onSwitch" />
            <text class="status-pill" :class="{ off: form.status !== '0' }">{{ form.status === '0' ? '正常' : '停用' }}</text>
          </view>
        </view>

        <view class="form-actions">
          <view class="btn plain" @tap="closeForm">取消</view>
          <view class="btn" @tap="submitForm">保存</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed, reactive } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { listMenu, createMenu, updateMenu, deleteMenu } from '@/api/menu'
import { trimStr, isNonNegInt } from '@/utils/validate'

const tree = ref([])
const loading = ref(false)
const expanded = ref(new Set())
const showForm = ref(false)
const form = reactive({ menuId: 0, parentId: 0, menuName: '', orderNum: 1, path: '', component: '', menuType: 'C', perms: '', icon: '', visible: '0', status: '0' })

const typeOptions = [
  { value: 'M', label: '目录' },
  { value: 'C', label: '菜单' },
  { value: 'F', label: '按钮' }
]
const visibleOptions = [
  { value: '0', label: '显示' },
  { value: '1', label: '隐藏' }
]
const typeIndex = computed(() => Math.max(0, typeOptions.findIndex((o) => o.value === form.menuType)))
const visibleIndex = computed(() => Math.max(0, visibleOptions.findIndex((o) => o.value === form.visible)))

function hasChild(n) {
  return !!(n.children && n.children.length)
}

function typeText(t) {
  return t === 'M' ? '目录' : t === 'C' ? '页面' : '按钮'
}

const rows = computed(() => {
  const out = []
  function walk(nodes, level) {
    for (const n of nodes || []) {
      out.push({ node: n, level })
      if (hasChild(n) && expanded.value.has(n.menuId)) {
        walk(n.children, level + 1)
      }
    }
  }
  walk(tree.value, 0)
  return out
})

const parentOptions = computed(() => {
  const out = [{ value: 0, label: '顶级菜单' }]
  const skipId = form.menuId
  function walk(nodes, level) {
    for (const n of nodes || []) {
      // 编辑时排除自身及其子孙，避免形成环
      if (skipId && n.menuId === skipId) continue
      out.push({ value: n.menuId, label: `${'　'.repeat(level)}${n.menuName}` })
      if (hasChild(n)) walk(n.children, level + 1)
    }
  }
  walk(tree.value, 0)
  return out
})
const parentIndex = computed(() => Math.max(0, parentOptions.value.findIndex((o) => o.value === form.parentId)))

function toggle(node) {
  if (!hasChild(node)) return
  const s = new Set(expanded.value)
  if (s.has(node.menuId)) s.delete(node.menuId)
  else s.add(node.menuId)
  expanded.value = s
}

async function fetchTree() {
  loading.value = true
  try {
    const data = await listMenu()
    if (data.code === 200) {
      tree.value = data.data || []
      const s = new Set()
      for (const n of tree.value) {
        if (hasChild(n)) s.add(n.menuId)
      }
      expanded.value = s
    } else {
      uni.showToast({ title: data.msg || '查询失败', icon: 'none' })
    }
  } catch (e) {
  } finally {
    loading.value = false
  }
}

// ---- 表单 ----
function openForm(m) {
  if (m) {
    Object.assign(form, {
      menuId: m.menuId,
      parentId: m.parentId || 0,
      menuName: m.menuName || '',
      orderNum: m.orderNum ?? 1,
      path: m.path || '',
      component: m.component || '',
      menuType: m.menuType || 'C',
      perms: m.perms || '',
      icon: m.icon || '',
      visible: m.visible ?? '0',
      status: m.status || '0'
    })
  } else {
    Object.assign(form, { menuId: 0, parentId: 0, menuName: '', orderNum: 1, path: '', component: '', menuType: 'C', perms: '', icon: '', visible: '0', status: '0' })
  }
  showForm.value = true
}

function closeForm() {
  showForm.value = false
}

function onSwitch(e) {
  form.status = e.detail.value ? '0' : '1'
}

function onParentChange(e) {
  form.parentId = parentOptions.value[Number(e.detail.value)].value
}

function onTypeChange(e) {
  form.menuType = typeOptions[Number(e.detail.value)].value
}

function onVisibleChange(e) {
  form.visible = visibleOptions[Number(e.detail.value)].value
}

async function submitForm() {
  const menuName = trimStr(form.menuName)
  const path = trimStr(form.path)
  const component = trimStr(form.component)
  const perms = trimStr(form.perms)
  const icon = trimStr(form.icon)
  if (!menuName) {
    uni.showToast({ title: '请输入菜单名称', icon: 'none' })
    return
  }
  if (menuName.length > 50) {
    uni.showToast({ title: '菜单名称不能超过 50 个字符', icon: 'none' })
    return
  }
  if (!isNonNegInt(form.orderNum)) {
    uni.showToast({ title: '显示排序需为不小于 0 的整数', icon: 'none' })
    return
  }
  if (form.menuType !== 'F' && !path) {
    uni.showToast({ title: '请输入路由地址', icon: 'none' })
    return
  }
  if (path && /\s/.test(path)) {
    uni.showToast({ title: '路由地址不能包含空格', icon: 'none' })
    return
  }
  if (form.menuType === 'C' && !component) {
    uni.showToast({ title: '请输入组件路径', icon: 'none' })
    return
  }
  if (component && /\s/.test(component)) {
    uni.showToast({ title: '组件路径不能包含空格', icon: 'none' })
    return
  }
  if (form.menuType === 'F' && !perms) {
    uni.showToast({ title: '按钮菜单必须填写权限标识', icon: 'none' })
    return
  }
  if (form.menuId && form.parentId && String(form.parentId) === String(form.menuId)) {
    uni.showToast({ title: '上级菜单不能选择自身', icon: 'none' })
    return
  }
  Object.assign(form, { menuName, path, component, perms, icon })
  const payload = {
    parentId: form.parentId,
    menuName,
    orderNum: Number(form.orderNum) || 0,
    path,
    component,
    menuType: form.menuType,
    perms,
    icon,
    visible: form.visible,
    status: form.status
  }
  const data = form.menuId ? await updateMenu({ ...payload, menuId: form.menuId }) : await createMenu(payload)
  if (data.code === 200) {
    closeForm()
    uni.showToast({ title: form.menuId ? '修改成功' : '新增成功', icon: 'success', duration: 1200 })
    setTimeout(() => fetchTree(), 400)
  } else {
    uni.showToast({ title: data.msg || '操作失败', icon: 'none' })
  }
}

// ---- 状态切换 ----
function toggleStatus(m) {
  const next = m.status === '0' ? '1' : '0'
  openForm({ ...m, status: next })
}

// ---- 删除 ----
function onDelete(m) {
  uni.showModal({
    title: '提示',
    content: `确定删除菜单「${m.menuName}」吗？`,
    confirmText: '删除',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      const data = await deleteMenu(m.menuId)
      if (data.code === 200) {
        uni.showToast({ title: '删除成功', icon: 'success' })
        fetchTree()
      } else {
        uni.showToast({ title: data.msg || '删除失败', icon: 'none' })
      }
    }
  })
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '菜单管理' })
  fetchTree()
})

onPullDownRefresh(async () => {
  try {
    await fetchTree()
    uni.showToast({ title: '已刷新', icon: 'none', duration: 800 })
  } catch (e) {
    uni.showToast({ title: '刷新失败', icon: 'none' })
  } finally {
    uni.stopPullDownRefresh()
  }
})
</script>

<style scoped>
.page {
  padding: 28rpx 28rpx 60rpx;
}

.top-bar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 20rpx;
}

.add-btn {
  padding: 0 26rpx;
  height: 68rpx;
  line-height: 68rpx;
  background: #67c23a;
  color: #fff;
  font-size: 26rpx;
  border-radius: 18rpx;
}

.menu-list {
  background: #fff;
  border-radius: 18rpx;
  padding: 0 28rpx;
  box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06);
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 26rpx 0;
}

.menu-item + .menu-item {
  border-top: 1rpx solid #f0f3f7;
}

.arrow {
  font-size: 24rpx;
  color: #c3ccd8;
  width: 32rpx;
  flex-shrink: 0;
  text-align: center;
}

.menu-main {
  flex: 1;
  min-width: 0;
}

.menu-line {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.menu-name {
  font-size: 28rpx;
  color: #1f2d3d;
}

.type-tag {
  font-size: 20rpx;
  padding: 2rpx 14rpx;
  border-radius: 18rpx;
  flex-shrink: 0;
}

.type-tag.M {
  color: #409eff;
  background: rgba(64, 158, 255, 0.1);
}

.type-tag.C {
  color: #67c23a;
  background: rgba(103, 194, 58, 0.1);
}

.type-tag.F {
  color: #909399;
  background: rgba(144, 147, 153, 0.1);
}

.status {
  font-size: 20rpx;
  padding: 2rpx 14rpx;
  border-radius: 18rpx;
  flex-shrink: 0;
}

.status.on {
  color: #67c23a;
  background: rgba(103, 194, 58, 0.1);
}

.status.off {
  color: #f56c6c;
  background: rgba(245, 108, 108, 0.1);
}

.menu-line.sub {
  margin-top: 6rpx;
  font-size: 22rpx;
  color: #9aa6b8;
}

.menu-perms {
  color: #8a94a6;
}

.menu-perms::before {
  content: '·';
  margin: 0 8rpx;
}

.op {
  width: 60rpx;
  height: 60rpx;
  line-height: 60rpx;
  text-align: center;
  font-size: 28rpx;
  color: #409eff;
  background: rgba(64, 158, 255, 0.08);
  border-radius: 50%;
  flex-shrink: 0;
}

.op.danger {
  color: #f56c6c;
  background: rgba(245, 108, 108, 0.08);
}

.load-hint {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  padding: 24rpx 0;
  font-size: 24rpx;
  color: #9aa6b8;
}

.load-spin {
  width: 28rpx;
  height: 28rpx;
  border: 4rpx solid #e6ebf2;
  border-top-color: #409eff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.empty {
  margin-top: 24rpx;
  padding: 100rpx 0;
  text-align: center;
  color: #9aa6b8;
  font-size: 26rpx;
  background: #fff;
  border-radius: 18rpx;
}

/* 弹层 */
.mask {
  position: fixed;
  inset: 0;
  background: rgba(31, 45, 61, 0.45);
  z-index: 99;
  display: flex;
  align-items: flex-end;
  justify-content: center;
}

.sheet {
  width: 100%;
  max-height: 86vh;
  overflow-y: auto;
  background: #fff;
  border-radius: 16rpx 16rpx 0 0;
  padding: 32rpx 32rpx 48rpx;
  box-sizing: border-box;
}

.sheet-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1f2d3d;
  text-align: center;
  margin-bottom: 28rpx;
}

.form-item {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 18rpx 0;
  border-bottom: 1rpx solid #f0f3f7;
}

.label {
  width: 140rpx;
  font-size: 26rpx;
  color: #5a6678;
  flex-shrink: 0;
}

.input {
  flex: 1;
  height: 64rpx;
  font-size: 26rpx;
}

.picker-value {
  height: 64rpx;
  line-height: 64rpx;
  font-size: 26rpx;
  color: #1f2d3d;
}

.switch-text {
  font-size: 24rpx;
  color: #8a94a6;
}

.form-actions {
  display: flex;
  gap: 20rpx;
  margin-top: 40rpx;
}

.btn {
  flex: 1;
  height: 84rpx;
  line-height: 84rpx;
  text-align: center;
  background: #409eff;
  color: #fff;
  font-size: 28rpx;
  border-radius: 18rpx;
}

.btn.plain {
  background: #f5f7fa;
  color: #5a6678;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* ===== 弹层精修（新版） ===== */
.sheet {
  padding: 0;
  max-height: 90vh;
  border-radius: 28rpx 28rpx 0 0;
  overflow-y: auto;
  background: #fff;
}

.sheet-head {
  position: sticky;
  top: 0;
  z-index: 5;
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 30rpx 32rpx 24rpx;
  background: #fff;
  border-bottom: 1rpx solid #f2f5fb;
}

.sheet-avatar {
  width: 88rpx;
  height: 88rpx;
  border-radius: 24rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36rpx;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
  background: linear-gradient(135deg, #6aa4ff, #3a6fe0);
  box-shadow: 0 10rpx 24rpx rgba(58, 111, 224, 0.26);
}

.sheet-head-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.sheet-head .sheet-title {
  font-size: 32rpx;
  font-weight: 700;
  color: #1c2738;
  text-align: left;
  margin: 0;
}

.sheet-sub {
  font-size: 22rpx;
  color: #9aa6b8;
}

.sheet-close {
  width: 56rpx;
  height: 56rpx;
  line-height: 56rpx;
  text-align: center;
  font-size: 32rpx;
  color: #9aa6b8;
  background: #f3f6fb;
  border-radius: 50%;
  flex-shrink: 0;
}

.group {
  padding: 28rpx 32rpx 0;
}

.group-head {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 20rpx;
}

.group-bar {
  width: 6rpx;
  height: 26rpx;
  border-radius: 4rpx;
  background: linear-gradient(180deg, #6aa4ff, #3a6fe0);
}

.group-title {
  font-size: 26rpx;
  font-weight: 600;
  color: #1c2738;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  margin-bottom: 24rpx;
}

.field-label {
  font-size: 24rpx;
  color: #6b7a90;
}

.field-input {
  width: 100%;
  height: 84rpx;
  padding: 0 24rpx;
  box-sizing: border-box;
  background: #f5f7fb;
  border-radius: 14rpx;
  font-size: 27rpx;
  color: #1c2738;
}

.field-select {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 84rpx;
  padding: 0 24rpx;
  background: #f5f7fb;
  border-radius: 14rpx;
}

.field-select-text {
  flex: 1;
  min-width: 0;
  font-size: 27rpx;
  color: #1c2738;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.field-select-arrow {
  font-size: 30rpx;
  color: #b9c3d3;
}

.field-card {
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 24rpx;
  margin-bottom: 24rpx;
  background: #f5f7fb;
  border-radius: 16rpx;
}

.field-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}

.field-label-inline {
  font-size: 26rpx;
  font-weight: 500;
  color: #1c2738;
}

.field-desc {
  font-size: 21rpx;
  color: #9aa6b8;
}

.status-pill {
  font-size: 20rpx;
  padding: 4rpx 16rpx;
  border-radius: 20rpx;
  color: #3fae7a;
  background: rgba(63, 174, 122, 0.12);
  flex-shrink: 0;
}

.status-pill.off {
  color: #f56c6c;
  background: rgba(245, 108, 108, 0.12);
}

.form-actions {
  position: sticky;
  bottom: 0;
  z-index: 5;
  display: flex;
  gap: 20rpx;
  padding: 20rpx 32rpx 40rpx;
  background: #fff;
  border-top: 1rpx solid #f2f5fb;
}

.btn {
  flex: 1;
  height: 92rpx;
  line-height: 92rpx;
  text-align: center;
  color: #fff;
  font-size: 28rpx;
  font-weight: 600;
  border-radius: 18rpx;
  background: linear-gradient(135deg, #4a7df0, #3457c9);
  box-shadow: 0 12rpx 26rpx rgba(52, 87, 201, 0.24);
}

.btn.plain {
  background: #f3f6fb;
  color: #6b7a90;
  box-shadow: none;
}
</style>
