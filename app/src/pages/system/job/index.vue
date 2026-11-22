<template>
  <view class="page">
    <view class="search-bar">
      <input v-model="keyword" class="search-input" placeholder="输入任务名称搜索" placeholder-class="ph" confirm-type="search" @confirm="onSearch" />
      <view class="search-btn" @tap="onSearch">搜索</view>
      <view class="search-btn plain" @tap="onReset">重置</view>
    </view>
    <view class="tool-bar">
      <view class="tool-btn" @tap="goLog">执行日志</view>
      <view v-if="$hasPerm('monitor:job:add')" class="tool-btn add" @tap="openForm()">＋ 新增任务</view>
    </view>

    <view v-if="list.length" class="job-list">
      <view v-for="j in list" :key="j.jobId" class="job-item">
        <view class="job-line">
          <text class="job-name">{{ j.jobName }}</text>
          <text class="status" :class="j.status === '0' ? 'on' : 'off'" @tap="toggleStatus(j)">{{ j.status === '0' ? '运行中' : '已暂停' }}</text>
        </view>
        <view class="job-line sub">
          <text>{{ j.jobGroup || 'DEFAULT' }}</text>
          <text class="cron">{{ j.cronExpression }}</text>
        </view>
        <view class="job-line sub">
          <text class="target">{{ j.invokeTarget }}</text>
        </view>
        <view class="job-ops">
          <text v-if="$hasPerm('monitor:job:run')" class="op" @tap="onRun(j)">执行一次</text>
          <text v-if="$hasPerm('monitor:job:edit')" class="op" @tap="openForm(j)">编辑</text>
          <text v-if="$hasPerm('monitor:job:remove')" class="op danger" @tap="onDelete(j)">删除</text>
        </view>
      </view>
      <view v-if="loading" class="load-hint"><text class="load-spin"></text><text>加载中…</text></view>
      <view v-else-if="finished" class="load-hint done">没有更多了</view>
    </view>
    <view v-else-if="!loading" class="empty">暂无定时任务</view>
    <view v-else class="empty">加载中…</view>

    <!-- 新增/编辑弹层 -->
    <view v-if="showForm" class="mask" @tap="closeForm">
      <view class="sheet" @tap.stop>
        <view class="sheet-head">
          <view class="sheet-avatar">{{ (form.jobName || 'J').charAt(0).toUpperCase() }}</view>
          <view class="sheet-head-main">
            <text class="sheet-title">{{ form.jobId ? '编辑任务' : '新增任务' }}</text>
            <text class="sheet-sub">{{ form.jobId ? '调整调度规则与执行策略' : '创建一个后台定时任务' }}</text>
          </view>
          <view class="sheet-close" @tap="closeForm">×</view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">基本信息</text>
          </view>
          <view class="field">
            <text class="field-label">任务名称</text>
            <input v-model="form.jobName" class="field-input" placeholder="必填，如 数据同步" placeholder-class="ph" />
          </view>
          <view class="field">
            <text class="field-label">任务分组</text>
            <input v-model="form.jobGroup" class="field-input" placeholder="默认 DEFAULT" placeholder-class="ph" />
          </view>
          <view class="field">
            <text class="field-label">备注</text>
            <input v-model="form.remark" class="field-input" placeholder="选填，说明任务用途" placeholder-class="ph" />
          </view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">调度配置</text>
          </view>
          <view class="field">
            <view class="field-row">
              <text class="field-label">调用目标</text>
              <text class="link" @tap="showExecutors = !showExecutors">{{ showExecutors ? '收起' : '选择内置任务' }}</text>
            </view>
            <input v-model="form.invokeTarget" class="field-input" placeholder="内置执行器名或 http(s):// 地址" placeholder-class="ph" />
            <view v-if="showExecutors && executors.length" class="chips">
              <text v-for="e in executors" :key="e.name" class="chip" @tap="pickExecutor(e)">{{ e.desc || e.name }}</text>
            </view>
          </view>
          <view class="field">
            <text class="field-label">Cron 表达式（6 段：秒 分 时 日 月 周）</text>
            <input v-model="form.cronExpression" class="field-input" placeholder="如 0 */1 * * * *" placeholder-class="ph" />
            <view class="chips">
              <text v-for="p in CRON_PRESETS" :key="p.value" class="chip" @tap="form.cronExpression = p.value">{{ p.label }}</text>
            </view>
          </view>
        </view>

        <view class="group">
          <view class="group-head">
            <view class="group-bar"></view>
            <text class="group-title">执行设置</text>
          </view>
          <view class="field-card">
            <view class="field-main">
              <text class="field-label field-label-inline">允许并发</text>
              <text class="field-desc">允许同一任务同时运行多个实例</text>
            </view>
            <switch :checked="form.concurrent === '0'" color="#3a6fe0" style="transform: scale(0.7)" @change="onConcurrentSwitch" />
            <text class="status-pill" :class="{ off: form.concurrent !== '0' }">{{ form.concurrent === '0' ? '允许' : '禁止' }}</text>
          </view>
          <view class="field-card">
            <view class="field-main">
              <text class="field-label field-label-inline">任务状态</text>
              <text class="field-desc">运行中的任务将按 Cron 持续触发</text>
            </view>
            <switch :checked="form.status === '0'" color="#3a6fe0" style="transform: scale(0.7)" @change="onStatusSwitch" />
            <text class="status-pill" :class="{ off: form.status !== '0' }">{{ form.status === '0' ? '运行中' : '已暂停' }}</text>
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
import { ref, reactive } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { listJob, listJobExecutors, createJob, updateJob, deleteJob, changeJobStatus, runJob } from '@/api/job'
import { trimStr, isCron } from '@/utils/validate'

