<template>
  <view class="page">
    <view class="search-bar">
      <input v-model="keyword" class="search-input" placeholder="输入表名搜索" placeholder-class="ph" confirm-type="search" @confirm="onSearch" />
      <view class="search-btn" @tap="onSearch">搜索</view>
      <view class="search-btn plain" @tap="onReset">重置</view>
    </view>
    <view class="tool-bar">
      <view v-if="$hasPerm('tool:gen:import')" class="tool-btn add" @tap="openImport">＋ 导入表</view>
    </view>

    <view v-if="list.length" class="gen-list">
      <view v-for="g in list" :key="g.tableId" class="gen-item">
        <view class="gen-line">
          <text class="gen-name">{{ g.tableName }}</text>
          <text class="gen-comment">{{ g.tableComment || '无描述' }}</text>
        </view>
        <view class="gen-line sub">
          <text>类名 {{ g.className || '—' }}</text>
          <text class="time">{{ g.createTime || '—' }}</text>
        </view>
        <view class="gen-ops">
          <text v-if="$hasPerm('tool:gen:edit')" class="op" @tap="openConfig(g)">配置</text>
          <text v-if="$hasPerm('tool:gen:preview')" class="op" @tap="onPreview(g)">预览代码</text>
          <text v-if="$hasPerm('tool:gen:code')" class="op warn" @tap="onDownload(g)">下载代码</text>
          <text v-if="$hasPerm('tool:gen:remove')" class="op danger" @tap="onDelete(g)">删除</text>
        </view>
      </view>
      <view v-if="loading" class="load-hint"><text class="load-spin"></text><text>加载中…</text></view>
      <view v-else-if="finished" class="load-hint done">没有更多了</view>
    </view>
    <view v-else-if="!loading" class="empty">暂无生成配置，请先导入数据表</view>
    <view v-else class="empty">加载中…</view>

    <!-- 导入表 -->
    <view v-if="showImport" class="mask" @tap="showImport = false">
      <view class="sheet" @tap.stop>
        <view class="sheet-title">导入数据库表</view>
        <view class="db-search">
          <input v-model="dbKeyword" class="search-input" placeholder="输入表名筛选" placeholder-class="ph" />
          <view class="search-btn" @tap="loadDbTables">查询</view>
        </view>
        <scroll-view scroll-y class="db-scroll">
          <view v-for="t in dbTables" :key="t.tableName" class="db-item" :class="{ selected: selected.includes(t.tableName) }" @tap="toggleSelect(t.tableName)">
            <view class="db-main">
              <text class="db-name">{{ t.tableName }}</text>
              <text class="db-comment">{{ t.tableComment || '无描述' }}</text>
            </view>
            <text class="checkbox">{{ selected.includes(t.tableName) ? '✓' : '' }}</text>
          </view>
          <view v-if="!dbTables.length" class="hint">没有可导入的业务表</view>
        </scroll-view>
        <view class="form-actions">
          <view class="btn plain" @tap="showImport = false">取消</view>
          <view class="btn" @tap="submitImport">导入（{{ selected.length }}）</view>
        </view>
      </view>
    </view>

    <!-- 代码预览 -->
    <view v-if="showPreview" class="mask" @tap="showPreview = false">
      <view class="preview-panel" @tap.stop>
        <view class="preview-head">
          <text class="preview-title">{{ previewTable.tableName }} · 代码预览</text>
          <text class="close" @tap="showPreview = false">关闭</text>
        </view>
        <scroll-view scroll-x class="file-tabs">
          <text v-for="f in fileNames" :key="f" class="file-tab" :class="{ active: activeFile === f }" @tap="activeFile = f">{{ shortName(f) }}</text>
        </scroll-view>
        <scroll-view scroll-y class="code-box">
          <text class="code">{{ previewFiles[activeFile] || '' }}</text>
        </scroll-view>
      </view>
    </view>
    <!-- 生成配置 -->
    <view v-if="showConfig" class="mask" @tap="showConfig = false">
      <view class="sheet config-sheet" @tap.stop>
        <view class="sheet-title">{{ configForm.tableName }} · 生成配置</view>
        <scroll-view scroll-y class="config-scroll">
          <view class="form-item">
            <text class="label">表描述</text>
            <input v-model="configForm.tableComment" class="input" placeholder="表描述" placeholder-class="ph" />
          </view>
          <view class="form-item">
            <text class="label">实体类名</text>
            <input v-model="configForm.className" class="input" placeholder="如 User" placeholder-class="ph" />
          </view>
          <view class="form-item">
            <text class="label">模块名</text>
            <input v-model="configForm.moduleName" class="input" placeholder="如 system" placeholder-class="ph" />
          </view>
          <view class="form-item">
            <text class="label">包名</text>
            <input v-model="configForm.packageName" class="input" placeholder="如 user" placeholder-class="ph" />
          </view>
          <view class="form-item">
            <text class="label">功能名</text>
            <input v-model="configForm.functionName" class="input" placeholder="如 用户" placeholder-class="ph" />
          </view>
          <view class="form-item">
            <text class="label">作者</text>
            <input v-model="configForm.author" class="input" placeholder="作者" placeholder-class="ph" />
          </view>
          <view class="form-item">
            <text class="label">备注</text>
            <input v-model="configForm.remark" class="input" placeholder="备注" placeholder-class="ph" />
          </view>

          <view class="config-section">
            <text class="section-label">列配置（共 {{ configForm.columns.length }} 列）</text>
            <view v-for="(c, i) in configForm.columns" :key="c.columnId || i" class="col-card">
              <view class="col-head">
                <text class="col-name">{{ c.columnName }}</text>
                <text class="col-type">{{ c.columnComment || c.goType || '—' }}</text>
              </view>
              <view class="col-switches">
                <view class="col-switch">
                  <switch :checked="c.isList === 1" color="#409eff" style="transform: scale(0.6)" @change="c.isList = $event.detail.value ? 1 : 0" />
                  <text>列表</text>
                </view>
                <view class="col-switch">
                  <switch :checked="c.isQuery === 1" color="#409eff" style="transform: scale(0.6)" @change="c.isQuery = $event.detail.value ? 1 : 0" />
                  <text>查询</text>
                </view>
                <view class="col-switch">
                  <switch :checked="c.isForm === 1" color="#409eff" style="transform: scale(0.6)" @change="c.isForm = $event.detail.value ? 1 : 0" />
                  <text>表单</text>
                </view>
                <view class="col-switch">
                  <switch :checked="c.isRequired === 1" color="#409eff" style="transform: scale(0.6)" @change="c.isRequired = $event.detail.value ? 1 : 0" />
                  <text>必填</text>
                </view>
              </view>
              <view class="col-pickers">
                <picker :range="QUERY_TYPES" @change="c.queryType = QUERY_TYPES[$event.detail.value]">
                  <view class="picker-value">查询方式 {{ c.queryType || 'EQ' }} ▾</view>
                </picker>
                <picker :range="FORM_TYPES" @change="c.formType = FORM_TYPES[$event.detail.value]">
                  <view class="picker-value">表单类型 {{ c.formType || 'input' }} ▾</view>
                </picker>
              </view>
            </view>
          </view>
        </scroll-view>
        <view class="form-actions">
          <view class="btn plain" @tap="showConfig = false">取消</view>
          <view class="btn" @tap="submitConfig">保存</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { listGenTables, listDbTables, importGenTables, deleteGenTable, previewGenTable, getGenTable, saveGenTable } from '@/api/tool'