const CRON_PRESETS = [
  { label: '每分钟', value: '0 */1 * * * *' },
  { label: '每小时', value: '0 0 * * * *' },
  { label: '每天零点', value: '0 0 0 * * *' },
  { label: '每天 8:30', value: '0 30 8 * * *' }
]

const keyword = ref('')
const list = ref([])
const pageNum = ref(1)
const pageSize = 10
const total = ref(0)
const loading = ref(false)
const finished = ref(false)
const showForm = ref(false)
const showExecutors = ref(false)
const executors = ref([])
const form = reactive({ jobId: 0, jobName: '', jobGroup: 'DEFAULT', invokeTarget: '', cronExpression: '', misfirePolicy: '3', concurrent: '1', status: '0', remark: '' })

async function fetchList(reset) {
  if (loading.value) return
  if (reset) { pageNum.value = 1; finished.value = false }
  else if (finished.value) return
  loading.value = true
  try {
    const data = await listJob({ pageNum: pageNum.value, pageSize, jobName: keyword.value || undefined })
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

async function loadExecutors() {
  try {
    const data = await listJobExecutors()
    executors.value = data.data || []
  } catch (e) {}
}

function onSearch() { fetchList(true) }
function onReset() { keyword.value = ''; fetchList(true) }
function goLog() { uni.navigateTo({ url: '/pages/system/joblog/index' }) }

function pickExecutor(e) {
  form.invokeTarget = e.name
  showExecutors.value = false
}

function openForm(j) {
  showExecutors.value = false
  if (j) {
    Object.assign(form, {
      jobId: j.jobId, jobName: j.jobName || '', jobGroup: j.jobGroup || 'DEFAULT',
      invokeTarget: j.invokeTarget || '', cronExpression: j.cronExpression || '',
      misfirePolicy: j.misfirePolicy || '3', concurrent: j.concurrent || '1',
      status: j.status || '0', remark: j.remark || ''
    })
  } else {
    Object.assign(form, { jobId: 0, jobName: '', jobGroup: 'DEFAULT', invokeTarget: '', cronExpression: '', misfirePolicy: '3', concurrent: '1', status: '0', remark: '' })
  }
  showForm.value = true
  if (!executors.value.length) loadExecutors()
}

function closeForm() { showForm.value = false }
function onConcurrentSwitch(e) { form.concurrent = e.detail.value ? '0' : '1' }
function onStatusSwitch(e) { form.status = e.detail.value ? '0' : '1' }

async function submitForm() {
  const jobName = trimStr(form.jobName)
  const jobGroup = trimStr(form.jobGroup) || 'DEFAULT'
  const invokeTarget = trimStr(form.invokeTarget)
  const cronExpression = trimStr(form.cronExpression)
  if (!jobName) {
    uni.showToast({ title: '请输入任务名称', icon: 'none' })
    return
  }
  if (jobName.length > 64) {
    uni.showToast({ title: '任务名称不能超过 64 个字符', icon: 'none' })
    return
  }
  if (!/^[A-Za-z0-9_-]+$/.test(jobGroup)) {
    uni.showToast({ title: '任务分组仅支持字母、数字、下划线或中划线', icon: 'none' })
    return
  }
  if (!invokeTarget) {
    uni.showToast({ title: '请输入调用目标', icon: 'none' })
    return
  }
  if (!cronExpression) {
    uni.showToast({ title: '请输入 Cron 表达式', icon: 'none' })
    return
  }
  if (!isCron(cronExpression)) {
    uni.showToast({ title: 'Cron 表达式需为 6 段（秒 分 时 日 月 周）', icon: 'none' })
    return
  }
  Object.assign(form, { jobName, jobGroup, invokeTarget, cronExpression })
  const payload = {
    jobName, jobGroup, invokeTarget,
    cronExpression, misfirePolicy: form.misfirePolicy || '3',
    concurrent: form.concurrent, status: form.status, remark: trimStr(form.remark)
  }
  try {
    if (form.jobId) {
      await updateJob({ ...payload, jobId: form.jobId })
      const idx = list.value.findIndex((it) => it.jobId === form.jobId)
      if (idx >= 0) list.value[idx] = { ...list.value[idx], ...payload }
      uni.showToast({ title: '修改成功', icon: 'success' })
    } else {
      await createJob(payload)
      uni.showToast({ title: '新增成功', icon: 'success' })
      setTimeout(() => fetchList(true), 400)
    }
    closeForm()
  } catch (e) {}
}

async function toggleStatus(j) {
  if (!$hasPerm('monitor:job:changeStatus')) return
  const next = j.status === '0' ? '1' : '0'
  try {
    await changeJobStatus({ jobId: j.jobId, status: next })
    j.status = next
    uni.showToast({ title: next === '0' ? '已启用' : '已暂停', icon: 'none' })
  } catch (e) {}
}

function onRun(j) {
  uni.showModal({
    title: '提示',
    content: `确定立即执行任务「${j.jobName}」吗？`,
    confirmText: '执行',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await runJob(j.jobId)
        uni.showToast({ title: '已触发执行，可稍后查看执行日志', icon: 'none', duration: 1500 })
      } catch (e) {}
    }
  })
}