import { exportFile } from '@/utils/download'
import { trimStr } from '@/utils/validate'

const keyword = ref('')
const list = ref([])
const pageNum = ref(1)
const pageSize = 10
const total = ref(0)
const loading = ref(false)
const finished = ref(false)

const showImport = ref(false)
const dbKeyword = ref('')
const dbTables = ref([])
const selected = ref([])

const showPreview = ref(false)
const previewTable = ref({})
const previewFiles = ref({})
const fileNames = ref([])
const activeFile = ref('')

const showConfig = ref(false)
const configForm = reactive({
  tableId: 0, tableName: '', tableComment: '', className: '', moduleName: '',
  packageName: '', functionName: '', author: '', remark: '', columns: []
})
const QUERY_TYPES = ['EQ', 'NE', 'GT', 'LT', 'LIKE']
const FORM_TYPES = ['input', 'textarea', 'select', 'date']

async function fetchList(reset) {
  if (loading.value) return
  if (reset) { pageNum.value = 1; finished.value = false }
  else if (finished.value) return
  loading.value = true
  try {
    const data = await listGenTables({ pageNum: pageNum.value, pageSize, tableName: keyword.value || undefined })
    total.value = data.total
    const rows = data.rows || []
    list.value = reset ? rows : list.value.concat(rows)
    finished.value = list.value.length >= data.total
    if (rows.length) pageNum.value += 1
  } catch (e) {
  } finally {
    loading.value = false
  }
}