function onDelete(j) {
  uni.showModal({
    title: '提示',
    content: `确定删除任务「${j.jobName}」吗？`,
    confirmText: '删除',
    confirmColor: '#f56c6c',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await deleteJob(j.jobId)
        list.value = list.value.filter((it) => it.jobId !== j.jobId)
        total.value = Math.max(0, total.value - 1)
        uni.showToast({ title: '删除成功', icon: 'success' })
        if (!list.value.length) fetchList(true)
      } catch (e) {}
    }
  })
}

onLoad(() => {
  uni.setNavigationBarTitle({ title: '定时任务' })
  fetchList(true)
  loadExecutors()
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
.tool-bar { display: flex; justify-content: flex-end; gap: 16rpx; margin-top: 16rpx; }
.tool-btn { padding: 0 22rpx; height: 64rpx; line-height: 64rpx; background: #fff; color: #5a6678; font-size: 26rpx; border-radius: 18rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.tool-btn.add { background: #67c23a; color: #fff; }
.job-list { margin-top: 24rpx; background: #fff; border-radius: 18rpx; padding: 0 28rpx; box-shadow: 0 12rpx 30rpx rgba(30, 48, 80, 0.06); }
.job-item { padding: 26rpx 0; }
.job-item + .job-item { border-top: 1rpx solid #f0f3f7; }
.job-line { display: flex; align-items: center; gap: 12rpx; }
.job-name { flex: 1; min-width: 0; font-size: 28rpx; color: #1f2d3d; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.status { font-size: 20rpx; padding: 2rpx 14rpx; border-radius: 18rpx; flex-shrink: 0; }
.status.on { color: #67c23a; background: rgba(103, 194, 58, 0.1); }
.status.off { color: #f56c6c; background: rgba(245, 108, 108, 0.1); }
.job-line.sub { margin-top: 8rpx; font-size: 22rpx; color: #8a94a6; }
.cron { color: #409eff; }
.cron::before { content: '·'; margin: 0 8rpx; color: #9aa6b8; }
.target { color: #8a94a6; word-break: break-all; }
.job-ops { display: flex; gap: 16rpx; margin-top: 16rpx; }
.op { font-size: 24rpx; padding: 8rpx 22rpx; border-radius: 18rpx; color: #409eff; background: rgba(64, 158, 255, 0.08); }
.op.danger { color: #f56c6c; background: rgba(245, 108, 108, 0.08); }
.load-hint { display: flex; align-items: center; justify-content: center; gap: 12rpx; padding: 24rpx 0; font-size: 24rpx; color: #9aa6b8; }
.load-spin { width: 28rpx; height: 28rpx; border: 4rpx solid #e6ebf2; border-top-color: #409eff; border-radius: 50%; animation: spin 0.8s linear infinite; }
.load-hint.done { color: #c3ccd8; }
.empty { margin-top: 24rpx; padding: 100rpx 0; text-align: center; color: #9aa6b8; font-size: 26rpx; background: #fff; border-radius: 18rpx; }
.mask { position: fixed; inset: 0; background: rgba(31, 45, 61, 0.45); z-index: 99; display: flex; align-items: flex-end; justify-content: center; }
.sheet { width: 100%; max-height: 88vh; overflow-y: auto; background: #fff; border-radius: 16rpx 16rpx 0 0; padding: 32rpx 32rpx 48rpx; box-sizing: border-box; }
.sheet-title { font-size: 32rpx; font-weight: 600; color: #1f2d3d; text-align: center; margin-bottom: 28rpx; }
.form-item { display: flex; align-items: center; gap: 20rpx; padding: 18rpx 0; border-bottom: 1rpx solid #f0f3f7; }
.form-item.column { flex-direction: column; align-items: stretch; gap: 12rpx; }
.label-row { display: flex; align-items: center; justify-content: space-between; }
.label { width: 140rpx; font-size: 26rpx; color: #5a6678; flex-shrink: 0; }
.form-item.column .label { width: auto; }
.link { font-size: 24rpx; color: #409eff; }
.input { flex: 1; height: 64rpx; font-size: 26rpx; }
.input.full { padding: 0 16rpx; background: #f5f7fa; border-radius: 18rpx; }
.switch-text { font-size: 24rpx; color: #8a94a6; }
.chips { display: flex; flex-wrap: wrap; gap: 12rpx; margin-top: 8rpx; }
.chip { font-size: 22rpx; padding: 8rpx 18rpx; border-radius: 18rpx; color: #409eff; background: rgba(64, 158, 255, 0.08); }
.form-actions { display: flex; gap: 20rpx; margin-top: 40rpx; }
.btn { flex: 1; height: 84rpx; line-height: 84rpx; text-align: center; background: #409eff; color: #fff; font-size: 28rpx; border-radius: 18rpx; }
.btn.plain { background: #f5f7fa; color: #5a6678; }
@keyframes spin { to { transform: rotate(360deg); } }

/* ===== 弹层精修（新版） ===== */
.sheet { padding: 0; max-height: 90vh; border-radius: 28rpx 28rpx 0 0; overflow-y: auto; background: #fff; }
.sheet-head { position: sticky; top: 0; z-index: 5; display: flex; align-items: center; gap: 20rpx; padding: 30rpx 32rpx 24rpx; background: #fff; border-bottom: 1rpx solid #f2f5fb; }
.sheet-avatar { width: 88rpx; height: 88rpx; border-radius: 24rpx; display: flex; align-items: center; justify-content: center; font-size: 36rpx; font-weight: 700; color: #fff; flex-shrink: 0; background: linear-gradient(135deg, #6aa4ff, #3a6fe0); box-shadow: 0 10rpx 24rpx rgba(58, 111, 224, 0.26); }
.sheet-head-main { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 8rpx; }
.sheet-head .sheet-title { font-size: 32rpx; font-weight: 700; color: #1c2738; text-align: left; margin: 0; }
.sheet-sub { font-size: 22rpx; color: #9aa6b8; }
.sheet-close { width: 56rpx; height: 56rpx; line-height: 56rpx; text-align: center; font-size: 32rpx; color: #9aa6b8; background: #f3f6fb; border-radius: 50%; flex-shrink: 0; }
.group { padding: 28rpx 32rpx 0; }
.group-head { display: flex; align-items: center; gap: 12rpx; margin-bottom: 20rpx; }
.group-bar { width: 6rpx; height: 26rpx; border-radius: 4rpx; background: linear-gradient(180deg, #6aa4ff, #3a6fe0); }
.group-title { font-size: 26rpx; font-weight: 600; color: #1c2738; }
.field { display: flex; flex-direction: column; gap: 12rpx; margin-bottom: 24rpx; }
.field-row { display: flex; align-items: center; justify-content: space-between; }
.field-label { font-size: 24rpx; color: #6b7a90; }
.field-input { width: 100%; height: 84rpx; padding: 0 24rpx; box-sizing: border-box; background: #f5f7fb; border-radius: 14rpx; font-size: 27rpx; color: #1c2738; }
.field-select { display: flex; align-items: center; justify-content: space-between; height: 84rpx; padding: 0 24rpx; background: #f5f7fb; border-radius: 14rpx; }
.field-select-text { flex: 1; min-width: 0; font-size: 27rpx; color: #1c2738; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.field-select-arrow { font-size: 30rpx; color: #b9c3d3; }
.field-card { display: flex; align-items: center; gap: 16rpx; padding: 24rpx; margin-bottom: 24rpx; background: #f5f7fb; border-radius: 16rpx; }
.field-main { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 6rpx; }
.field-label-inline { font-size: 26rpx; font-weight: 500; color: #1c2738; }
.field-desc { font-size: 21rpx; color: #9aa6b8; }
.status-pill { font-size: 20rpx; padding: 4rpx 16rpx; border-radius: 20rpx; color: #3fae7a; background: rgba(63, 174, 122, 0.12); flex-shrink: 0; }
.status-pill.off { color: #f56c6c; background: rgba(245, 108, 108, 0.12); }
.link { font-size: 22rpx; color: #3a6fe0; }
.chips { display: flex; flex-wrap: wrap; gap: 12rpx; margin-top: 4rpx; }
.chip { font-size: 21rpx; padding: 8rpx 20rpx; border-radius: 20rpx; color: #3a6fe0; background: rgba(58, 111, 224, 0.09); }
.form-actions { position: sticky; bottom: 0; z-index: 5; display: flex; gap: 20rpx; padding: 20rpx 32rpx 40rpx; background: #fff; border-top: 1rpx solid #f2f5fb; }
.btn { flex: 1; height: 92rpx; line-height: 92rpx; text-align: center; color: #fff; font-size: 28rpx; font-weight: 600; border-radius: 18rpx; background: linear-gradient(135deg, #4a7df0, #3457c9); box-shadow: 0 12rpx 26rpx rgba(52, 87, 201, 0.24); }
.btn.plain { background: #f3f6fb; color: #6b7a90; box-shadow: none; }
</style>