function onSearch() { fetchList(true) }
function onReset() { keyword.value = ''; fetchList(true) }

function openImport() {
  selected.value = []
  dbKeyword.value = ''
  dbTables.value = []
  showImport.value = true
  loadDbTables()
}

async function loadDbTables() {
  try {
    const data = await listDbTables(dbKeyword.value)
    dbTables.value = data.data || []
  } catch (e) {}
}

function toggleSelect(name) {
  const idx = selected.value.indexOf(name)
  if (idx >= 0) selected.value.splice(idx, 1)
  else selected.value.push(name)
}

async function submitImport() {
  if (!selected.value.length) {
    uni.showToast({ title: '请先选择要导入的表', icon: 'none' })
    return
  }
  try {
    await importGenTables(selected.value)
    showImport.value = false
    uni.showToast({ title: '导入成功', icon: 'success' })
    fetchList(true)
  } catch (e) {}
}

function onDelete(g) {
  uni.showModal({
    title: '提示',
    content: `确定删除表「${g.tableName}」的生成配置吗？`,
    confirmText: '删除',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await deleteGenTable(g.tableId)
        list.value = list.value.filter((it) => it.tableId !== g.tableId)
        total.value = Math.max(0, total.value - 1)
        uni.showToast({ title: '删除成功', icon: 'success' })
        if (!list.value.length) fetchList(true)
      } catch (e) {}
    }
  })
}

async function onPreview(g) {
  try {
    const data = await previewGenTable(g.tableId)
    previewFiles.value = data.data || {}
    fileNames.value = Object.keys(previewFiles.value)
    activeFile.value = fileNames.value[0] || ''
    previewTable.value = g
    showPreview.value = true
  } catch (e) {}
}

function shortName(f) {
  return String(f).split(/[\\/]/).pop()
}

// 下载生成的代码压缩包
function onDownload(g) {
  exportFile(`/api/tool/gen/code/${g.tableId}`, {}, `${g.tableName || 'code'}.zip`, 'application/zip')
}

// ---- 生成配置 ----
async function openConfig(g) {
  try {
    const data = await getGenTable(g.tableId)
    const t = (data && data.data) || {}
    Object.assign(configForm, {
      tableId: t.tableId,
      tableName: t.tableName || '',
      tableComment: t.tableComment || '',
      className: t.className || '',
      moduleName: t.moduleName || '',
      packageName: t.packageName || '',
      functionName: t.functionName || '',
      author: t.author || '',
      remark: t.remark || '',
      columns: (t.columns || []).map((c) => ({ ...c }))
    })
    showConfig.value = true
  } catch (e) {}
}

async function submitConfig() {
  const className = trimStr(configForm.className)
  const tableComment = trimStr(configForm.tableComment)
  const moduleName = trimStr(configForm.moduleName)
  const packageName = trimStr(configForm.packageName)
  const functionName = trimStr(configForm.functionName)
  const author = trimStr(configForm.author)
  if (!className) {
    uni.showToast({ title: '请填写实体类名', icon: 'none' })
    return
  }
  if (!/^[A-Za-z][A-Za-z0-9_]*$/.test(className)) {
    uni.showToast({ title: '实体类名需以字母开头，仅含字母、数字和下划线', icon: 'none' })
    return
  }
  if (moduleName && !/^[A-Za-z][A-Za-z0-9_-]*$/.test(moduleName)) {
    uni.showToast({ title: '模块名仅支持字母、数字、下划线或中划线', icon: 'none' })
    return
  }
  if (packageName && !/^[A-Za-z][A-Za-z0-9_.]*$/.test(packageName)) {
    uni.showToast({ title: '包名仅支持字母、数字、点和下划线', icon: 'none' })
    return
  }
  try {
    await saveGenTable({
      tableId: configForm.tableId,
      tableName: configForm.tableName,
      tableComment,
      className,
      moduleName,
      packageName,
      functionName,
      author,
      remark: trimStr(configForm.remark),
      columns: configForm.columns
    })
    showConfig.value = false
    uni.showToast({ title: '配置已保存', icon: 'success' })
    fetchList(true)
  } catch (e) {}
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '代码生成' })
  fetchList(true)
})

onPullDownRefresh(async () => {
  try {
    await fetchList(true)
    uni.showToast({ title: '已刷新', icon: 'none', duration: 800 })
  } catch (e) {
  } finally {
    uni.stopPullDownRefresh()
  }
})

onReachBottom(() => { fetchList(false) })
</script>

<style scoped>
.page { padding: 28rpx 28rpx 60rpx; }
.search-bar { display: flex; align-items: center; gap: 12rpx; padding: 20rpx 24rpx; background: #fff; border-radius: 18rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.search-input { flex: 1; min-width: 0; height: 68rpx; padding: 0 20rpx; background: #f5f7fa; border-radius: 18rpx; font-size: 26rpx; }
.ph { color: #9aa6b8; }
.search-btn { padding: 0 22rpx; height: 68rpx; line-height: 68rpx; background: #409eff; color: #fff; font-size: 26rpx; border-radius: 18rpx; flex-shrink: 0; }
.search-btn.plain { background: #f5f7fa; color: #5a6678; }
.tool-bar { display: flex; justify-content: flex-end; margin-top: 16rpx; }
.tool-btn { padding: 0 22rpx; height: 64rpx; line-height: 64rpx; background: #67c23a; color: #fff; font-size: 26rpx; border-radius: 18rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.gen-list { margin-top: 24rpx; background: #fff; border-radius: 18rpx; padding: 0 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.gen-item { padding: 26rpx 0; }
.gen-item + .gen-item { border-top: 1rpx solid #f0f3f7; }
.gen-line { display: flex; align-items: center; gap: 12rpx; }
.gen-name { font-size: 28rpx; color: #1f2d3d; font-weight: 600; word-break: break-all; }
.gen-comment { margin-left: auto; font-size: 22rpx; color: #8a94a6; flex-shrink: 0; max-width: 40%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.gen-line.sub { margin-top: 8rpx; font-size: 22rpx; color: #8a94a6; }
.time { margin-left: auto; }
.gen-ops { display: flex; gap: 16rpx; margin-top: 16rpx; }
.op { font-size: 24rpx; padding: 8rpx 22rpx; border-radius: 18rpx; color: #409eff; background: rgba(64, 158, 255, 0.08); }
.op.danger { color: #f56c6c; background: rgba(245, 108, 108, 0.08); }
.op.warn { color: #e6a23c; background: rgba(230, 162, 60, 0.12); }
.load-hint { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 24rpx 0; font-size: 24rpx; color: #9aa6b8; }
.load-spin { width: 28rpx; height: 28rpx; border: 4rpx solid #e6ebf2; border-top-color: #409eff; border-radius: 50%; animation: spin 0.8s linear infinite; }
.load-hint.done { color: #c3ccd8; }
.empty { margin-top: 24rpx; padding: 100rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
.mask { position: fixed; inset: 0; background: rgba(31, 45, 61, 0.45); z-index: 99; display: flex; align-items: flex-end; justify-content: center; }
.sheet { width: 100%; max-height: 88vh; background: #fff; border-radius: 16rpx 16rpx 0 0; padding: 32rpx 32rpx 48rpx; box-sizing: border-box; display: flex; flex-direction: column; }
.sheet-title { font-size: 32rpx; font-weight: 600; color: #1f2d3d; text-align: center; }
.db-search { display: flex; gap: 12rpx; margin-top: 24rpx; }
.db-scroll { flex: 1; max-height: 52vh; margin-top: 16rpx; }
.db-item { display: flex; align-items: center; gap: 16rpx; padding: 20rpx; border-radius: 18rpx; background: #f7f9fc; margin-bottom: 12rpx; }
.db-item.selected { background: rgba(64, 158, 255, 0.1); }
.db-main { flex: 1; min-width: 0; }
.db-name { display: block; font-size: 26rpx; color: #1f2d3d; word-break: break-all; }
.db-comment { display: block; margin-top: 4rpx; font-size: 22rpx; color: #8a94a6; }
.checkbox { width: 40rpx; height: 40rpx; line-height: 40rpx; text-align: center; border-radius: 50%; border: 2rpx solid #c3ccd8; color: #fff; font-size: 24rpx; flex-shrink: 0; }
.db-item.selected .checkbox { background: #409eff; border-color: #409eff; }
.hint { padding: 30rpx 0; text-align: center; font-size: 24rpx; color: #9aa6b8; }
.form-actions { display: flex; gap: 20rpx; margin-top: 24rpx; }
.btn { flex: 1; height: 84rpx; line-height: 84rpx; text-align: center; background: #409eff; color: #fff; font-size: 28rpx; border-radius: 18rpx; }
.btn.plain { background: #f5f7fa; color: #5a6678; }

.config-sheet { max-height: 92vh; }
.config-scroll { margin-top: 20rpx; max-height: 62vh; }
.form-item { display: flex; align-items: center; gap: 16rpx; padding: 14rpx 0; }
.label { width: 130rpx; font-size: 26rpx; color: #5a6678; flex-shrink: 0; }
.input { flex: 1; min-width: 0; height: 64rpx; padding: 0 18rpx; background: #f5f7fa; border-radius: 18rpx; font-size: 26rpx; }
.config-section { margin-top: 16rpx; }
.section-label { font-size: 26rpx; color: #1f2d3d; font-weight: 600; }
.col-card { margin-top: 16rpx; padding: 18rpx 20rpx; background: #f7f9fc; border-radius: 8rpx; }
.col-head { display: flex; align-items: baseline; gap: 12rpx; }
.col-name { font-size: 26rpx; color: #1f2d3d; font-weight: 600; }
.col-type { font-size: 22rpx; color: #8a94a6; }
.col-switches { display: flex; align-items: center; gap: 8rpx; margin-top: 6rpx; }
.col-switch { display: flex; align-items: center; font-size: 22rpx; color: #5a6678; }
.col-pickers { display: flex; gap: 16rpx; margin-top: 10rpx; }
.picker-value { padding: 8rpx 18rpx; background: #fff; border-radius: 18rpx; font-size: 22rpx; color: #409eff; }
.preview-panel { width: 92%; height: 82vh; margin-bottom: 6vh; background: #fff; border-radius: 12rpx; padding: 28rpx; box-sizing: border-box; display: flex; flex-direction: column; }
.preview-head { display: flex; align-items: center; gap: 16rpx; }
.preview-title { flex: 1; min-width: 0; font-size: 28rpx; font-weight: 600; color: #1f2d3d; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.close { font-size: 24rpx; color: #8a94a6; flex-shrink: 0; }
.file-tabs { white-space: nowrap; margin-top: 20rpx; max-height: 70rpx; }
.file-tab { display: inline-block; font-size: 22rpx; padding: 8rpx 18rpx; margin-right: 12rpx; border-radius: 18rpx; color: #5a6678; background: #f5f7fa; }
.file-tab.active { color: #409eff; background: rgba(64, 158, 255, 0.12); }
.code-box { flex: 1; margin-top: 16rpx; background: #1f2d3d; border-radius: 18rpx; padding: 20rpx; }
.code { font-family: Consolas, Monaco, monospace; font-size: 22rpx; line-height: 1.6; color: #e6ebf2; white-space: pre; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
